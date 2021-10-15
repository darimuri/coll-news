package naver

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/darimuri/coll-news/pkg/naver/pc"
	"github.com/darimuri/coll-news/pkg/test"
	"github.com/darimuri/coll-news/pkg/types"
)

var _ = Describe("naver news pc", func() {
	var browser *rod.Browser
	var cut types.Collector

	BeforeEach(func() {
		url, err := launcher.New().
			Headless(test.LaunchHeadless).
			Devtools(false).
			Set("start-maximized").
			Launch()

		Expect(err).Should(BeNil())

		browser = rod.New()
		err = browser.
			ControlURL(url).
			//Trace(true).
			//Slowmotion(2 * time.Second).
			Connect()
		Expect(err).Should(BeNil())

		cut, err = NewPortal(browser, types.PC(), pc.New(), "../../test/naver/pc", endCache)
		Expect(err).Should(BeNil())
	})

	AfterEach(func() {
		cut.Cleanup()
		_ = browser.Close()
	})

	Context("collect", func() {
		It("top news", func() {
			cut.Top()

			newsList, err := cut.GetTopNewsList()
			Expect(err).Should(BeNil())
			Expect(newsList).ShouldNot(BeNil())
			Expect(newsList).ShouldNot(BeEmpty())

			for idx := range newsList {
				n := newsList[idx]
				Expect(n.Title).ShouldNot(BeEmpty())
				Expect(n.URL).ShouldNot(BeEmpty())
				Expect(n.NewsPage).Should(BeNumerically(">", 0))
				Expect(n.Order).Should(BeNumerically(">=", 0))
				Expect(n.FullHTML).ShouldNot(BeEmpty())
				Expect(n.FullScreenShot).ShouldNot(BeEmpty())
				Expect(n.TabScreenShot).ShouldNot(BeEmpty())

				err = cut.GetNewsEnd(&n)
				Expect(err).Should(BeNil(), "err getting top news end %v", n)
			}
		})

		It("news home news", func() {
			cut.NewsHome()

			newsList, err := cut.GetNewsHomeNewsList()
			Expect(err).Should(BeNil())
			Expect(newsList).ShouldNot(BeNil())
			Expect(newsList).ShouldNot(BeEmpty())

			for idx := range newsList {
				n := newsList[idx]
				Expect(n.Title).ShouldNot(BeEmpty())
				Expect(n.URL).ShouldNot(BeEmpty())
				Expect(n.NewsPage).Should(BeNumerically(">", 0))
				Expect(n.Order).Should(BeNumerically(">=", 0))
				Expect(n.FullHTML).ShouldNot(BeEmpty())
				Expect(n.FullScreenShot).ShouldNot(BeEmpty())
				Expect(n.TabScreenShot).ShouldNot(BeEmpty())

				err = cut.GetNewsEnd(&n)
				Expect(err).Should(BeNil(), "err getting new home news end %v", n)
			}
		})
	})
})
