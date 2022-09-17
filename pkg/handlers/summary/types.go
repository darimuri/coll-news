package summary

type Result struct {
	Platforms []Platform
}

type Platform struct {
	Platform string
	Types    []Type
}

type Type struct {
	Type   string
	Error  string
	Yearly []Yearly
}

type Yearly struct {
	Year  string
	Daily []Daily
}

type Daily struct {
}

type Sum struct {
	//daum/[pc|mobile]/dump/2022/yyyyMMdd/yyyyMMdd-hhmmss.json.gz
}
