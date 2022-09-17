package common

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	rt "github.com/darimuri/go-lib/rodtemplate"

	"github.com/darimuri/coll-news/pkg/types"
	"github.com/darimuri/coll-news/pkg/util"
)

func MustHeaderMeta(p *rt.PageTemplate, propertyName string) *string {
	metaProviderSelector := fmt.Sprintf(`head > meta[property="%s"]`, propertyName)
	if p.Has(metaProviderSelector) {
		return p.El(metaProviderSelector).MustAttribute("content")
	}

	return nil
}

func MustNumComment(block *rt.ElementTemplate) *uint64 {
	var numComment *uint64

	counterSelector := "button.btn_cmt"
	if true == block.Has(counterSelector) {
		counterBlock := block.El(counterSelector)
		textVal := counterBlock.El("span.num_cmt").MustTextAsUInt64()
		numComment = &textVal
	}

	return numComment
}

func ParseEmotions(articleBlock *rt.ElementTemplate, n *types.News) error {
	emotionBoxSelector := "div.emotion_wrap > div.emotion_list > div#alex_action_emotion > div > div"
	if true == articleBlock.Has(emotionBoxSelector) {
		n.End.Emotions = make([]types.Emotion, 0)

		emotionBox := articleBlock.El(emotionBoxSelector)
		for _, e := range emotionBox.Els("button") {
			emotionName := util.EmptyIfNilString(e.MustAttribute("data-tiara-action-name"))
			emotionCount := strings.TrimSpace(e.El("span.🎬_count_label").MustText())

			emotionName = strings.Replace(emotionName, "액션_", "", 1)

			if emotionCount == "" {
				log.Println("skip emotion collection of", emotionName, "for empty emotionCount string in", emotionBox.MustHTML())
			}

			if count, err := strconv.ParseInt(emotionCount, 10, 64); err != nil {
				n.End.Emotions = append(n.End.Emotions, types.Emotion{Name: emotionName, CountString: emotionCount})
			} else {
				n.End.Emotions = append(n.End.Emotions, types.Emotion{Name: emotionName, Count: count})
			}
		}
	}
	return nil
}
