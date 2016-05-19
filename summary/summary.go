package summary

import (
    "fmt"
    "strconv"
    "sort"
    "io/ioutil"

    "github.com/fsnotify/fsnotify"

    "github.com/AstromechZA/inotify-spy/fileevents"
    "github.com/AstromechZA/inotify-spy/eventbox"
)

var opNameLookup = map[fsnotify.Op]string {
    fsnotify.Create: "Create",                  // 1
    fsnotify.Write: "Write",                    // 2
    fsnotify.Remove: "Remove",                  // 4
    fsnotify.Rename: "Rename",                  // 8
    fsnotify.Chmod: "Chmod",                    // 16
    fsnotify.Open: "Open",                      // 32
}

func DoSummary(box *eventbox.EventBox, recordMask uint, sortByName bool, exportCSV string) error {

    fmt.Println()

    strColumn := "%-7s"
    numColumn := "%-7d"
    if recordMask & uint(fsnotify.Create) == uint(fsnotify.Create) {
        fmt.Printf(strColumn, "Create")
    }
    if recordMask & uint(fsnotify.Write) == uint(fsnotify.Write) {
        fmt.Printf(strColumn, "Write")
    }
    if recordMask & uint(fsnotify.Remove) == uint(fsnotify.Remove) {
        fmt.Printf(strColumn, "Remove")
    }
    if recordMask & uint(fsnotify.Rename) == uint(fsnotify.Rename) {
        fmt.Printf(strColumn, "Rename")
    }
    if recordMask & uint(fsnotify.Chmod) == uint(fsnotify.Chmod) {
        fmt.Printf(strColumn, "Chmod")
    }
    if recordMask & uint(fsnotify.Open) == uint(fsnotify.Open) {
        fmt.Printf(strColumn, "Open")
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

        if recordMask & uint(fsnotify.Create) == uint(fsnotify.Create) {
            fmt.Printf(numColumn, v.Events[fsnotify.Create])
        }
        if recordMask & uint(fsnotify.Write) == uint(fsnotify.Write) {
            fmt.Printf(numColumn, v.Events[fsnotify.Write])
        }
        if recordMask & uint(fsnotify.Remove) == uint(fsnotify.Remove) {
            fmt.Printf(numColumn, v.Events[fsnotify.Remove])
        }
        if recordMask & uint(fsnotify.Rename) == uint(fsnotify.Rename) {
            fmt.Printf(numColumn, v.Events[fsnotify.Rename])
        }
        if recordMask & uint(fsnotify.Chmod) == uint(fsnotify.Chmod) {
            fmt.Printf(numColumn, v.Events[fsnotify.Chmod])
        }
        if recordMask & uint(fsnotify.Open) == uint(fsnotify.Open) {
            fmt.Printf(numColumn, v.Events[fsnotify.Open])
        }
        fmt.Println(v.Name)
    }

    if len(fevents) == 0 {
        fmt.Println("No events recorded.")
    }

    if exportCSV != "" {
        fmt.Println("Writing CSV to", exportCSV)

        content := ""

        if recordMask & uint(fsnotify.Create) == uint(fsnotify.Create) {
            content += "Create,"
        }
        if recordMask & uint(fsnotify.Write) == uint(fsnotify.Write) {
            content += "Write,"
        }
        if recordMask & uint(fsnotify.Remove) == uint(fsnotify.Remove) {
            content += "Remove,"
        }
        if recordMask & uint(fsnotify.Rename) == uint(fsnotify.Rename) {
            content += "Rename,"
        }
        if recordMask & uint(fsnotify.Chmod) == uint(fsnotify.Chmod) {
            content += "Chmod,"
        }
        if recordMask & uint(fsnotify.Open) == uint(fsnotify.Open) {
            content += "Open,"
        }
        content += "Path\n"

        for _, v := range fevents {
            if recordMask & uint(fsnotify.Create) == uint(fsnotify.Create) {
                content += strconv.Itoa(int(v.Events[fsnotify.Create])) + ","
            }
            if recordMask & uint(fsnotify.Write) == uint(fsnotify.Write) {
                content += strconv.Itoa(int(v.Events[fsnotify.Write])) + ","
            }
            if recordMask & uint(fsnotify.Remove) == uint(fsnotify.Remove) {
                content += strconv.Itoa(int(v.Events[fsnotify.Remove])) + ","
            }
            if recordMask & uint(fsnotify.Rename) == uint(fsnotify.Rename) {
                content += strconv.Itoa(int(v.Events[fsnotify.Rename])) + ","
            }
            if recordMask & uint(fsnotify.Chmod) == uint(fsnotify.Chmod) {
                content += strconv.Itoa(int(v.Events[fsnotify.Chmod])) + ","
            }
            if recordMask & uint(fsnotify.Open) == uint(fsnotify.Open) {
                content += strconv.Itoa(int(v.Events[fsnotify.Open])) + ","
            }
            content += v.Name + "\n"
        }

        contentBytes := []byte(content)
        err := ioutil.WriteFile(exportCSV, contentBytes, 0644)
        if err != nil {
            return err
        }
    }
    return nil
}
