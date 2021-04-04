package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "version",
	Short: "Print the version number news coll",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Portal News Collector v0.1 -- HEAD")
	},
}
