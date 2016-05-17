package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "path/filepath"
    "sort"
    "bufio"

    "github.com/fsnotify/fsnotify"

    "github.com/AstromechZA/inotify-spy/fileevents"
    "github.com/AstromechZA/inotify-spy/eventbox"
)

const versionString =
`Version: 1.0
          ____
     _[]_/____\__n_
    |_____.--.__()_|
    |LI  //# \\    |
    |    \\__//    |
    |     '--'     |
    '--------------'

Project: https://github.com/AstromechZA/inotify-spy
`

const usageString =
`inotify-spy is a simple binary for doing inotify watching on Linux.
It will allow you to watch a directory or directory tree for file events:

- Create
- Write
- Remove
- Rename
- Chmod
- Open

Because this tool uses inotify events, it has to create one for each directory
that it watches. On most systems there is a limit to the number of inotify
watches or files a process is allowed to create at once. You might hit these
limits if you try to watch a very large tree of directories. On most systems
there are ways to increase these limits if required.

This tool will continue to watch the tree until you stop it using SIGINT (Ctrl-C)
which will cause it to print a sorted summary of the files touched.

Usage: inotify-spy [-live] [-mute-errors] [-recursive] directory

`

func addDirWatchers(w *fsnotify.Watcher, wyes *int, wno *int, mute bool) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if err != nil { return nil }

        if info.IsDir() {
            e := w.Add(path)
            if e != nil {
                if mute == false {
                    fmt.Printf("Failed to watch %v: %v\n", path, e.Error())
                }
                (*wno)++
                return nil
            }
            (*wyes)++
        }
        return nil
    }
}

func safeAbsolutePath(path string) string {
    abspath, err := filepath.Abs(path)
    if err == nil { return abspath }
    return path
}

var opNameLookup = map[fsnotify.Op]string {
    fsnotify.Create: "Create",
    fsnotify.Write: "Write",
    fsnotify.Remove: "Remove",
    fsnotify.Rename: "Rename",
    fsnotify.Chmod: "Chmod",
    fsnotify.Open: "Open",
}

func doSummary(box *eventbox.EventBox, recordMask uint64, sortByName bool, exportCSV string) {

    fmt.Println()

    column := "%-6s"
    if recordMask & uint64(fsnotify.Create) == uint64(fsnotify.Create) {
        fmt.Printf(column, "Create")
    }
    if recordMask & uint64(fsnotify.Write) == uint64(fsnotify.Write) {
        fmt.Printf(column, "Write")
    }
    if recordMask & uint64(fsnotify.Remove) == uint64(fsnotify.Remove) {
        fmt.Printf(column, "Remove")
    }
    if recordMask & uint64(fsnotify.Rename) == uint64(fsnotify.Rename) {
        fmt.Printf(column, "Rename")
    }
    if recordMask & uint64(fsnotify.Chmod) == uint64(fsnotify.Chmod) {
        fmt.Printf(column, "Chmod")
    }
    if recordMask & uint64(fsnotify.Open) == uint64(fsnotify.Open) {
        fmt.Printf(column, "Open")
    }
    fmt.Println("Path")

    var fevents []fileevents.FileWithEvents
    for _, v := range (*box).Data {
        fevents = append(fevents, v)
    }
    if sortByName {
        sort.Sort(fileevents.ByName(fevents))
    } else {
        sort.Sort(fileevents.ByEventTotal(fevents))
    }

    for _, v := range fevents {
        fmt.Printf("%-6d %-6d %-6d %-6d %-6d %-6d %s\n",
            v.Events[fsnotify.Create],
            v.Events[fsnotify.Write],
            v.Events[fsnotify.Remove],
            v.Events[fsnotify.Rename],
            v.Events[fsnotify.Chmod],
            v.Events[fsnotify.Open],
            v.Name,
        )

        if recordMask & uint64(fsnotify.Create) == uint64(fsnotify.Create) {
            fmt.Printf(column, v.Events[fsnotify.Create])
        }
        if recordMask & uint64(fsnotify.Write) == uint64(fsnotify.Write) {
            fmt.Printf(column, v.Events[fsnotify.Write])
        }
        if recordMask & uint64(fsnotify.Remove) == uint64(fsnotify.Remove) {
            fmt.Printf(column, v.Events[fsnotify.Remove])
        }
        if recordMask & uint64(fsnotify.Rename) == uint64(fsnotify.Rename) {
            fmt.Printf(column, v.Events[fsnotify.Rename])
        }
        if recordMask & uint64(fsnotify.Chmod) == uint64(fsnotify.Chmod) {
            fmt.Printf(column, v.Events[fsnotify.Chmod])
        }
        if recordMask & uint64(fsnotify.Open) == uint64(fsnotify.Open) {
            fmt.Printf(column, v.Events[fsnotify.Open])
        }
        fmt.Println(v.Name)
    }
}

func main() {

    // flag args
    recursiveFlag := flag.Bool("recursive", false, "Recursively watch target directory")
    liveFlag := flag.Bool("live", false, "Show events live, not just as a summary at the end")
    muteErrorsFlag := flag.Bool("mute-errors", false, "Mute error messages related to setting up watches")
    versionFlag := flag.Bool("version", false, "Print version information")

    // summary flags
    sortByNameFlag := flag.Bool("sort-name", false, "Sort summary by file path rather than most events")
    exportCSVFlag := flag.String("export-csv", "", "Export summary as csv to the given path")

    // record options
    dontRecordCreate := flag.Bool("dont-record-create", false, "Don't record create events")
    dontRecordWrite := flag.Bool("dont-record-write", false, "Don't record write events")
    dontRecordRemove := flag.Bool("dont-record-remove", false, "Don't record remove events")
    dontRecordRename := flag.Bool("dont-record-rename", false, "Don't record rename events")
    dontRecordChmod := flag.Bool("dont-record-chmod", false, "Don't record chmod events")
    dontRecordOpen := flag.Bool("dont-record-open", false, "Don't record open events")

    // ignore prefix
    //ignorePrefix := flag.String("ignore-prefix-file", "", "Ignore any directories that start with the given prefixes found in this file")

    flag.Usage = func() {
        os.Stderr.WriteString(usageString)
        flag.PrintDefaults()
    }

    // parse them
    flag.Parse()

    if (*versionFlag) {
        fmt.Print(versionString)
        os.Exit(0)
    }

    // make sure we have our single positional arg
    if len(flag.Args()) != 1 {
        flag.Usage()
        os.Exit(1)
    }

    targetDir := flag.Args()[0]

    // setup watcher
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        fmt.Printf("Failed to setup fsnotify watcher: %v\n", err.Error())
    }
    // make sure we close it
    defer watcher.Close()

    // get mute value
    mustMute := *muteErrorsFlag

    var watchedCounter int
    var notWatchedCounter int
    box := eventbox.NewEventBox()
    readyChannel := make(chan bool)

    var recordMask uint64
    if (*dontRecordCreate) == false { recordMask -= uint64(fsnotify.Create) }
    if (*dontRecordWrite) == false { recordMask -= uint64(fsnotify.Write) }
    if (*dontRecordRemove) == false { recordMask -= uint64(fsnotify.Remove) }
    if (*dontRecordRename) == false { recordMask -= uint64(fsnotify.Rename) }
    if (*dontRecordChmod) == false { recordMask -= uint64(fsnotify.Chmod) }
    if (*dontRecordOpen) == false { recordMask -= uint64(fsnotify.Open) }

    fmt.Println("Beginning to watch events..")
    go func(live bool, box *eventbox.EventBox) {
        ready := false
        for {
            select {
            case event := <- watcher.Events:
                if ready {
                    if recordMask & uint64(event.Op) == uint64(event.Op) {
                        event.Name = safeAbsolutePath(event.Name)
                        if live {
                            fmt.Printf("event: %v\n", event.String())
                        }
                        box.Add(&event)
                    }
                }
                // otherwise ignore it
            case <- readyChannel:
                ready = true
            case err := <- watcher.Errors:
                fmt.Printf("error: %v\n", err)
            }
        }
    }(*liveFlag, box)

    if (*recursiveFlag) {
        err = filepath.Walk(targetDir, addDirWatchers(watcher, &watchedCounter, &notWatchedCounter, mustMute))
        if err != nil {
            fmt.Printf("Could not walk %v: %v\n", targetDir, err.Error())
            os.Exit(1)
        }
    } else {
        err = watcher.Add(targetDir)
        if err != nil {
            fmt.Printf("Could not watch %v: %v\n", targetDir, err.Error())
            os.Exit(1)
        } else {
            watchedCounter++
        }
    }

    fmt.Printf("Watching %d directories..\n", watchedCounter)
    if notWatchedCounter > 0 {
        fmt.Printf("Could not watch %d directories.\n", notWatchedCounter)
        fmt.Println("If you got 'permission denied errors', try running as root.")
        fmt.Println("If you got 'too many open files' or 'no space left on device' you probably need to increase the number of inotify watches you're allowed.")
    }

    fmt.Println("Press enter to start recording:")
    reader := bufio.NewReader(os.Stdin)
    reader.ReadString('\n')

    // now tell goroutine to start recording things
    fmt.Println("Beginning to record events. Press Ctrl-C to stop..")
    readyChannel <- true

    // instead of sitting in a for loop or something, we wait for sigint
    signalChannel := make(chan os.Signal, 1)
    // notify that we are going to handle interrupts
    signal.Notify(signalChannel, os.Interrupt)
    for sig := range signalChannel {
        fmt.Printf("Received %v signal. Stopping.\n", sig)

        doSummary(box, recordMask, *sortByNameFlag, *exportCSVFlag)

        os.Exit(0)
    }
}
