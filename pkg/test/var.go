package test

import (
	"os"
)

var LaunchHeadless bool = false

func init() {
	if "" != os.Getenv("TEST_HEADLESS") {
		LaunchHeadless = true
	}
}
