package daum

import (
	"testing"

	"github.com/darimuri/coll-news/pkg/cache"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Daum Test Suite")
}

var endCache cache.Cache

var _ = BeforeSuite(func() {
	endCache = cache.NewLargeCache()
})
