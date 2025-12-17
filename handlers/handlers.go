package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"

	"llmaget/config"
	"llmaget/models"
	"llmaget/services"
)

// Handler HTTP 处理器
type Handler struct {
	ff14Svc *services.FF14Service
	state   *config.AppState
}

// NewHandler 创建处理器实例
func NewHandler(ff14Svc *services.FF14Service) *Handler {
	return &Handler{
		ff14Svc: ff14Svc,
		state:   config.GetState(),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/llmaget")
	{
		api.GET("/ff_info", h.GetFFInfo)
		api.GET("/status", h.GetStatus)
		api.GET("/refresh", h.Refresh)
		api.GET("/sign_in", h.SignIn)
		api.GET("/config", h.GetConfig)
		api.POST("/config", h.UpdateConfig)
		api.GET("/set", h.SetConfigPage)
		api.GET("/search", h.SearchUserInfo)
		api.GET("/get_sign_reward", h.GetSignReward)
		api.GET("/sign_reward_list", h.SignRewardList)
		api.GET("/sign_and_get_sign_reward", h.SignAndGetSignReward)
	}
}

// GetFFInfo 获取 FF14 角色信息
// @Summary 获取 FF14 角色信息
// @Router /llmaget/ff_info [get]
func (h *Handler) GetFFInfo(c *gin.Context) {
	data, err := h.ff14Svc.ParseFFInfo()
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewError(404, "数据尚未获取，请先配置Cookie后刷新"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccess("success", data))
}

// 领取签到奖励
func (h *Handler) GetSignReward(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewError(400, "错误的请求参数"))
		return
	}

	data, err := h.ff14Svc.GetSignReward(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewError(500, "获取数据发生错误"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccess("success", string(data)))
}

// 获取签到奖励列表
func (h *Handler) SignRewardList(c *gin.Context) {
	data, err := h.ff14Svc.SignRewardList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewError(500, "获取数据发生错误"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccess("success", data))
}

// 签到并领取奖励
func (h *Handler) SignAndGetSignReward(c *gin.Context) {
	result, err := h.ff14Svc.SignAndGetSignReward()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewError(500, "签到并领取奖励过程中发生错误"))
		return
	}

	var data any
	if err := sonic.Unmarshal(result, &data); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewError(500, "解析响应失败"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccess("success", data))
}

// GetStatus 获取服务状态
// @Summary 获取服务状态
// @Router /llmaget/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	lastFetch := h.state.GetLastFetchAt()
	var nextFetch time.Time
	if !lastFetch.IsZero() {
		nextFetch = lastFetch.Add(config.FetchInterval)
	}

	data := models.StatusData{
		HasData:       h.state.HasData(),
		HasCookie:     h.state.HasCookie(),
		LastFetchAt:   formatTime(lastFetch),
		NextFetchAt:   formatTime(nextFetch),
		FetchInterval: config.FetchInterval.String(),
	}

	c.JSON(http.StatusOK, models.Response{
		Code: 10000,
		Msg:  "服务运行中",
		Data: data,
	})
}

// Refresh 手动刷新数据
// @Summary 手动刷新数据
// @Router /llmaget/refresh [get]
func (h *Handler) Refresh(c *gin.Context) {
	if !h.state.HasCookie() {
		c.JSON(http.StatusBadRequest, models.NewError(400, "请先配置Cookie"))
		return
	}

	go h.ff14Svc.SaveMyBaseInfo()

	c.JSON(http.StatusOK, models.NewSuccess("刷新任务已触发，请稍后查询结果", nil))
}

// SignIn 执行签到
// @Summary 执行签到
// @Router /llmaget/sign_in [get]
func (h *Handler) SignIn(c *gin.Context) {
	if !h.state.HasCookie() {
		c.JSON(http.StatusBadRequest, models.NewError(400, "请先配置Cookie"))
		return
	}

	result, err := h.ff14Svc.SignIn()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewError(500, "打卡失败: "+err.Error()))
		return
	}

	var data any
	if err := sonic.Unmarshal(result, &data); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewError(500, "解析响应失败"))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccess("success", data))
}

// GetConfig 获取配置
// @Summary 获取配置
// @Router /llmaget/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	data := models.ConfigData{
		HasCookie: h.state.HasCookie(),
	}
	c.JSON(http.StatusOK, models.NewSuccess("success", data))
}

// UpdateConfig 更新配置
// @Summary 更新配置
// @Router /llmaget/config [post]
func (h *Handler) UpdateConfig(c *gin.Context) {
	var req models.ConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.NewError(400, "请求格式错误: "+err.Error()))
		return
	}

	if err := h.state.SetConfig(config.Config{
		UserAgent: req.UserAgent,
		Cookie:    req.Cookie,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, models.NewError(500, "保存配置失败: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.NewSuccess("配置更新成功", nil))
}

// SetConfigPage 配置页面（支持 GET 方式设置配置）
// @Summary 配置页面
// @Router /llmaget/set [get]
func (h *Handler) SetConfigPage(c *gin.Context) {
	cookie := c.Query("cookie")
	userAgent := c.Query("ua")

	if cookie == "" && userAgent == "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, configPageHTML)
		return
	}

	// 更新配置
	if err := h.state.SetConfig(config.Config{
		UserAgent: userAgent,
		Cookie:    cookie,
	}); err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusInternalServerError, errorPageHTML(err.Error()))
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, successPageHTML)
}

// SearchUserInfo 搜索用户信息页面
// @Summary 搜索用户信息
// @Router /llmaget/search [get]
func (h *Handler) SearchUserInfo(c *gin.Context) {
	name := c.Query("name")
	serverName := c.Query("server_name")

	// 如果没有查询参数，显示搜索页面
	if name == "" && serverName == "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, searchPageHTML)
		return
	}

	// 检查Cookie配置
	if !h.state.HasCookie() {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, searchPageHTML+`
			<script>
				alert("请先配置Cookie");
				window.location.href = "/llmaget/set";
			</script>
		`)
		return
	}

	// 执行搜索
	result, err := h.ff14Svc.SearchUser(name, serverName)
	if err != nil {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, searchResultPageHTML(name, serverName, nil, err.Error()))
		return
	}

	// 显示搜索结果
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, searchResultPageHTML(name, serverName, result, ""))
}

// formatTime 格式化时间
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

// HTML 模板
const configPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>FF14 石之家 - 配置设置</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif; 
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            padding: 40px 20px;
            color: #e8e8e8;
        }
        .container {
            max-width: 560px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 40px;
            border: 1px solid rgba(255, 255, 255, 0.1);
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.3);
        }
        h1 { 
            color: #00d4ff; 
            margin-bottom: 30px;
            font-size: 28px;
            display: flex;
            align-items: center;
            gap: 12px;
        }
        .form-group { margin: 24px 0; }
        label { 
            display: block; 
            margin-bottom: 10px; 
            color: #b8b8b8;
            font-weight: 500;
        }
        input, textarea { 
            width: 100%; 
            padding: 14px 16px; 
            border: 2px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            background: rgba(0, 0, 0, 0.3);
            color: #fff;
            font-size: 14px;
            transition: all 0.3s ease;
        }
        input:focus, textarea:focus {
            outline: none;
            border-color: #00d4ff;
            box-shadow: 0 0 20px rgba(0, 212, 255, 0.2);
        }
        textarea { height: 120px; resize: vertical; }
        button { 
            background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
            color: #000; 
            border: none; 
            padding: 16px 32px; 
            border-radius: 12px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            margin-top: 16px;
            width: 100%;
            transition: all 0.3s ease;
        }
        button:hover { 
            transform: translateY(-2px);
            box-shadow: 0 10px 30px rgba(0, 212, 255, 0.3);
        }
        .hint { 
            font-size: 12px; 
            color: #666; 
            margin-top: 8px;
            line-height: 1.6;
        }
        .links {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
            display: flex;
            flex-wrap: wrap;
            gap: 16px;
        }
        .links a {
            color: #00d4ff;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 8px;
            background: rgba(0, 212, 255, 0.1);
            transition: all 0.3s ease;
        }
        .links a:hover {
            background: rgba(0, 212, 255, 0.2);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🎮 FF14 石之家配置</h1>
        <form method="GET" action="/llmaget/set">
            <div class="form-group">
                <label>Cookie (ff14risingstones 的值)</label>
                <textarea name="cookie" placeholder="粘贴 ff14risingstones cookie 值..."></textarea>
                <div class="hint">💡 在浏览器登录石之家后，F12 → Application → Cookies → 复制 ff14risingstones 的值</div>
            </div>
            <div class="form-group">
                <label>User-Agent (可选)</label>
                <input type="text" name="ua" placeholder="留空使用默认值">
            </div>
            <button type="submit">💾 保存配置</button>
        </form>
        <div class="links">
            <a href="/llmaget/search">🔍 搜索用户</a>
            <a href="/llmaget/refresh">🔄 刷新数据</a>
            <a href="/llmaget/status">📊 查看状态</a>
            <a href="/llmaget/ff_info">📄 查看数据</a>
            <a href="/llmaget/sign_in">✍️ 打卡</a>
        </div>
    </div>
</body>
</html>`

const successPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>配置保存成功</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif; 
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            color: #e8e8e8;
        }
        .container {
            text-align: center;
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 50px;
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        h1 { color: #00ff88; font-size: 32px; margin-bottom: 16px; }
        p { color: #aaa; margin-bottom: 30px; }
        .btn {
            display: inline-block;
            background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
            color: #000;
            padding: 14px 28px;
            border-radius: 10px;
            text-decoration: none;
            font-weight: 600;
            margin: 8px;
            transition: all 0.3s ease;
        }
        .btn:hover { transform: translateY(-2px); }
        .btn-secondary {
            background: transparent;
            color: #00d4ff;
            border: 2px solid rgba(0, 212, 255, 0.3);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>✅ 配置保存成功!</h1>
        <p>配置已更新，可以刷新数据了</p>
        <a href="/llmaget/refresh" class="btn">🔄 立即刷新数据</a>
        <a href="/llmaget/set" class="btn btn-secondary">← 返回配置页</a>
    </div>
</body>
</html>`

func errorPageHTML(errMsg string) string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>保存失败</title>
    <meta charset="utf-8">
    <style>
        body { 
            font-family: sans-serif;
            text-align: center;
            padding: 50px;
            background: #1a1a2e;
            color: #fff;
        }
        h1 { color: #ff4444; }
        a { color: #00d4ff; }
    </style>
</head>
<body>
    <h1>❌ 保存失败</h1>
    <p>` + errMsg + `</p>
    <a href="/llmaget/set">返回</a>
</body>
</html>`
}

// searchPageHTML 搜索页面HTML
const searchPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>FF14 石之家 - 搜索用户</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif; 
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            padding: 40px 20px;
            color: #e8e8e8;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 40px;
            border: 1px solid rgba(255, 255, 255, 0.1);
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.3);
        }
        h1 { 
            color: #00d4ff; 
            margin-bottom: 30px;
            font-size: 28px;
            display: flex;
            align-items: center;
            gap: 12px;
        }
        .form-group { margin: 24px 0; }
        label { 
            display: block; 
            margin-bottom: 10px; 
            color: #b8b8b8;
            font-weight: 500;
        }
        input { 
            width: 100%; 
            padding: 14px 16px; 
            border: 2px solid rgba(255, 255, 255, 0.1);
            border-radius: 12px;
            background: rgba(0, 0, 0, 0.3);
            color: #fff;
            font-size: 14px;
            transition: all 0.3s ease;
        }
        input:focus {
            outline: none;
            border-color: #00d4ff;
            box-shadow: 0 0 20px rgba(0, 212, 255, 0.2);
        }
        button { 
            background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
            color: #000; 
            border: none; 
            padding: 16px 32px; 
            border-radius: 12px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            margin-top: 16px;
            width: 100%;
            transition: all 0.3s ease;
        }
        button:hover { 
            transform: translateY(-2px);
            box-shadow: 0 10px 30px rgba(0, 212, 255, 0.3);
        }
        .hint { 
            font-size: 12px; 
            color: #888; 
            margin-top: 8px;
            line-height: 1.6;
        }
        .links {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
            display: flex;
            flex-wrap: wrap;
            gap: 16px;
        }
        .links a {
            color: #00d4ff;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 8px;
            background: rgba(0, 212, 255, 0.1);
            transition: all 0.3s ease;
        }
        .links a:hover {
            background: rgba(0, 212, 255, 0.2);
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔍 搜索用户</h1>
        <form method="GET" action="/llmaget/search">
            <div class="form-group">
                <label>角色名称 *</label>
                <input type="text" name="name" placeholder="请输入角色名称" required>
                <div class="hint">💡 请输入要搜索的FF14角色名称</div>
            </div>
            <div class="form-group">
                <label>服务器名称 (可选)</label>
                <input type="text" name="server_name" placeholder="请输入服务器名称或区名称">
                <div class="hint">💡 可输入服务器名称进行精确搜索，留空则搜索所有服务器</div>
            </div>
            <button type="submit">🔍 开始搜索</button>
        </form>
        <div class="links">
            <a href="/llmaget/set">⚙️ 配置设置</a>
            <a href="/llmaget/refresh">🔄 刷新数据</a>
            <a href="/llmaget/status">📊 查看状态</a>
            <a href="/llmaget/ff_info">📄 查看数据</a>
            <a href="/llmaget/sign_in">✍️ 打卡</a>
        </div>
    </div>
</body>
</html>`

// searchResultPageHTML 搜索结果页面HTML
func searchResultPageHTML(name, serverName string, result *models.UserInfo, errorMsg string) string {
	var content string

	if errorMsg != "" {
		// 显示错误信息
		content = `
        <div class="result-container error">
            <h2>❌ 搜索失败</h2>
            <p class="error-msg">` + errorMsg + `</p>
            <a href="/llmaget/search" class="btn">🔍 重新搜索</a>
        </div>`
	} else if result != nil {
		// 显示搜索结果
		serverDisplay := result.AreaName
		if serverDisplay == "" {
			serverDisplay = "未指定"
		}

		content = `
        <div class="result-container success">
            <h2>✅ 找到用户</h2>
            <div class="user-info">
                <div class="info-item">
                    <span class="label">角色名称:</span>
                    <span class="value">` + result.UserName + `</span>
                </div>
                <div class="info-item">
                    <span class="label">服务器:</span>
                    <span class="value">` + result.GroupName + `</span>
                </div>
                <div class="info-item">
                    <span class="label">区域:</span>
                    <span class="value">` + serverDisplay + `</span>
                </div>
                <div class="info-item">
                    <span class="label">UUID:</span>
                    <span class="value uuid">` + result.UUID + `</span>
                </div>
            </div>
            <div class="actions">
                <a href="/llmaget/search" class="btn">🔍 继续搜索</a>
                <button onclick="copyUUID('` + result.UUID + `')" class="btn btn-secondary">📋 复制UUID</button>
            </div>
        </div>`
	} else {
		content = `
        <div class="result-container">
            <h2>🔍 搜索结果</h2>
            <p>未找到匹配的用户</p>
            <a href="/llmaget/search" class="btn">🔍 重新搜索</a>
        </div>`
	}

	return `<!DOCTYPE html>
<html>
<head>
    <title>FF14 石之家 - 搜索结果</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif; 
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            padding: 40px 20px;
            color: #e8e8e8;
        }
        .container {
            max-width: 700px;
            margin: 0 auto;
            background: rgba(255, 255, 255, 0.05);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            padding: 40px;
            border: 1px solid rgba(255, 255, 255, 0.1);
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.3);
        }
        h1 { 
            color: #00d4ff; 
            margin-bottom: 30px;
            font-size: 28px;
        }
        .result-container {
            margin: 20px 0;
        }
        .result-container h2 {
            color: #00d4ff;
            margin-bottom: 20px;
            font-size: 24px;
        }
        .result-container.error h2 {
            color: #ff4444;
        }
        .result-container.success h2 {
            color: #00ff88;
        }
        .error-msg {
            color: #ff6666;
            background: rgba(255, 68, 68, 0.1);
            padding: 16px;
            border-radius: 12px;
            border: 1px solid rgba(255, 68, 68, 0.3);
            margin: 16px 0;
        }
        .user-info {
            background: rgba(0, 0, 0, 0.2);
            border-radius: 12px;
            padding: 24px;
            margin: 20px 0;
        }
        .info-item {
            display: flex;
            padding: 12px 0;
            border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        .info-item:last-child {
            border-bottom: none;
        }
        .label {
            color: #b8b8b8;
            font-weight: 500;
            min-width: 100px;
        }
        .value {
            color: #fff;
            flex: 1;
        }
        .value.uuid {
            font-family: 'Courier New', monospace;
            font-size: 12px;
            word-break: break-all;
            color: #00d4ff;
        }
        .actions {
            margin-top: 30px;
            display: flex;
            gap: 16px;
            flex-wrap: wrap;
        }
        .btn {
            display: inline-block;
            background: linear-gradient(135deg, #00d4ff 0%, #0099cc 100%);
            color: #000;
            padding: 14px 28px;
            border-radius: 10px;
            text-decoration: none;
            font-weight: 600;
            transition: all 0.3s ease;
            border: none;
            cursor: pointer;
            font-size: 14px;
        }
        .btn:hover { 
            transform: translateY(-2px);
            box-shadow: 0 10px 30px rgba(0, 212, 255, 0.3);
        }
        .btn-secondary {
            background: transparent;
            color: #00d4ff;
            border: 2px solid rgba(0, 212, 255, 0.3);
        }
        .btn-secondary:hover {
            background: rgba(0, 212, 255, 0.1);
        }
        .links {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid rgba(255, 255, 255, 0.1);
            display: flex;
            flex-wrap: wrap;
            gap: 16px;
        }
        .links a {
            color: #00d4ff;
            text-decoration: none;
            padding: 8px 16px;
            border-radius: 8px;
            background: rgba(0, 212, 255, 0.1);
            transition: all 0.3s ease;
        }
        .links a:hover {
            background: rgba(0, 212, 255, 0.2);
        }
    </style>
    <script>
        function copyUUID(uuid) {
            navigator.clipboard.writeText(uuid).then(function() {
                alert('UUID已复制到剪贴板: ' + uuid);
            }, function(err) {
                console.error('复制失败:', err);
                // 降级方案
                var textArea = document.createElement('textarea');
                textArea.value = uuid;
                document.body.appendChild(textArea);
                textArea.select();
                try {
                    document.execCommand('copy');
                    alert('UUID已复制到剪贴板: ' + uuid);
                } catch (err) {
                    alert('复制失败，请手动复制: ' + uuid);
                }
                document.body.removeChild(textArea);
            });
        }
    </script>
</head>
<body>
    <div class="container">
        <h1>🔍 搜索结果</h1>
        ` + content + `
        <div class="links">
            <a href="/llmaget/set">⚙️ 配置设置</a>
            <a href="/llmaget/refresh">🔄 刷新数据</a>
            <a href="/llmaget/status">📊 查看状态</a>
            <a href="/llmaget/ff_info">📄 查看数据</a>
            <a href="/llmaget/sign_in">✍️ 打卡</a>
        </div>
    </div>
</body>
</html>`
}
