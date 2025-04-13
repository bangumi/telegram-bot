package main

import (
	"fmt"
	"html"
	"text/template"
)

type Cfg struct {
	URL    string
	Anchor string
	ID     int
	Temp   *template.Template
	Hash   int
	Merge  bool
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
		URL:    "https://bgm.tv/group/topic",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在你的小组话题 <b>{{.Title | html}}</b> 中发表了新回复"),
		ID:     1,
		Hash:   1,
		Merge:  true,
	},
	2: {
		URL:    "https://bgm.tv/group/topic",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在小组话题 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     2,
		Hash:   1,
		Merge:  true,
	},
	3: {
		URL:    "https://bgm.tv/subject/topic",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在你的条目讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		ID:     3,
		Hash:   3,
		Merge:  true,
	},
	4: {
		URL:    "https://bgm.tv/subject/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在条目讨论 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     4,
		Hash:   3,
		Merge:  true,
	},
	5: {
		URL:    "https://bgm.tv/character/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在角色讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		ID:     5,
		Hash:   5,
		Merge:  true,
	},
	6: {
		URL:    "https://bgm.tv/character/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在角色 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     6,
		Hash:   5,
		Merge:  true,
	},
	7: {
		URL:    "/blog/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在你的日志 <b>{{.Title | html}}</b> 中发表了新回复"),
		ID:     7,
		Hash:   7,
		Merge:  true,
	},
	8: {
		URL:    "https://bgm.tv/blog/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在日志 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     8,
		Hash:   7,
		Merge:  true,
	},
	9: {
		URL:    "https://bgm.tv/subject/ep/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在章节讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		ID:     9,
		Hash:   9,
		Merge:  true,
	},
	10: {
		URL:    "https://bgm.tv/subject/ep/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在章节讨论 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     10,
		Hash:   9,
		Merge:  true,
	},
	11: {
		URL:    "https://bgm.tv/index/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在目录 <b>{{.Title | html}}</b> 中给你留言了"),
		ID:     11,
		Hash:   11,
		Merge:  true,
	},
	12: {
		URL:    "https://bgm.tv/index/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在目录 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     12,
		Hash:   11,
		Merge:  true,
	},
	13: {
		URL:    "https://bgm.tv/person/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在人物 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     13,
		Hash:   13,
		Merge:  true,
	},
	14: {
		URL:    "https://bgm.tv/user/",
		Anchor: "#",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 请求与你成为好友"),
		ID:     14,
		Hash:   14,
		Merge:  false,
	},
	15: {
		URL:    "https://bgm.tv/user/",
		Anchor: "#",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 通过了你的好友请求"),
		ID:     15,
		Hash:   14,
		Merge:  false,
	},
	17: {
		URL:    "DOUJIN_URL/club/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在你的社团讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		ID:     17,
		Hash:   17,
		Merge:  true,
	},
	18: {
		URL:    "DOUJIN_URL/club/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在社团讨论 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     18,
		Hash:   17,
		Merge:  true,
	},
	19: {
		URL:    "DOUJIN_URL/subject/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在同人作品 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     19,
		Hash:   19,
		Merge:  true,
	},
	20: {
		URL:    "DOUJIN_URL/event/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在你的展会讨论 <b>{{.Title | html}}</b> 中发表了新回复"),
		ID:     20,
		Hash:   20,
		Merge:  true,
	},
	21: {
		URL:    "DOUJIN_URL/event/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在展会讨论 <b>{{.Title | html}}</b> 中回复了你"),
		ID:     21,
		Hash:   20,
		Merge:  true,
	},
	22: {
		URL:    "https://bgm.tv/user/chobits_user/timeline/status/",
		Anchor: "#post_",
		Temp:   tmpl(`<code>{{.FromNickname | html}}</code> 回复了你的吐槽`),
		ID:     22,
		Hash:   22,
		Merge:  true,
	},
	23: {
		URL:    "https://bgm.tv/group/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在小组话题 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     23,
		Hash:   1,
		Merge:  true,
	},
	24: {
		URL:    "https://bgm.tv/subject/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在条目讨论 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     24,
		Hash:   3,
		Merge:  true,
	},
	25: {
		URL:    "https://bgm.tv/character/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在角色 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     25,
		Hash:   5,
		Merge:  true,
	},
	26: {
		URL:    "https://bgm.tv/person/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在人物讨论 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     26,
		Hash:   5,
		Merge:  true,
	},
	27: {
		URL:    "https://bgm.tv/index/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在目录 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     27,
		Hash:   11,
		Merge:  true,
	},
	28: {
		URL:    "https://bgm.tv/user/chobits_user/timeline/status/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在吐槽中提到了你"),
		ID:     28,
		Hash:   22,
		Merge:  true,
	},
	29: {
		URL:    "https://bgm.tv/blog/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在日志 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     29,
		Hash:   7,
		Merge:  true,
	},
	30: {
		URL:    "https://bgm.tv/subject/ep/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在章节讨论 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     30,
		Hash:   9,
		Merge:  true,
	},
	31: {
		URL:    "DOUJIN_URL/club/",
		Anchor: "/shoutbox#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在社团 <b>{{.Title | html}}</b> 的留言板中提到了你"),
		ID:     31,
		Hash:   31,
		Merge:  true,
	},
	32: {
		URL:    "DOUJIN_URL/club/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在社团讨论 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     32,
		Hash:   17,
		Merge:  true,
	},
	33: {
		URL:    "DOUJIN_URL/subject/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在同人作品 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     33,
		Hash:   19,
		Merge:  true,
	},
	34: {
		URL:    "DOUJIN_URL/event/topic/",
		Anchor: "#post_",
		Temp:   tmpl("<code>{{.FromNickname | html}}</code> 在展会讨论 <b>{{.Title | html}}</b> 中提到了你"),
		ID:     34,
		Hash:   20,
		Merge:  true,
	},
}
