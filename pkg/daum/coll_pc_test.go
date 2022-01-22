package daum

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/darimuri/coll-news/pkg/adaptor"

	"github.com/darimuri/coll-news/pkg/daum/pc"
	"github.com/darimuri/coll-news/pkg/test"
	"github.com/darimuri/coll-news/pkg/types"
)

var _ = Describe("daum news pc", func() {
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

		cut, err = NewPortal(browser, types.PC(), pc.New(), "../../test/daum/pc", endCache)
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
				_, typedError := err.(adaptor.TypedError)
				if false == typedError {
					Expect(err).Should(BeNil(), "error getting top news end %v", n)
				}
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
				_, typedError := err.(adaptor.TypedError)
				if false == typedError {
					Expect(err).Should(BeNil(), "error getting top news end %v", n)
				}
			}
		})

		It("new end ModifiedAt is correct for author is 고수정", func() {
			cut.Top()
			n := types.News{URL: "https://news.v.daum.net/v/20211016040021618"}
			err := cut.GetNewsEnd(&n)
			_, typedError := err.(adaptor.TypedError)
			if false == typedError {
				Expect(err).Should(BeNil(), "error getting top news end %v", n)
			}

			Expect(n.End.ModifiedAt).Should(BeEmpty())
		})

		It("new end causes no error div[id=kakaoContent] block is missing", func() {
			cut.Top()
			n := types.News{URL: "https://content.v.daum.net/v/kWGY0DyI9E"}
			err := cut.GetNewsEnd(&n)
			_, typedError := err.(adaptor.TypedError)
			if false == typedError {
				Expect(err).Should(BeNil(), "error getting top news end %v", n)
			}
		})
	})
})
