package eventbox

import (
    "sync"
    "github.com/fsnotify/fsnotify"

    "github.com/AstromechZA/inotify-spy/fileevents"
)

type EventBox struct {
    lock sync.Mutex
    Data map[string]fileevents.FileWithEvents
}

func NewEventBox() *EventBox {
    return &EventBox{
        lock: sync.Mutex{},
        Data: make(map[string]fileevents.FileWithEvents),
    }
}

func (b *EventBox) Add(e *fsnotify.Event) {
    b.lock.Lock()
    defer b.lock.Unlock()

    fevent, ok := b.Data[e.Name]
    if ok == false {
        fevent = fileevents.FileWithEvents{
            Name: e.Name,
            Events: make(map[fsnotify.Op]int),
            Total: 0,
        }
    }

    count := fevent.Events[e.Op]
    fevent.Events[e.Op] = count + 1
    fevent.Total++
    b.Data[e.Name] = fevent
}
