package coll

import (
	"time"

	"github.com/spf13/cobra"
)

const (
	mobile = "mobile"
	pc     = "pc"

	daum  = "daum"
	naver = "naver"
)

var (
	collectDuration time.Duration
	collectType     string
	collectPortal   string
	collectSavePath string
	enableHeadless  bool
)

var Command = &cobra.Command{
	Use:   "coll",
	Short: "Collect portal news in a given duration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	Command.PersistentFlags().DurationVarP(&collectDuration, "duration", "d", time.Minute, "duration of every news collection")
	Command.PersistentFlags().StringVarP(&collectType, "type", "t", "", "collect news type(pc/mobile)")
	Command.PersistentFlags().StringVarP(&collectType, "portal", "p", "", "collect news portal(daum/naver)")
	Command.PersistentFlags().StringVarP(&collectSavePath, "save-path", "s", "", "save path for collected data")
	Command.PersistentFlags().BoolVarP(&enableHeadless, "headless", "", true, "collect in headless mode")

	Command.MarkFlagRequired("type")
	Command.MarkFlagRequired("portal")
	Command.MarkFlagRequired("save-path")
}
