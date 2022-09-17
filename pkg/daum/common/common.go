package common

import (
	"fmt"

	rt "github.com/darimuri/go-lib/rodtemplate"
)

func MustHeaderMeta(p *rt.PageTemplate, propertyName string) *string {
	metaProviderSelector := fmt.Sprintf(`head > meta[property="%s"]`, propertyName)
	if p.Has(metaProviderSelector) {
		return p.El(metaProviderSelector).MustAttribute("content")
	}

	return nil
}
