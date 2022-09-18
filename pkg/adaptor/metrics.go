package adaptor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var totalNewsCollected = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "news_total",
		Help: "Number of news collected.",
	},
	[]string{"type", "source", "location"},
)

var newsEndStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "news_end_status",
		Help: "Status of news end call.",
	},
	[]string{"type", "source", "status"},
)

var newsEndDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "news_end_duration",
	Help: "Duration of news end call.",
}, []string{"type", "source"})

var newsListDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "news_list_duration",
	Help: "Duration of news list call.",
}, []string{"type", "source", "location"})

func init() {
	prometheus.Register(totalNewsCollected)
	prometheus.Register(newsEndStatus)
	prometheus.Register(newsEndDuration)
	prometheus.Register(newsListDuration)
}
