package pc

import (
	"fmt"
	"log"
	"strings"

	"github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

func (_ *pc) GetNewsEnd(p *rodtemplate.PageTemplate, n *types.News) error {
	var contentBlock *rodtemplate.ElementTemplate

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

	selfChain := rodtemplate.NewInspectChain(mainBlock).ForOne("div[id=mArticle]", true, true, func(mArticleBlock *rodtemplate.ElementTemplate) error {
		return nil
	}).SelfChain()

	var endProcessed bool

	selfChain.ForOne("div[data-cloud-area=article]", false, true, func(articleBlock *rodtemplate.ElementTemplate) error {
		if articleBlock == nil {
			return nil
		}

		endProcessed = true

		n.End = &types.End{}
		n.End.Category = p.SelectOrPanic("h2[id=kakaoBody]").MustText()

		headBlock := contentBlock.SelectOrPanic("div[class=head_view]")

		if headBlock.Has("em[class=info_cp] > a[class=link_cp]") {
			n.End.Provider = util.ImgALT(headBlock.El("em[class=info_cp] > a[class=link_cp]"))
		} else if headBlock.Has("a.link_issue > strong.tit_thumb") {
			n.End.Provider = headBlock.El("a.link_issue > strong.tit_thumb").MustText()
		}

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
	}).ForOne("div[id=videoWrap]", false, true, func(videoBlock *rodtemplate.ElementTemplate) error {
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
	}).ForOne("div[class=photo_view]", false, true, func(photoBlock *rodtemplate.ElementTemplate) error {
		if photoBlock == nil {
			return nil
		}

		endProcessed = true

		log.Println("skip collect end of photo view")

		return nil
	}).ForOne("div[class=view_vod]", false, true, func(vodBlock *rodtemplate.ElementTemplate) error {
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

	rodtemplate.NewInspectChain(contentBlock).ForOne("div[id=cFeature]", false, true, func(featureBlock *rodtemplate.ElementTemplate) error {
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
