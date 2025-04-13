package main

// 导入 fmt 包用于可能的错误处理或日志记录

// Cfg 结构体定义，对应 Python 的 dataclass
type Cfg struct {
	URL       string
	URLMobile *string // 使用指针类型以支持 nil (对应 Python 的 None)
	Anchor    string
	Prefix    string
	Suffix    string
	ID        int
	Hash      int
	Merge     bool
}

// stringPtr 是一个辅助函数，用于创建字符串指针
func stringPtr(s string) *string {
	return &s
}

// getNotifyConfig 根据通知 ID 获取对应的配置
// 返回 Cfg 和一个布尔值，表示是否找到了对应的配置
func getNotifyConfig(notifyID int) (Cfg, bool) {
	switch notifyID {
	case 1:
		return Cfg{
			URL:       "https://bgm.tv/group/topic",
			URLMobile: stringPtr("MOBILE_URL/topic/group/"),
			Anchor:    "#post_",
			Prefix:    "在你的小组话题",
			Suffix:    "中发表了新回复",
			ID:        1,
			Hash:      1,
			Merge:     true,
		}, true
	case 2:
		return Cfg{
			URL:       "https://bgm.tv/group/topic",
			URLMobile: stringPtr("MOBILE_URL/topic/group/"),
			Anchor:    "#post_",
			Prefix:    "在小组话题",
			Suffix:    "中回复了你",
			ID:        2,
			Hash:      1,
			Merge:     true,
		}, true
	case 3:
		return Cfg{
			URL:       "https://bgm.tv/subject/topic",
			URLMobile: stringPtr("/topic/subject"),
			Anchor:    "#post_",
			Prefix:    "在你的条目讨论",
			Suffix:    "中发表了新回复",
			ID:        3,
			Hash:      3,
			Merge:     true,
		}, true
	case 4:
		return Cfg{
			URL:       "https://bgm.tv/subject/topic/",
			URLMobile: stringPtr("MOBILE_URL/topic/subject/"),
			Anchor:    "#post_",
			Prefix:    "在条目讨论",
			Suffix:    "中回复了你",
			ID:        4,
			Hash:      3,
			Merge:     true,
		}, true
	case 5:
		return Cfg{
			URL:       "https://bgm.tv/character/",
			URLMobile: stringPtr("MOBILE_URL/topic/crt/"),
			Anchor:    "#post_",
			Prefix:    "在角色讨论",
			Suffix:    "中发表了新回复",
			ID:        5,
			Hash:      5,
			Merge:     true,
		}, true
	case 6:
		return Cfg{
			URL:       "https://bgm.tv/character/",
			URLMobile: stringPtr("MOBILE_URL/topic/crt/"),
			Anchor:    "#post_",
			Prefix:    "在角色",
			Suffix:    "中回复了你",
			ID:        6,
			Hash:      5,
			Merge:     true,
		}, true
	case 7:
		return Cfg{
			URL:       "/blog/",
			URLMobile: nil, // 对应 Python 的 None
			Anchor:    "#post_",
			Prefix:    "在你的日志",
			Suffix:    "中发表了新回复",
			ID:        7,
			Hash:      7,
			Merge:     true,
		}, true
	case 8:
		return Cfg{
			URL:       "https://bgm.tv/blog/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在日志",
			Suffix:    "中回复了你",
			ID:        8,
			Hash:      7,
			Merge:     true,
		}, true
	case 9:
		return Cfg{
			URL:       "https://bgm.tv/subject/ep/",
			URLMobile: stringPtr("MOBILE_URL/topic/ep/"),
			Anchor:    "#post_",
			Prefix:    "在章节讨论",
			Suffix:    "中发表了新回复",
			ID:        9,
			Hash:      9,
			Merge:     true,
		}, true
	case 10:
		return Cfg{
			URL:       "https://bgm.tv/subject/ep/",
			URLMobile: stringPtr("MOBILE_URL/topic/ep/"),
			Anchor:    "#post_",
			Prefix:    "在章节讨论",
			Suffix:    "中回复了你",
			ID:        10,
			Hash:      9,
			Merge:     true,
		}, true
	case 11:
		return Cfg{
			URL:       "https://bgm.tv/index/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在目录",
			Suffix:    "中给你留言了",
			ID:        11,
			Hash:      11,
			Merge:     true,
		}, true
	case 12:
		return Cfg{
			URL:       "https://bgm.tv/index/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在目录",
			Suffix:    "中回复了你",
			ID:        12,
			Hash:      11,
			Merge:     true,
		}, true
	case 13:
		return Cfg{
			URL:       "https://bgm.tv/person/",
			URLMobile: stringPtr("MOBILE_URL/topic/prsn/"),
			Anchor:    "#post_",
			Prefix:    "在人物",
			Suffix:    "中回复了你",
			ID:        13,
			Hash:      13,
			Merge:     true,
		}, true
	case 14:
		return Cfg{
			URL:       "https://bgm.tv/user/",
			URLMobile: nil,
			Anchor:    "#",
			Prefix:    "请求与你成为好友",
			Suffix:    "",
			ID:        14,
			Hash:      14,
			Merge:     false,
		}, true
	case 15:
		return Cfg{
			URL:       "https://bgm.tv/user/",
			URLMobile: nil,
			Anchor:    "#",
			Prefix:    "通过了你的好友请求",
			Suffix:    "",
			ID:        15,
			Hash:      14,
			Merge:     false,
		}, true
	case 17:
		return Cfg{
			URL:       "DOUJIN_URL/club/topic/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在你的社团讨论",
			Suffix:    "中发表了新回复",
			ID:        17,
			Hash:      17,
			Merge:     true,
		}, true
	case 18:
		return Cfg{
			URL:       "DOUJIN_URL/club/topic/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在社团讨论",
			Suffix:    "中回复了你",
			ID:        18,
			Hash:      17,
			Merge:     true,
		}, true
	case 19:
		return Cfg{
			URL:       "DOUJIN_URL/subject/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在同人作品",
			Suffix:    "中回复了你",
			ID:        19,
			Hash:      19,
			Merge:     true,
		}, true
	case 20:
		return Cfg{
			URL:       "DOUJIN_URL/event/topic/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在你的展会讨论",
			Suffix:    "中发表了新回复",
			ID:        20,
			Hash:      20,
			Merge:     true,
		}, true
	case 21:
		return Cfg{
			URL:       "DOUJIN_URL/event/topic/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在展会讨论",
			Suffix:    "中回复了你",
			ID:        21,
			Hash:      20,
			Merge:     true,
		}, true
	case 22:
		return Cfg{
			URL:       "https://bgm.tv/user/chobits_user/timeline/status/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    `回复了你的 <a href="%2$s%3$s" class="nt_link link_%4$s" target="_blank">吐槽</a>`,
			Suffix:    "",
			ID:        22,
			Hash:      22,
			Merge:     true,
		}, true
	case 23:
		return Cfg{
			URL:       "https://bgm.tv/group/topic/",
			URLMobile: stringPtr("MOBILE_URL/topic/group/"),
			Anchor:    "#post_",
			Prefix:    "在小组话题",
			Suffix:    "中提到了你",
			ID:        23,
			Hash:      1,
			Merge:     true,
		}, true
	case 24:
		return Cfg{
			URL:       "https://bgm.tv/subject/topic/",
			URLMobile: stringPtr("MOBILE_URL/topic/subject/"),
			Anchor:    "#post_",
			Prefix:    "在条目讨论",
			Suffix:    "中提到了你",
			ID:        24,
			Hash:      3,
			Merge:     true,
		}, true
	case 25:
		return Cfg{
			URL:       "https://bgm.tv/character/",
			URLMobile: stringPtr("MOBILE_URL/topic/crt/"),
			Anchor:    "#post_",
			Prefix:    "在角色",
			Suffix:    "中提到了你",
			ID:        25,
			Hash:      5,
			Merge:     true,
		}, true
	case 26:
		return Cfg{
			URL:       "https://bgm.tv/person/",
			URLMobile: stringPtr("MOBILE_URL/topic/prsn/"),
			Anchor:    "#post_",
			Prefix:    "在人物讨论",
			Suffix:    "中提到了你",
			ID:        26,
			Hash:      5,
			Merge:     true,
		}, true
	case 27:
		return Cfg{
			URL:       "https://bgm.tv/index/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在目录",
			Suffix:    "中提到了你",
			ID:        28, // 注意 ID
			Hash:      11,
			Merge:     true,
		}, true
	case 28:
		return Cfg{
			URL:       "https://bgm.tv/user/chobits_user/timeline/status/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在",
			Suffix:    "中提到了你",
			ID:        28,
			Hash:      22,
			Merge:     true,
		}, true
	case 29:
		return Cfg{
			URL:       "https://bgm.tv/blog/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在日志",
			Suffix:    "中提到了你",
			ID:        29,
			Hash:      7,
			Merge:     true,
		}, true
	case 30:
		return Cfg{
			URL:       "https://bgm.tv/subject/ep/",
			URLMobile: stringPtr("MOBILE_URL/topic/ep/"),
			Anchor:    "#post_",
			Prefix:    "在章节讨论",
			Suffix:    "中提到了你",
			ID:        30,
			Hash:      9,
			Merge:     true,
		}, true
	case 31:
		return Cfg{
			URL:       "DOUJIN_URL/club/",
			URLMobile: nil,
			Anchor:    "/shoutbox#post_",
			Prefix:    "在社团",
			Suffix:    "的留言板中提到了你",
			ID:        31,
			Hash:      31,
			Merge:     true,
		}, true
	case 32:
		return Cfg{
			URL:       "DOUJIN_URL/club/topic/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在社团讨论",
			Suffix:    "中提到了你",
			ID:        32,
			Hash:      17,
			Merge:     true,
		}, true
	case 33:
		return Cfg{
			URL:       "DOUJIN_URL/subject/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在同人作品",
			Suffix:    "中提到了你",
			ID:        33,
			Hash:      19,
			Merge:     true,
		}, true
	case 34:
		return Cfg{
			URL:       "DOUJIN_URL/event/topic/",
			URLMobile: nil,
			Anchor:    "#post_",
			Prefix:    "在展会讨论",
			Suffix:    "中提到了你",
			ID:        34,
			Hash:      20,
			Merge:     true,
		}, true
	default:
		// 如果没有匹配的 ID，返回一个零值的 Cfg 和 false
		return Cfg{}, false
	}
}
