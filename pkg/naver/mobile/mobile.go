package mobile

import (
	rt "github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
)

var _ types.TypedCollector = (*mobile)(nil)

type mobile struct {
}

func New() *mobile {
	return &mobile{}
}

func (_ mobile) Source() string {
	return types.Naver
}

func (_ mobile) Type() string {
	return types.Mobile
}

func (_ mobile) PrepareNewsHomeScreenShot(p *rt.PageTemplate) {
	panic("implement me")
}

func (_ mobile) GetNewsHomeNewsList(p *rt.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	panic("implement me")
}

func (_ mobile) GetTopNewsList(p *rt.PageTemplate, dd types.DumpDirectory) ([]types.News, error) {
	panic("implement me")
}

func (_ mobile) GetNewsEnd(p *rt.PageTemplate, n *types.News) error {
	panic("implement me")
}
