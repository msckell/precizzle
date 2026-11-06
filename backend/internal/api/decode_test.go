package api

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
)

func TestDecodeFields(t *testing.T) {
	allowed := []string{"a", "b"}
	tests := []struct {
		name        string
		contentType string
		body        string
		wantCode    string
		wantField   string
		wantFields  []string
	}{
		{name: "two strings", contentType: "application/json", body: `{"a":"1","b":"2"}`, wantFields: []string{"a", "b"}},
		{name: "charset parameter", contentType: "application/json; charset=utf-8", body: `{"a":"1"}`, wantFields: []string{"a"}},
		{name: "upper-case media type", contentType: "Application/JSON", body: `{}`},
		{name: "surrounding whitespace", contentType: "application/json", body: " \n{\"a\": \"1\"}\n", wantFields: []string{"a"}},
		{name: "null counts as missing", contentType: "application/json", body: `{"a":null,"b":"2"}`, wantFields: []string{"b"}},
		{name: "numbers are kept for the caller", contentType: "application/json", body: `{"a":1}`, wantFields: []string{"a"}},

		{name: "no content type", body: `{}`, wantCode: codeUnsupportedMediaType},
		{name: "text/plain", contentType: "text/plain", body: `{}`, wantCode: codeUnsupportedMediaType},
		{name: "malformed content type", contentType: "application/json; =", body: `{}`, wantCode: codeUnsupportedMediaType},
		{name: "media type wins over invalid JSON", contentType: "text/plain", body: `{`, wantCode: codeUnsupportedMediaType},

		{name: "empty body", contentType: "application/json", body: ``, wantCode: codeInvalidJSON},
		{name: "syntax error", contentType: "application/json", body: `{"a":"1",}`, wantCode: codeInvalidJSON},
		{name: "data after the object", contentType: "application/json", body: `{"a":"1"} {}`, wantCode: codeInvalidJSON},
		{name: "duplicate member", contentType: "application/json", body: `{"a":"1","a":"2"}`, wantCode: codeInvalidJSON},
		{name: "invalid UTF-8", contentType: "application/json", body: "{\"a\":\"\xff\"}", wantCode: codeInvalidJSON},
		{name: "top-level array", contentType: "application/json", body: `["1","2"]`, wantCode: codeInvalidJSON},
		{name: "top-level null", contentType: "application/json", body: `null`, wantCode: codeInvalidJSON},
		{name: "invalid JSON wins over an unknown field", contentType: "application/json", body: `{"c":"1",`, wantCode: codeInvalidJSON},

		{name: "unknown field", contentType: "application/json", body: `{"a":"1","c":"2"}`, wantCode: codeUnknownField, wantField: "c"},
		{name: "names are case-sensitive", contentType: "application/json", body: `{"A":"1"}`, wantCode: codeUnknownField, wantField: "A"},
		{name: "first unknown field in document order", contentType: "application/json", body: `{"z":"1","c":"2"}`, wantCode: codeUnknownField, wantField: "z"},
		{name: "unknown field wins over a JSON number", contentType: "application/json", body: `{"a":1,"c":"2"}`, wantCode: codeUnknownField, wantField: "c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/add", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			fields, p := decodeFields(httptest.NewRecorder(), req, allowed)
			if tt.wantCode != "" {
				if p == nil || p.code != tt.wantCode || p.field != tt.wantField {
					t.Fatalf("problem = %+v, want code %q and field %q", p, tt.wantCode, tt.wantField)
				}
				return
			}
			if p != nil {
				t.Fatalf("problem = %+v, want none", p)
			}
			if len(fields) != len(tt.wantFields) {
				t.Errorf("fields = %v, want %v", fields, tt.wantFields)
			}
			for _, name := range tt.wantFields {
				if _, ok := fields[name]; !ok {
					t.Errorf("field %q is missing from %v", name, fields)
				}
			}
		})
	}
}

func TestDecodeFieldsUnreadableBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/add", iotest.ErrReader(errors.New("connection reset")))
	req.Header.Set("Content-Type", "application/json")
	if _, p := decodeFields(httptest.NewRecorder(), req, []string{"a"}); p == nil || p.code != codeInvalidJSON {
		t.Errorf("problem = %+v, want %s", p, codeInvalidJSON)
	}
}

func TestDecodeFieldsBodyLimit(t *testing.T) {
	object := `{"a":"1","b":"2"}`
	tests := []struct {
		name     string
		size     int
		wantCode string
	}{
		{"exactly 64 KiB", maxBodyBytes, ""},
		{"one byte over 64 KiB", maxBodyBytes + 1, codePayloadTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := "{" + strings.Repeat(" ", tt.size-len(object)) + object[1:]
			req := httptest.NewRequest("POST", "/api/v1/add", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			_, p := decodeFields(httptest.NewRecorder(), req, []string{"a", "b"})
			if tt.wantCode == "" && p != nil {
				t.Errorf("problem = %+v, want none", p)
			}
			if tt.wantCode != "" && (p == nil || p.code != tt.wantCode) {
				t.Errorf("problem = %+v, want %s", p, tt.wantCode)
			}
		})
	}
}
