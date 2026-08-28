package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// game_global_config.gkey 常量。
const (
	GameGlobalConfigKeyControlConfig = "control_config"
	GameGlobalConfigKeyBetConfig     = "bet_config"
)

// GameConfigModel app_game / app_info 的 config_json 结构。
//
// 示例：
//
//	{
//	  "control_config": { "control_max": { "rate": 500, "win": -1 } },
//	  "bet_config": {}
//	}
type GameConfigModel struct {
	ControlConfig *ControlConfig `json:"control_config,omitempty"`
	BetConfig     *BetConfig     `json:"bet_config,omitempty"`
}

// ControlConfig 风控/派彩上限配置。
// 对应 game_global_config 中 gkey=control_config 的 gvalue 结构。
type ControlConfig struct {
	ControlMax *ControlMaxConfig `json:"control_max,omitempty"`
}

// ControlMaxConfig 单局最大派彩限制；rate/win 为 -1 表示不限制。
type ControlMaxConfig struct {
	Rate int64   `json:"rate"` // 单局最大派彩倍数
	Win  float64 `json:"win"`  // 单局最大派彩金额
}

// BetConfig 下注档位配置。
// 对应 game_global_config 中 gkey=bet_config 的 gvalue 结构。
type BetConfig struct {
	MinBet      float64   `json:"minBet,omitempty"`      // 最小下注金额
	MaxBet      float64   `json:"maxBet,omitempty"`      // 最大下注金额
	DefaultBet  float64   `json:"defaultBet,omitempty"`  // 默认下注金额
	MaxWinLimit float64   `json:"maxWinLimit,omitempty"` // 最大中奖金额
	BetOptions  []float64 `json:"betOptions,omitempty"`  // 下注选项
}

// ParseConfigJson 解析 config_json 字符串为 GameConfigModel。
// raw 为空时返回空配置，不报错。
func ParseConfigJson(raw string) (*GameConfigModel, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &GameConfigModel{}, nil
	}
	var cfg GameConfigModel
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("parse config_json: %w", err)
	}
	return &cfg, nil
}

// MarshalConfigJson 将 GameConfigModel 序列化为 JSON 字符串。
// m 为 nil 时返回 "{}"。
func MarshalConfigJson(m *GameConfigModel) (string, error) {
	if m == nil {
		return "{}", nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("marshal config_json: %w", err)
	}
	return string(data), nil
}

// ParseControlConfigJSON 解析 control_config 的 gvalue。
func ParseControlConfigJSON(raw string) (*ControlConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &ControlConfig{}, nil
	}
	var cfg ControlConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("parse control_config: %w", err)
	}
	return &cfg, nil
}

// ParseBetConfigJSON 解析 bet_config 的 gvalue。
func ParseBetConfigJSON(raw string) (*BetConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &BetConfig{}, nil
	}
	var cfg BetConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("parse bet_config: %w", err)
	}
	return &cfg, nil
}

// MarshalControlConfig 序列化 ControlConfig。
func MarshalControlConfig(cfg *ControlConfig) (string, error) {
	if cfg == nil {
		return "{}", nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal control_config: %w", err)
	}
	return string(data), nil
}

// MarshalBetConfig 序列化 BetConfig。
func MarshalBetConfig(cfg *BetConfig) (string, error) {
	if cfg == nil {
		return "{}", nil
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal bet_config: %w", err)
	}
	return string(data), nil
}

// IsRateUnlimited 是否不限制单局最大派彩倍数。
func (c *ControlMaxConfig) IsRateUnlimited() bool {
	return c == nil || c.Rate < 0
}

// IsWinUnlimited 是否不限制单局最大派彩金额。
func (c *ControlMaxConfig) IsWinUnlimited() bool {
	return c == nil || c.Win < 0
}

// GetConfig 解析 AppGame.config_json。
func (a *AppGame) GetConfig() (*GameConfigModel, error) {
	if a == nil {
		return &GameConfigModel{}, nil
	}
	return ParseConfigJson(a.ConfigJson)
}

// SetConfig 将配置写回 AppGame.config_json。
func (a *AppGame) SetConfig(cfg *GameConfigModel) error {
	if a == nil {
		return fmt.Errorf("app game is nil")
	}
	raw, err := MarshalConfigJson(cfg)
	if err != nil {
		return err
	}
	a.ConfigJson = raw
	return nil
}

// GetConfig 解析 AppInfo.config_json。
func (a *AppInfo) GetConfig() (*GameConfigModel, error) {
	if a == nil {
		return &GameConfigModel{}, nil
	}
	return ParseConfigJson(a.ConfigJson)
}

// SetConfig 将配置写回 AppInfo.config_json。
func (a *AppInfo) SetConfig(cfg *GameConfigModel) error {
	if a == nil {
		return fmt.Errorf("app info is nil")
	}
	raw, err := MarshalConfigJson(cfg)
	if err != nil {
		return err
	}
	a.ConfigJson = raw
	return nil
}

// ParseValue 按 gkey 解析 GameGlobalConfig.gvalue。
func (g *GameGlobalConfig) ParseValue() (any, error) {
	if g == nil {
		return nil, fmt.Errorf("game global config is nil")
	}
	switch g.GKey {
	case GameGlobalConfigKeyControlConfig:
		return g.ParseControlConfig()
	case GameGlobalConfigKeyBetConfig:
		return g.ParseBetConfig()
	default:
		return nil, fmt.Errorf("unsupported game global config gkey: %s", g.GKey)
	}
}

// ParseControlConfig 解析 gkey=control_config 的 gvalue。
func (g *GameGlobalConfig) ParseControlConfig() (*ControlConfig, error) {
	if g == nil {
		return &ControlConfig{}, nil
	}
	return ParseControlConfigJSON(g.GValue)
}

// ParseBetConfig 解析 gkey=bet_config 的 gvalue。
func (g *GameGlobalConfig) ParseBetConfig() (*BetConfig, error) {
	if g == nil {
		return &BetConfig{}, nil
	}
	return ParseBetConfigJSON(g.GValue)
}

// SetControlConfig 设置 gkey=control_config 的 gvalue。
func (g *GameGlobalConfig) SetControlConfig(cfg *ControlConfig) error {
	if g == nil {
		return fmt.Errorf("game global config is nil")
	}
	raw, err := MarshalControlConfig(cfg)
	if err != nil {
		return err
	}
	g.GKey = GameGlobalConfigKeyControlConfig
	g.GValue = raw
	return nil
}

// SetBetConfig 设置 gkey=bet_config 的 gvalue。
func (g *GameGlobalConfig) SetBetConfig(cfg *BetConfig) error {
	if g == nil {
		return fmt.Errorf("game global config is nil")
	}
	raw, err := MarshalBetConfig(cfg)
	if err != nil {
		return err
	}
	g.GKey = GameGlobalConfigKeyBetConfig
	g.GValue = raw
	return nil
}
