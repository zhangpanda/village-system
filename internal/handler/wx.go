package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"village-system/internal/middleware"
	"village-system/internal/model"
)

var (
	wxAppID     = os.Getenv("WX_APPID")
	wxAppSecret = os.Getenv("WX_APPSECRET")
)

type wxLoginReq struct {
	Code      string `json:"code"`
	NickName  string `json:"nick_name"`
	AvatarURL string `json:"avatar_url"`
}

type code2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func (h *Handler) WxLogin(w http.ResponseWriter, r *http.Request) {
	var req wxLoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errJSON(w, 400, "请求格式错误")
		return
	}
	if req.Code == "" {
		errJSON(w, 400, "code不能为空")
		return
	}

	openid, err := code2Session(req.Code)
	if err != nil {
		errJSON(w, 401, err.Error())
		return
	}

	// 查找已有用户或自动注册
	u, err := h.User.GetByOpenID(openid)
	if err != nil {
		name := req.NickName
		if name == "" {
			name = fmt.Sprintf("微信用户%04d", time.Now().UnixMilli()%10000)
		}
		u = &model.User{
			Phone:     "wx_" + openid,
			Name:      name,
			Role:      "villager",
			OpenID:    openid,
			AvatarURL: req.AvatarURL,
		}
		if err := h.User.CreateWxUser(u); err != nil {
			errJSON(w, 500, "创建用户失败")
			return
		}
	}

	token, err := middleware.GenerateToken(u.ID, u.Role)
	if err != nil {
		errJSON(w, 500, "生成令牌失败")
		return
	}
	JSON(w, 200, map[string]any{"token": token, "user": u})
}

// WxGetPhone 通过微信 getPhoneNumber 按钮返回的 code 获取手机号并绑定
func (h *Handler) WxGetPhone(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		errJSON(w, 400, "code不能为空")
		return
	}
	if wxAppID == "" {
		errJSON(w, 400, "未配置微信AppID")
		return
	}

	// 复用带缓存的 access_token
	accessToken, err := getAccessToken()
	if err != nil {
		errJSON(w, 500, "获取access_token失败")
		return
	}

	// 用 code 换手机号
	phoneURL := fmt.Sprintf(
		"https://api.weixin.qq.com/wxa/business/getuserphonenumber?access_token=%s",
		accessToken,
	)
	body, _ := json.Marshal(map[string]string{"code": req.Code})
	phoneResp, err := http.Post(phoneURL, "application/json", io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		errJSON(w, 500, "获取手机号失败")
		return
	}
	defer phoneResp.Body.Close()
	var phoneResult struct {
		ErrCode   int `json:"errcode"`
		PhoneInfo struct {
			PhoneNumber string `json:"phoneNumber"`
		} `json:"phone_info"`
	}
	json.NewDecoder(phoneResp.Body).Decode(&phoneResult)
	if phoneResult.ErrCode != 0 || phoneResult.PhoneInfo.PhoneNumber == "" {
		errJSON(w, 400, "获取手机号失败，请手动填写")
		return
	}

	phone := phoneResult.PhoneInfo.PhoneNumber
	uid := getUID(r)
	// 检查手机号是否已被占用
	existing, _ := h.User.GetByPhoneOnly(phone)
	if existing != nil && existing.ID != uid {
		errJSON(w, 400, "该手机号已被其他用户绑定")
		return
	}
	h.User.UpdatePhone(uid, phone)
	u, _ := h.User.GetByID(uid)
	if u != nil && strings.HasPrefix(u.Account, "wx_") {
		h.User.UpdateAccount(uid, phone)
	}
	JSON(w, 200, map[string]string{"ok": "绑定成功", "phone": phone})
}

func code2Session(code string) (string, error) {
	if wxAppID == "" {
		return "test_user", nil // 测试模式：固定 openid，保证同一用户
	}
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		wxAppID, wxAppSecret, code,
	)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("微信接口请求失败")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result code2SessionResp
	json.Unmarshal(body, &result)
	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信登录失败: %s", result.ErrMsg)
	}
	return result.OpenID, nil
}
