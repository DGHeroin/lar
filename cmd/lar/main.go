package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/lar"
)

var (
    larFile    = flag.String("f", "", "")
    entryFile  = flag.String("e", "", "")
    runType    = flag.Int("t", 0, "run type [0]")
    searchPath = flag.String("s", "", "search path")
)

func main() {
    flag.Parse()
    L := lar.New()
    L.AddSearchPath(*searchPath)

    switch *runType {
    case 0:
        if err := L.LoadFiles(*larFile); err != nil {
            fmt.Println(err)
        }
        if err := L.DoFile(*entryFile); err != nil {
            fmt.Println(err)
        }
    case 1:
        if err := L.RunDiskFile(*entryFile); err != nil {
            fmt.Println(err)
        }
    default:
        flag.Usage()
    }

}
