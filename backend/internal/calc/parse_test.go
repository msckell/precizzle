package calc

import (
	"errors"
	"strings"
	"testing"
)

func TestParseCanonicalForms(t *testing.T) {
	tests := []struct {
		in   string
		want Result
	}{
		{"12", Result{"12", "12", Exact}},
		{"-12.5", Result{"-12.5", "-12.5", Exact}},
		{"0.1", Result{"0.1", "0.1", Exact}},
		{"1e-3", Result{"0.001", "0.001", Exact}},
		{"2.5E10", Result{"25000000000", "25000000000", Exact}},
		{"1.5E+3", Result{"1500", "1500", Exact}},
		{"1.10", Result{"1.1", "1.1", Exact}},
		{"007", Result{"7", "7", Exact}},
		{"0.50e1", Result{"5", "5", Exact}},
		{"-0", Result{"0", "0", Exact}},
		{"-0.0e5", Result{"0", "0", Exact}},
		{"0e99999", Result{"0", "0", Exact}},
		{"0e99999999999999999999", Result{"0", "0", Exact}},
		{"1/3", Result{"1/3", "0." + strings.Repeat("3", 100), Repeating}},
		{"-2/7", Result{"-2/7", "-0." + strings.Repeat("285714", 16) + "2857", Repeating}},
		{"4/2", Result{"2", "2", Exact}},
		{"-2/4", Result{"-0.5", "-0.5", Exact}},
		{"1/4", Result{"0.25", "0.25", Exact}},
		{"0/5", Result{"0", "0", Exact}},
		{"~1.4142135623730950488", Result{"~1.4142135623730950488", "1.4142135623730950488", Approximate}},
		{"~-3.2e-25", Result{"~-3.2e-25", "-3.2e-25", Approximate}},
		{"~0", Result{"~0", "0", Approximate}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			n, err := Parse(tt.in)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tt.in, err)
			}
			got, err := newResult(n)
			if err != nil {
				t.Fatalf("newResult(Parse(%q)) error = %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseRejectsInvalidSyntax(t *testing.T) {
	tests := []string{
		"", "-", "+1", ".5", "5.", "1,5", "1 000", " 1", "1 ", "0x10", "0b1", "1e", "1e+", "1e1.5", "1.2.3",
		"--1", "1-", "e5", "1/0", "1/00", "-1/-3", "1/-3", "+1/3", "1/3.0", "1.5/2", "1/2/3", "/3", "1/",
		"~1/3", "~~1", "~", "~+1", "~.5", "~ 1", "−1", "١", "Inf", "NaN", "1_000",
	}
	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			if _, err := Parse(in); !errors.Is(err, ErrInvalidNumber) {
				t.Errorf("Parse(%q) error = %v, want %v", in, err, ErrInvalidNumber)
			}
		})
	}
}

func TestParseLimits(t *testing.T) {
	thousandZeros := strings.Repeat("0", 1000)
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"1e39", "1e39", true},
		{"forty nines", strings.Repeat("9", 40), true},
		{"just below 1e40", "9." + strings.Repeat("9", 999) + "e39", true},
		{"negative forty nines", "-" + strings.Repeat("9", 40), true},
		{"1e40", "1e40", false},
		{"-1e40", "-1e40", false},
		{"forty-one digits", "1" + strings.Repeat("0", 40), false},
		{"1e99999", "1e99999", false},
		{"huge exponent", "1e99999999999999999999", false},
		{"1000 fractional digits", "0." + strings.Repeat("0", 999) + "1", true},
		{"1e-1000", "1e-1000", true},
		{"1001 fractional digits", "0." + thousandZeros + "1", false},
		{"1e-1001", "1e-1001", false},
		{"tiny exponent", "1e-99999999999999999999", false},
		{"trailing zeros do not count", "0." + thousandZeros + "0000", true},
		{"terminating fraction at 1000 digits", "1/1" + thousandZeros, true},
		{"terminating fraction at 1001 digits", "1/2" + thousandZeros, false},
		{"repeating denominator below 10^1000", "1/" + strings.Repeat("9", 1000), true},
		{"repeating denominator above 10^1000", "1/1" + strings.Repeat("0", 999) + "1", false},
		{"fraction too large", "1" + strings.Repeat("0", 40) + "/1", false},
		{"fraction just below 1e40", strings.Repeat("9", 41) + "/10", true},
		{"approximate at 1000 digits", "~0." + strings.Repeat("0", 999) + "1", true},
		{"approximate past 1000 digits", "~0." + thousandZeros + "1", false},
		{"approximate too large", "~1e40", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.in)
			if tt.ok && err != nil {
				t.Errorf("Parse error = %v, want nil", err)
			}
			if !tt.ok && !errors.Is(err, ErrOperandOutOfRange) {
				t.Errorf("Parse error = %v, want %v", err, ErrOperandOutOfRange)
			}
		})
	}
}

func TestNumberResult(t *testing.T) {
	tests := []struct {
		operand string
		want    Result
	}{
		{"5.0", Result{"5", "5", Exact}},
		{"-0", Result{"0", "0", Exact}},
		{"-0.50e1", Result{"-5", "-5", Exact}},
		{"4/2", Result{"2", "2", Exact}},
		{"1/4", Result{"0.25", "0.25", Exact}},
		{"-2/6", Result{"-1/3", "-0." + strings.Repeat("3", 100), Repeating}},
		{"~1.50", Result{"~1.5", "1.5", Approximate}},
		{"~-3.2e-25", Result{"~-3.2e-25", "-3.2e-25", Approximate}},
	}
	for _, tt := range tests {
		if got := num(t, tt.operand).Result(); got != tt.want {
			t.Errorf("Parse(%q).Result() = %+v, want %+v", tt.operand, got, tt.want)
		}
	}
}
