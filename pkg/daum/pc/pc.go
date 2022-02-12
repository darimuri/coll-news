package pc

import (
	"strconv"
	"strings"

	"github.com/darimuri/go-lib/rodtemplate"
	rt "github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

var _ types.TypedCollector = (*pc)(nil)

type pc struct {
}

func New() *pc {
	return &pc{}
}

func (*pc) PrepareNewsHomeScreenShot(_ *rt.PageTemplate) {
}

func (*pc) GetNewsHomeNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)
	pageNum := 1

	rt.NewInspectChain(p).ForOne("ul.list_newsissue", false, true, func(el *rt.ElementTemplate) error {
		return nil
	}).ForEach("li", true, true, func(idx int, li *rt.ElementTemplate) error {
		news := extractIssueHomeNews(li)
		news.SetContextData(pageNum, idx, 0, dd, true)
		newsList = append(newsList, news)

		return nil
	})

	return newsList, nil
}

func (*pc) GetTopNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)

	mediaTabSelector := "div[id=mediaTab]"
	if false == p.Has(mediaTabSelector) {
		return newsList, nil
	}

	newMapByTab := make(map[int]string, 0)

	mediaBlock := p.El(mediaTabSelector)

	newsPagerBlock := mediaBlock.El("div[class=page_tabcont]")

	for i := 0; i < 10; i++ {
		mediaBlock.MustWaitLoad()
		mediaBlock.MustWaitStable()
		mediaBlock.MustWaitVisible()

		currentNewsPage := newsPagerBlock.El("strong[class=screen_out]").MustText()
		currentMediaPage := mediaBlock.El("strong[class=num_index]").MustText()

		pageNum, err := strconv.Atoi(currentMediaPage)
		if err != nil {
			return nil, err
		}

		if _, ok := newMapByTab[pageNum]; ok {
			newsPagerBlock.MustClick()
			continue
		}

		p.ScreenShot(mediaBlock, dd.TabScreenShot(pageNum), 0)

		groupNewsChain := rt.NewInspectChain(mediaBlock).ForOne("div[class=group_news]", true, true, func(el *rt.ElementTemplate) error {
			return nil
		})

		groupNewsChain.ForEach("ul[class=list_thumb] > li", false, true, func(idx int, item *rt.ElementTemplate) error {
			news := types.News{
				URL:   util.AnchorHREF(item),
				Image: util.ImgSrc(item),
				Title: item.El("div[class=cont_item] > strong[class=tit_item]").MustText(),
			}
			news.SetContextData(pageNum, idx, 0, dd, true)
			newsList = append(newsList, news)

			return nil
		})

		groupNewsChain.ForEach("ul[class=list_txt] > li", false, true, func(idx int, item *rt.ElementTemplate) error {
			a := item.El("a")
			news := types.News{
				URL:   util.EmptyIfNilString(a.MustAttribute("href")),
				Title: a.MustText(),
			}
			news.SetContextData(pageNum, idx, 0, dd, true)
			newsList = append(newsList, news)
			return nil
		})

		newMapByTab[pageNum] = currentNewsPage

		if len(newMapByTab) == 2 {
			break
		}

		newsPagerBlock.MustClick()
	}

	return newsList, nil
}

func extractIssueRelate(div *rt.ElementTemplate) types.News {
	a := div.El("a")
	rspan := div.El("span[class=info_news]")

	rnews := types.News{
		URL:       util.EmptyIfNilString(a.MustAttribute("href")),
		Title:     a.MustText(),
		Publisher: rspan.MustText(),
	}
	return rnews
}

func extractIssueHomeNews(li *rodtemplate.ElementTemplate) types.News {
	var src string
	if li.Has("div.item_issue") {
		src = util.ImgSrc(li.El("div.item_issue"))
	}

	contItem := li.El("div[class=cont_thumb]")
	a := contItem.El("strong > a")
	imgPublisher := contItem.El("span.info_thumb > span.logo_cp > img.thumb_g")
	spanCategory := contItem.El("span.txt_category")

	news := types.News{
		URL:       util.EmptyIfNilString(a.MustAttribute("href")),
		Image:     src,
		Title:     a.MustText(),
		Category:  spanCategory.MustText(),
		Publisher: *imgPublisher.MustAttribute("alt"),
	}
	return news
}

func extractIssue(li *rt.ElementTemplate) types.News {
	src := util.ImgSrc(li.El("div[class=item_issue]"))

	contItem := li.El("div[class=cont_thumb]")
	a := contItem.El("strong > a")
	span := contItem.El("span[class=info_thumb]")

	news := types.News{
		URL:       util.EmptyIfNilString(a.MustAttribute("href")),
		Image:     src,
		Title:     a.MustText(),
		Publisher: span.MustText(),
	}
	return news
}

func extractHeadlineSub(li *rt.ElementTemplate) types.News {
	classAttr := util.EmptyIfNilString(li.MustAttribute("class"))

	var href, title, src, publisher string
	if strings.Contains(classAttr, "item_main") {
		href = util.AnchorHREF(li)
		title = li.El("strong[class=tit_g]").MustText()
		src = util.ImgSrc(li)
	} else {
		a := li.El("a")
		href = util.EmptyIfNilString(a.MustAttribute("href"))
		title = a.MustText()
		publisher = li.El("span[class=info_news]").MustText()
	}

	news := types.News{
		URL:       href,
		Image:     src,
		Title:     title,
		Publisher: publisher,
	}
	return news
}

func extractPopNewses(dd types.DumpDirectory, et *rodtemplate.ElementTemplate, popSelector string, pageNum int, order int) []types.News {
	myNewsList := make([]types.News, 0)

	for idx, li := range et.El(popSelector).Els("li") {
		a := li.El("a")
		href := util.EmptyIfNilString(a.MustAttribute("href"))
		title := a.MustText()

		publisher := ""
		if li.Has("span[class=info_news]") {
			publisher = li.El("span[class=info_news]").MustText()
		}

		news := types.News{
			URL:       href,
			Title:     title,
			Publisher: publisher,
		}
		news.SetContextData(pageNum, order, idx, dd, true)

		myNewsList = append(myNewsList, news)
	}
	return myNewsList
}
