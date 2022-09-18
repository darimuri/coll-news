package mobile

import (
	"fmt"
	"log"
	"strings"

	"github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/adaptor"
	"github.com/darimuri/coll-news/pkg/daum/common"
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

		providerFromMeta := common.MustHeaderMeta(p, "og:article:author")
		if providerFromMeta != nil {
			n.End.Provider = *providerFromMeta
		}

		n.End.Category = contentBlock.SelectOrPanic("h2[class=screen_out]").MustText()

		articleBlock := mArticleBlock.SelectOrPanic(articleSelector)
		headBlock := contentBlock.SelectOrPanic("div[class=head_view]")

		if n.End.Provider == "" {
			cpBlockSelector := "em[class=info_cp] > a[class=link_cp] > picture"

			if false == headBlock.Has(cpBlockSelector) {
				return adaptor.CPBlockNotFound
			}

			cpBlock := headBlock.El(cpBlockSelector)

			n.End.Provider = util.ImgALT(cpBlock)
			if n.End.Provider == "" {
				n.End.Provider = util.ImgAltTryFromHTML(cpBlock)
			}
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

		utilBlock := headBlock.SelectOrPanic("div[class=util_wrap]")

		numComment := common.MustNumComment(utilBlock)
		if numComment != nil {
			n.End.NumComment = *numComment
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

		err := common.ParseEmotions(articleBlock, n)
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
		log.Println("skip collect end of photo view", n.URL)
		return adaptor.EndSkippedOnPurpose
	} else if true == mArticleBlock.Has("div.box_g") {
		log.Println("skip collect end of gallery view", n.URL)
		return adaptor.EndSkippedOnPurpose
	} else if true == contentBlock.Has("div[class=view_vod]") {
		log.Println("skip to collect news end for view_vod", n.URL)
		return adaptor.EndSkippedOnPurpose
	} else if true == contentBlock.Has("div[class=cont_vod]") {
		log.Println("skip to collect news end for cont_vod", n.URL)
		return adaptor.EndSkippedOnPurpose
	} else if true == contentBlock.Has("div[data-tiara-layer=c_viewcontents]") {
		log.Println("skip to collect news end for c_viewcontents", n.URL)
		return adaptor.EndSkippedOnPurpose
	} else {
		log.Println("failed to collect new end", n.URL)
		return adaptor.EndSkippedUnexpectedly
	}

	return nil
}
