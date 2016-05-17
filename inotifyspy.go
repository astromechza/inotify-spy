package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "path/filepath"
    "sort"

    "github.com/fsnotify/fsnotify"

    "github.com/AstromechZA/inotify-spy/fileevents"
    "github.com/AstromechZA/inotify-spy/eventbox"
)

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

func printSummary(box *eventbox.EventBox) {

    fmt.Println()

    fmt.Printf("%-6s %-6s %-6s %-6s %-6s %-6s %s\n", "Create", "Write", "Remove", "Rename", "Chmod", "Open", "Path")

    var fevents []fileevents.FileWithEvents
    for _, v := range (*box).Data {
        fevents = append(fevents, v)
    }
    sort.Sort(fileevents.ByEventTotal(fevents))

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
    }
}

func main() {

    // flag args
    recursiveFlag := flag.Bool("recursive", false, "Recursively watch target directory")
    liveFlag := flag.Bool("live", false, "Show events live, not as a summary")
    muteErrors := flag.Bool("mute-errors", false, "Mute error messages related to setting up watches")

    // parse them
    flag.Parse()

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
    mustMute := *muteErrors

    var watchedCounter int
    var notWatchedCounter int
    err = watcher.Add(targetDir)
    if err != nil {
        if mustMute == false {
            fmt.Printf("Could not watch %v: %v\n", targetDir, err.Error())
        }
        notWatchedCounter++
    } else {
        watchedCounter++
    }

    if (*recursiveFlag) {
        err = filepath.Walk(targetDir, addDirWatchers(watcher, &watchedCounter, &notWatchedCounter, mustMute))
        if err != nil {
            fmt.Printf("Could not walk %v: %v\n", targetDir, err.Error())
            os.Exit(1)
        }
    }

    fmt.Printf("Watching %d directories..\n", watchedCounter)
    if notWatchedCounter > 0 {
        fmt.Printf("Could not watch %d directories.\n", notWatchedCounter)
        fmt.Println("If you got 'permission denied errors', try running as root.")
        fmt.Println("If you got 'too many open files' or 'no space left on device' you probably need to increase the number of inotify watches you're allowed.")
    }

    box := eventbox.NewEventBox()

    fmt.Println("Beginning to watch events..")
    go func(live bool, box *eventbox.EventBox) {
        for {
            select {
            case event := <-watcher.Events:
                event.Name = safeAbsolutePath(event.Name)
                if live {
                    fmt.Printf("event: %v\n", event.String())
                }
                box.Add(&event)
            case err := <-watcher.Errors:
                fmt.Printf("error: %v\n", err)
            }
        }
    }(*liveFlag, box)

    // instead of sitting in a for loop or something, we wait for sigint
    signalChannel := make(chan os.Signal, 1)
    // notify that we are going to handle interrupts
    signal.Notify(signalChannel, os.Interrupt)
    for sig := range signalChannel {
        fmt.Printf("Received %v signal. Stopping.\n", sig)

        printSummary(box)

        os.Exit(0)
    }
}
