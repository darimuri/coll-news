package daum

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	rt "github.com/dormael/go-lib/rodtemplate"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/cdp"

	"github.com/darimuri/coll-news/pkg/types"
)

const (
	topNewsURL  = "https://www.daum.net/" //https://m.daum.net/?nil_top=mobile
	newsHomeURL = "https://news.daum.net/"
)

var _ types.Collector = (*Portal)(nil)

type Portal struct {
	*rt.BrowserTemplate
	*rt.PageTemplate

	profile   types.Profile
	collector types.TypedCollector
	dumpRoot  string
}

func (p *Portal) Cleanup() {
	for _, pg := range p.MustPages() {
		pg.MustClose()
	}
}

func (p *Portal) Top() {
	p.open(topNewsURL)
}

func (p *Portal) NewsHome() {
	p.open(newsHomeURL)
}

func (p *Portal) open(url string) {
	page := p.BrowserTemplate.MustPage(url)
	p.PageTemplate = rt.NewPageTemplate(page)
	p.SetViewport(p.profile.Width, p.profile.Height)

	if err := page.WaitLoad(); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "occurred occasionally but has no problem")
	}

	if err := page.WaitIdle(time.Minute * 10); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "failed to wait idle for 10m, but has no problem")
	}
}

func (p *Portal) openTab(url string) {
	page := p.BrowserTemplate.MustPages().First().MustNavigate(url)
	p.PageTemplate = rt.NewPageTemplate(page)
	p.SetViewport(p.profile.Width, p.profile.Height)

	if err := page.WaitLoad(); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "occurred occasionally but has no problem")
	}

	if err := page.WaitIdle(time.Minute * 10); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "failed to wait idle for 10m, but has no problem")
	}
}

func (p *Portal) GetTopNewsList() (news []types.News, retErr error) {
	defer func() {
		v := recover()
		if v == nil {
			return
		}

		switch t := v.(type) {
		case error:
			retErr = t
		default:
			retErr = fmt.Errorf("errorless panic %+v", v)
		}
	}()

	p.ScrollBottomHuman()
	p.WaitLoadAndIdle()

	dd := types.DumpDirectory{RootPath: p.dumpRoot, Source: "top", DumpTime: time.Now()}
	if err := dd.Init(); err != nil {
		return nil, err
	}

	p.ScreenShotFull(dd.FullScreenShot())

	err := ioutil.WriteFile(dd.FullHTML(), []byte(p.PageTemplate.HTML()), 0644)
	if err != nil {
		return nil, err
	}

	return p.collector.GetTopNewsList(p.PageTemplate, dd)
}

func (p *Portal) GetNewsHomeNewsList() (news []types.News, retErr error) {
	defer func() {
		v := recover()
		if v == nil {
			return
		}

		switch t := v.(type) {
		case error:
			retErr = t
		default:
			retErr = fmt.Errorf("errorless panic %+v", v)
		}
	}()

	dd := types.DumpDirectory{RootPath: p.dumpRoot, Source: "news", DumpTime: time.Now()}
	if err := dd.Init(); err != nil {
		return nil, err
	}

	p.collector.PrepareNewsHomeScreenShot(p.PageTemplate)
	p.PageTemplate.ScreenShotFull(dd.FullScreenShot())

	err := ioutil.WriteFile(dd.FullHTML(), []byte(p.PageTemplate.HTML()), 0644)
	if err != nil {
		return nil, err
	}

	return p.collector.GetNewsHomeNewsList(p.PageTemplate, dd)
}

func (p *Portal) GetNewsEnd(n *types.News) (retErr error) {
	p.openTab(n.URL)

	defer func() {
		v := recover()
		if v == nil {
			return
		}

		switch t := v.(type) {
		case error:
			retErr = t
		default:
			retErr = fmt.Errorf("errorless panic %+v", v)
		}

	}()

	return p.collector.GetNewsEnd(p.PageTemplate, n)
}

func NewPortal(browser *rod.Browser, profile types.Profile, collector types.TypedCollector, dumpRoot string) (types.Collector, error) {
	s := &Portal{BrowserTemplate: rt.NewBrowserTemplate(browser), profile: profile, collector: collector, dumpRoot: dumpRoot}

	return s, nil
}
