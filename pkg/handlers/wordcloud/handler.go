package wordcloud

import "github.com/labstack/echo"

type WordCloud struct {
	collectPath string
}

func (h *WordCloud) Handle(ctx echo.Context) error {
	return nil
}

func NewWordCloud(collectPath string) *WordCloud {
	return &WordCloud{collectPath: collectPath}
}
