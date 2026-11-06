package main

import "testing"

func TestListenAddr(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "default", env: map[string]string{}, want: "127.0.0.1:8080"},
		{name: "empty variable", env: map[string]string{"PRECIZZLE_ADDR": ""}, want: "127.0.0.1:8080"},
		{name: "custom port", env: map[string]string{"PRECIZZLE_ADDR": "127.0.0.1:8081"}, want: "127.0.0.1:8081"},
		{name: "all interfaces", env: map[string]string{"PRECIZZLE_ADDR": "0.0.0.0:8080"}, want: "0.0.0.0:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string { return tt.env[key] }
			if got := listenAddr(getenv); got != tt.want {
				t.Errorf("listenAddr = %q, want %q", got, tt.want)
			}
		})
	}
}
