package pack

import (
    "github.com/DGHeroin/lar/larc"
    "github.com/spf13/cobra"
)

var (
    Cmd = &cobra.Command{
        Use:   "pack <args>",
        Short: "pack a dir",
        RunE: func(cmd *cobra.Command, args []string) error {
            return doPack()
        },
    }
)

func init() {
    Cmd.PersistentFlags().StringVar(&outputFile, "output", "output.lar", "output lar file name")
    Cmd.PersistentFlags().StringVar(&inputDir, "dir", "scripts", "pack dir")
}

var (
    outputFile string
    inputDir   string
)

func doPack() error {
    lar := larc.New()
    if err := lar.Pack(outputFile, inputDir); err != nil {
        return err
    }
    return nil
}
