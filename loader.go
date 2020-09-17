package lar

import (
    "archive/zip"
    "bytes"
    "io"
    "io/ioutil"
)

type loader struct {
    z           *zip.ReadCloser
    readiedFile map[string][]byte
    fileReaders map[string]io.ReadCloser
}

func newLarLoader() *loader {
    l := &loader{}
    l.readiedFile = make(map[string][]byte)
    l.fileReaders = make(map[string]io.ReadCloser)
    return l
}


// close lar
func (l *loader) Close() {
    if l.z != nil {
        _ = l.z.Close()
    }
    l.z = nil
}


// init lar from file
func (l *loader) preloadLarFile(file string) error {
    zr, err := zip.OpenReader(file)
    if err != nil {
        return err
    }
    for _, file := range zr.File {
        if file.FileInfo().IsDir() {
            continue
        }
        fr, err := file.Open()
        if err != nil {
            return err
        }
        l.fileReaders[file.Name] = fr
    }

    return nil
}

// load lar form memory
func (l *loader) LoadMemory(data []byte) error {
    r := bytes.NewReader(data)
    return l.preloadLarMemory(r, r.Size())
}

// init memory lar
func (l *loader) preloadLarMemory(r io.ReaderAt, sz int64) error {
    zr, err := zip.NewReader(r, sz)
    if err != nil {
        return err
    }
    for _, file := range zr.File {
        if file.FileInfo().IsDir() {
            continue
        }
        fr, err := file.Open()
        if err != nil {
            return err
        }
        l.fileReaders[file.Name] = fr
    }

    return nil
}

// get bytes data in lar file
func (l *loader) getBytes(filename string) ([]byte, error) {
    if cacheByte, ok := l.readiedFile[filename]; ok {
        return cacheByte, nil
    }
    if fr, ok := l.fileReaders[filename]; ok {
        if data, err := ioutil.ReadAll(fr); err != nil {
            return nil, err
        } else {
            l.readiedFile[filename] = data
            return data, nil
        }
    }
    return nil, ErrorLuaFileNotFound
}
