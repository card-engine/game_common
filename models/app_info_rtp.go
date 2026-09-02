package models

import (
	"errors"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

// GlobalRtpTiers 全局允许的 RTP 档位（升序，含合并档位）
var GlobalRtpTiers = []int{50, 65, 75, 85, 90, 92, 95, 96, 97, 100, 150, 250, 500}

// MergedRtpTier 合并档位：由相邻两个基础档位按权重随机合并。
// 较低档位概率 P(B) = (A-C)/(B-C)，较高档位概率 P(C) = 1 - P(B)，其中 B < A < C。
type MergedRtpTier struct {
	Tier      int
	LowerTier int
	UpperTier int
}

var mergedRtpTiers = []MergedRtpTier{
	{Tier: 92, LowerTier: 90, UpperTier: 95},
	{Tier: 96, LowerTier: 95, UpperTier: 97},
}

func lookupMergedRtpTier(tier int) (MergedRtpTier, bool) {
	for _, m := range mergedRtpTiers {
		if m.Tier == tier {
			return m, true
		}
	}
	return MergedRtpTier{}, false
}

// IsMergedRtpTier 判断是否为合并档位（无独立 RTP 配置，需解析到基础档位）。
func IsMergedRtpTier(tier int) bool {
	_, ok := lookupMergedRtpTier(tier)
	return ok
}

// MergedRtpTierWeights 返回合并档位对应两个基础档位的选取权重。
func MergedRtpTierWeights(tier int) (lowerWeight, upperWeight float64, ok bool) {
	m, ok := lookupMergedRtpTier(tier)
	if !ok {
		return 0, 0, false
	}
	// B < A < C => (A-C)/(B-C) ∈ (0,1)
	lowerWeight = float64(m.Tier-m.UpperTier) / float64(m.LowerTier-m.UpperTier)
	upperWeight = 1 - lowerWeight
	return lowerWeight, upperWeight, true
}

// ResolveMergedRtpTier 将合并档位随机解析为基础档位；非合并档位原样返回。
func ResolveMergedRtpTier(tier int) int {
	m, ok := lookupMergedRtpTier(tier)
	if !ok {
		return tier
	}
	lowerWeight, _, _ := MergedRtpTierWeights(tier)
	if rand.Float64() < lowerWeight {
		return m.LowerTier
	}
	return m.UpperTier
}

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
