package mobile

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/darimuri/go-lib/rodtemplate"
	rt "github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

const (
	topNewsTabSelector = "div[id=channel_news1_top]"
)

var _ types.TypedCollector = (*mobile)(nil)

type mobile struct {
}

func New() *mobile {
	return &mobile{}
}

func (_ mobile) PrepareNewsHomeScreenShot(p *rt.PageTemplate) {
	mainBlockSelector := "main[id=kakaoContent]"
	if false == p.Has(mainBlockSelector) {
		return
	}
	mainBlock := p.SelectOrPanic(mainBlockSelector)
	sectionMainBlock := mainBlock.SelectOrPanic("div.section_main")
	mainNewsSelector := "div[data-tiara-layer=MAIN_NEWS]"
	if true == sectionMainBlock.Has(mainNewsSelector) {
		mainNewsBlock := mainBlock.SelectOrPanic(mainNewsSelector)

		p.ScrollTo(mainNewsBlock)
		p.WaitRepaint()

		moreSelector := "a.link_more"
		if true == mainNewsBlock.Has(moreSelector) {
			for mainNewsBlock.El(moreSelector).MustVisible() {
				mainNewsBlock.El(moreSelector).MustClick()
				p.WaitRepaint()

				time.Sleep(100 * time.Microsecond)
			}
		}
	}
}

func (_ mobile) GetNewsHomeNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)
	pageNum := 1

	mainBlockSelector := "main[id=kakaoContent]"
	if false == p.Has(mainBlockSelector) {
		return newsList, nil
	}

	mainBlock := p.SelectOrPanic(mainBlockSelector)

	mainSectionSelfChain := rt.NewInspectChain(mainBlock).ForOne("div.section_main", true, true, func(sectionMainBlock *rt.ElementTemplate) error {
		return nil
	}).SelfChain()

	mainSectionSelfChain.ForOne("div.box_homeissue", false, true, func(homeIssueBlock *rt.ElementTemplate) error {
		if homeIssueBlock == nil {
			return nil
		}

		rt.NewInspectChain(homeIssueBlock).ForEach("ul.list_homeissue > li", false, true, func(idx int, b *rt.ElementTemplate) error {
			thumb := extractThumb(b)
			thumb.SetContextData(pageNum, idx, 0, dd, false)
			newsList = append(newsList, thumb)

			contSubSelector := "div.cont_sub"
			if true == b.Has(contSubSelector) {
				thumbSub := extractThumbSub(b, contSubSelector)
				thumbSub.SetContextData(pageNum, idx, 1, dd, false)
				newsList = append(newsList, thumbSub)
			}

			return nil
		})

		return nil
	})

	pageNum++

	mainSectionSelfChain.ForOne("div[data-tiara-layer=MAIN_NEWS]", false, true, func(mainNewsBlock *rt.ElementTemplate) error {
		if mainNewsBlock == nil {
			return nil
		}

		p.ScrollTo(mainNewsBlock)
		p.WaitRepaint()

		order := newsList[len(newsList)-1].Order + 1

		newsList = append(newsList, extractGenericNewsList(mainNewsBlock, pageNum, order, dd)...)

		return nil
	})

	subSectionSelectors := []string{
		"div[data-tiara-layer=POPULAR]",
		"div[data-tiara-layer=DRI]",
		"div.box_cmtrank",
	}

	rt.NewInspectChain(mainBlock).ForOne("div.section_sub", true, true, func(sectionSubBlock *rt.ElementTemplate) error {
		for _, selector := range subSectionSelectors {
			pageNum++
			if true == sectionSubBlock.Has(selector) {
				popularBlock := sectionSubBlock.SelectOrPanic(selector)

				p.ScrollTo(popularBlock)
				p.WaitRepaint()

				order := newsList[len(newsList)-1].Order + 1

				items := extractGenericNewsList(popularBlock, pageNum, order, dd)
				newsList = append(newsList, items...)
			}
		}

		rt.NewInspectChain(sectionSubBlock).ForEach("div.box_agenews > ul > li", false, true, func(_ int, t *rt.ElementTemplate) error {
			t.MustClick()
			p.WaitRepaint()

			order := newsList[len(newsList)-1].Order + 1

			for idx, b := range sectionSubBlock.Els("div.tab_slide > div.slide > div.panel > ul.list_news > div.slide > div.panel > li") {
				n := types.News{
					Title: b.MustText(),
					URL:   util.AnchorHREF(b),
				}
				n.SetContextData(pageNum, order, idx, dd, false)
				newsList = append(newsList, n)
			}

			return nil
		})

		return nil
	})

	return newsList, nil
}

func (_ mobile) GetTopNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)

	if false == p.Has(topNewsTabSelector) {
		return newsList, nil
	}

	newsBlocksSelector := "div._box_feed_news1"
	if false == p.Has(newsBlocksSelector) {
		return newsList, nil
	}

	pageNum := 1
	textListSelector := "ul.list_txt"
	thumbListSelector := "ul.list_thumb"
	horizonBlockSelector := "ul.list_horizon"
	themeBlockSelector := "ul.list_theme"
	//adBlockSelector := "div.mtop_adfit_channel_news1"

	//yDelta := 0.0
	//adBlocks := p.Els(adBlockSelector)

	newsBlocks := p.Els(newsBlocksSelector)
	for _, b := range newsBlocks {
		var extractor func(item *rodtemplate.ElementTemplate, idx int) types.News
		var items rodtemplate.ElementsTemplate
		if true == b.Has(textListSelector) {
			items = b.El(textListSelector).Els("li")
			extractor = extractTextItem
		} else if true == b.Has(horizonBlockSelector) {
			items = b.El(horizonBlockSelector).Els("li")
			extractor = extractHorizonItem
		} else if true == b.Has(thumbListSelector) {
			items = b.El(thumbListSelector).Els("li")
			extractor = extractThumbItem
		} else if true == b.Has(themeBlockSelector) {
			themeBlocks := b.Els(themeBlockSelector)
			items = make([]*rodtemplate.ElementTemplate, 0)
			for _, tb := range themeBlocks {
				if false == tb.Has("li") {
					continue
				}
				for _, item := range tb.Els("li") {
					items = append(items, item)
				}
			}
			extractor = extractThemeItem
		} else {
			log.Println("failed to get news items from news block", b.MustHTML())
			continue
		}

		if extractor != nil && items != nil {
			p.ScrollTo(b)
			p.WaitRepaint()

			for idx, item := range items {
				news := extractor(item, idx)

				news.NewsPage = pageNum
				news.FullHTML = dd.FullHTML()
				news.FullScreenShot = dd.FullScreenShot()
				//news.TabScreenShot = dd.TabScreenShot(pageNum)

				newsList = append(newsList, news)
			}

			//if bidx > 0 && bidx < len(adBlocks) {
			//	if adBlocks[bidx-1].MustVisible() {
			//		yDelta += adBlocks[bidx-1].Height()
			//	}
			//}
			//
			//myDelta := 0.0
			//if bidx == 2 {
			//	myDelta += p.GetVisibleHeight("div.d_head")
			//	myDelta += p.GetVisibleHeight("div.disaster_weather")
			//	myDelta += p.GetVisibleHeight("div.slidebox_menu") * 2
			//	myDelta += p.GetVisibleHeight("div.box_promotion")
			//	myDelta += p.GetVisibleHeight("div.ibox_issue")
			//	myDelta += p.GetVisibleHeight("div.tb_txt")
			//	myDelta += p.GetVisibleHeight("div.box_direct")
			//}
			//yDelta += myDelta
			//
			//p.ScreenShot(b, dd.TabScreenShot(pageNum), yDelta)
			pageNum++
		}
	}

	return newsList, nil
}

func extractThumbSub(b *rt.ElementTemplate, contSubSelector string) types.News {
	subBlock := b.El(contSubSelector)
	contSub := subBlock.El("span.inner_link")
	n2 := types.News{
		Publisher: contSub.El("span.txt_cp").MustText(),
		Title:     contSub.El("span.tit_sub").MustText(),
		URL:       util.AnchorHREF(subBlock),
	}
	return n2
}

func extractThumb(b *rt.ElementTemplate) types.News {
	contBlock := b.El("div.cont_thumb")
	contMain := contBlock.El("span.inner_link")
	n1 := types.News{
		Publisher: contMain.El("span.txt_cp").MustText(),
		Title:     contMain.El("strong.tit_thumb").MustText(),
		URL:       util.AnchorHREF(contBlock),
		Image:     util.ImgSrc(b.El("div.wrap_thumb")),
	}
	return n1
}

func extractGenericNewsList(listBlock *rt.ElementTemplate, pageNum int, order int, dd types.DumpDirectory) []types.News {
	items := make([]types.News, 0)

	numBanners := 0
	for idx, b := range listBlock.Els("ul > li") {
		if b.MustAttribute("class") != nil && strings.Contains(*b.MustAttribute("class"), "item_bnr") {
			numBanners++
			continue
		}
		var contBlock *rt.ElementTemplate

		if b.Has("div.cont_thumb > strong.tit_thumb") {
			contBlock = b.El("div.cont_thumb > strong.tit_thumb")
		} else if b.Has("div.cont_thumb > strong.tit_news") {
			contBlock = b.El("div.cont_thumb > strong.tit_news")
		} else if b.Has("div.item_cmtrank > strong.tit_cmtrank") {
			contBlock = b.El("div.item_cmtrank > strong.tit_cmtrank")
		} else if b.Has("a.link_correction") {
			log.Println("skip link correction list html", b.MustHTML())
			numBanners++
			continue
		} else {
			panic(fmt.Errorf("failed to find content block from html %s", b.MustHTML()))
		}

		var myPublisher, myTitle string
		if contBlock.Has("span.txt_cp") {
			myPublisher = contBlock.El("span.txt_cp").MustText()
		}

		if contBlock.Has("span.txt_g") {
			myTitle = contBlock.El("span.txt_g").MustText()
		} else {
			myTitle = strings.TrimSpace(contBlock.MustText())
		}

		n := types.News{
			Publisher: myPublisher,
			Title:     myTitle,
			URL:       util.AnchorHREF(b),
		}
		n.SetContextData(pageNum, order, idx-numBanners, dd, false)

		if true == b.Has("div.wrap_thumb") {
			n.Image = util.ImgSrc(b.El("div.wrap_thumb"))
		}

		items = append(items, n)
	}
	return items
}

func extractThemeItem(item *rodtemplate.ElementTemplate, idx int) types.News {
	contentBlock := item.El("div[class=cont_item]")
	titleBlock := contentBlock.El("strong[class=tit_item]")

	var title, seriesTitle string
	if titleBlock.Has("em") {
		seriesBlock := titleBlock.El("em")
		seriesTitle = seriesBlock.MustText()

		title = strings.ReplaceAll(titleBlock.MustText(), seriesTitle, "")
	} else {
		title = titleBlock.MustText()
	}

	news := types.News{
		URL:         util.AnchorHREF(item),
		Title:       title,
		SeriesTitle: seriesTitle,
		Publisher:   contentBlock.El("span").MustText(),
		Order:       idx,
	}
	return news
}

func extractThumbItem(item *rodtemplate.ElementTemplate, idx int) types.News {
	news := types.News{
		URL:   util.AnchorHREF(item),
		Image: util.ImgSrc(item),
		Title: item.El("div[class=cont_item] > strong[class=tit_item]").MustText(),
		Order: idx,
	}
	return news
}

func extractHorizonItem(item *rodtemplate.ElementTemplate, idx int) types.News {
	a := item.El("a")

	var title, imgSrc string
	if true == a.Has("div.wrap_thumb") {
		imgSrc = util.ImgSrc(a.El("div.wrap_thumb"))
		title = a.El("strong.tit_item").MustText()
	} else if true == a.Has("span.link_txt") {
		title = a.El("span.tit_news").MustText()
	}

	news := types.News{
		URL:   util.EmptyIfNilString(a.MustAttribute("href")),
		Image: imgSrc,
		Title: title,
		Order: idx,
	}
	return news
}

func extractTextItem(item *rodtemplate.ElementTemplate, idx int) types.News {
	a := item.El("a")
	news := types.News{
		URL:   util.EmptyIfNilString(a.MustAttribute("href")),
		Title: a.MustText(),
		Order: idx,
	}
	return news
}
