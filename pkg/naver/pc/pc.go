package pc

import (
	"github.com/darimuri/coll-news/pkg/types"
	rt "github.com/darimuri/go-lib/rodtemplate"
)

var _ types.TypedCollector = (*pc)(nil)

type pc struct {
}

func (_ pc) PrepareNewsHomeScreenShot(p *rt.PageTemplate) {
	panic("implement me")
}

func (_ pc) GetNewsHomeNewsList(p *rt.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	panic("implement me")
}

func (_ pc) GetTopNewsList(p *rt.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	panic("implement me")
}

func (_ pc) GetNewsEnd(p *rt.PageTemplate, n *types.News) error {
	panic("implement me")
}

func New() *pc {
	return &pc{}
}
