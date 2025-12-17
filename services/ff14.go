package services

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"

	"llmaget/config"
	"llmaget/models"
)

// FF14Service FF14 石之家服务
type FF14Service struct {
	client *resty.Client
	state  *config.AppState
}

// NewFF14Service 创建 FF14 服务实例
func NewFF14Service() *FF14Service {
	client := resty.New().
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	return &FF14Service{
		client: client,
		state:  config.GetState(),
	}
}

// buildURL 构建完整 URL
func (s *FF14Service) buildURL(path string) string {
	return fmt.Sprintf("%s://%s%s", config.Scheme, config.BaseURL, path)
}

// setCommonHeaders 设置通用请求头
func (s *FF14Service) setCommonHeaders(req *resty.Request) *resty.Request {
	cfg := s.state.GetConfig()
	return req.
		SetHeader("User-Agent", cfg.UserAgent).
		SetHeader("Cookie", fmt.Sprintf("ff14risingstones=%s", cfg.Cookie)).
		SetHeader("Accept", "application/json, text/plain, */*").
		SetHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
}

// GetBindInfo 获取角色绑定信息
func (s *FF14Service) GetBindInfo() error {
	log.Println("🚀 开始获取数据...")

	if !s.state.HasCookie() {
		log.Println("⚠️ Cookie未配置，跳过数据获取")
		return fmt.Errorf("cookie未配置")
	}

	req := s.setCommonHeaders(s.client.R())

	resp, err := req.
		SetQueryParams(map[string]string{
			"platform": "2",
			"tempsuid": uuid.New().String(),
		}).
		Get(s.buildURL(config.BindInfoPath))

	if err != nil {
		log.Printf("❌ 请求失败: %v", err)
		return fmt.Errorf("请求失败: %w", err)
	}

	log.Printf("📥 收到响应 (状态码: %d, 长度: %d)", resp.StatusCode(), len(resp.Body()))

	if err := s.saveResponse(resp.Body()); err != nil {
		log.Printf("❌ 保存响应失败: %v", err)
		return fmt.Errorf("保存响应失败: %w", err)
	}

	log.Printf("✅ 数据获取完成! 结果已保存到 %s", config.OutputFile)
	return nil
}

func (s *FF14Service) SaveMyBaseInfo() error {

	infoResp, err := s.GetUserInfo("")
	if err != nil {
		log.Printf("❌ 获取数据失败: %v", err)
		return fmt.Errorf("获取数据失败: %w", err)
	}

	if err := s.saveBaseInfo(infoResp); err != nil {
		log.Printf("❌ 保存响应失败: %v", err)
		return fmt.Errorf("保存响应失败: %w", err)
	}

	log.Printf("✅ 数据获取完成! 结果已保存到 %s", config.OutputFile)
	return nil
}

func (s *FF14Service) GetUserInfo(userId string) (*models.UserInfoResp, error) {
	log.Println("🚀 开始获取数据...")

	if !s.state.HasCookie() {
		log.Println("⚠️ Cookie未配置，跳过数据获取")
		return nil, fmt.Errorf("cookie未配置")
	}

	req := s.setCommonHeaders(s.client.R())

	params := map[string]string{
		"tempsuid": uuid.New().String(),
	}
	if userId != "" {
		params["uuid"] = userId
	} else {
		log.Printf("未提供用户id，获取当前登录用户信息")
	}

	resp, err := req.
		SetQueryParams(params).
		Get(s.buildURL(config.UserInfoPath))

	if err != nil {
		log.Printf("❌ 请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	log.Printf("📥 收到响应 (状态码: %d, 长度: %d)", resp.StatusCode(), len(resp.Body()))
	var userInfoResp models.UserInfoResp
	if err := sonic.Unmarshal(resp.Body(), &userInfoResp); err != nil {
		log.Printf("❌ 解析响应失败: %v", err)
	}

	return &userInfoResp, nil
}

func (s *FF14Service) SignAndGetSignReward() ([]byte, error) {
	log.Printf("开始签到并检测奖励...")
	_, err := s.SignIn()
	if err != nil {
		log.Printf("❌ 签到时发生错误")
		return nil, err
	}

	rewardsBody, err := s.SignRewardList()
	if err != nil {
		log.Printf("❌ 获取奖励列表时发生错误")
		return nil, err
	}

	respMap := map[string][]string{
		"unavailable": {},
		"available":   {},
		"claimed":     {},
		"success":     {},
		"fail":        {},
	}

	for _, reward := range rewardsBody.Data {
		if reward.IsGet == 0 {
			respMap["available"] = append(respMap["available"], reward.ItemName)
			log.Printf("奖励 %s 可领取！", reward.ItemName)
			resp, err := s.GetSignReward(reward.ID)
			if err != nil {
				respMap["fail"] = append(respMap["fail"], reward.ItemName)
				log.Printf("❌ 奖励 %s 领取失败！错误：%s 响应：%s", reward.ItemName, err, string(resp))
				return nil, err
			} else {
				respMap["success"] = append(respMap["success"], reward.ItemName)
				log.Printf("✅ 奖励 %s 领取成功！响应：%s", reward.ItemName, string(resp))
			}
			continue
		} else if reward.IsGet == 1 {
			respMap["claimed"] = append(respMap["claimed"], reward.ItemName)
			log.Printf("奖励 %s 已领取，跳过...", reward.ItemName)
			continue
		} else {
			respMap["unavailable"] = append(respMap["unavailable"], reward.ItemName)
			log.Printf("奖励 %s 暂未达到领取条件，跳过...", reward.ItemName)
			continue
		}
	}
	log.Printf("奖励领取处理完成")
	resp, err := sonic.Marshal(respMap)
	if err != nil {
		log.Fatalf("map转换json失败 %v", &err)
		return nil, err
	}
	return resp, nil
}

// SignIn 执行签到
func (s *FF14Service) SignIn() ([]byte, error) {
	log.Printf("📝 开始尝试打卡...")

	if !s.state.HasCookie() {
		log.Println("⚠️ Cookie未配置")
		return nil, fmt.Errorf("cookie未配置")
	}

	req := s.setCommonHeaders(s.client.R())

	resp, err := req.
		SetQueryParam("tempsuid", uuid.New().String()).
		Post(s.buildURL(config.SignInPath))

	if err != nil {
		log.Printf("❌ 请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	log.Printf("📔 签到响应: %s", string(resp.Body()))

	return resp.Body(), nil
}

func (s *FF14Service) SignRewardList() (*models.SignInRewards, error) {
	log.Printf("📝 获取签到奖励列表...")

	if !s.state.HasCookie() {
		log.Println("⚠️ Cookie未配置")
		return nil, fmt.Errorf("cookie未配置")
	}

	req := s.setCommonHeaders(s.client.R())

	resp, err := req.
		SetQueryParams(map[string]string{
			"tempsuid": uuid.New().String(),
			"month":    time.Now().Format("2006-01"),
		}).
		Get(s.buildURL(config.SignRewardsPath))

	if err != nil {
		log.Printf("❌ 请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	log.Printf("📔 签到奖励列表响应: %s", string(resp.Body()))
	var result models.SignInRewards
	if err := sonic.Unmarshal(resp.Body(), &result); err != nil {
		log.Printf("❌ 解析响应失败: %v", err)
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

func (s *FF14Service) GetSignReward(id int) ([]byte, error) {
	log.Printf("🎁 领取签到奖励... 奖励id %d", id)

	if !s.state.HasCookie() {
		log.Println("⚠️ Cookie未配置")
		return nil, fmt.Errorf("cookie未配置")
	}

	req := s.setCommonHeaders(s.client.R())

	reqBody := map[string]any{
		"id":    id,
		"month": time.Now().Format("2006-01"),
	}
	resp, err := req.
		SetBody(reqBody).
		Post(s.buildURL(config.GetSignRewardPath))

	if err != nil {
		log.Printf("❌ 请求失败: %v", err)
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	body := resp.Body()
	log.Printf("🎁 领取签到奖励响应: %s", string(body))

	return body, nil
}

// SearchUser 搜索用户
func (s *FF14Service) SearchUser(name string, groupName string) (*models.UserInfo, error) {
	areaName := GetAreaName(groupName)

	log.Printf("🔍 开始搜索用户: %s", name)

	if !s.state.HasCookie() {
		return nil, fmt.Errorf("cookie未配置")
	}

	for page := 1; page <= 30; page++ {
		req := s.setCommonHeaders(s.client.R())

		resp, err := req.
			SetQueryParams(map[string]string{
				"tempsuid": uuid.New().String(),
				"type":     "6",
				"orderBy":  "comment",
				"keywords": name,
				"limit":    "60",
				"page":     strconv.Itoa(page),
			}).
			Get(s.buildURL(config.SearchUserPath))

		if err != nil {
			log.Printf("❌ 请求失败: %v", err)
			return nil, fmt.Errorf("请求失败: %w", err)
		}

		var result models.SearchResponse
		if err := sonic.Unmarshal(resp.Body(), &result); err != nil {
			log.Printf("❌ 解析响应失败: %v", err)
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}
		if result.Code != 10000 {
			return nil, fmt.Errorf("错误码%d", result.Code)
		}

		var data []models.UserProfile
		if err := sonic.Unmarshal(result.Data, &data); err != nil {
			return nil, fmt.Errorf("解析Data数据失败: %w", err)
		}

		if len(data) == 0 {
			break
		}

		for _, user := range data {
			if user.CharacterName == name {
				if groupName == "" && areaName == "" {
					return s.parseUserInfo(user), nil
				} else if groupName != "" && user.GroupName == groupName {
					return s.parseUserInfo(user), nil
				} else if areaName != "" && user.AreaName == areaName {
					return s.parseUserInfo(user), nil
				}
			}
		}
	}

	return nil, fmt.Errorf("未找到用户: %s", name)
}

// parseUserInfo 解析用户信息
func (s *FF14Service) parseUserInfo(user models.UserProfile) *models.UserInfo {
	return &models.UserInfo{
		UUID:      user.UUID,
		UserName:  user.CharacterName,
		GroupName: user.GroupName,
		AreaName:  user.AreaName,
	}
}

// saveResponse 保存响应数据
func (s *FF14Service) saveResponse(body []byte) error {
	var data []byte

	// 尝试格式化 JSON
	var prettyJSON map[string]any
	if err := sonic.Unmarshal(body, &prettyJSON); err != nil {
		data = body
	} else {
		formatted, err := sonic.MarshalIndent(prettyJSON, "", "  ")
		if err != nil {
			data = body
		} else {
			data = formatted
		}
	}

	log.Printf("📄 响应内容预览:\n%s", truncateString(string(data), 500))

	// 保存到内存
	s.state.SetResponseData(data)

	// 保存到文件
	return os.WriteFile(config.OutputFile, data, 0644)
}

func (s *FF14Service) saveBaseInfo(infoResp *models.UserInfoResp) error {
	b, err := sonic.Marshal(infoResp)
	if err != nil {
		return fmt.Errorf("编码infoResp失败: %w", err)
	}
	s.state.SetResponseData(b)
	return os.WriteFile(config.OutputFile, b, 0644)
}

// ParseFFInfo 获取处理后的 FF 信息
func (s *FF14Service) ParseFFInfo() (*models.FFInfoData, error) {
	data := s.state.GetResponseData()

	if len(data) == 0 {
		// 尝试从文件读取
		fileData, err := os.ReadFile(config.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("数据尚未获取")
		}
		data = fileData
	}

	var apiResp models.UserInfoResp
	if err := sonic.Unmarshal(data, &apiResp); err != nil {
		return nil, fmt.Errorf("数据解析失败: %w", err)
	}

	playTimeMinutes := ParsePlayTimeToMinutes(apiResp.Data.CharacterDetail[0].PlayTime)

	return &models.FFInfoData{
		CharacterName: apiResp.Data.CharacterName,
		PlayTime:      playTimeMinutes,
	}, nil
}

// ParsePlayTimeToMinutes 将 "X天Y小时Z分钟" 格式转换为分钟数
func ParsePlayTimeToMinutes(playTime string) int {
	totalMinutes := 0

	dayRe := regexp.MustCompile(`(\d+)天`)
	if matches := dayRe.FindStringSubmatch(playTime); len(matches) > 1 {
		days, _ := strconv.Atoi(matches[1])
		totalMinutes += days * 24 * 60
	}

	hourRe := regexp.MustCompile(`(\d+)小时`)
	if matches := hourRe.FindStringSubmatch(playTime); len(matches) > 1 {
		hours, _ := strconv.Atoi(matches[1])
		totalMinutes += hours * 60
	}

	minRe := regexp.MustCompile(`(\d+)分钟`)
	if matches := minRe.FindStringSubmatch(playTime); len(matches) > 1 {
		mins, _ := strconv.Atoi(matches[1])
		totalMinutes += mins
	}

	return totalMinutes
}

// GetAreaName 获取区域名称
func GetAreaName(serverName string) string {
	switch serverName {
	case "n", "鸟", "陆行鸟":
		return "陆行鸟"
	case "m", "猫", "猫小胖":
		return "猫小胖"
	case "g", "狗", "豆豆柴":
		return "豆豆柴"
	case "z", "猪", "莫古力":
		return "莫古力"
	default:
		return ""
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
