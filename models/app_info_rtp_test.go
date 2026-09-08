package models

import (
	"errors"
	"testing"
)

func TestFixRtp(t *testing.T) {
	tests := []struct {
		name      string
		app       *AppInfo
		rtp       string
		wantFixed string
		wantAdj   bool
		wantErr   error
	}{
		{
			name:      "不限区间，合法档位",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "95",
			wantFixed: "95",
			wantAdj:   false,
		},
		{
			name:      "不限区间，合并档位 80",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "80",
			wantFixed: "80",
			wantAdj:   false,
		},
		{
			name:      "不限区间，合并档位 88",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "88",
			wantFixed: "88",
			wantAdj:   false,
		},
		{
			name:      "不限区间，合并档位 92",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "92",
			wantFixed: "92",
			wantAdj:   false,
		},
		{
			name:      "不限区间，合并档位 93",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "93",
			wantFixed: "93",
			wantAdj:   false,
		},
		{
			name:      "不限区间，合并档位 96",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "96",
			wantFixed: "96",
			wantAdj:   false,
		},
		{
			name:      "区间 50-97，合法",
			app:       &AppInfo{RtpMin: 50, RtpMax: 97},
			rtp:       "90",
			wantFixed: "90",
			wantAdj:   false,
		},
		{
			name:    "区间 50-97，超上限",
			app:     &AppInfo{RtpMin: 50, RtpMax: 97},
			rtp:     "100",
			wantErr: ErrRtpOutOfRange,
		},
		{
			name:    "区间 50-97，超下限",
			app:     &AppInfo{RtpMin: 50, RtpMax: 97},
			rtp:     "40",
			wantErr: ErrRtpOutOfRange,
		},
		{
			name:      "区间 50-97，合并档位 96",
			app:       &AppInfo{RtpMin: 50, RtpMax: 97},
			rtp:       "96",
			wantFixed: "96",
			wantAdj:   false,
		},
		{
			name:      "区间 50-500，高档位",
			app:       &AppInfo{RtpMin: 50, RtpMax: 500},
			rtp:       "500",
			wantFixed: "500",
			wantAdj:   false,
		},
		{
			name:    "非法字符串",
			app:     &AppInfo{},
			rtp:     "abc",
			wantErr: ErrInvalidRtp,
		},
		{
			name:      "nil AppInfo",
			app:       nil,
			rtp:       "95",
			wantFixed: "95",
			wantAdj:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixed, adjusted, err := tt.app.FixRtp(tt.rtp)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("FixRtp() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FixRtp() unexpected err = %v", err)
			}
			if fixed != tt.wantFixed {
				t.Errorf("FixRtp() fixed = %q, want %q", fixed, tt.wantFixed)
			}
			if adjusted != tt.wantAdj {
				t.Errorf("FixRtp() adjusted = %v, want %v", adjusted, tt.wantAdj)
			}
		})
	}
}

func TestMergedRtpTier(t *testing.T) {
	for _, tier := range []int{80, 88, 92, 93, 96} {
		if !IsMergedRtpTier(tier) {
			t.Fatalf("%d should be a merged rtp tier", tier)
		}
		if IsBaseRtpTier(tier) {
			t.Fatalf("%d should not be a base rtp tier", tier)
		}
	}
	for _, tier := range []int{75, 85, 90, 95, 97} {
		if IsMergedRtpTier(tier) {
			t.Fatalf("%d should not be a merged rtp tier", tier)
		}
		if !IsBaseRtpTier(tier) {
			t.Fatalf("%d should be a base rtp tier", tier)
		}
	}

	cases := []struct {
		tier        int
		lowerWeight float64
		upperWeight float64
		lowerTier   int
		upperTier   int
	}{
		{80, 0.5, 0.5, 75, 85},
		{88, 0.4, 0.6, 85, 90},
		{92, 0.6, 0.4, 90, 95},
		{93, 0.4, 0.6, 90, 95},
		{96, 0.5, 0.5, 95, 97},
	}
	for _, c := range cases {
		m, ok := lookupMergedRtpTier(c.tier)
		if !ok {
			t.Fatalf("lookupMergedRtpTier(%d) should succeed", c.tier)
		}
		if m.LowerTier != c.lowerTier || m.UpperTier != c.upperTier {
			t.Fatalf("lookupMergedRtpTier(%d) neighbors = (%d, %d), want (%d, %d)",
				c.tier, m.LowerTier, m.UpperTier, c.lowerTier, c.upperTier)
		}
		if !IsBaseRtpTier(m.LowerTier) || !IsBaseRtpTier(m.UpperTier) {
			t.Fatalf("merged tier %d neighbors must be base tiers, got (%d, %d)",
				c.tier, m.LowerTier, m.UpperTier)
		}

		lower, upper, ok := MergedRtpTierWeights(c.tier)
		if !ok {
			t.Fatalf("MergedRtpTierWeights(%d) should succeed", c.tier)
		}
		if lower != c.lowerWeight || upper != c.upperWeight {
			t.Fatalf("MergedRtpTierWeights(%d) = (%v, %v), want (%v, %v)",
				c.tier, lower, upper, c.lowerWeight, c.upperWeight)
		}

		resolved := make(map[int]int)
		for i := 0; i < 1000; i++ {
			resolved[ResolveMergedRtpTier(c.tier)]++
		}
		if resolved[c.lowerTier] == 0 || resolved[c.upperTier] == 0 {
			t.Fatalf("ResolveMergedRtpTier(%d) should hit both %d and %d, got %v",
				c.tier, c.lowerTier, c.upperTier, resolved)
		}
	}
	if ResolveMergedRtpTier(95) != 95 {
		t.Fatal("non-merged tier should be returned as-is")
	}
}

func TestAllowedRtpTiers(t *testing.T) {
	app := &AppInfo{RtpMin: 50, RtpMax: 97}
	want := []int{50, 65, 75, 80, 85, 88, 90, 92, 93, 95, 96, 97}
	got := app.AllowedRtpTiers()
	if len(got) != len(want) {
		t.Fatalf("AllowedRtpTiers() len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AllowedRtpTiers()[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestHasRtpLimit(t *testing.T) {
	if (&AppInfo{}).HasRtpLimit() {
		t.Error("0/0 should not have limit")
	}
	if !(&AppInfo{RtpMin: 50}).HasRtpLimit() {
		t.Error("RtpMin > 0 should have limit")
	}
	var nilApp *AppInfo
	if nilApp.HasRtpLimit() {
		t.Error("nil should not have limit")
	}
}
