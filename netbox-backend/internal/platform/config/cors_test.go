package config

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestCORSAllowedOrigins(t *testing.T) {
	t.Parallel()

	accepted := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "empty", raw: "", want: []string{}},
		{name: "space and tab only", raw: " \t ", want: []string{}},
		{name: "one HTTPS origin", raw: "https://example.com", want: []string{"https://example.com"}},
		{
			name: "multiple origins preserve order",
			raw:  "\thttps://FIRST.example\t, http://second.example:8080 ",
			want: []string{"https://first.example", "http://second.example:8080"},
		},
		{
			name: "scheme hostname and default ports canonicalize",
			raw:  "HTTP://LOCALHOST:80,HTTPS://Example.COM:443",
			want: []string{"http://localhost", "https://example.com"},
		},
		{
			name: "nondefault ports remain",
			raw:  "http://example.com:8080,https://example.com:8443",
			want: []string{"http://example.com:8080", "https://example.com:8443"},
		},
		{name: "strict IPv4", raw: "http://192.0.2.10", want: []string{"http://192.0.2.10"}},
		{
			name: "IPv6 canonicalizes",
			raw:  "https://[2001:0DB8:0000:0000:0000:0000:0000:0001]",
			want: []string{"https://[2001:db8::1]"},
		},
		{
			name: "IPv6 with port",
			raw:  "http://[2001:db8::2]:8080",
			want: []string{"http://[2001:db8::2]:8080"},
		},
		{name: "ASCII punycode", raw: "https://xn--bcher-kva.example", want: []string{"https://xn--bcher-kva.example"}},
	}

	for _, test := range accepted {
		t.Run("accept "+test.name, func(t *testing.T) {
			got, err := parseCORSAllowedOrigins(test.raw)
			if err != nil {
				t.Fatalf("parseCORSAllowedOrigins() error = %v", err)
			}
			if !equalStrings(got, test.want) {
				t.Fatalf("parseCORSAllowedOrigins() = %#v, want %#v", got, test.want)
			}
		})
	}

	rejected := []struct {
		name string
		raw  string
	}{
		{name: "empty first element", raw: ",https://example.com"},
		{name: "empty middle element", raw: "https://a.example,,https://b.example"},
		{name: "empty final element", raw: "https://example.com,"},
		{name: "wildcard", raw: "*"},
		{name: "partial wildcard", raw: "https://*.example.com"},
		{name: "regular expression", raw: "^https://example\\.com$"},
		{name: "null", raw: "null"},
		{name: "duplicate", raw: "https://example.com,https://example.com"},
		{name: "hostname case duplicate", raw: "https://EXAMPLE.com,https://example.COM"},
		{name: "default port duplicate", raw: "http://example.com:80,http://example.com"},
		{
			name: "IPv6 canonical duplicate",
			raw:  "https://[2001:0db8:0:0:0:0:0:1],https://[2001:db8::1]",
		},
		{name: "missing scheme", raw: "example.com"},
		{name: "missing host", raw: "https://"},
		{name: "unsupported file scheme", raw: "file://example.com"},
		{name: "unsupported websocket scheme", raw: "wss://example.com"},
		{name: "opaque form", raw: "https:example.com"},
		{name: "multiple serialized origins", raw: "https://a.example https://b.example"},
		{name: "userinfo", raw: "https://user@example.com"},
		{name: "password", raw: "https://user:password@example.com"},
		{name: "path", raw: "https://example.com/path"},
		{name: "trailing slash", raw: "https://example.com/"},
		{name: "query", raw: "https://example.com?key=value"},
		{name: "empty query", raw: "https://example.com?"},
		{name: "fragment", raw: "https://example.com#fragment"},
		{name: "empty fragment", raw: "https://example.com#"},
		{name: "port zero", raw: "https://example.com:0"},
		{name: "signed port", raw: "https://example.com:+443"},
		{name: "negative port", raw: "https://example.com:-1"},
		{name: "leading zero port", raw: "https://example.com:0443"},
		{name: "nonnumeric port", raw: "https://example.com:https"},
		{name: "port above range", raw: "https://example.com:65536"},
		{name: "empty port", raw: "https://example.com:"},
		{name: "backslash", raw: "https://example.com\\path"},
		{name: "leading DNS hyphen", raw: "https://-example.com"},
		{name: "trailing DNS hyphen", raw: "https://example-.com"},
		{name: "empty DNS label", raw: "https://example..com"},
		{name: "trailing DNS dot", raw: "https://example.com."},
		{name: "DNS label too long", raw: "https://" + strings.Repeat("a", 64) + ".example"},
		{name: "DNS hostname too long", raw: "https://" + strings.Repeat("a.", 126) + "aa"},
		{name: "all numeric host", raw: "https://12345"},
		{name: "numeric final DNS label", raw: "https://example.123"},
		{name: "legacy compact IPv4", raw: "https://2130706433"},
		{name: "legacy shortened IPv4", raw: "https://127.1"},
		{name: "legacy octal IPv4", raw: "https://0177.0.0.1"},
		{name: "legacy hexadecimal IPv4", raw: "https://0x7f000001"},
		{name: "internal space", raw: "https://exa mple.com"},
		{name: "internal tab", raw: "https://exa\tmple.com"},
		{name: "NUL", raw: "https://example.com\x00"},
		{name: "unit separator", raw: "https://example.com\x1f"},
		{name: "DEL", raw: "https://example.com\x7f"},
		{name: "CR", raw: "https://example.com\r"},
		{name: "LF", raw: "https://example.com\n"},
		{name: "percent escaped host", raw: "https://%65xample.com"},
		{name: "non ASCII host", raw: "https://b\u00fccher.example"},
		{name: "nonstandard IPv4", raw: "https://192.168.001.1"},
		{name: "unbracketed IPv6", raw: "https://2001:db8::1"},
		{name: "IPv6 zone", raw: "https://[fe80::1%25eth0]"},
	}

	safeError := regexp.MustCompile(`^NETBOX_CORS_ALLOWED_ORIGINS\[[0-9]+\]: [a-z0-9 -]+$`)
	for _, test := range rejected {
		t.Run("reject "+test.name, func(t *testing.T) {
			got, err := parseCORSAllowedOrigins(test.raw)
			if err == nil {
				t.Fatalf("parseCORSAllowedOrigins() = %#v, want error", got)
			}
			if got != nil {
				t.Fatalf("parseCORSAllowedOrigins() returned partial origins %#v", got)
			}
			if !safeError.MatchString(err.Error()) {
				t.Fatalf("error %q does not match the non-disclosing format", err)
			}
		})
	}

	t.Run("reject every C0 control except tab before trimming", func(t *testing.T) {
		for control := byte(0); control < 0x20; control++ {
			if control == '\t' {
				continue
			}
			t.Run("byte "+strconv.Itoa(int(control)), func(t *testing.T) {
				got, err := parseCORSAllowedOrigins("https://example.com" + string([]byte{control}))
				if err == nil {
					t.Fatalf("parseCORSAllowedOrigins() = %#v, want error", got)
				}
				if got != nil {
					t.Fatalf("parseCORSAllowedOrigins() returned partial origins %#v", got)
				}
				const want = "NETBOX_CORS_ALLOWED_ORIGINS[0]: control character not allowed"
				if err.Error() != want {
					t.Fatalf("error = %q, want %q", err, want)
				}
			})
		}
	})

	t.Run("later invalid element reports its zero based index without disclosure", func(t *testing.T) {
		const raw = "https://valid.example,https://later-user:LATER-ELEMENT-SENTINEL@example.com"
		got, err := parseCORSAllowedOrigins(raw)
		if err == nil {
			t.Fatalf("parseCORSAllowedOrigins() = %#v, want error", got)
		}
		if got != nil {
			t.Fatalf("parseCORSAllowedOrigins() returned partial origins %#v", got)
		}
		const want = "NETBOX_CORS_ALLOWED_ORIGINS[1]: user information is not allowed"
		if err.Error() != want {
			t.Fatalf("error = %q, want %q", err, want)
		}
		for _, forbidden := range []string{"later-user", "LATER-ELEMENT-SENTINEL", "SENTINEL", "example.com"} {
			if strings.Contains(err.Error(), forbidden) {
				t.Fatalf("error %q disclosed forbidden input fragment %q", err, forbidden)
			}
		}
	})
}

func TestCORSAllowedOriginsDefensiveCopy(t *testing.T) {
	t.Setenv(corsAllowedOriginsEnvironment, "https://first.example,https://second.example")

	runtime, err := LoadHTTPRuntimeFromEnvironment()
	if err != nil {
		t.Fatalf("LoadHTTPRuntimeFromEnvironment() error = %v", err)
	}

	first := runtime.CORSAllowedOrigins()
	first[0] = "https://mutated.example"

	want := []string{"https://first.example", "https://second.example"}
	if got := runtime.CORSAllowedOrigins(); !equalStrings(got, want) {
		t.Fatalf("CORSAllowedOrigins() after caller mutation = %#v, want %#v", got, want)
	}
}

func TestLoadHTTPRuntimeFromEnvironment(t *testing.T) {
	t.Run("unset is empty", func(t *testing.T) {
		unsetEnvironment(t, corsAllowedOriginsEnvironment)

		runtime, err := LoadHTTPRuntimeFromEnvironment()
		if err != nil {
			t.Fatalf("LoadHTTPRuntimeFromEnvironment() error = %v", err)
		}
		if got := runtime.CORSAllowedOrigins(); len(got) != 0 {
			t.Fatalf("CORSAllowedOrigins() = %#v, want empty", got)
		}
	})

	t.Run("explicit empty is empty", func(t *testing.T) {
		t.Setenv(corsAllowedOriginsEnvironment, "")

		runtime, err := LoadHTTPRuntimeFromEnvironment()
		if err != nil {
			t.Fatalf("LoadHTTPRuntimeFromEnvironment() error = %v", err)
		}
		if got := runtime.CORSAllowedOrigins(); len(got) != 0 {
			t.Fatalf("CORSAllowedOrigins() = %#v, want empty", got)
		}
	})

	t.Run("loads canonical origins", func(t *testing.T) {
		t.Setenv(corsAllowedOriginsEnvironment, "HTTPS://CONSOLE.example:443, http://localhost:3000")

		runtime, err := LoadHTTPRuntimeFromEnvironment()
		if err != nil {
			t.Fatalf("LoadHTTPRuntimeFromEnvironment() error = %v", err)
		}
		want := []string{"https://console.example", "http://localhost:3000"}
		if got := runtime.CORSAllowedOrigins(); !equalStrings(got, want) {
			t.Fatalf("CORSAllowedOrigins() = %#v, want %#v", got, want)
		}
	})

	t.Run("invalid value returns no runtime policy and does not disclose input", func(t *testing.T) {
		const secretLikeValue = "https://cors-user:CORS-SENTINEL-PASSWORD@example.com"
		t.Setenv(corsAllowedOriginsEnvironment, secretLikeValue)

		runtime, err := LoadHTTPRuntimeFromEnvironment()
		if err == nil {
			t.Fatal("LoadHTTPRuntimeFromEnvironment() error = nil, want error")
		}
		if got := runtime.CORSAllowedOrigins(); len(got) != 0 {
			t.Fatalf("CORSAllowedOrigins() = %#v after error, want empty", got)
		}
		for _, forbidden := range []string{
			secretLikeValue,
			"cors-user",
			"CORS-SENTINEL-PASSWORD",
			"SENTINEL",
			"PASSWORD",
			"example.com",
		} {
			if strings.Contains(err.Error(), forbidden) {
				t.Fatalf("error %q disclosed forbidden input fragment %q", err, forbidden)
			}
		}
		if got, want := err.Error(), "NETBOX_CORS_ALLOWED_ORIGINS[0]: user information is not allowed"; got != want {
			t.Fatalf("error = %q, want %q", got, want)
		}
	})

	t.Run("missing after populated is fresh and empty", func(t *testing.T) {
		t.Setenv(corsAllowedOriginsEnvironment, "https://populated.example")
		populated, err := LoadHTTPRuntimeFromEnvironment()
		if err != nil {
			t.Fatalf("populated LoadHTTPRuntimeFromEnvironment() error = %v", err)
		}
		if got := populated.CORSAllowedOrigins(); len(got) != 1 {
			t.Fatalf("populated CORSAllowedOrigins() = %#v, want one origin", got)
		}

		if err := os.Unsetenv(corsAllowedOriginsEnvironment); err != nil {
			t.Fatalf("Unsetenv() error = %v", err)
		}
		empty, err := LoadHTTPRuntimeFromEnvironment()
		if err != nil {
			t.Fatalf("empty LoadHTTPRuntimeFromEnvironment() error = %v", err)
		}
		if got := empty.CORSAllowedOrigins(); len(got) != 0 {
			t.Fatalf("empty CORSAllowedOrigins() = %#v, want empty", got)
		}
	})
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func unsetEnvironment(t *testing.T, key string) {
	t.Helper()

	value, present := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Unsetenv(%q) error = %v", key, err)
	}
	t.Cleanup(func() {
		if present {
			if err := os.Setenv(key, value); err != nil {
				t.Errorf("Setenv(%q) during cleanup error = %v", key, err)
			}
			return
		}
		if err := os.Unsetenv(key); err != nil {
			t.Errorf("Unsetenv(%q) during cleanup error = %v", key, err)
		}
	})
}
