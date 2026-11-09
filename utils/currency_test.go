package utils

import (
	"strings"
	"testing"
)

func TestGetCurrencySymbol(t *testing.T) {
	if got := GetCurrencySymbol("mmk"); got != "K" {
		t.Fatalf("GetCurrencySymbol(mmk)=%q want K", got)
	}
	if got := GetCurrencySymbol("XYZ"); got != "" {
		t.Fatalf("GetCurrencySymbol(XYZ)=%q want empty", got)
	}
}

func TestCurrencySymbolAndNormalize(t *testing.T) {
	if got := CurrencySymbol(""); got != "$" {
		t.Fatalf("CurrencySymbol(\"\")=%q", got)
	}
	if got := NormalizeCurrency(""); got != "USD" {
		t.Fatalf("NormalizeCurrency(\"\")=%q", got)
	}
	info, ok := LookupCurrency("USD")
	if !ok || info.Name != "美元" {
		t.Fatalf("LookupCurrency(USD)=%+v ok=%v", info, ok)
	}
	for _, c := range ListCurrencies() {
		if strings.TrimSpace(c.Name) == "" {
			t.Fatalf("%s missing name", c.Code)
		}
	}
}
