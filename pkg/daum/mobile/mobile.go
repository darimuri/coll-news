package mobile

import (
	"log"

	"github.com/dormael/go-lib/rodtemplate"

	_interface "github.com/darimuri/coll-news/pkg/daum/interface"
	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

const (
	topNewsTabSelector = "div[id=channel_news1_top]"
)

var _ _interface.Collector = (*mobile)(nil)

type mobile struct {
}

func (_ mobile) GetTopNews(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)

	if false == p.Has(topNewsTabSelector) {
		return newsList, nil
	}

	newsBlocksSelector := "div._box_feed_news1"
	if false == p.Has(newsBlocksSelector) {
		return newsList, nil
	}

	pageNum := 0
	textListSelector := "ul.list_txt"
	thumbListSelector := "ul.list_thumb"
	horizonBlockSelector := "ul.list_horizon"
	//adBlockSelector := "div.mtop_adfit_channel_news1"

	//yDelta := 0.0
	//adBlocks := p.Els(adBlockSelector)

	newsBlocks := p.Els(newsBlocksSelector)
	for _, b := range newsBlocks {
		var parser func(item *rodtemplate.ElementTemplate, idx int) types.News
		var items rodtemplate.ElementsTemplate
		if true == b.Has(textListSelector) {
			items = b.El(textListSelector).Els("li")
			parser = parseTextItem
		} else if true == b.Has(horizonBlockSelector) {
			items = b.El(horizonBlockSelector).Els("li")
			parser = parseHorizonItem
		} else if true == b.Has(thumbListSelector) {
			items = b.El(thumbListSelector).Els("li")
			parser = parseThumbItem

		} else {
			log.Println("failed to get news items from news block", b.MustHTML())
			continue
		}

		if parser != nil && items != nil {
			p.ScrollTo(b)
			p.WaitRepaint()

			for idx, item := range items {
				news := parser(item, idx)

				news.NewsPage = pageNum
				news.FullHTML = dd.FullHTML()
				news.FullScreenShot = dd.FullScreenShot()
				news.TabScreenShot = dd.TabScreenShot(pageNum)

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

func parseThumbItem(item *rodtemplate.ElementTemplate, idx int) types.News {
	news := types.News{
		URL:   util.AnchorHREF(item),
		Image: util.ImgSrc(item),
		Title: item.El("div[class=cont_item] > strong[class=tit_item]").MustText(),
		Order: idx,
	}
	return news
}

func parseHorizonItem(item *rodtemplate.ElementTemplate, idx int) types.News {
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

func parseTextItem(item *rodtemplate.ElementTemplate, idx int) types.News {
	a := item.El("a")
	news := types.News{
		URL:   util.EmptyIfNilString(a.MustAttribute("href")),
		Title: a.MustText(),
		Order: idx,
	}
	return news
}

func (_ mobile) GetNewsHomeNews(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	panic("implement me")
}

func New() *mobile {
	return &mobile{}
}
