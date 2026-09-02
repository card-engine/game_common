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
			name:      "不限区间，合并档位 92",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "92",
			wantFixed: "92",
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
	if !IsMergedRtpTier(92) || !IsMergedRtpTier(96) {
		t.Fatal("92 and 96 should be merged rtp tiers")
	}
	if IsMergedRtpTier(90) || IsMergedRtpTier(95) || IsMergedRtpTier(97) {
		t.Fatal("90, 95 and 97 should not be merged rtp tiers")
	}

	lower92, upper92, ok := MergedRtpTierWeights(92)
	if !ok {
		t.Fatal("MergedRtpTierWeights(92) should succeed")
	}
	if lower92 != 0.6 || upper92 != 0.4 {
		t.Fatalf("MergedRtpTierWeights(92) = (%v, %v), want (0.6, 0.4)", lower92, upper92)
	}

	lower96, upper96, ok := MergedRtpTierWeights(96)
	if !ok {
		t.Fatal("MergedRtpTierWeights(96) should succeed")
	}
	if lower96 != 0.5 || upper96 != 0.5 {
		t.Fatalf("MergedRtpTierWeights(96) = (%v, %v), want (0.5, 0.5)", lower96, upper96)
	}

	resolved92 := make(map[int]int)
	for i := 0; i < 1000; i++ {
		resolved92[ResolveMergedRtpTier(92)]++
	}
	if resolved92[90] == 0 || resolved92[95] == 0 {
		t.Fatalf("ResolveMergedRtpTier(92) should hit both 90 and 95, got %v", resolved92)
	}

	resolved96 := make(map[int]int)
	for i := 0; i < 1000; i++ {
		resolved96[ResolveMergedRtpTier(96)]++
	}
	if resolved96[95] == 0 || resolved96[97] == 0 {
		t.Fatalf("ResolveMergedRtpTier(96) should hit both 95 and 97, got %v", resolved96)
	}
	if ResolveMergedRtpTier(95) != 95 {
		t.Fatal("non-merged tier should be returned as-is")
	}
}

func TestAllowedRtpTiers(t *testing.T) {
	app := &AppInfo{RtpMin: 50, RtpMax: 97}
	want := []int{50, 65, 75, 85, 90, 92, 95, 96, 97}
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
