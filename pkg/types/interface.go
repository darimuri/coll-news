package types

import (
	"github.com/darimuri/go-lib/rodtemplate"
)

type Collector interface {
	Top()
	NewsHome()
	GetTopNewsList() ([]News, error)
	GetNewsHomeNewsList() ([]News, error)
	GetNewsEnd(n *News) error
	Cleanup()
}

type TypedCollector interface {
	PrepareNewsHomeScreenShot(p *rodtemplate.PageTemplate)
	GetNewsHomeNewsList(p *rodtemplate.PageTemplate, dd DumpDirectory) ([]News, error)
	GetTopNewsList(p *rodtemplate.PageTemplate, dd DumpDirectory) ([]News, error)
	GetNewsEnd(p *rodtemplate.PageTemplate, n *News) error
}
