package summary

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/labstack/echo"

	"github.com/darimuri/coll-news/pkg/coll"
)

type Summary struct {
	collectPath string
}

func (h *Summary) Handle(c echo.Context) error {
	files, err := ioutil.ReadDir(h.collectPath)
	if err != nil {
		return err
	}

	platformDir := make([]string, 0)
	for _, f := range files {
		if !f.IsDir() {
			continue
		} else if f.Name() == "." || f.Name() == ".." {
			continue
		}
		platformDir = append(platformDir, f.Name())
	}

	wg := sync.WaitGroup{}

	platforms := make([]Platform, len(platformDir))
	for i, p := range platformDir {
		wg.Add(1)
		go func(myIdx int) {
			defer wg.Done()

			platforms[myIdx] = h.getPlatformSummary(p)
		}(i)
	}

	wg.Wait()

	c.JSON(http.StatusOK, Result{Platforms: platforms})

	return nil
}

func (h *Summary) getPlatformSummary(platform string) Platform {
	res := Platform{Platform: platform, Types: make([]Type, len(coll.GetTypes()))}
	platformPath := filepath.Join(h.collectPath, platform)

	wg := sync.WaitGroup{}

	for i, t := range coll.GetTypes() {
		wg.Add(1)
		go func(myIdx int, t string) {
			defer wg.Done()

			res.Types[myIdx] = Type{
				Type: t,
			}

			typePath := filepath.Join(platformPath, t)
			dumpPath := filepath.Join(typePath, coll.DumpDir)
			dirs, err := ioutil.ReadDir(dumpPath)

			if err != nil {
				if !os.IsNotExist(err) {
					res.Types[myIdx].Error = fmt.Sprintf("failed to get summary of type %s from %s cause by %s", t, platformPath, err.Error())
				}
				return
			}

			for _, d := range dirs {
				if !d.IsDir() {
					continue
				} else if d.Name() == "." || d.Name() == ".." {
					continue
				}

				//TODO: collect yearly directory summary
			}

		}(i, t)
	}

	wg.Wait()

	return res
}

func NewSummary(collectPath string) *Summary {
	return &Summary{collectPath: collectPath}
}
