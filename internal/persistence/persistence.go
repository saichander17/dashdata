package persistence

import (
    "encoding/gob"
    "os"
    "sync"
    "time"

    "github.com/saichander17/dashdata/internal/store"
)

type Persister struct {
    store     store.Store
    filename  string
    interval  time.Duration
    mutex     sync.Mutex
}

func NewPersister(s store.Store, filename string, interval time.Duration) *Persister {
    return &Persister{
        store:    s,
        filename: filename,
        interval: interval,
    }
}

func (p *Persister) Start() {
    ticker := time.NewTicker(p.interval)
    go func() {
        for range ticker.C {
            p.SaveToDisk()
        }
    }()
}

func (p *Persister) SaveToDisk() error {
    p.mutex.Lock()
    defer p.mutex.Unlock()

    file, err := os.Create(p.filename)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := gob.NewEncoder(file)
    return encoder.Encode(p.store.GetAll())
}

func (p *Persister) LoadFromDisk() error {
    file, err := os.Open(p.filename)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // No persistence file yet, not an error
        }
        return err
    }
    defer file.Close()

    decoder := gob.NewDecoder(file)
    var data map[string]string
    if err := decoder.Decode(&data); err != nil {
        return err
    }

    for k, v := range data {
        p.store.Set(k, v)
    }
    return nil
}
/*
Gob Encoding. What does it do?
How is the data being stored in a file?
What if the single file gets super large?
What if the service goes down before data being written to disk?
*/
