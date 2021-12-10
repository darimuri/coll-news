package mobile

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/adaptor"
	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

func (_ mobile) GetNewsEnd(p *rodtemplate.PageTemplate, n *types.News) error {
	var contentBlock *rodtemplate.ElementTemplate

	daumDivSelector := "div[id=daumContent]"
	kakaoSelector := "main[id=kakaoContent]"
	daumSelector := "main[id=daumContent]"
	if p.Has(daumDivSelector) {
		contentBlock = p.SelectOrPanic(daumDivSelector)
	} else if p.Has(kakaoSelector) {
		contentBlock = p.SelectOrPanic(kakaoSelector)
	} else if p.Has(daumSelector) {
		contentBlock = p.SelectOrPanic(daumSelector)
	} else if p.Has("main[class=doc-main]") {
		return adaptor.NewTypedError("main[class=doc-main] is not supported content block")
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

		cpBlockSelector := "em[class=info_cp] > a[class=link_cp] > picture"

		if false == headBlock.Has(cpBlockSelector) {
			return adaptor.CPBlockNotFound
		}

		cpBlock := headBlock.El(cpBlockSelector)

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

func parseEmotions(articleBlock *rodtemplate.ElementTemplate, n *types.News) error {
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
				n.End.Emotions = append(n.End.Emotions, types.Emotion{Name: emotionName, CountString: emotionCount})
			} else {
				n.End.Emotions = append(n.End.Emotions, types.Emotion{Name: emotionName, Count: count})
			}
		}
	}
	return nil
}
