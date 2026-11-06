package calc

import (
	"errors"
	"strings"
	"testing"
)

func TestPercentage(t *testing.T) {
	tests := []struct {
		in   string
		want Result
	}{
		{"15", Result{"0.15", "0.15", Exact}},
		{"-15", Result{"-0.15", "-0.15", Exact}},
		{"0", Result{"0", "0", Exact}},
		{"1/3", Result{"1/300", "0.00" + strings.Repeat("3", 98), Repeating}},
		{"~15", Result{"~0.15", "0.15", Approximate}},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := Percentage(num(t, tt.in))
			if err != nil {
				t.Fatalf("Percentage error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Percentage = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestApplyPercentage(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		op      PercentOperator
		percent string
		want    Result
		wantErr error
	}{
		{name: "200 + 15%", base: "200", op: PercentAdd, percent: "15", want: Result{"230", "230", Exact}},
		{name: "200 - 15%", base: "200", op: PercentSubtract, percent: "15", want: Result{"170", "170", Exact}},
		{name: "200 × 15%", base: "200", op: PercentMultiply, percent: "15", want: Result{"30", "30", Exact}},
		{name: "200 ÷ 15%", base: "200", op: PercentDivide, percent: "15", want: Result{"4000/3", "1333." + strings.Repeat("3", 100), Repeating}},
		{name: "200 ÷ 0%", base: "200", op: PercentDivide, percent: "0", wantErr: ErrDivisionByZero},
		{name: "0 ÷ 0%", base: "0", op: PercentDivide, percent: "0", wantErr: ErrDivisionByZero},
		{name: "negative base", base: "-200", op: PercentAdd, percent: "15", want: Result{"-230", "-230", Exact}},
		{name: "negative percent", base: "200", op: PercentAdd, percent: "-15", want: Result{"170", "170", Exact}},
		{name: "100% off", base: "200", op: PercentSubtract, percent: "100", want: Result{"0", "0", Exact}},
		{name: "fraction base", base: "1/3", op: PercentMultiply, percent: "300", want: Result{"1", "1", Exact}},
		{name: "approximate base", base: "~200", op: PercentAdd, percent: "15", want: Result{"~230", "230", Approximate}},
		{name: "approximate percent", base: "200", op: PercentDivide, percent: "~15", want: Result{"~1333." + strings.Repeat("3", 20), "1333." + strings.Repeat("3", 20), Approximate}},
		{name: "too large", base: "9e39", op: PercentAdd, percent: "20", wantErr: ErrResultTooLarge},
		{name: "too precise", base: "1e-1000", op: PercentMultiply, percent: "1", wantErr: ErrResultTooPrecise},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyPercentage(num(t, tt.base), tt.op, num(t, tt.percent))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ApplyPercentage error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ApplyPercentage = %+v, want %+v", got, tt.want)
			}
		})
	}
}
