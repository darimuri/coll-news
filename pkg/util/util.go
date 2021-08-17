package util

import (
	"strings"

	"github.com/darimuri/go-lib/rodtemplate"
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

func ImgAltTryFromHTML(item *rodtemplate.ElementTemplate) string {
	return GetElementAttributeFromHTML(item, "img", "alt")
}

func GetElementAttributeFromHTML(item *rodtemplate.ElementTemplate, selector string, attribute string) string {
	el := item.El(selector)
	html := el.MustHTML()

	if strings.Contains(html, attribute) {
		split := strings.Split(html, " ")
		for _, s := range split {
			if false == strings.Contains(s, "=") {
				continue
			}
			kv := strings.Split(s, "=")
			if len(kv) != 2 {
				continue
			}
			key := kv[0]
			if strings.TrimSpace(key) == attribute {
				val := kv[1]
				if strings.HasSuffix(val, ">") {
					val = val[0 : len(val)-1]
				}
				valLen := len(val)
				if valLen < 3 {
					return val
				}
				return val[1 : valLen-1]
			}
		}
	}

	return ""
}
