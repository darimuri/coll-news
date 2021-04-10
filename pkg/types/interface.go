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
	GetTopNews(pageTemplate *rodtemplate.PageTemplate, dd DumpDirectory) ([]News, error)
	GetNewsHomeNews(pageTemplate *rodtemplate.PageTemplate, dd DumpDirectory) ([]News, error)
}
