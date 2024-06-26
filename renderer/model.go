package renderer

import (
	"bloat/mastodon"
	"bloat/model"
)

type Context struct {
	HideAttachments  bool
	MaskNSFW         bool
	FluorideMode     bool
	ThreadInNewTab   bool
	DarkMode         bool
	CSRFToken        string
	UserID           string
	AntiDopamineMode bool
	UserCSS          string
	Referrer         string
}

type CommonData struct {
	Title           string
	CustomCSS       string
	CSRFToken       string
	Count           int
	RefreshInterval int
	Target          string
}

type NavData struct {
	CommonData  *CommonData
	User        *mastodon.Account
	PostContext model.PostContext
}

type ErrorData struct {
	*CommonData
	Err        string
	Retry      bool
	SessionErr bool
}

type HomePageData struct {
	*CommonData
}

type SigninData struct {
	*CommonData
}

type RootData struct {
	*CommonData
}

type TimelineData struct {
	*CommonData
	Title    string
	Type     string
	Instance string
	Statuses []*mastodon.Status
	NextLink string
	PrevLink string
}

type ListsData struct {
	*CommonData
	Lists []*mastodon.List
}

type ListData struct {
	*CommonData
	List           *mastodon.List
	Accounts       []*mastodon.Account
	Q              string
	SearchAccounts []*mastodon.Account
}

type ThreadData struct {
	*CommonData
	Statuses    []*mastodon.Status
	PostContext model.PostContext
	ReplyMap    map[string][]mastodon.ReplyInfo
}

type QuickReplyData struct {
	*CommonData
	Ancestor    *mastodon.Status
	Status      *mastodon.Status
	PostContext model.PostContext
}

type NotificationData struct {
	*CommonData
	Notifications []*mastodon.Notification
	UnreadCount   int
	ReadID        string
	NextLink      string
}

type UserData struct {
	*CommonData
	User     *mastodon.Account
	Type     string
	Users    []*mastodon.Account
	Statuses []*mastodon.Status
	NextLink string
}

type UserSearchData struct {
	*CommonData
	User     *mastodon.Account
	Q        string
	Statuses []*mastodon.Status
	NextLink string
}

type AboutData struct {
	*CommonData
}

type EmojiData struct {
	*CommonData
	Emojis []*mastodon.Emoji
}

type LikedByData struct {
	*CommonData
	Users    []*mastodon.Account
	NextLink string
}

type RetweetedByData struct {
	*CommonData
	Users    []*mastodon.Account
	NextLink string
}

type SearchData struct {
	*CommonData
	Q        string
	Type     string
	Users    []*mastodon.Account
	Statuses []*mastodon.Status
	NextLink string
}

type SettingsData struct {
	*CommonData
	Settings    *model.Settings
	PostFormats []model.PostFormat
}

type FiltersData struct {
	*CommonData
	Filters []*mastodon.Filter
}

type ProfileData struct {
	*CommonData
	User *mastodon.Account
}

type MuteData struct {
	*CommonData
	User *mastodon.Account
}
