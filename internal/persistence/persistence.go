package persistence

import (
    "encoding/gob"
    "os"
    "sync"
    "time"

    "github.com/saichander17/dashdata/internal/store"
    "github.com/saichander17/dashdata/internal/wal"
)

type Persister struct {
    store     store.Store
    filename  string
    interval  time.Duration
    mutex     sync.Mutex
}

type SnapshotData struct {
    Timestamp time.Time
    Data      map[string]string
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

    tempFile := p.filename + ".tmp"
    file, err := os.Create(tempFile)
    if err != nil {
        return err
    }

    snapshot := SnapshotData{
        Timestamp: time.Now(),
        Data:      p.store.GetAll(),
    }

    encoder := gob.NewEncoder(file)
    if err := encoder.Encode(snapshot); err != nil {
        file.Close()
        os.Remove(tempFile)
        return err
    }

    if err := file.Sync(); err != nil {
        file.Close()
        os.Remove(tempFile)
        return err
    }

    if err := file.Close(); err != nil {
        os.Remove(tempFile)
        return err
    }

    return os.Rename(tempFile, p.filename)
}

func (p *Persister) LoadFromDisk(wal *wal.WAL) error {
    file, err := os.Open(p.filename)
    if err != nil {
        if os.IsNotExist(err) {
            return nil // No persistence file yet, not an error
        }
        return err
    }
    defer file.Close()

    var snapshot SnapshotData
    decoder := gob.NewDecoder(file)
    if err := decoder.Decode(&snapshot); err != nil {
        return err
    }

    for k, v := range snapshot.Data {
        p.store.Set(k, v)
    }

    // Apply WAL entries after snapshot timestamp
    return wal.ApplyEntriesAfter(snapshot.Timestamp, p.store)
}
/*
Gob Encoding. What does it do?
How is the data being stored in a file?
What if the single file gets super large?
What if the service goes down before data being written to disk?
*/
