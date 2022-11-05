package adaptor

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Adaptor Test Suite")
}

var _ = Describe("xxx", func() {
	It("convertToDataFormat", func() {
		converted := convertToDataFormat("2022. 11. 5. 20:30")
		Expect(converted).Should(Equal("2022-11-05T20:30:00+09:00"))

		converted = convertToDataFormat("2022. 11.11. 20:30")
		Expect(converted).Should(Equal("2022-11-11T20:30:00+09:00"))
	})
})
