package cmd

import (
    "fmt"
    "github.com/DGHeroin/lar/cmd/pack"
    "github.com/DGHeroin/lar/cmd/run"
    "github.com/spf13/cobra"
)

var (
    Cmd = &cobra.Command{}
)

func Execute() {
    Cmd.AddCommand(run.Cmd, pack.Cmd)
    if e := Cmd.Execute(); e != nil {
        fmt.Println(e)
    }
}
