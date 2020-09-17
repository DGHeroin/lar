package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/lar"
    "path/filepath"
    "strings"
)

var (
    dir = flag.String("d", "", "dir")
    name = flag.String("o", "", "output lar file name")
    root = flag.String("r", "", "root dir of scripts")
)
func main()  {
    flag.Parse()
    var dst string
    if *name == "" {
        dst = filepath.Base(*dir) + ".lar"
    } else {
        dst := *name
        if !strings.HasSuffix(dst, ".lar") {
            dst = dst + ".lar"
        }
    }
    lar := lar.New()
    if err := lar.Pack(dst, *dir, *root); err != nil{
        fmt.Println(err)
    }
}

