package types

import (
	"github.com/dormael/go-lib/rodtemplate"
)

type Collector interface {
	Top()
	NewsHome()
	GetTopNews() ([]News, error)
	GetNewsHomeNews() ([]News, error)
	GetNewsEnd(n *News) error
	Cleanup()
}

type TypedCollector interface {
	GetTopNews(p *rodtemplate.PageTemplate, dd DumpDirectory) ([]News, error)
	GetNewsHomeNews(p *rodtemplate.PageTemplate, dd DumpDirectory) ([]News, error)
	GetNewsEnd(p *rodtemplate.PageTemplate, n *News) error
}
