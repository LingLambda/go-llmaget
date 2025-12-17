package models

import (
	"github.com/bytedance/sonic"
)

// APIResponse FF14 API 原始响应结构
type APIResponse struct {
	Code int `json:"code"`
	Data struct {
		CharacterName   string `json:"character_name"`
		CharacterDetail struct {
			CharacterName string `json:"character_name"`
			PlayTime      string `json:"play_time"`
		} `json:"characterDetail"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// UserInfo 用户信息
type UserInfo struct {
	UUID      string `json:"uuid"`
	UserName  string `json:"user_name"`
	GroupName string `json:"group_name"`
	AreaName  string `json:"area_name"`
}

type UserInfoResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ID                   int    `json:"id"`
		UUID                 string `json:"uuid"`
		CharacterName        string `json:"character_name"`
		AreaID               int    `json:"area_id"`
		AreaName             string `json:"area_name"`
		GroupID              int    `json:"group_id"`
		GroupName            string `json:"group_name"`
		Avatar               string `json:"avatar"`
		Profile              string `json:"profile"`
		WeekdayTime          string `json:"weekday_time"`
		WeekendTime          string `json:"weekend_time"`
		Qq                   string `json:"qq"`
		CareerPublish        int    `json:"career_publish"`
		GuildPublish         int    `json:"guild_publish"`
		CreateTimePublish    int    `json:"create_time_publish"`
		LastLoginTimePublish int    `json:"last_login_time_publish"`
		PlayTimePublish      int    `json:"play_time_publish"`
		HouseInfoPublish     int    `json:"house_info_publish"`
		WashingNumPublish    int    `json:"washing_num_publish"`
		AchievePublish       int    `json:"achieve_publish"`
		ResentlyPublish      int    `json:"resently_publish"`
		Experience           string `json:"experience"`
		ThemeID              any    `json:"theme_id"`
		TestLimitedBadge     int    `json:"test_limited_badge"`
		Posts2CreatorBadge   int    `json:"posts2_creator_badge"`
		AdminTag             int    `json:"admin_tag"`
		PublishTab           string `json:"publish_tab"`
		AchieveTab           string `json:"achieve_tab"`
		TreasureTimesPublish int    `json:"treasure_times_publish"`
		KillTimesPublish     int    `json:"kill_times_publish"`
		NewrankPublish       int    `json:"newrank_publish"`
		CrystalRankPublish   int    `json:"crystal_rank_publish"`
		FishTimesPublish     int    `json:"fish_times_publish"`
		CollapseBadge        int    `json:"collapse_badge"`
		FeifeiBadge          int    `json:"feifei_badge"`
		Badge                string `json:"badge"`
		AchieveInfo          []struct {
			MedalID       string `json:"medal_id"`
			MedalType     string `json:"medal_type"`
			AchieveID     string `json:"achieve_id"`
			AchieveTime   string `json:"achieve_time"`
			GroupID       string `json:"group_id"`
			CharacterName string `json:"character_name"`
			MedalTypeID   string `json:"medal_type_id"`
			AchieveName   string `json:"achieve_name"`
			AreaID        string `json:"area_id"`
			AchieveDetail string `json:"achieve_detail"`
			PartDate      string `json:"part_date"`
		} `json:"achieveInfo"`
		AchieveTopInfo []struct {
			MedalID       string `json:"medal_id"`
			MedalType     string `json:"medal_type"`
			AchieveTime   string `json:"achieve_time"`
			AAchieveID    string `json:"a.achieve_id"`
			GroupID       string `json:"group_id"`
			CharacterName string `json:"character_name"`
			MedalTypeID   string `json:"medal_type_id"`
			AchieveName   string `json:"achieve_name"`
			AreaID        string `json:"area_id"`
			AchieveDetail string `json:"achieve_detail"`
			PartDate      string `json:"part_date"`
		} `json:"achieveTopInfo"`
		CareerLevel []struct {
			Career         string `json:"career"`
			CharacterLevel string `json:"character_level"`
			PartDate       string `json:"part_date"`
			UpdateDate     string `json:"update_date"`
			CareerType     string `json:"career_type"`
		} `json:"careerLevel"`
		CharacterDetail []struct {
			CreateTime    string `json:"create_time"`
			Gender        string `json:"gender"`
			LastLoginTime string `json:"last_login_time"`
			Race          string `json:"race"`
			CharacterName string `json:"character_name"`
			AreaID        string `json:"area_id"`
			PlayTime      string `json:"play_time"`
			GroupID       string `json:"group_id"`
			GuildName     string `json:"guild_name"`
			FcID          string `json:"fc_id"`
			Tribe         string `json:"tribe"`
			GuildTag      string `json:"guild_tag"`
			WashingNum    string `json:"washing_num"`
			PartDate      string `json:"part_date"`
			TreasureTimes string `json:"treasure_times"`
			KillTimes     string `json:"kill_times"`
			Newrank       string `json:"newrank"`
			CrystalRank   string `json:"crystal_rank"`
			FishTimes     string `json:"fish_times"`
		} `json:"characterDetail"`
		FollowFansiNum struct {
			FollowNum int `json:"followNum"`
			FansNum   int `json:"fansNum"`
		} `json:"followFansiNum"`
		InteractNum string `json:"interactNum"`
		BeLikedNum  string `json:"beLikedNum"`
		Relation    int    `json:"relation"`
		NewFens     int    `json:"newFens"`
	} `json:"data"`
}

type SignInRewards struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		ID        int    `json:"id"`
		BeginDate string `json:"begin_date"`
		EndDate   string `json:"end_date"`
		Rule      int    `json:"rule"`
		ItemName  string `json:"item_name"`
		ItemPic   string `json:"item_pic"`
		Num       int    `json:"num"`
		ItemDesc  string `json:"item_desc"`
		IsGet     int    `json:"is_get"`
	} `json:"data"`
}

// Response 统一响应结构
type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// FFInfoData FF 信息数据
type FFInfoData struct {
	CharacterName string `json:"character_name"`
	PlayTime      int    `json:"play_time"`
}

type SearchResponse struct {
	Code    int32
	Data    sonic.NoCopyRawMessage
	Message string
}

type UserProfile struct {
	AdminTag           int    `json:"admin_tag"`
	AreaName           string `json:"area_name"`
	Avatar             string `json:"avatar"`
	CharacterName      string `json:"character_name"`
	FansNum            int    `json:"fansNum"`
	GroupName          string `json:"group_name"`
	Posts2CreatorBadge int    `json:"posts2_creator_badge"`
	Profile            string `json:"profile"`
	Relation           int    `json:"relation"`
	TestLimitedBadge   int    `json:"test_limited_badge"`
	UUID               string `json:"uuid"`
}

// StatusData 状态数据
type StatusData struct {
	HasData       bool   `json:"has_data"`
	HasCookie     bool   `json:"has_cookie"`
	LastFetchAt   string `json:"last_fetch_at"`
	NextFetchAt   string `json:"next_fetch_at"`
	FetchInterval string `json:"fetch_interval"`
}

// ConfigRequest 配置请求
type ConfigRequest struct {
	UserAgent string `json:"user_agent"`
	Cookie    string `json:"cookie"`
}

// ConfigData 配置响应数据
type ConfigData struct {
	HasCookie bool `json:"has_cookie"`
}

// NewSuccess 创建成功响应
func NewSuccess(msg string, data any) Response {
	return Response{
		Code: 10000,
		Msg:  msg,
		Data: data,
	}
}

// NewError 创建错误响应
func NewError(code int, msg string) Response {
	return Response{
		Code: code,
		Msg:  msg,
	}
}
