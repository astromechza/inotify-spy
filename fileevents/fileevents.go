package fileevents

import (
    "strings"
    "github.com/fsnotify/fsnotify"
)

type FileWithEvents struct {
    Name string
    Events map[fsnotify.Op]int64
    Total int64
}

type ByEventTotal []FileWithEvents
func (a ByEventTotal) Len() int {return len(a)}
func (a ByEventTotal) Swap(i, j int) {a[i], a[j] = a[j], a[i]}
func (a ByEventTotal) Less(i, j int) bool {return a[i].Total > a[j].Total}

type ByName []FileWithEvents
func (a ByName) Len() int {return len(a)}
func (a ByName) Swap(i, j int) {a[i], a[j] = a[j], a[i]}
func (a ByName) Less(i, j int) bool {return strings.Compare(a[i].Name, a[j].Name) > 0}
