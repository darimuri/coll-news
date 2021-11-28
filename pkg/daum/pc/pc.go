package pc

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/darimuri/go-lib/rodtemplate"
	rt "github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

const (
	mediaTabSelector = "div[id=mediaTab]"
	newsTabSelector  = "div[class=page_tabcont]"
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

	rt.NewInspectChain(p.SelectOrPanic("#cSub")).ForOne("ul[class=list_issue]", false, true, func(subListBlock *rt.ElementTemplate) error {
		if subListBlock == nil {
			return nil
		}

		p.ScreenShot(subListBlock, dd.TabScreenShot(pageNum), 0)

		return nil
	}).ForEach("li", true, true, func(idx int, li *rt.ElementTemplate) error {
		news := extractIssue(li)
		news.SetContextData(pageNum, idx, 0, dd)
		newsList = append(newsList, news)

		rt.NewInspectChain(li).ForOne("div[class=relate_thumb]", false, true, func(el *rt.ElementTemplate) error {
			return nil
		}).ForEach("div[class=thumb_relate]", true, true, func(jdx int, div *rt.ElementTemplate) error {
			rnews := extractIssueRelate(div)
			rnews.SetContextData(pageNum, idx, jdx+1, dd)
			newsList = append(newsList, rnews)

			return nil
		})

		return nil
	})

	pageNum++

	yDelta := p.El("#wrapMinidaum").Height()
	yDelta += p.El("#kakaoHead").Height()
	yDelta -= 24

	newsMainBlock := p.SelectOrPanic("#cMain")
	articleChain := rt.NewInspectChain(newsMainBlock).ForOne("#mArticle", true, true, func(newsArticleBlock *rt.ElementTemplate) error {
		pageNum++

		yDelta += newsArticleBlock.El("div[class=box_photo]").Height()
		yDelta += 660

		return nil
	})

	selfChain := articleChain.SelfChain()
	selfChain.ForOne("div[class=box_peruse] > div[class='pop_news pop_cmt']", false, true, func(perBlock *rt.ElementTemplate) error {
		if perBlock == nil {
			return nil
		}

		p.ScreenShot(perBlock, dd.TabScreenShot(pageNum), yDelta)

		myNewsList := extractPopNews(dd, perBlock, "ol[class=list_popcmt]", pageNum, 0)
		newsList = append(newsList, myNewsList...)

		pageNum++

		return nil
	}).ForOne("div[class='box_g box_popnews'] > div[class='pop_news pop_cmt']", false, true, func(perBlock *rt.ElementTemplate) error {
		if perBlock == nil {
			return nil
		}

		p.ScreenShot(perBlock, dd.TabScreenShot(pageNum), yDelta)

		myNewsList := extractPopNews(dd, perBlock, "ol[class=list_popcmt]", pageNum, 1)
		newsList = append(newsList, myNewsList...)

		pageNum++

		yDelta += 60

		return nil
	}).ForOne("div[class='pop_news pop_age']", false, true, func(perBlock *rt.ElementTemplate) error {
		if perBlock == nil {
			return nil
		}

		p.ScreenShot(perBlock, dd.TabScreenShot(pageNum), yDelta)

		for idx, genderBlock := range perBlock.Els("div") {
			myNewsList := extractPopNews(dd, genderBlock, "ul", pageNum, 2+idx)
			newsList = append(newsList, myNewsList...)
		}

		return nil
	})

	articleChain.ForOne("div[class=box_headline]", false, true, func(headlineBlock *rt.ElementTemplate) error {
		if headlineBlock == nil {
			return nil
		}

		p.ScreenShot(headlineBlock, dd.TabScreenShot(pageNum), yDelta)

		return nil
	}).ForEach("ul[class=list_headline]", false, true, func(idx int, ul *rt.ElementTemplate) error {
		rt.NewInspectChain(ul).ForEach("li", true, true, func(jdx int, li *rt.ElementTemplate) error {
			news := extractHeadlineSub(li)
			news.SetContextData(pageNum, idx, jdx, dd)
			newsList = append(newsList, news)

			return nil
		})

		return nil
	})

	return newsList, nil
}

func (*pc) GetTopNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)

	if false == p.Has(mediaTabSelector) {
		return newsList, nil
	}

	newMapByTab := make(map[int]string, 0)

	mediaBlock := p.El(mediaTabSelector)

	newsPagerBlock := mediaBlock.El(newsTabSelector)

	for {
		mediaBlock.MustWaitLoad()
		mediaBlock.MustWaitStable()
		mediaBlock.MustWaitVisible()

		currentNewsPage := newsPagerBlock.El("strong[class=screen_out]").MustText()
		currentMediaPage := mediaBlock.El("strong[id=mediaPageNum]").MustText()

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
			news.SetContextData(pageNum, idx, 0, dd)
			newsList = append(newsList, news)

			return nil
		})

		groupNewsChain.ForEach("ul[class=list_txt] > li", false, true, func(idx int, item *rt.ElementTemplate) error {
			a := item.El("a")
			news := types.News{
				URL:   util.EmptyIfNilString(a.MustAttribute("href")),
				Title: a.MustText(),
			}
			news.SetContextData(pageNum, idx, 0, dd)
			newsList = append(newsList, news)
			return nil
		})

		newMapByTab[pageNum] = currentNewsPage

		if len(newMapByTab) == 3 {
			break
		}

		newsPagerBlock.MustClick()
	}

	return newsList, nil
}

func (_ *pc) GetNewsEnd(p *rt.PageTemplate, n *types.News) error {
	var contentBlock *rt.ElementTemplate

	divDaumContentSelector := "div[id=daumContent]"
	mainDaumContentSelector := "main[id=daumContent]"
	divKakaoContentSelector := "div[id=kakaoContent]"
	if p.Has(divDaumContentSelector) {
		contentBlock = p.SelectOrPanic(divDaumContentSelector)
	} else if p.Has(mainDaumContentSelector) {
		contentBlock = p.SelectOrPanic(mainDaumContentSelector)
	} else if p.Has(divKakaoContentSelector) {
		contentBlock = p.SelectOrPanic(divKakaoContentSelector)
	} else {
		contentBlock = p.SelectOrPanic("main[id=kakaoContent]")
	}

	mainBlockSelector := "div[id=cMain]"
	if false == contentBlock.Has(mainBlockSelector) {
		log.Printf("main block %s is missing in %s\n", mainBlockSelector, n.URL)
		return nil
	}

	mainBlock := contentBlock.SelectOrPanic(mainBlockSelector)

	selfChain := rt.NewInspectChain(mainBlock).ForOne("div[id=mArticle]", true, true, func(mArticleBlock *rt.ElementTemplate) error {
		return nil
	}).SelfChain()

	var endProcessed bool

	selfChain.ForOne("div[data-cloud-area=article]", false, true, func(articleBlock *rt.ElementTemplate) error {
		if articleBlock == nil {
			return nil
		}

		endProcessed = true

		n.End = &types.End{}
		n.End.Category = p.SelectOrPanic("h2[id=kakaoBody]").MustText()

		headBlock := contentBlock.SelectOrPanic("div[class=head_view]")

		n.End.Provider = util.ImgALT(headBlock.El("em[class=info_cp] > a[class=link_cp]"))
		n.End.Title = headBlock.SelectOrPanic("h3[class=tit_view]").MustText()

		infoBlock := headBlock.SelectOrPanic("span[class=info_view]")

		spanS := infoBlock.Els("span[class=txt_info]")
		for idx := range spanS {
			spText := spanS[idx].MustText()
			if strings.Contains(spText, "입력 ") {
				n.End.PostedAt = strings.TrimSpace(strings.ReplaceAll(spText, "입력 ", ""))
			} else if strings.Contains(spText, "수정 ") {
				n.End.ModifiedAt = strings.TrimSpace(strings.ReplaceAll(spText, "수정 ", ""))
			} else {
				n.End.Author = strings.TrimSpace(spText)
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

		return nil
	}).ForOne("div[id=videoWrap]", false, true, func(videoBlock *rt.ElementTemplate) error {
		if videoBlock == nil {
			return nil
		}

		endProcessed = true

		n.End = &types.End{}
		innerBlock := videoBlock.SelectOrPanic("div[class=inner_view]")
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

		return nil
	}).ForOne("div[class=photo_view]", false, true, func(photoBlock *rt.ElementTemplate) error {
		if photoBlock == nil {
			return nil
		}

		endProcessed = true

		log.Println("skip collect end of photo view")

		return nil
	}).ForOne("div[class=view_vod]", false, true, func(vodBlock *rt.ElementTemplate) error {
		if vodBlock == nil {
			return nil
		}

		endProcessed = true

		log.Println("skip to collect news end for", n.URL)

		return nil
	})

	if endProcessed {
		return nil
	}

	rt.NewInspectChain(contentBlock).ForOne("div[id=cFeature]", false, true, func(featureBlock *rt.ElementTemplate) error {
		if featureBlock == nil {
			return nil
		}

		endProcessed = true

		log.Println("skip to collect news end for", n.URL)

		return nil
	})

	if endProcessed {
		return nil
	}

	return fmt.Errorf("failed to collect new end for %s", n.URL)
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

func extractPopNews(dd types.DumpDirectory, et *rodtemplate.ElementTemplate, popSelector string, pageNum int, order int) []types.News {
	myNewsList := make([]types.News, 0)

	for jdx, li := range et.El(popSelector).Els("li") {
		a := li.El("a")
		href := util.EmptyIfNilString(a.MustAttribute("href"))
		title := a.MustText()

		publisher := ""
		if li.Has("span[class=info_news]") {
			publisher = li.El("span[class=info_news]").MustText()
		}

		news := types.News{
			URL:            href,
			Title:          title,
			NewsPage:       pageNum,
			Order:          order,
			SubOrder:       jdx,
			FullHTML:       dd.FullHTML(),
			FullScreenShot: dd.FullScreenShot(),
			TabScreenShot:  dd.TabScreenShot(pageNum),
			Publisher:      publisher,
		}

		myNewsList = append(myNewsList, news)
	}
	return myNewsList
}
