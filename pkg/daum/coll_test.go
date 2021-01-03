package daum

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/darimuri/coll-news/pkg/test"
)

var _ = Describe("Purchase", func() {
	var browser *rod.Browser
	var cut *Portal

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
	})

	AfterEach(func() {
		browser.Close()
	})

	Context("portal", func() {

		BeforeEach(func() {
			var err error
			cut, err = NewPortal(browser, "../../test")
			Expect(err).Should(BeNil())
		})

		It("top news", func() {
			cut.Top()

			newsList, err := cut.GetTopNews()
			Expect(err).Should(BeNil())
			Expect(newsList).ShouldNot(BeNil())
			Expect(newsList).ShouldNot(BeEmpty())

			for _, n := range newsList {
				Expect(n.Title).ShouldNot(BeEmpty())
				Expect(n.URL).ShouldNot(BeEmpty())
				Expect(n.NewsPage).Should(BeNumerically(">", 0))
				Expect(n.Order).Should(BeNumerically(">=", 0))
				Expect(n.FullHTML).ShouldNot(BeEmpty())
				Expect(n.FullScreenShot).ShouldNot(BeEmpty())
				Expect(n.TabScreenShot).ShouldNot(BeEmpty())
			}
		})

		FIt("news home news", func() {
			cut.NewsHome()

			newsList, err := cut.GetNewsHomeNews()
			Expect(err).Should(BeNil())
			Expect(newsList).ShouldNot(BeNil())
			Expect(newsList).ShouldNot(BeEmpty())

			for _, n := range newsList {
				Expect(n.Title).ShouldNot(BeEmpty())
				Expect(n.URL).ShouldNot(BeEmpty())
				Expect(n.NewsPage).Should(BeNumerically(">", 0))
				Expect(n.Order).Should(BeNumerically(">=", 0))
				Expect(n.FullHTML).ShouldNot(BeEmpty())
				Expect(n.FullScreenShot).ShouldNot(BeEmpty())
				Expect(n.TabScreenShot).ShouldNot(BeEmpty())
			}
		})
	})
})
