package mobile

import (
	"log"
	"strings"

	"github.com/darimuri/go-lib/rodtemplate"
	rt "github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

const (
	topNewsTabSelector = "div[id=channel_news_top]"
)

var _ types.TypedCollector = (*mobile)(nil)

type mobile struct {
}

func New() *mobile {
	return &mobile{}
}

func (_ mobile) Source() string {
	return types.Daum
}

func (_ mobile) Type() string {
	return types.Mobile
}

func (_ mobile) PrepareNewsHomeScreenShot(p *rt.PageTemplate) {
	rt.NewInspectChain(p).ForOne("div.box_g", true, true, func(el *rt.ElementTemplate) error {
		for {
			more := el.El("a.link_more")

			if !more.MustVisible() {
				break
			}

			p.ScrollTo(more)
			more.MustClick()

			p.WaitIdle()
		}

		return nil
	})
}

func (_ mobile) GetNewsHomeNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)
	pageNum := 1

	rt.NewInspectChain(p).ForOne("ul.list_column", true, true, func(el *rt.ElementTemplate) error {
		return nil
	}).ForEach("li", false, true, func(idx int, li *rt.ElementTemplate) error {
		news := extractIssueHomeNews(li)
		news.SetContextData(pageNum, idx, 0, dd, true)
		newsList = append(newsList, news)

		return nil
	})

	return newsList, nil
}

func (_ mobile) GetTopNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)

	if false == p.Has(topNewsTabSelector) {
		return newsList, nil
	}

	newsBlocksSelector := "div.box_rtnews"
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

func extractIssueHomeNews(li *rodtemplate.ElementTemplate) types.News {
	var src string
	if li.Has("div.wrap_thumb") {
		src = util.ImgSrc(li.El("div.wrap_thumb"))
	}

	contItem := li.El("div.cont_thumb > div.inner_thumb > div.thumb_wrap")
	a := li.El("a")
	title := contItem.El("strong.tit_g")
	publisher := contItem.El("span.info_thumb")

	news := types.News{
		URL:       util.EmptyIfNilString(a.MustAttribute("href")),
		Image:     src,
		Title:     title.MustText(),
		Publisher: publisher.MustText(),
	}
	return news
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
