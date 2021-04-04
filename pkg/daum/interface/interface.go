package _interface

import (
	"github.com/dormael/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
)

type Collector interface {
	GetTopNews(pageTemplate *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error)
	GetNewsHomeNews(pageTemplate *rodtemplate.PageTemplate, dd types.DumpDirectory) ([]types.News, error)
}
