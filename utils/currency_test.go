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
	want := map[string]string{
		"USD": "$",
		"CNY": "¥",
		"JPY": "JP¥",
		"CAD": "CA$",
		"NIO": "C$",
		"MXN": "MX$",
		"RON": "lei",
		"IQD": "د.ع.",
		"BHD": "د.ب.",
		"VES": "Bs.",
		"SYP": "£S",
	}
	for code, symbol := range want {
		if got := GetCurrencySymbol(code); got != symbol {
			t.Errorf("GetCurrencySymbol(%s)=%q want %q", code, got, symbol)
		}
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

func TestExchangeRatesCoverCatalog(t *testing.T) {
	if ExchangeRates["USD"] != 1 {
		t.Fatalf("USD rate = %v want 1", ExchangeRates["USD"])
	}
	for _, code := range []string{"GC", "SC"} {
		if ExchangeRates[code] != ExchangeRates["USD"] {
			t.Errorf("%s rate = %v want USD %v", code, ExchangeRates[code], ExchangeRates["USD"])
		}
		info, ok := LookupCurrency(code)
		if !ok {
			t.Errorf("LookupCurrency(%s) missing", code)
			continue
		}
		if info.Symbol != GetCurrencySymbol("USD") {
			t.Errorf("%s symbol = %q want USD %q", code, info.Symbol, GetCurrencySymbol("USD"))
		}
	}
	for _, c := range ListCurrencies() {
		rate, ok := ExchangeRates[c.Code]
		if !ok {
			t.Errorf("ExchangeRates missing catalog currency %s", c.Code)
			continue
		}
		if rate <= 0 {
			t.Errorf("ExchangeRates[%s]=%v want > 0", c.Code, rate)
		}
	}
}
