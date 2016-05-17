package fileevents

import (
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
func (a ByEventTotal) Less(i, j int) bool {return a[i].Total < a[j].Total}
