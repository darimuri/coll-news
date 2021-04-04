package util

import (
	"github.com/dormael/go-lib/rodtemplate"
)

func EmptyIfNilString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func AnchorHREF(item *rodtemplate.ElementTemplate) string {
	return EmptyIfNilString(item.ElementAttribute("a", "href"))
}

func ImgSrc(item *rodtemplate.ElementTemplate) string {
	return EmptyIfNilString(item.ElementAttribute("img", "src"))
}

func ImgALT(item *rodtemplate.ElementTemplate) string {
	return EmptyIfNilString(item.ElementAttribute("img", "alt"))
}
