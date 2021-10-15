package naver

import (
	rt "github.com/darimuri/go-lib/rodtemplate"
	"github.com/go-rod/rod"

	"github.com/darimuri/coll-news/pkg/adaptor"
	"github.com/darimuri/coll-news/pkg/cache"
	"github.com/darimuri/coll-news/pkg/types"
)

const (
	topNewsURL  = "https://www.naver.com/"
	newsHomeURL = "https://news.naver.com/"
)

var _ types.Collector = (*Collector)(nil)

type Collector struct {
	*adaptor.Adaptor
}

func (c *Collector) Top() {
	c.Open(topNewsURL)
}

func (c *Collector) NewsHome() {
	c.Open(newsHomeURL)
}

func NewPortal(browser *rod.Browser, profile types.Profile, collector types.TypedCollector, dumpRoot string, endCache cache.Cache) (types.Collector, error) {
	s := &Collector{
		Adaptor: &adaptor.Adaptor{BrowserTemplate: rt.NewBrowserTemplate(browser), Profile: profile, Collector: collector, DumpRoot: dumpRoot, Cache: endCache},
	}

	return s, nil
}
