import pydantic


class Cfg(pydantic.BaseModel):
    url: str
    url_mobile: str | None = None
    anchor: str
    prefix: str
    suffix: str
    id: int
    hash: int
    merge: bool = False


notify_types = {
    1: Cfg(
        url="https://bgm.tv/group/topic",
        url_mobile="MOBILE_URL/topic/group/",
        anchor="#post_",
        prefix="在你的小组话题",
        suffix="中发表了新回复",
        id=1,
        hash=1,
        merge=True,
    ),
    2: Cfg(
        url="https://bgm.tv/group/topic",
        url_mobile="MOBILE_URL/topic/group/",
        anchor="#post_",
        prefix="在小组话题",
        suffix="中回复了你",
        id=2,
        hash=1,
        merge=True,
    ),
    3: Cfg(
        url="https://bgm.tv/subject/topic",
        url_mobile="/topic/subject",
        anchor="#post_",
        prefix="在你的条目讨论",
        suffix="中发表了新回复",
        id=3,
        hash=3,
        merge=True,
    ),
    4: Cfg(
        url="https://bgm.tv/subject/topic/",
        url_mobile="MOBILE_URL/topic/subject/",
        anchor="#post_",
        prefix="在条目讨论",
        suffix="中回复了你",
        id=4,
        hash=3,
        merge=True,
    ),
    5: Cfg(
        url="https://bgm.tv/character/",
        url_mobile="MOBILE_URL/topic/crt/",
        anchor="#post_",
        prefix="在角色讨论",
        suffix="中发表了新回复",
        id=5,
        hash=5,
        merge=True,
    ),
    6: Cfg(
        url="https://bgm.tv/character/",
        url_mobile="MOBILE_URL/topic/crt/",
        anchor="#post_",
        prefix="在角色",
        suffix="中回复了你",
        id=6,
        hash=5,
        merge=True,
    ),
    7: Cfg(
        url="/blog/",
        url_mobile=None,
        anchor="#post_",
        prefix="在你的日志",
        suffix="中发表了新回复",
        id=7,
        hash=7,
        merge=True,
    ),
    8: Cfg(
        url="https://bgm.tv/blog/",
        url_mobile=None,
        anchor="#post_",
        prefix="在日志",
        suffix="中回复了你",
        id=8,
        hash=7,
        merge=True,
    ),
    9: Cfg(
        url="https://bgm.tv/subject/ep/",
        url_mobile="MOBILE_URL/topic/ep/",
        anchor="#post_",
        prefix="在章节讨论",
        suffix="中发表了新回复",
        id=9,
        hash=9,
        merge=True,
    ),
    10: Cfg(
        url="https://bgm.tv/subject/ep/",
        url_mobile="MOBILE_URL/topic/ep/",
        anchor="#post_",
        prefix="在章节讨论",
        suffix="中回复了你",
        id=10,
        hash=9,
        merge=True,
    ),
    11: Cfg(
        url="https://bgm.tv/index/",
        url_mobile=None,
        anchor="#post_",
        prefix="在目录",
        suffix="中给你留言了",
        id=11,
        hash=11,
        merge=True,
    ),
    12: Cfg(
        url="https://bgm.tv/index/",
        url_mobile=None,
        anchor="#post_",
        prefix="在目录",
        suffix="中回复了你",
        id=12,
        hash=11,
        merge=True,
    ),
    13: Cfg(
        url="https://bgm.tv/person/",
        url_mobile="MOBILE_URL/topic/prsn/",
        anchor="#post_",
        prefix="在人物",
        suffix="中回复了你",
        id=13,
        hash=13,
        merge=True,
    ),
    14: Cfg(
        url="https://bgm.tv/user/",
        url_mobile=None,
        anchor="#",
        prefix="请求与你成为好友",
        suffix="",
        id=14,
        hash=14,
        merge=False,
    ),
    15: Cfg(
        url="https://bgm.tv/user/",
        url_mobile=None,
        anchor="#",
        prefix="通过了你的好友请求",
        suffix="",
        id=15,
        hash=14,
        merge=False,
    ),
    17: Cfg(
        url="DOUJIN_URL/club/topic/",
        url_mobile=None,
        anchor="#post_",
        prefix="在你的社团讨论",
        suffix="中发表了新回复",
        id=17,
        hash=17,
        merge=True,
    ),
    18: Cfg(
        url="DOUJIN_URL/club/topic/",
        url_mobile=None,
        anchor="#post_",
        prefix="在社团讨论",
        suffix="中回复了你",
        id=18,
        hash=17,
        merge=True,
    ),
    19: Cfg(
        url="DOUJIN_URL/subject/",
        url_mobile=None,
        anchor="#post_",
        prefix="在同人作品",
        suffix="中回复了你",
        id=19,
        hash=19,
        merge=True,
    ),
    20: Cfg(
        url="DOUJIN_URL/event/topic/",
        url_mobile=None,
        anchor="#post_",
        prefix="在你的展会讨论",
        suffix="中发表了新回复",
        id=20,
        hash=20,
        merge=True,
    ),
    21: Cfg(
        url="DOUJIN_URL/event/topic/",
        url_mobile=None,
        anchor="#post_",
        prefix="在展会讨论",
        suffix="中回复了你",
        id=21,
        hash=20,
        merge=True,
    ),
    22: Cfg(
        url="https://bgm.tv/user/chobits_user/timeline/status/",
        url_mobile=None,
        anchor="#post_",
        prefix='回复了你的 <a href="%2$s%3$s" class="nt_link link_%4$s" target="_blank">吐槽</a>',
        suffix="",
        id=22,
        hash=22,
        merge=True,
    ),
    23: Cfg(
        url="https://bgm.tv/group/topic/",
        url_mobile="MOBILE_URL/topic/group/",
        anchor="#post_",
        prefix="在小组话题",
        suffix="中提到了你",
        id=23,
        hash=1,
        merge=True,
    ),
    24: Cfg(
        url="https://bgm.tv/subject/topic/",
        url_mobile="MOBILE_URL/topic/subject/",
        anchor="#post_",
        prefix="在条目讨论",
        suffix="中提到了你",
        id=24,
        hash=3,
        merge=True,
    ),
    25: Cfg(
        url="https://bgm.tv/character/",
        url_mobile="MOBILE_URL/topic/crt/",
        anchor="#post_",
        prefix="在角色",
        suffix="中提到了你",
        id=25,
        hash=5,
        merge=True,
    ),
    26: Cfg(
        url="https://bgm.tv/person/",
        url_mobile="MOBILE_URL/topic/prsn/",
        anchor="#post_",
        prefix="在人物讨论",
        suffix="中提到了你",
        id=26,
        hash=5,
        merge=True,
    ),
    27: Cfg(
        url="https://bgm.tv/index/",
        url_mobile=None,
        anchor="#post_",
        prefix="在目录",
        suffix="中提到了你",
        id=28,
        hash=11,
        merge=True,
    ),
    28: Cfg(
        url="https://bgm.tv/user/chobits_user/timeline/status/",
        url_mobile=None,
        anchor="#post_",
        prefix="在",
        suffix="中提到了你",
        id=28,
        hash=22,
        merge=True,
    ),
    29: Cfg(
        url="https://bgm.tv/blog/",
        url_mobile=None,
        anchor="#post_",
        prefix="在日志",
        suffix="中提到了你",
        id=29,
        hash=7,
        merge=True,
    ),
    30: Cfg(
        url="https://bgm.tv/subject/ep/",
        url_mobile="MOBILE_URL/topic/ep/",
        anchor="#post_",
        prefix="在章节讨论",
        suffix="中提到了你",
        id=30,
        hash=9,
        merge=True,
    ),
    31: Cfg(
        url="DOUJIN_URL/club/",
        url_mobile=None,
        anchor="/shoutbox#post_",
        prefix="在社团",
        suffix="的留言板中提到了你",
        id=31,
        hash=31,
        merge=True,
    ),
    32: Cfg(
        url="DOUJIN_URL/club/topic/",
        url_mobile=None,
        anchor="#post_",
        prefix="在社团讨论",
        suffix="中提到了你",
        id=32,
        hash=17,
        merge=True,
    ),
    33: Cfg(
        url="DOUJIN_URL/subject/",
        url_mobile=None,
        anchor="#post_",
        prefix="在同人作品",
        suffix="中提到了你",
        id=33,
        hash=19,
        merge=True,
    ),
    34: Cfg(
        url="DOUJIN_URL/event/topic/",
        url_mobile=None,
        anchor="#post_",
        prefix="在展会讨论",
        suffix="中提到了你",
        id=34,
        hash=20,
        merge=True,
    ),
}