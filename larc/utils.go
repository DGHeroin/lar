package larc

import (
    "errors"
    "os"
)

func FileExist(filename string) bool {
    if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
        return false
    }
    return true
}
