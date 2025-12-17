package config

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
)

const (
	ConfigFile    = "config.json"
	OutputFile    = "response.json"
	ServerPort    = ":8080"
	SignInterval  = 24 * time.Hour
	FetchInterval = 12 * time.Hour
)

// FF14 API 相关常量
const (
	Scheme            = "https"
	BaseURL           = "apiff14risingstones.web.sdo.com"
	UserInfoPath      = "/api/home/userInfo/getUserInfo"
	SignRewardsPath   = "/api/home/sign/signRewardList"
	GetSignRewardPath = "/api/home/sign/getSignReward" // POST
	BindInfoPath      = "/api/home/groupAndRole/getCharacterBindInfo"
	SignInPath        = "/api/home/sign/signIn" // POST
	SearchUserPath    = "/api/common/search"
)

// Config 存储配置信息
type Config struct {
	UserAgent string `json:"user_agent"`
	Cookie    string `json:"cookie"`
}

// AppState 应用状态
type AppState struct {
	mu           sync.RWMutex
	config       Config
	responseData []byte
	lastFetchAt  time.Time
}

var (
	state = &AppState{}
)

// GetState 获取应用状态单例
func GetState() *AppState {
	return state
}

// Load 从文件加载配置
func (s *AppState) Load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		log.Printf("⚠️ 配置文件不存在，使用默认配置")
		s.config = Config{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Cookie:    "",
		}
		s.saveUnsafe()
		return
	}

	if err := sonic.Unmarshal(data, &s.config); err != nil {
		log.Printf("⚠️ 配置文件解析失败: %v，使用默认配置", err)
		s.config = Config{
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Cookie:    "",
		}
		return
	}

	log.Printf("✅ 配置加载成功")
}

// saveUnsafe 保存配置（不加锁，内部使用）
func (s *AppState) saveUnsafe() error {
	data, err := sonic.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, data, 0644)
}

// Save 保存配置到文件
func (s *AppState) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.saveUnsafe()
}

// GetConfig 获取配置副本
func (s *AppState) GetConfig() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetConfig 更新配置
func (s *AppState) SetConfig(cfg Config) error {
	s.mu.Lock()
	if cfg.UserAgent != "" {
		s.config.UserAgent = cfg.UserAgent
	}
	if cfg.Cookie != "" {
		s.config.Cookie = cfg.Cookie
	}
	s.mu.Unlock()
	return s.Save()
}

// HasCookie 检查是否配置了 Cookie
func (s *AppState) HasCookie() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.Cookie != ""
}

// GetResponseData 获取响应数据
func (s *AppState) GetResponseData() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.responseData
}

// SetResponseData 设置响应数据
func (s *AppState) SetResponseData(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.responseData = data
	s.lastFetchAt = time.Now()
}

// GetLastFetchAt 获取最后获取时间
func (s *AppState) GetLastFetchAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastFetchAt
}

// HasData 检查是否有数据
func (s *AppState) HasData() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.responseData) > 0
}
