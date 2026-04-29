package sanitize

import "github.com/microcosm-cc/bluemonday"

// NoticeContent 对公告富文本（如 Quill 输出）做白名单净化，移除脚本与危险 URL，降低 XSS 风险。
// 参数 s 可为原始 HTML 或纯文本；返回可安全用于前端 innerHTML 的字符串。
func NoticeContent(s string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(s)
}
