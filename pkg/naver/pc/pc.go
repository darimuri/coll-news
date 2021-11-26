package pc

import (
	"github.com/darimuri/coll-news/pkg/types"
	rt "github.com/darimuri/go-lib/rodtemplate"
)

var _ types.TypedCollector = (*pc)(nil)

type pc struct {
}

func (_ pc) PrepareNewsHomeScreenShot(p *rt.PageTemplate) {
}

func (_ pc) GetNewsHomeNewsList(p *rt.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	newsList := make([]types.News, 0)

	//timelateBlock := p.SelectOrPanic("div.home_timelate")
	// containerBlock := p.SelectOrPanic("div#container")
	// mainBlock := containerBlock.SelectOrPanic("div#main_content")
	// asideBlock := containerBlock.SelectOrPanic("div#main_aside")

	// headlineBlock := mainBlock.SelectOrPanic("div#today_main_news")
	// politicsBlock := mainBlock.SelectOrPanic("div#section_politics")
	// economyBlock := mainBlock.SelectOrPanic("div#section_economy")
	// societyBlock := mainBlock.SelectOrPanic("div#section_society")
	// lifeBlock := mainBlock.SelectOrPanic("div#section_life")
	// worldBlock := mainBlock.SelectOrPanic("div#section_world")
	// itBlock := mainBlock.SelectOrPanic("div#section_it")

	// rankingBlock := asideBlock.SelectOrPanic("div.section")

	// flickBlock := headlineBlock.SelectOrPanic("div.hdline_flick")
	// articleListBlock := headlineBlock.SelectOrPanic("ul.hdline_article_list")

	return newsList, nil
}

func (_ pc) GetTopNewsList(p *rt.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	return []types.News{}, nil
}

func (_ pc) GetNewsEnd(p *rt.PageTemplate, n *types.News) error {
	panic("implement me")
}

func New() *pc {
	return &pc{}
}
