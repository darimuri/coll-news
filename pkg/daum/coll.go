package daum

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/darimuri/coll-news/pkg/cache"
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
	cache cache.Cache

	profile   types.Profile
	collector types.TypedCollector
	dumpRoot  string
}

func (p *Portal) Cleanup() {
	for _, pg := range p.MustPages() {
		pg.MustClose()
	}
	p.Browser.MustClose()
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
		retErr = panicAsError(v)
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
		retErr = panicAsError(v)
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

func panicAsError(v interface{}) (retErr error) {
	if v == nil {
		return
	}

	switch t := v.(type) {
	case error:
		retErr = fmt.Errorf(fmt.Sprintf("panic '%s' from stack:\n%s", t.Error(), string(debug.Stack())))
	default:
		retErr = fmt.Errorf("errorless panic %+v from %s", v, string(debug.Stack()))
	}
	return
}

func (p *Portal) GetNewsEnd(n *types.News) (retErr error) {
	var end interface{}

	defer func() {
		v := recover()
		if v == nil {
			return
		}

		log.Println("panic occurred while getting new end for url", n.URL)
		switch t := v.(type) {
		case error:
			log.Println("panic message", t)
		}
		panic(v)
	}()

	cacheKey := asKey(n.URL)
	end, retErr = p.cache.Get(cacheKey, &types.End{})
	if retErr != nil {
		return
	}

	if end != nil {
		n.End = end.(*types.End)
		return
	}

	p.openTab(n.URL)

	collectedAt := time.Now()

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

	retErr = p.collector.GetNewsEnd(p.PageTemplate, n)
	if retErr != nil {
		return
	}

	if n.End != nil {
		n.End.CollectedAt = collectedAt.Format(types.DataDateTimeFormat)
		n.End.PostedAt = convertToDataFormat(n.End.PostedAt)
		n.End.ModifiedAt = convertToDataFormat(n.End.ModifiedAt)
	}

	retErr = p.cache.Set(cacheKey, n.End, time.Minute*3)

	return
}

func convertToDataFormat(at string) string {
	if at == "" {
		return ""
	}

	layout := "2006. 01. 02. 15:04"
	layoutFallback := "2006.01.02 15:04"
	dt, err := time.ParseInLocation(layout, at, time.Local)
	if err != nil {
		log.Println("failed to parse", at, "with", layout, "for", err.Error())
		dt, err = time.Parse(layoutFallback, at)
		if err != nil {
			log.Println("failed to parse", at, "with", layoutFallback, "for", err.Error())
			return at
		}
	}

	return dt.Format(types.DataDateTimeFormat)
}

func asKey(k string) string {
	u, err := url.Parse(k)
	if err != nil {
		panic(fmt.Errorf("failed to parse url %s for error: %v", k, err))
	}

	return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
}

func NewPortal(browser *rod.Browser, profile types.Profile, collector types.TypedCollector, dumpRoot string, endCache cache.Cache) (types.Collector, error) {
	s := &Portal{BrowserTemplate: rt.NewBrowserTemplate(browser), profile: profile, collector: collector, dumpRoot: dumpRoot, cache: endCache}

	return s, nil
}
