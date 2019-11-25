package types

// 资讯表
type Information struct {
	Id         uint64             `json:"id"`
	Author     string             `json:"author"`  // 作者
	Title      string             `json:"title"`   // 标题
	Lang       InfomationLanguage `json:"lang"`    // 语言类型：1/中文，2/英文
	Context    string             `json:"context"` // 正文
	CreateTime int64              `json:"create_time"`
}

type InfomationLanguage uint8

const (
	InformationChinese InfomationLanguage = iota + 1
	InformationEnglish
)

const (
	Author = "admin"
)
