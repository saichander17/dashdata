package wal

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "time"
)

type WAL struct {
    file *os.File
    writer *bufio.Writer
}

type LogEntry struct {
    Timestamp  time.Time
    Operation  string
    Key        string
    Value      string
}

type StoreOperations interface {
    Set(key, value string)
    Delete(key string)
}

func NewWAL(filename string) (*WAL, error) {
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return nil, err
    }
    return &WAL{
        file: file,
        writer: bufio.NewWriter(file),
    }, nil
}

func (w *WAL) Log(operation, key, value string) error {
    entry := LogEntry{
        Timestamp: time.Now(),
        Operation: operation,
        Key:       key,
        Value:     value,
    }
    data, err := json.Marshal(entry)
    if err != nil {
        return err
    }
    _, err = w.writer.WriteString(string(data) + "\n")
    if err != nil {
        return err
    }
    return w.writer.Flush()
}

func (w *WAL) Close() error {
    return w.file.Close()
}

func (w *WAL) ApplyEntriesAfter(timestamp time.Time, store StoreOperations) error {
    _, err := w.file.Seek(0, 0)
    if err != nil {
        return fmt.Errorf("failed to seek to beginning of WAL file: %v", err)
    }

    scanner := bufio.NewScanner(w.file)
    for scanner.Scan() {
        var entry LogEntry
        if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
            return fmt.Errorf("failed to unmarshal log entry: %v", err)
        }
        if entry.Timestamp.After(timestamp) {
            switch entry.Operation {
            case "SET":
                store.Set(entry.Key, entry.Value)
            case "DELETE":
                store.Delete(entry.Key)
            }
        }
    }
    return scanner.Err()
}
