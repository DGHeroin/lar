package main

import (
    "flag"
    "fmt"
    "github.com/DGHeroin/lar"
)

var (
    larFile = flag.String("f", "", "")
    entryFile = flag.String("e", "", "")
)

func main()  {
    flag.Parse()
    L := lar.New()
    if err := L.Load(*larFile, *entryFile); err != nil {
        fmt.Println(err)
    }
}
