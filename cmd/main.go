package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/darimuri/coll-news/cmd/api"
	"github.com/darimuri/coll-news/cmd/coll"
	"github.com/darimuri/coll-news/cmd/version"
)

var rootCmd = &cobra.Command{
	Use:   "news",
	Short: "Collect portal news in a given period",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Args: cobra.MinimumNArgs(1),
}

func main() {
	rootCmd.AddCommand(coll.Command, api.Command, version.Command)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
