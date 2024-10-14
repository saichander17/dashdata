package wal

import (
    "bufio"
    "fmt"
    "os"
)

type WAL struct {
    file *os.File
    writer *bufio.Writer
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
    _, err := w.writer.WriteString(fmt.Sprintf("%s %s %s\n", operation, key, value))
    if err != nil {
        return err
    }
    return w.writer.Flush()
}

func (w *WAL) Close() error {
    return w.file.Close()
}
