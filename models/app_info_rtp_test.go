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
			name:      "不限区间，向下取档",
			app:       &AppInfo{RtpMin: 0, RtpMax: 0},
			rtp:       "96",
			wantFixed: "95",
			wantAdj:   true,
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
			name:      "区间 50-97，取档",
			app:       &AppInfo{RtpMin: 50, RtpMax: 97},
			rtp:       "96",
			wantFixed: "95",
			wantAdj:   true,
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

func TestAllowedRtpTiers(t *testing.T) {
	app := &AppInfo{RtpMin: 50, RtpMax: 97}
	want := []int{50, 65, 75, 85, 90, 95, 97}
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
