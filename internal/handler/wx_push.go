package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// ==================== 微信模板消息推送 ====================

var (
	wxAccessToken     string
	wxTokenExpireAt   time.Time
	wxTokenMu         sync.Mutex
)

// getAccessToken 获取微信 access_token（带缓存）
func getAccessToken() (string, error) {
	wxTokenMu.Lock()
	defer wxTokenMu.Unlock()

	if wxAccessToken != "" && time.Now().Before(wxTokenExpireAt) {
		return wxAccessToken, nil
	}

	appid := os.Getenv("WX_APPID")
	secret := os.Getenv("WX_APPSECRET")
	if appid == "" || secret == "" {
		return "", fmt.Errorf("未配置 WX_APPID/WX_APPSECRET")
	}

	resp, err := http.Get(fmt.Sprintf(
		"https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		appid, secret,
	))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信错误: %d %s", result.ErrCode, result.ErrMsg)
	}

	wxAccessToken = result.AccessToken
	wxTokenExpireAt = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second) // 提前5分钟过期
	return wxAccessToken, nil
}

// SendWxTemplateMsg 发送微信订阅消息
// 需要用户在小程序内订阅对应模板
func (h *Handler) SendWxTemplateMsg(openID, templateID string, data map[string]interface{}, page string) error {
	token, err := getAccessToken()
	if err != nil {
		log.Printf("获取 access_token 失败: %v", err)
		return err
	}

	body := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"data":        data,
	}
	if page != "" {
		body["page"] = page
	}

	jsonBody, _ := json.Marshal(body)
	resp, err := http.Post(
		"https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token="+token,
		"application/json", bytes.NewReader(jsonBody),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	if result.ErrCode != 0 {
		log.Printf("微信模板消息发送失败: %d %s (openid=%s)", result.ErrCode, result.ErrMsg, openID)
		return fmt.Errorf("%d: %s", result.ErrCode, result.ErrMsg)
	}
	return nil
}

// NotifyViaWechat 尝试通过微信推送通知（如果用户绑定了 openid）
// 在站内通知的基础上额外推送微信消息
func (h *Handler) NotifyViaWechat(userID int64, title, content, refType string, refID int64) {
	u, err := h.User.GetByID(userID)
	if err != nil || u.OpenID == "" {
		return // 用户未绑定微信，跳过
	}

	templateID := os.Getenv("WX_TPL_NOTIFY")
	if templateID == "" {
		return // 未配置模板ID，跳过
	}

	data := map[string]interface{}{
		"thing1":  map[string]string{"value": truncate(title, 20)},
		"thing2":  map[string]string{"value": truncate(content, 20)},
		"time3":   map[string]string{"value": time.Now().Format("2006-01-02 15:04")},
	}

	page := ""
	if refType != "" && refID > 0 {
		page = fmt.Sprintf("pages/%s-detail/%s-detail?id=%d", refType, refType, refID)
	}

	go h.SendWxTemplateMsg(u.OpenID, templateID, data, page)
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "..."
}
