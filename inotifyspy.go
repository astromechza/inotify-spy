package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "path/filepath"
    "bufio"
    "io/ioutil"
    "strings"

    "github.com/fsnotify/fsnotify"

    "github.com/AstromechZA/inotify-spy/eventbox"
    "github.com/AstromechZA/inotify-spy/summary"
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

func safeAbsolutePath(path string) string {
    abspath, err := filepath.Abs(path)
    if err == nil { return abspath }
    return path
}

func mustIgnorePath(path string, ignore *[]string) bool {
    for _, s := range *ignore {
        if strings.HasPrefix(path, s) {
            return true
        }
    }
    return false
}

func addDirWatchers(w *fsnotify.Watcher, wyes *int, wno *int, mute bool, ignorePrefixes *[]string) filepath.WalkFunc {
    return func(path string, info os.FileInfo, err error) error {
        if err != nil { return nil }
        if info.IsDir() {
            path = safeAbsolutePath(path)

            if mustIgnorePath(path, ignorePrefixes) {
                fmt.Printf("Not watching %v or its children since it matches an ignore prefix\n", path)
                return filepath.SkipDir
            }

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

    // ignore prefixes
    ignorePrefixFlag := flag.String("ignore-prefixes", "", "File to read ignore prefixes from")

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

    // read ignore prefixes if required
    var ignorePrefixes []string
    ignorePrefixFile := *ignorePrefixFlag
    if ignorePrefixFile != "" {
        fmt.Printf("Loading ignore prefixes from %v\n", ignorePrefixFile)
        prefixes, err := ioutil.ReadFile(ignorePrefixFile)
        if err != nil {
            fmt.Printf("Could not open ignore prefixes file %v: %v\n", ignorePrefixFile, err.Error())
            os.Exit(1)
        }
        ignorePrefixes = strings.Split(strings.TrimSpace(string(prefixes)), "\n")
        fmt.Printf("Loaded %d ignore prefixes\n", len(ignorePrefixes))
    }

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

    var recordMask uint64 = 63
    if (*dontRecordCreate) == true { recordMask -= uint64(fsnotify.Create) }
    if (*dontRecordWrite) == true { recordMask -= uint64(fsnotify.Write) }
    if (*dontRecordRemove) == true { recordMask -= uint64(fsnotify.Remove) }
    if (*dontRecordRename) == true { recordMask -= uint64(fsnotify.Rename) }
    if (*dontRecordChmod) == true { recordMask -= uint64(fsnotify.Chmod) }
    if (*dontRecordOpen) == true { recordMask -= uint64(fsnotify.Open) }

    fmt.Println("Beginning to watch events..")
    readyChannel := make(chan bool)
    stopChannel := make(chan bool)
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
            case <- stopChannel:
                return
            case err := <- watcher.Errors:
                fmt.Printf("error: %v\n", err)
            }
        }
    }(*liveFlag, box)

    if (*recursiveFlag) {
        err = filepath.Walk(targetDir, addDirWatchers(watcher, &watchedCounter, &notWatchedCounter, mustMute, &ignorePrefixes))
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
        stopChannel <- true

        fmt.Printf("Stopping inotify watcher..\n")
        watcher.Close()

        // print and output summary infos
        err := summary.DoSummary(box, recordMask, *sortByNameFlag, *exportCSVFlag)
        if err != nil {
            fmt.Printf("Error: %s\n", err.Error())
            os.Exit(1)
        }

        os.Exit(0)
    }
}
