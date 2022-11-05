package adaptor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	rt "github.com/darimuri/go-lib/rodtemplate"
	"github.com/go-rod/rod/lib/cdp"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/darimuri/coll-news/pkg/cache"
	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

type Status string

const (
	StatusCPBlockNotFound        = Status("CPBlockNotFound")
	StatusEndSkippedOnPurpose    = Status("EndSkippedOnPurpose")
	StatusEndSkippedUnexpectedly = Status("EndSkippedUnexpectedly")
	StatusOK                     = Status("OK")
	StatusUndefinedError         = Status("UndefinedError")
)

var _ error = (*TypedError)(nil)

type TypedError struct {
	status Status
	err    error
}

func NewTypedError(err string) TypedError {
	return TypedError{err: errors.New(err)}
}

func (t TypedError) Error() string {
	return t.err.Error()
}

var (
	CPBlockNotFound        = TypedError{err: errors.New("content provider block is missing"), status: StatusCPBlockNotFound}
	EndSkippedOnPurpose    = TypedError{err: errors.New("skip to collect news end on purpose"), status: StatusEndSkippedOnPurpose}
	EndSkippedUnexpectedly = TypedError{err: errors.New("skip to collect news end unexpectedly"), status: StatusEndSkippedUnexpectedly}
)

type Adaptor struct {
	*rt.BrowserTemplate
	*rt.PageTemplate
	Cache cache.Cache

	Profile   types.Profile
	Collector types.TypedCollector
	DumpRoot  string
}

func (a *Adaptor) Cleanup() {
	for _, pg := range a.MustPages() {
		pg.MustClose()
	}
	a.Browser.MustClose()
}

func (a *Adaptor) Open(url string) {
	page := a.BrowserTemplate.MustPage(url)

	log.Println("open url", url)

	a.PageTemplate = rt.NewPageTemplate(page)
	a.SetViewport(a.Profile.Width, a.Profile.Height)

	if err := page.Timeout(time.Second * 15).WaitLoad(); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "occurred occasionally but has no problem")
	}

	if err := page.WaitIdle(time.Second * 10); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "failed to wait idle for 10m, but has no problem")
	}
}

func (a *Adaptor) OpenTab(url string) {
	page := a.BrowserTemplate.MustPages().First().MustNavigate(url)

	log.Println("open tab url", url)

	a.PageTemplate = rt.NewPageTemplate(page)
	a.SetViewport(a.Profile.Width, a.Profile.Height)

	if err := page.Timeout(time.Second * 15).WaitLoad(); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "occurred occasionally but has no problem")
	}

	if err := page.WaitIdle(time.Second * 10); err != nil {
		if false == cdp.ErrCtxDestroyed.Is(err) {
			panic(err)
		}
		log.Println(err.Error(), "failed to wait idle for 10m, but has no problem")
	}
}

func (a *Adaptor) GetTopNewsList() (news []types.News, retErr error) {
	defer func() {
		v := recover()
		retErr = util.PanicAsError(v)
	}()

	a.ScrollBottomHuman()
	a.WaitLoadAndIdle()

	dd := types.DumpDirectory{RootPath: a.DumpRoot, Source: "top", DumpTime: time.Now()}
	if err := dd.Init(); err != nil {
		return nil, err
	}

	a.ScreenShotFull(dd.FullScreenShot())

	err := ioutil.WriteFile(dd.FullHTML(), []byte(a.PageTemplate.HTML()), 0644)
	if err != nil {
		return nil, err
	}

	timer := prometheus.NewTimer(newsListDuration.WithLabelValues(a.Collector.Type(), a.Collector.Source(), types.Top))
	defer func() {
		timer.ObserveDuration()
		newsLen := len(news)
		if newsLen > 0 {
			totalNewsCollected.WithLabelValues(a.Collector.Type(), a.Collector.Source(), types.Top).Add(float64(newsLen))
		}
	}()

	return a.Collector.GetTopNewsList(a.PageTemplate, dd)
}

func (a *Adaptor) GetNewsHomeNewsList() (news []types.News, retErr error) {
	defer func() {
		v := recover()
		retErr = util.PanicAsError(v)
	}()

	dd := types.DumpDirectory{RootPath: a.DumpRoot, Source: "news", DumpTime: time.Now()}
	if err := dd.Init(); err != nil {
		return nil, err
	}

	a.Collector.PrepareNewsHomeScreenShot(a.PageTemplate)
	a.PageTemplate.ScreenShotFull(dd.FullScreenShot())

	err := ioutil.WriteFile(dd.FullHTML(), []byte(a.PageTemplate.HTML()), 0644)
	if err != nil {
		return nil, err
	}

	timer := prometheus.NewTimer(newsListDuration.WithLabelValues(a.Collector.Type(), a.Collector.Source(), types.Home))
	defer func() {
		timer.ObserveDuration()
		newsLen := len(news)
		if newsLen > 0 {
			totalNewsCollected.WithLabelValues(a.Collector.Type(), a.Collector.Source(), types.Home).Add(float64(newsLen))
		}
	}()

	return a.Collector.GetNewsHomeNewsList(a.PageTemplate, dd)
}

func (a *Adaptor) GetNewsEnd(n *types.News) (retErr error) {
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
			retErr = t
			return
		}
		panic(v)
	}()

	cacheKey := asKey(n.URL)
	end, retErr = a.Cache.Get(cacheKey, &types.End{})
	if retErr != nil {
		return
	}

	if end != nil {
		n.End = end.(*types.End)
		return
	}

	timer := prometheus.NewTimer(newsEndDuration.WithLabelValues(a.Collector.Type(), a.Collector.Source()))
	defer func() {
		timer.ObserveDuration()
		status := StatusOK
		if retErr != nil {
			if te, ok := retErr.(TypedError); ok {
				status = te.status
			} else {
				status = StatusUndefinedError
			}
		}

		newsEndStatus.WithLabelValues(a.Collector.Type(), a.Collector.Source(), string(status)).Inc()
	}()

	a.OpenTab(n.URL)

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

	retErr = a.Collector.GetNewsEnd(a.PageTemplate, n)
	if retErr != nil {
		return
	}

	if n.End != nil {
		n.End.CollectedAt = collectedAt.Format(types.DataDateTimeFormat)
		n.End.PostedAt = convertToDataFormat(n.End.PostedAt)
		n.End.ModifiedAt = convertToDataFormat(n.End.ModifiedAt)
	}

	retErr = a.Cache.Set(cacheKey, n.End, time.Minute*3)

	return
}

func convertToDataFormat(at string) string {
	if at == "" {
		return ""
	}

	layouts := []string{
		"2006.01.02. 15:04",
		"2006.01.02 15:04",
		"2006. 01. 2. 15:04",
		"2006. 01.2. 15:04",
	}

	var lastError error
	for _, l := range layouts {
		dt, err := time.ParseInLocation(l, at, time.Local)

		if err == nil {
			return dt.Format(types.DataDateTimeFormat)
		}

		lastError = err
	}

	if lastError != nil {
		log.Println("failed to parse", at)
	}

	return at
}

func asKey(k string) string {
	u, err := url.Parse(k)
	if err != nil {
		panic(fmt.Errorf("failed to parse url %s for error: %v", k, err))
	}

	return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
}
