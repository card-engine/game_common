package models

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

// GlobalRtpTiers 全局允许的 RTP 档位（升序）
var GlobalRtpTiers = []int{50, 65, 75, 85, 90, 95, 97, 100, 150, 500}

var (
	ErrInvalidRtp       = errors.New("invalid rtp value")
	ErrRtpOutOfRange    = errors.New("rtp out of allowed range")
	ErrNoAllowedRtpTier = errors.New("no allowed rtp tier in range")
)

func (a *AppInfo) HasRtpLimit() bool {
	if a == nil {
		return false
	}
	return a.RtpMin > 0 || a.RtpMax > 0
}

func (a *AppInfo) AllowedRtpTiers() []int {
	if a == nil || !a.HasRtpLimit() {
		return append([]int(nil), GlobalRtpTiers...)
	}
	min, max := a.rtpBounds()
	var tiers []int
	for _, t := range GlobalRtpTiers {
		if t >= min && t <= max {
			tiers = append(tiers, t)
		}
	}
	return tiers
}

func (a *AppInfo) rtpBounds() (min, max int) {
	min = a.RtpMin
	max = a.RtpMax
	if min <= 0 {
		min = GlobalRtpTiers[0]
	}
	if max <= 0 {
		max = GlobalRtpTiers[len(GlobalRtpTiers)-1]
	}
	if min > max {
		min, max = max, min
	}
	return min, max
}

func (a *AppInfo) IsRtpInRange(v int) bool {
	if a == nil || !a.HasRtpLimit() {
		return true
	}
	min, max := a.rtpBounds()
	return v >= min && v <= max
}

// FixRtp 校验商户 RTP 并向下取到最近合法档位。
// 返回：修正后的 RTP 字符串、是否发生过取档、错误。
func (a *AppInfo) FixRtp(rtp string) (fixed string, adjusted bool, err error) {
	v, err := strconv.Atoi(strings.TrimSpace(rtp))
	if err != nil {
		return "", false, ErrInvalidRtp
	}
	if !a.IsRtpInRange(v) {
		return "", false, ErrRtpOutOfRange
	}

	tiers := a.AllowedRtpTiers()
	if len(tiers) == 0 {
		return "", false, ErrNoAllowedRtpTier
	}

	tierSet := make(map[int]struct{}, len(tiers))
	for _, t := range tiers {
		tierSet[t] = struct{}{}
	}
	if _, ok := tierSet[v]; ok {
		return strconv.Itoa(v), false, nil
	}

	idx := sort.SearchInts(tiers, v+1) - 1
	if idx < 0 {
		return "", false, ErrRtpOutOfRange
	}
	fixedV := tiers[idx]
	return strconv.Itoa(fixedV), fixedV != v, nil
}

func (a *AppInfo) ValidateRtp(rtp string) error {
	_, _, err := a.FixRtp(rtp)
	return err
}
