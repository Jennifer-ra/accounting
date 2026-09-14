package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	appauth "github.com/Jennifer-ra/accounting/services/go-api/internal/auth"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/config"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/model"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg config.Config
	db  *gorm.DB
}

type LoginInput struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

type LoginResult struct {
	Token string     `json:"token"`
	User  model.User `json:"user"`
	Book  model.Book `json:"book"`
}

type wechatSession struct {
	OpenID     string `json:"openid"`
	UnionID    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func NewAuthService(cfg config.Config, db *gorm.DB) *AuthService {
	return &AuthService{cfg: cfg, db: db}
}

func (s *AuthService) LoginWithWeChat(input LoginInput) (LoginResult, error) {
	input.Code = strings.TrimSpace(input.Code)
	if input.Code == "" {
		return LoginResult{}, errors.New("code is required")
	}
	session, err := s.fetchWeChatSession(input.Code)
	if err != nil {
		return LoginResult{}, err
	}

	var user model.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("open_id = ?", session.OpenID).First(&user).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			user = model.User{OpenID: session.OpenID, UnionID: session.UnionID, Nickname: input.Nickname, AvatarURL: input.AvatarURL}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		} else {
			updates := map[string]interface{}{}
			if session.UnionID != "" && user.UnionID == "" {
				updates["union_id"] = session.UnionID
			}
			if input.Nickname != "" {
				updates["nickname"] = input.Nickname
			}
			if input.AvatarURL != "" {
				updates["avatar_url"] = input.AvatarURL
			}
			if len(updates) > 0 {
				if err := tx.Model(&user).Updates(updates).Error; err != nil {
					return err
				}
				if err := tx.First(&user, user.ID).Error; err != nil {
					return err
				}
			}
		}
		_, err := ensureDefaultBook(tx, user.ID)
		return err
	})
	if err != nil {
		return LoginResult{}, err
	}

	book, err := ensureDefaultBook(s.db, user.ID)
	if err != nil {
		return LoginResult{}, err
	}
	token, err := appauth.Sign(s.cfg.SessionSecret, user.ID, time.Duration(s.cfg.SessionTTLHours)*time.Hour)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: token, User: user, Book: book}, nil
}

func (s *AuthService) UserFromToken(token string) (model.User, error) {
	claims, err := appauth.Parse(s.cfg.SessionSecret, token)
	if err != nil {
		return model.User{}, err
	}
	var user model.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *AuthService) fetchWeChatSession(code string) (wechatSession, error) {
	if s.cfg.WeChatMockLogin {
		return wechatSession{OpenID: "mock_developer"}, nil
	}
	if s.cfg.MiniProgramAppID == "" || s.cfg.MiniProgramSecret == "" {
		return wechatSession{}, errors.New("wechat appid and secret are required")
	}
	endpoint := "https://api.weixin.qq.com/sns/jscode2session"
	query := url.Values{}
	query.Set("appid", s.cfg.MiniProgramAppID)
	query.Set("secret", s.cfg.MiniProgramSecret)
	query.Set("js_code", code)
	query.Set("grant_type", "authorization_code")
	resp, err := http.Get(endpoint + "?" + query.Encode())
	if err != nil {
		return wechatSession{}, err
	}
	defer resp.Body.Close()
	var session wechatSession
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return wechatSession{}, err
	}
	if session.ErrCode != 0 {
		return wechatSession{}, fmt.Errorf("wechat login failed: %d %s", session.ErrCode, session.ErrMsg)
	}
	if session.OpenID == "" {
		return wechatSession{}, errors.New("wechat openid is empty")
	}
	return session, nil
}
