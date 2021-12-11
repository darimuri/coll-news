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

func (_ *pc) PrepareNewsHomeScreenShot(_ *rt.PageTemplate) {
}

func (_ *pc) GetNewsHomeNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)

	newsSubBlock := p.SelectOrPanic("#cSub")
	newsMainBlock := p.SelectOrPanic("#cMain")
	newsArticleBlock := newsMainBlock.SelectOrPanic("#mArticle")

	pageNum := 1
	subListSelector := "ul[class=list_issue]"
	if true == newsSubBlock.Has(subListSelector) {
		subListBlock := newsSubBlock.El(subListSelector)
		p.ScreenShot(subListBlock, dd.TabScreenShot(pageNum), 0)

		for idx, li := range subListBlock.Els("li") {
			imageItem := li.El("div[class=item_issue]")
			src := util.ImgSrc(imageItem)

			contItem := li.El("div[class=cont_thumb]")
			a := contItem.El("strong > a")
			span := contItem.El("span[class=info_thumb]")

			news := types.News{
				URL:            util.EmptyIfNilString(a.MustAttribute("href")),
				Image:          src,
				Title:          a.MustText(),
				NewsPage:       pageNum,
				Order:          idx,
				SubOrder:       0,
				FullHTML:       dd.FullHTML(),
				FullScreenShot: dd.FullScreenShot(),
				TabScreenShot:  dd.TabScreenShot(pageNum),
				Publisher:      span.MustText(),
			}

			newsList = append(newsList, news)

			relateSelector := "div[class=relate_thumb]"

			if li.Has(relateSelector) {
				for jdx, div := range li.El(relateSelector).Els("div[class=thumb_relate]") {
					a := div.El("a")
					rspan := div.El("span[class=info_news]")

					rnews := types.News{
						URL:            util.EmptyIfNilString(a.MustAttribute("href")),
						Title:          a.MustText(),
						NewsPage:       pageNum,
						Order:          idx,
						SubOrder:       jdx + 1,
						FullHTML:       dd.FullHTML(),
						FullScreenShot: dd.FullScreenShot(),
						TabScreenShot:  dd.TabScreenShot(pageNum),
						Publisher:      rspan.MustText(),
					}

					newsList = append(newsList, rnews)
				}
			}
		}
	}

	pageNum++

	yDelta := p.El("#wrapMinidaum").Height()
	yDelta += p.El("#kakaoHead").Height()
	yDelta -= 24

	headlineSelector := "div[class=box_headline]"
	if true == newsArticleBlock.Has(headlineSelector) {
		headlineBlock := newsArticleBlock.El(headlineSelector)
		p.ScreenShot(headlineBlock, dd.TabScreenShot(pageNum), yDelta)

		for idx, ul := range headlineBlock.Els("ul[class=list_headline]") {
			for jdx, li := range ul.Els("li") {
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
					URL:            href,
					Image:          src,
					Title:          title,
					NewsPage:       pageNum,
					Order:          idx,
					SubOrder:       jdx,
					FullHTML:       dd.FullHTML(),
					FullScreenShot: dd.FullScreenShot(),
					TabScreenShot:  dd.TabScreenShot(pageNum),
					Publisher:      publisher,
				}

				newsList = append(newsList, news)
			}
		}
	}

	pageNum++

	yDelta += newsArticleBlock.El("div[class=box_photo]").Height()
	yDelta += 660

	perUseSelector := "div[class=box_peruse] > div[class='pop_news pop_cmt']"
	if true == newsArticleBlock.Has(perUseSelector) {
		perBlock := newsArticleBlock.El(perUseSelector)
		p.ScreenShot(perBlock, dd.TabScreenShot(pageNum), yDelta)

		myNewsList := extractPopNews(dd, perBlock, "ol[class=list_popcmt]", pageNum, 0)
		newsList = append(newsList, myNewsList...)
	}

	pageNum++

	perCmtSelector := "div[class='box_g box_popnews'] > div[class='pop_news pop_cmt']"
	if true == newsArticleBlock.Has(perCmtSelector) {
		perBlock := newsArticleBlock.El(perCmtSelector)
		p.ScreenShot(perBlock, dd.TabScreenShot(pageNum), yDelta)

		myNewsList := extractPopNews(dd, perBlock, "ol[class=list_popcmt]", pageNum, 1)
		newsList = append(newsList, myNewsList...)
	}

	pageNum++

	yDelta += 60

	perAgeSelector := "div[class='pop_news pop_age']"
	if true == newsArticleBlock.Has(perAgeSelector) {
		perBlock := newsArticleBlock.El(perAgeSelector)
		p.ScreenShot(perBlock, dd.TabScreenShot(pageNum), yDelta)

		for idx, genderBlock := range perBlock.Els("div") {
			myNewsList := extractPopNews(dd, genderBlock, "ul", pageNum, 2+idx)
			newsList = append(newsList, myNewsList...)
		}
	}

	return newsList, nil
}

func (_ *pc) GetTopNewsList(p *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
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

		groupNews := mediaBlock.El("div[class=group_news]")
		for idx, item := range groupNews.Els("ul[class=list_thumb] > li") {
			news := types.News{
				URL:            util.AnchorHREF(item),
				Image:          util.ImgSrc(item),
				Title:          item.El("div[class=cont_item] > strong[class=tit_item]").MustText(),
				NewsPage:       pageNum,
				Order:          idx,
				FullHTML:       dd.FullHTML(),
				FullScreenShot: dd.FullScreenShot(),
				TabScreenShot:  dd.TabScreenShot(pageNum),
			}

			newsList = append(newsList, news)
		}

		for idx, item := range groupNews.Els("ul[class=list_txt] > li") {
			a := item.El("a")
			news := types.News{
				URL:            util.EmptyIfNilString(a.MustAttribute("href")),
				Title:          a.MustText(),
				NewsPage:       pageNum,
				Order:          idx,
				FullHTML:       dd.FullHTML(),
				FullScreenShot: dd.FullScreenShot(),
				TabScreenShot:  dd.TabScreenShot(pageNum),
			}

			newsList = append(newsList, news)
		}

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
	mArticleBlock := contentBlock.SelectOrPanic(mainBlockSelector).SelectOrPanic("div[id=mArticle]")

	articleSelector := "div[data-cloud-area=article]"
	videoSelector := "div[id=videoWrap]"

	if true == mArticleBlock.Has(articleSelector) {
		n.End = &types.End{}
		n.End.Category = p.SelectOrPanic("h2[id=kakaoBody]").MustText()

		articleBlock := mArticleBlock.SelectOrPanic(articleSelector)
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
	} else if true == mArticleBlock.Has(videoSelector) {
		n.End = &types.End{}
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
		log.Println("skip to collect news end for", n.URL)
	} else {
		return fmt.Errorf("failed to collect new end for %s", n.URL)
	}

	return nil
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
