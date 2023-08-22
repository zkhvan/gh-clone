package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdRepos "github.com/zkhvan/gh-clone/internal/cmd/repos"
)

func main() {
	rootCmd := &cobra.Command{
		Use:           "clone",
		Short:         "Clone repositories.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(cmdRepos.NewCmdRepos(nil))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
