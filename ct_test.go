package ct

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type NextHandlerLoadConfig struct{}

func (NextHandlerLoadConfig) ServeHTTP(w http.ResponseWriter, _ *http.Request) error {
	body := `passwd:
  users:
    - name: core
      ssh_authorized_keys:
        - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIVloJO7jus6ib2Vl4diiTOGBE9gYoVyvXC9UopvSMKt`

	w.Header().Set("Content-Type", "application/x-yaml")
	_, err := w.Write([]byte(body))

	return err
}

func TestTranspiler(t *testing.T) {
	ct := Ct{Strict: true}
	if err := ct.Provision(caddy.Context{}); err != nil {
		t.Fatalf("could not provision module: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/config.yaml", nil)

	expect := `{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.3.0"},"networkd":{},` +
		`"passwd":{"users":[{"name":"core","sshAuthorizedKeys":["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIIVloJO7j` +
		`us6ib2Vl4diiTOGBE9gYoVyvXC9UopvSMKt"]}]},"storage":{},"systemd":{}}`

	if err := ct.ServeHTTP(rec, req, NextHandlerLoadConfig{}); err != nil {
		t.Fatalf("serving HTTP failed: %v", err)
	}

	if rec.Code != 200 {
		t.Fatalf("wrong response code: %d", rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/vnd.coreos.ignition+json" {
		t.Fatalf("content type header is wrong: %s", rec.Header().Get("Content-Type"))
	}

	if rec.Body.String() != expect {
		t.Fatalf("Unexpected body content:\n====== Got ======\n%s\n====== Want ======\n%s", rec.Body.String(), expect)
	}
}

func TestUnmarshalCandyfile(t *testing.T) {
	directive := `ct {
  strict
  mime application/x-yaml
  platform ec2
}`

	d := caddyfile.NewTestDispenser(directive)
	ct := Ct{}
	if err := ct.UnmarshalCaddyfile(d); err != nil {
		t.Fatalf("failed parsing Candyfile %v", err)
	}

	expect := Ct{
		Strict:    true,
		MIMETypes: []string{"application/x-yaml"},
		Platform:  "ec2",
	}

	if !reflect.DeepEqual(ct, expect) {
		t.Fatal("unexpected configuration in module")
	}
}
