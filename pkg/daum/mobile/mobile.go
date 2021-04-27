package mobile

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dormael/go-lib/rodtemplate"
	rt "github.com/dormael/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

const (
	topNewsTabSelector = "div[id=channel_news1_top]"
)

var _ types.TypedCollector = (*mobile)(nil)

type mobile struct {
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

	mainBlockSelector := "main[id=kakaoContent]"
	if false == p.Has(mainBlockSelector) {
		return newsList, nil
	}

	pageNum := 1
	mainBlock := p.SelectOrPanic(mainBlockSelector)
	sectionMainBlock := mainBlock.SelectOrPanic("div.section_main")
	sectionSubBlock := mainBlock.SelectOrPanic("div.section_sub")

	homeIssueSelector := "div.box_homeissue"
	if true == sectionMainBlock.Has(homeIssueSelector) {
		homeIssueBlock := sectionMainBlock.SelectOrPanic(homeIssueSelector)
		for idx, b := range homeIssueBlock.Els("ul.list_homeissue > li") {
			contBlock := b.El("div.cont_thumb")
			contMain := contBlock.El("span.inner_link")
			n1 := types.News{
				NewsPage:       pageNum,
				Order:          idx,
				SubOrder:       0,
				Publisher:      contMain.El("span.txt_cp").MustText(),
				Title:          contMain.El("strong.tit_thumb").MustText(),
				URL:            util.AnchorHREF(contBlock),
				Image:          util.ImgSrc(b.El("div.wrap_thumb")),
				FullHTML:       dd.FullHTML(),
				FullScreenShot: dd.FullScreenShot(),
			}
			newsList = append(newsList, n1)

			subBlock := b.El("div.cont_sub")
			contSub := subBlock.El("span.inner_link")
			n2 := types.News{
				NewsPage:       pageNum,
				Order:          idx,
				SubOrder:       1,
				Publisher:      contSub.El("span.txt_cp").MustText(),
				Title:          contSub.El("span.tit_sub").MustText(),
				URL:            util.AnchorHREF(subBlock),
				FullHTML:       dd.FullHTML(),
				FullScreenShot: dd.FullScreenShot(),
			}
			newsList = append(newsList, n2)
		}
	}

	pageNum++
	mainNewsSelector := "div[data-tiara-layer=MAIN_NEWS]"
	if true == sectionMainBlock.Has(mainNewsSelector) {
		mainNewsBlock := mainBlock.SelectOrPanic(mainNewsSelector)

		p.ScrollTo(mainNewsBlock)
		p.WaitRepaint()

		order := newsList[len(newsList)-1].Order + 1

		newsList = append(newsList, parseNewsList(mainNewsBlock, pageNum, order, dd)...)
	}

	subSectionSelectors := []string{
		"div[data-tiara-layer=POPULAR]",
		"div[data-tiara-layer=DRI]",
		"div.box_cmtrank",
	}

	for _, selector := range subSectionSelectors {
		pageNum++
		if true == sectionSubBlock.Has(selector) {
			popularBlock := sectionSubBlock.SelectOrPanic(selector)

			p.ScrollTo(popularBlock)
			p.WaitRepaint()

			order := newsList[len(newsList)-1].Order + 1

			items := parseNewsList(popularBlock, pageNum, order, dd)
			newsList = append(newsList, items...)
		}
	}

	if sectionSubBlock.Has("div.box_agenews > ul > li") {
		//this list is not order properly(should be appeared one step earlier)
		for _, t := range sectionSubBlock.Els("div.box_agenews > ul > li") {
			t.MustClick()
			p.WaitRepaint()

			order := newsList[len(newsList)-1].Order + 1

			for idx, b := range sectionSubBlock.Els("div.tab_slide > div.slide > div.panel > ul.list_news > div.slide > div.panel > li") {
				n := types.News{
					NewsPage:       pageNum,
					Order:          order,
					SubOrder:       idx,
					Title:          b.MustText(),
					URL:            util.AnchorHREF(b),
					FullHTML:       dd.FullHTML(),
					FullScreenShot: dd.FullScreenShot(),
				}
				newsList = append(newsList, n)
			}
		}
	}

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
			parser = parseThemeItem
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

func (_ mobile) GetNewsEnd(p *rodtemplate.PageTemplate, n *types.News) error {
	var contentBlock *rt.ElementTemplate

	daumDivSelector := "div[id=daumContent]"
	kakaoSelector := "main[id=kakaoContent]"
	daumSelector := "main[id=daumContent]"
	if p.Has(daumDivSelector) {
		contentBlock = p.SelectOrPanic(daumDivSelector)
	} else if p.Has(kakaoSelector) {
		contentBlock = p.SelectOrPanic(kakaoSelector)
	} else if p.Has(daumSelector) {
		contentBlock = p.SelectOrPanic(daumSelector)
	} else {
		return fmt.Errorf("failed to find content block from url %s", n.URL)
	}

	articleBlockSelector := "article[id=mArticle]"
	if false == contentBlock.Has(articleBlockSelector) {
		log.Printf("article block %s is missing in %s\n", articleBlockSelector, n.URL)
		return nil
	}
	mArticleBlock := contentBlock.SelectOrPanic(articleBlockSelector)

	articleSelector := "div[data-cloud-area=article]"
	videoSelector := "div[id=videoWrap]"

	if true == mArticleBlock.Has(articleSelector) {
		n.End = &types.End{}
		n.End.Category = contentBlock.SelectOrPanic("h2[class=screen_out]").MustText()

		articleBlock := mArticleBlock.SelectOrPanic(articleSelector)
		headBlock := contentBlock.SelectOrPanic("div[class=head_view]")

		cpBlock := headBlock.El("em[class=info_cp] > a[class=link_cp] > picture")

		n.End.Provider = util.ImgALT(cpBlock)
		if n.End.Provider == "" {
			n.End.Provider = util.ImgAltTryFromHTML(cpBlock)
		}

		n.End.Title = headBlock.SelectOrPanic("h3[class=tit_view]").MustText()

		infoBlock := headBlock.SelectOrPanic("div[class=info_view]")

		spanS := infoBlock.Els("span[class=txt_info]")
		for idx := range spanS {
			spText := spanS[idx].MustText()
			switch idx {
			case 0:
				n.End.PostedAt = strings.TrimSpace(strings.ReplaceAll(spText, "입력", ""))
			case 1:
				n.End.ModifiedAt = strings.TrimSpace(strings.ReplaceAll(spText, "수정", ""))
			}
		}

		authorSelector := "span[class=txt_author]"
		if infoBlock.Has(authorSelector) {
			n.End.Author = strings.TrimSpace(infoBlock.El(authorSelector).MustText())
		} else {
			n.End.Author = "NotFound"
		}

		counterSelector := "button[id=alexCounter]"
		if true == headBlock.Has(counterSelector) {
			counterBlock := headBlock.El(counterSelector)
			n.End.NumComment = counterBlock.El("span.alex-count-area").MustTextAsUInt64()
		}

		bodySelector := "div[data-cloud=article_body]"
		bodyBlock := articleBlock.El(bodySelector)
		n.End.Text = bodyBlock.El("div[class=article_view]").MustText()

		if true == bodyBlock.Has("figure") {
			figureBlock := bodyBlock.El("figure")
			figureText := figureBlock.MustText()

			if true == strings.HasPrefix(n.End.Text, figureText) {
				n.End.Text = strings.Replace(n.End.Text, figureText, "", 1)
			}
		}

		n.End.HTML = p.El("html").MustHTML()

		n.End.Images = make([]string, 0)
		for _, img := range articleBlock.Els("img[class=thumb_g_article]") {
			n.End.Images = append(n.End.Images, util.EmptyIfNilString(img.MustAttribute("src")))
		}

		err := parseEmotions(articleBlock, n)
		if err != nil {
			return err
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
		log.Println("skip to collect news end for view_vod", n.URL)
	} else if true == contentBlock.Has("div[class=cont_vod]") {
		log.Println("skip to collect news end for cont_vod", n.URL)
	} else if true == contentBlock.Has("div[data-tiara-layer=c_viewcontents]") {
		log.Println("skip to collect news end for c_viewcontents", n.URL)
	} else {
		return fmt.Errorf("failed to collect new end for %s", n.URL)
	}

	return nil
}

func parseNewsList(listBlock *rt.ElementTemplate, pageNum int, order int, dd types.DumpDirectory) []types.News {
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
			NewsPage:       pageNum,
			Order:          order,
			SubOrder:       idx - numBanners,
			Publisher:      myPublisher,
			Title:          myTitle,
			URL:            util.AnchorHREF(b),
			FullHTML:       dd.FullHTML(),
			FullScreenShot: dd.FullScreenShot(),
		}

		if true == b.Has("div.wrap_thumb") {
			n.Image = util.ImgSrc(b.El("div.wrap_thumb"))
		}

		items = append(items, n)
	}
	return items
}

func parseEmotions(articleBlock *rt.ElementTemplate, n *types.News) error {
	emotionBoxSelector := "div.emotion_wrap > div.emotion_list > div.alex-action > div > div.list-wrapper"
	if true == articleBlock.Has(emotionBoxSelector) {
		n.End.Emotions = make([]types.Emotion, 0)

		emotionBox := articleBlock.El(emotionBoxSelector)
		for _, e := range emotionBox.Els("div.selectionbox") {
			emotionName := util.EmptyIfNilString(e.MustAttribute("data-tiara-action-name"))
			emotionCount := strings.TrimSpace(e.El("span.count").MustText())

			emotionName = strings.Replace(emotionName, "액션_", "", 1)

			if emotionCount == "" {
				log.Println("skip emotion collection of", emotionName, "for empty emotionCount string in", emotionBox.MustHTML())
			}

			if count, err := strconv.ParseInt(emotionCount, 10, 64); err != nil {
				return err
			} else {
				n.End.Emotions = append(n.End.Emotions, types.Emotion{Name: emotionName, Count: count})
			}
		}
	}
	return nil
}

func parseThemeItem(item *rodtemplate.ElementTemplate, idx int) types.News {
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

func New() *mobile {
	return &mobile{}
}
