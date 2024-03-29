package run

import (
    "flag"
    "github.com/DGHeroin/lar/larc"
    "github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
    Use:   "run <args>",
    Short: "run lar",
    RunE: func(cmd *cobra.Command, args []string) error {
        return doRun()
    },
}

var (
    larFiles   []string
    entryFile  string
    searchPath []string
)

func init() {
    Cmd.PersistentFlags().StringArrayVar(&larFiles, "f", []string{}, "lar files")
    Cmd.PersistentFlags().StringVar(&entryFile, "e", "main.lua", "entry file")
    Cmd.PersistentFlags().StringArrayVar(&searchPath, "s", []string{"./", "scripts"}, "search path")
}

func doRun() error {
    flag.Parse()
    L := larc.New()
    if len(searchPath) > 0 {
        err := L.AddSearchPaths(searchPath...)
        if err != nil {
            return err
        }
    }
    if len(larFiles) > 0 {
        if err := L.LoadFiles(larFiles); err != nil {
            return err
        }
    }

    if err := L.DoFile(entryFile); err != nil {
        return err
    }
    return nil
}
