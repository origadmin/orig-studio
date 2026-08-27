package conf

import (
	"fmt"
	"testing"

	"github.com/go-kratos/kratos/v2/config"
	_ "github.com/go-kratos/kratos/v2/encoding/yaml"
)

func TestDebugConfigParsing(t *testing.T) {
	yamlContent := `
servers:
  configs:
    - name: user
      protocol: http
      http:
        addr: "${USER_HTTP_ADDR::8001}"
        timeout:
          seconds: 10
    - name: user
      protocol: grpc
      grpc:
        addr: "${USER_GRPC_ADDR::9001}"
        timeout:
          seconds: 10

security:
  authn:
    configs:
      - type: jwt
        jwt:
          signing_key: "${JWT_SIGNING_KEY:origstudio-secret-key-change-in-production}"
          signing_method: "${JWT_SIGNING_METHOD:HS256}"
          access_token_ttl: "${JWT_ACCESS_TOKEN_TTL:3600s}"
          refresh_token_ttl: "${JWT_REFRESH_TOKEN_TTL:7200s}"
`

	c := config.New(config.WithSource(&debugSource{data: yamlContent}))
	if err := c.Load(); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	var cfg Config
	if err := c.Scan(&cfg); err != nil {
		t.Fatalf("failed to scan: %v", err)
	}

	if cfg.Servers == nil {
		t.Fatal("Servers is nil!")
	}
	if len(cfg.Servers.Configs) == 0 {
		t.Fatal("Servers.Configs is empty!")
	}

	for i, s := range cfg.Servers.Configs {
		t.Logf("Server[%d] name=%s protocol=%s", i, s.GetName(), s.GetProtocol())
		t.Logf("  GetHttp()=%v (nil? %v)", s.GetHttp(), s.GetHttp() == nil)
		t.Logf("  GetGrpc()=%v (nil? %v)", s.GetGrpc(), s.GetGrpc() == nil)
		if s.GetHttp() != nil {
			t.Logf("  HTTP addr=%s timeout=%v", s.GetHttp().GetAddr(), s.GetHttp().GetTimeout())
		}
		if s.GetGrpc() != nil {
			t.Logf("  gRPC addr=%s timeout=%v", s.GetGrpc().GetAddr(), s.GetGrpc().GetTimeout())
		}
	}

	// KEY CHECK: oneof fields must NOT be nil after Scan
	for i, s := range cfg.Servers.Configs {
		switch s.GetProtocol() {
		case "http":
			if s.GetHttp() == nil {
				t.Errorf("Server[%d] protocol=http but GetHttp() is nil — proto oneof NOT filled by Scan!", i)
			}
		case "grpc":
			if s.GetGrpc() == nil {
				t.Errorf("Server[%d] protocol=grpc but GetGrpc() is nil — proto oneof NOT filled by Scan!", i)
			}
		}
	}

	t.Logf("Security.Authn.Configs len=%d: %+v", len(cfg.Security.Authn.Configs), cfg.Security.Authn.Configs)
}

// Debug: also try direct JSON round-trip to see how transportv1.Server oneof serializes
func TestDebugProtoOneofJSON(t *testing.T) {
	type httpCfg struct {
		Addr string `json:"addr"`
	}
	s := &struct {
		Name     string   `json:"name"`
		Protocol string   `json:"protocol"`
		Http     *httpCfg `json:"http"`
	}{
		Name:     "user",
		Protocol: "http",
		Http:     &httpCfg{Addr: ":8001"},
	}
	t.Logf("plain struct JSON-like: %+v", s)
	fmt.Printf("") // keep import
}

type debugSource struct {
	data string
}

func (d *debugSource) Load() ([]*config.KeyValue, error) {
	return []*config.KeyValue{
		{Key: "config.yaml", Value: []byte(d.data), Format: "yaml"},
	}, nil
}

func (d *debugSource) Watch() (config.Watcher, error) {
	return &debugWatcher{}, nil
}

type debugWatcher struct{}

func (w *debugWatcher) Next() ([]*config.KeyValue, error) { return nil, nil }
func (w *debugWatcher) Stop() error                       { return nil }
