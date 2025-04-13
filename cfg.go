package main

import (
	"fmt"
	"html"
	"html/template"
)

// 导入 fmt 包用于可能的错误处理或日志记录

// Cfg 结构体定义，对应 Python 的 dataclass
type Cfg struct {
	URL       string
	URLMobile *string // 使用指针类型以支持 nil (对应 Python 的 None)
	Anchor    string
	Prefix    string
	Suffix    string
	ID        int
	Temp      *template.Template
	Hash      int
	Merge     bool
}

// stringPtr 是一个辅助函数，用于创建字符串指针
func stringPtr(s string) *string {
	return &s
}

func tmpl(s string) *template.Template {
	t := template.New("")
	_, err := t.Parse(s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse %q as template, error: %s", s, err))
	}

	t.Funcs(template.FuncMap{
		"html": html.EscapeString,
	})

	return t
}

type TmplData struct {
	FromNickname string
	Title        string
}

// getNotifyConfig 根据通知 ID 获取对应的配置
// 返回 Cfg 和一个布尔值，表示是否找到了对应的配置
var notifyConfigs = map[int]Cfg{
	1: {
		URL:       "https://bgm.tv/group/topic",
		URLMobile: stringPtr("MOBILE_URL/topic/group/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在你的小组话题 <b>{{.Title | html}}</b> 中发表了新回复"),
		Prefix:    "在你的小组话题",
		Suffix:    "中发表了新回复",
		ID:        1,
		Hash:      1,
		Merge:     true,
	},
	2: {
		URL:       "https://bgm.tv/group/topic",
		URLMobile: stringPtr("MOBILE_URL/topic/group/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在小组话题 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在小组话题",
		Suffix:    "中回复了你",
		ID:        2,
		Hash:      1,
		Merge:     true,
	},
	3: {
		URL:       "https://bgm.tv/subject/topic",
		URLMobile: stringPtr("/topic/subject"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在你的条目讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		Prefix:    "在你的条目讨论",
		Suffix:    "中发表了新回复",
		ID:        3,
		Hash:      3,
		Merge:     true,
	},
	4: {
		URL:       "https://bgm.tv/subject/topic/",
		URLMobile: stringPtr("MOBILE_URL/topic/subject/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在条目讨论 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在条目讨论",
		Suffix:    "中回复了你",
		ID:        4,
		Hash:      3,
		Merge:     true,
	},
	5: {
		URL:       "https://bgm.tv/character/",
		URLMobile: stringPtr("MOBILE_URL/topic/crt/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在角色讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		Prefix:    "在角色讨论",
		Suffix:    "中发表了新回复",
		ID:        5,
		Hash:      5,
		Merge:     true,
	},
	6: {
		URL:       "https://bgm.tv/character/",
		URLMobile: stringPtr("MOBILE_URL/topic/crt/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在角色 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在角色",
		Suffix:    "中回复了你",
		ID:        6,
		Hash:      5,
		Merge:     true,
	},
	7: {
		URL:       "/blog/",
		URLMobile: nil, // 对应 Python 的 None
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在你的日志 <b>{{.Title | html}}</b> 中发表了新回复"),
		Prefix:    "在你的日志",
		Suffix:    "中发表了新回复",
		ID:        7,
		Hash:      7,
		Merge:     true,
	},
	8: {
		URL:       "https://bgm.tv/blog/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在日志 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在日志",
		Suffix:    "中回复了你",
		ID:        8,
		Hash:      7,
		Merge:     true,
	},
	9: {
		URL:       "https://bgm.tv/subject/ep/",
		URLMobile: stringPtr("MOBILE_URL/topic/ep/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在章节讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		Prefix:    "在章节讨论",
		Suffix:    "中发表了新回复",
		ID:        9,
		Hash:      9,
		Merge:     true,
	},
	10: {
		URL:       "https://bgm.tv/subject/ep/",
		URLMobile: stringPtr("MOBILE_URL/topic/ep/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在章节讨论 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在章节讨论",
		Suffix:    "中回复了你",
		ID:        10,
		Hash:      9,
		Merge:     true,
	},
	11: {
		URL:       "https://bgm.tv/index/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在目录 <b>{{.Title | html}}</b> 中给你留言了"),
		Prefix:    "在目录",
		Suffix:    "中给你留言了",
		ID:        11,
		Hash:      11,
		Merge:     true,
	},
	12: {
		URL:       "https://bgm.tv/index/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在目录 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在目录",
		Suffix:    "中回复了你",
		ID:        12,
		Hash:      11,
		Merge:     true,
	},
	13: {
		URL:       "https://bgm.tv/person/",
		URLMobile: stringPtr("MOBILE_URL/topic/prsn/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在人物 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在人物",
		Suffix:    "中回复了你",
		ID:        13,
		Hash:      13,
		Merge:     true,
	},
	14: {
		URL:       "https://bgm.tv/user/",
		URLMobile: nil,
		Anchor:    "#",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 请求与你成为好友"),
		Prefix:    "请求与你成为好友",
		Suffix:    "",
		ID:        14,
		Hash:      14,
		Merge:     false,
	},
	15: {
		URL:       "https://bgm.tv/user/",
		URLMobile: nil,
		Anchor:    "#",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 通过了你的好友请求"),
		Prefix:    "通过了你的好友请求",
		Suffix:    "",
		ID:        15,
		Hash:      14,
		Merge:     false,
	},
	17: {
		URL:       "DOUJIN_URL/club/topic/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在你的社团讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		Prefix:    "在你的社团讨论",
		Suffix:    "中发表了新回复",
		ID:        17,
		Hash:      17,
		Merge:     true,
	},
	18: {
		URL:       "DOUJIN_URL/club/topic/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在社团讨论 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在社团讨论",
		Suffix:    "中回复了你",
		ID:        18,
		Hash:      17,
		Merge:     true,
	},
	19: {
		URL:       "DOUJIN_URL/subject/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在同人作品 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在同人作品",
		Suffix:    "中回复了你",
		ID:        19,
		Hash:      19,
		Merge:     true,
	},
	20: {
		URL:       "DOUJIN_URL/event/topic/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在你的展会讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		Prefix:    "在你的展会讨论",
		Suffix:    "中发表了新回复",
		ID:        20,
		Hash:      20,
		Merge:     true,
	},
	21: {
		URL:       "DOUJIN_URL/event/topic/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在展会讨论 <b>{{.Title | html}}</b> 中回复了你"),
		Prefix:    "在展会讨论",
		Suffix:    "中回复了你",
		ID:        21,
		Hash:      20,
		Merge:     true,
	},
	22: {
		URL:       "https://bgm.tv/user/chobits_user/timeline/status/",
		URLMobile: nil,
		Anchor:    "#post_",
		// Note: The original Python code used string formatting here, which isn't directly possible in Go templates
		// in the same way during initialization. Assuming a simpler template for now.
		// If dynamic URL generation is needed, it should be handled when executing the template.
		Temp:   tmpl(`<code>{{.FromNickname | html}}</code> 回复了你的吐槽`),
		Prefix: `回复了你的吐槽`, // Simplified, original had HTML link
		Suffix: "",
		ID:     22,
		Hash:   22,
		Merge:  true,
	},
	23: {
		URL:       "https://bgm.tv/group/topic/",
		URLMobile: stringPtr("MOBILE_URL/topic/group/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在小组话题 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在小组话题",
		Suffix:    "中提到了你",
		ID:        23,
		Hash:      1,
		Merge:     true,
	},
	24: {
		URL:       "https://bgm.tv/subject/topic/",
		URLMobile: stringPtr("MOBILE_URL/topic/subject/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在条目讨论 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在条目讨论",
		Suffix:    "中提到了你",
		ID:        24,
		Hash:      3,
		Merge:     true,
	},
	25: {
		URL:       "https://bgm.tv/character/",
		URLMobile: stringPtr("MOBILE_URL/topic/crt/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在角色 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在角色",
		Suffix:    "中提到了你",
		ID:        25,
		Hash:      5,
		Merge:     true,
	},
	26: {
		URL:       "https://bgm.tv/person/",
		URLMobile: stringPtr("MOBILE_URL/topic/prsn/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在人物讨论 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在人物讨论",
		Suffix:    "中提到了你",
		ID:        26,
		Hash:      5, // Note: Original Python had hash 5, Go code had 13. Using 5 from Python.
		Merge:     true,
	},
	27: {
		URL:       "https://bgm.tv/index/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在目录 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在目录",
		Suffix:    "中提到了你",
		ID:        27, // Note: Original Python had ID 27, Go code had 28. Using 27 from Python.
		Hash:      11,
		Merge:     true,
	},
	28: {
		URL:       "https://bgm.tv/user/chobits_user/timeline/status/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在吐槽中提到了你"), // Simplified title
		Prefix:    "在",
		Suffix:    "中提到了你", // Assuming "吐槽" is the title context here
		ID:        28,
		Hash:      22,
		Merge:     true,
	},
	29: {
		URL:       "https://bgm.tv/blog/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在日志 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在日志",
		Suffix:    "中提到了你",
		ID:        29,
		Hash:      7,
		Merge:     true,
	},
	30: {
		URL:       "https://bgm.tv/subject/ep/",
		URLMobile: stringPtr("MOBILE_URL/topic/ep/"),
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在章节讨论 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在章节讨论",
		Suffix:    "中提到了你",
		ID:        30,
		Hash:      9,
		Merge:     true,
	},
	31: {
		URL:       "DOUJIN_URL/club/",
		URLMobile: nil,
		Anchor:    "/shoutbox#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在社团 <b>{{.Title | html}}</b> 的留言板中提到了你"),
		Prefix:    "在社团",
		Suffix:    "的留言板中提到了你",
		ID:        31,
		Hash:      31,
		Merge:     true,
	},
	32: {
		URL:       "DOUJIN_URL/club/topic/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在社团讨论 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在社团讨论",
		Suffix:    "中提到了你",
		ID:        32,
		Hash:      17,
		Merge:     true,
	},
	33: {
		URL:       "DOUJIN_URL/subject/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在同人作品 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在同人作品",
		Suffix:    "中提到了你",
		ID:        33,
		Hash:      19,
		Merge:     true,
	},
	34: {
		URL:       "DOUJIN_URL/event/topic/",
		URLMobile: nil,
		Anchor:    "#post_",
		Temp:      tmpl("<code>{{.FromNickname | html}}</code> 在展会讨论 <b>{{.Title | html}}</b> 中提到了你"),
		Prefix:    "在展会讨论",
		Suffix:    "中提到了你",
		ID:        34,
		Hash:      20,
		Merge:     true,
	},
}
