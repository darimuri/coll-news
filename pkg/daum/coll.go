package daum

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	rt "github.com/dormael/go-lib/rodtemplate"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/cdp"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
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

func (p *Portal) GetTopNews() (news []types.News, retErr error) {
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

	return p.collector.GetTopNews(p.PageTemplate, dd)
}

func (p *Portal) GetNewsHomeNews() (news []types.News, retErr error) {
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

	p.PageTemplate.ScreenShotFull(dd.FullScreenShot())

	err := ioutil.WriteFile(dd.FullHTML(), []byte(p.PageTemplate.HTML()), 0644)
	if err != nil {
		return nil, err
	}

	return p.collector.GetNewsHomeNews(p.PageTemplate, dd)
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

	var contentBlock *rt.ElementTemplate

	if p.Has("div[id=daumContent]") {
		contentBlock = p.SelectOrPanic("div[id=daumContent]")
	} else {
		contentBlock = p.SelectOrPanic("div[id=kakaoContent]")
	}

	mainBlockSelector := "div[id=cMain]"
	if false == contentBlock.Has(mainBlockSelector) {
		log.Printf("main block %s is missing in %s\n", mainBlockSelector, n.URL)
		return nil
	}
	mArticleBlock := contentBlock.SelectOrPanic(mainBlockSelector).SelectOrPanic("div[id=mArticle]")

	articleSelector := "div[data-cloud-area=article]"
	videoSelector := "div[id=videoWrap]"

	n.End = &types.End{}

	if true == mArticleBlock.Has(articleSelector) {
		n.End.Category = p.SelectOrPanic("h2[id=kakaoBody]").MustText()

		articleBlock := mArticleBlock.SelectOrPanic(articleSelector)
		headBlock := contentBlock.SelectOrPanic("div[class=head_view]")

		n.End.Provider = util.ImgALT(headBlock.El("em[class=info_cp] > a[class=link_cp]"))
		n.End.Title = headBlock.SelectOrPanic("h3[class=tit_view]").MustText()

		infoBlock := headBlock.SelectOrPanic("span[class=info_view]")

		spanS := infoBlock.Els("span[class=txt_info]")
		for idx := range spanS {
			spText := spanS[idx].MustText()
			switch idx {
			case 0:
				n.End.Author = strings.TrimSpace(spText)
			case 1:
				n.End.PostedAt = strings.TrimSpace(strings.ReplaceAll(spText, "입력", ""))
			case 2:
				n.End.ModifiedAt = strings.TrimSpace(strings.ReplaceAll(spText, "수정", ""))
			}
		}

		counterSelector := "button[id=alexCounter]"
		if true == infoBlock.Has(counterSelector) {
			counterBlock := infoBlock.El(counterSelector)
			n.End.NumComment = counterBlock.El("span[class=alex-count-area]").MustTextAsUInt64()
		}

		n.End.Text = articleBlock.MustText()
		n.End.HTML = p.El("html").MustHTML()

		n.End.Images = make([]string, 0)

		for _, img := range articleBlock.Els("img[class=thumb_g_article]") {
			n.End.Images = append(n.End.Images, util.EmptyIfNilString(img.MustAttribute("src")))
		}
	} else if true == mArticleBlock.Has(videoSelector) {
		innerBlock := mArticleBlock.El(videoSelector).SelectOrPanic("div[class=inner_view]")
		programBlock := innerBlock.SelectOrPanic("h3[class=tit_program]")

		n.End.Program = util.ImgALT(programBlock.SelectOrPanic("span[class=wrap_thumb]"))
		n.End.Provider = programBlock.SelectOrPanic("a[class=btn_allview]").SelectOrPanic("span").MustText()

		contBlock := innerBlock.SelectOrPanic("div[class=box_vod]").SelectOrPanic("div[class=cont_vod]")
		titleBlock := contBlock.SelectOrPanic("h4[class=tit_vod]")
		infoBlock := contentBlock.SelectOrPanic("div[class=info_vod]")

		n.End.Title = titleBlock.SelectOrPanic("span[class=inner_tit]").SelectOrPanic("span[class=inner_tit2]").MustText()

		spans := infoBlock.Els("span")
		for idx := range spans {
			switch idx {
			case 1:
				n.End.NumPlayed = spans[idx].MustTextAsUInt64()
			case 3:
				n.End.PostedAt = strings.TrimSpace(strings.ReplaceAll(spans[idx].MustText(), "등록", ""))
			}
		}
	} else if true == mArticleBlock.Has("div[class=photo_view]") {
		log.Println("skip collect end of photo view")
	} else if true == contentBlock.Has("div[class=view_vod]") {
		log.Println("skip to collect new end for", n.URL)
	} else {
		return fmt.Errorf("failed to collect new end for %s", n.URL)
	}

	return nil
}

func NewPortal(browser *rod.Browser, profile types.Profile, collector types.TypedCollector, dumpRoot string) (types.Collector, error) {
	s := &Portal{BrowserTemplate: rt.NewBrowserTemplate(browser), profile: profile, collector: collector, dumpRoot: dumpRoot}

	return s, nil
}
