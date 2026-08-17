// Package config owns validated, immutable runtime configuration that is not
// part of the generated YAML configuration surface.
package config

import (
	"fmt"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const corsAllowedOriginsEnvironment = "NETBOX_CORS_ALLOWED_ORIGINS"

// HTTPRuntime is the immutable, process-local HTTP security configuration.
// Construct it with LoadHTTPRuntimeFromEnvironment.
type HTTPRuntime struct {
	corsAllowedOrigins []string
}

// LoadHTTPRuntimeFromEnvironment validates the HTTP-only environment boundary.
// Invalid values return an empty runtime value and a non-disclosing error.
func LoadHTTPRuntimeFromEnvironment() (HTTPRuntime, error) {
	origins, err := parseCORSAllowedOrigins(os.Getenv(corsAllowedOriginsEnvironment))
	if err != nil {
		return HTTPRuntime{}, err
	}

	return HTTPRuntime{corsAllowedOrigins: cloneStrings(origins)}, nil
}

// CORSAllowedOrigins returns a defensive copy of the canonical exact-origin
// allowlist.
func (r HTTPRuntime) CORSAllowedOrigins() []string {
	return cloneStrings(r.corsAllowedOrigins)
}

func parseCORSAllowedOrigins(raw string) ([]string, error) {
	if index, ok := prohibitedControlIndex(raw); ok {
		return nil, corsConfigurationError(index, "control character not allowed")
	}

	raw = trimSpaceAndTab(raw)
	if raw == "" {
		return []string{}, nil
	}

	elements := strings.Split(raw, ",")
	origins := make([]string, 0, len(elements))
	seen := make(map[string]struct{}, len(elements))
	for index, element := range elements {
		origin := trimSpaceAndTab(element)
		if origin == "" {
			return nil, corsConfigurationError(index, "empty origin not allowed")
		}

		canonical, reason := canonicalOrigin(origin)
		if reason != "" {
			return nil, corsConfigurationError(index, reason)
		}
		if _, duplicate := seen[canonical]; duplicate {
			return nil, corsConfigurationError(index, "duplicate origin not allowed")
		}

		seen[canonical] = struct{}{}
		origins = append(origins, canonical)
	}

	return origins, nil
}

func canonicalOrigin(origin string) (string, string) {
	if strings.ContainsAny(origin, " \t") {
		return "", "internal whitespace not allowed"
	}
	if strings.Contains(origin, "*") {
		return "", "wildcard origin not allowed"
	}
	if strings.EqualFold(origin, "null") {
		return "", "null origin not allowed"
	}
	if strings.Contains(origin, "\\") {
		return "", "backslash not allowed"
	}
	if strings.Contains(origin, "?") {
		return "", "query not allowed"
	}
	if strings.Contains(origin, "#") {
		return "", "fragment not allowed"
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return "", "origin syntax is invalid"
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", "scheme must be http or https"
	}
	if parsed.Opaque != "" {
		return "", "opaque origin not allowed"
	}
	if parsed.User != nil {
		return "", "user information is not allowed"
	}
	if parsed.Path != "" || parsed.RawPath != "" {
		return "", "path not allowed"
	}
	if parsed.RawQuery != "" || parsed.ForceQuery {
		return "", "query not allowed"
	}
	if parsed.Fragment != "" || parsed.RawFragment != "" {
		return "", "fragment not allowed"
	}

	delimiter := strings.Index(origin, "://")
	if delimiter <= 0 || !strings.EqualFold(origin[:delimiter], scheme) {
		return "", "origin syntax is invalid"
	}
	authority := origin[delimiter+3:]
	if authority == "" || parsed.Host == "" {
		return "", "host is required"
	}
	if strings.Contains(authority, "%") {
		return "", "escaped host not allowed"
	}

	host, port, hasPort, bracketed, ok := splitOriginAuthority(authority)
	if !ok {
		return "", "host or port is invalid"
	}
	canonicalHost, ok := canonicalOriginHost(host, bracketed)
	if !ok {
		return "", "host is invalid"
	}

	canonicalPort := ""
	if hasPort {
		value, valid := canonicalPortNumber(port)
		if !valid {
			return "", "port is invalid"
		}
		defaultPort := (scheme == "http" && value == 80) || (scheme == "https" && value == 443)
		if !defaultPort {
			canonicalPort = ":" + port
		}
	}

	return scheme + "://" + canonicalHost + canonicalPort, ""
}

func splitOriginAuthority(authority string) (host, port string, hasPort, bracketed, ok bool) {
	if strings.HasPrefix(authority, "[") {
		closing := strings.IndexByte(authority, ']')
		if closing <= 1 {
			return "", "", false, false, false
		}
		host = authority[1:closing]
		remainder := authority[closing+1:]
		if remainder == "" {
			return host, "", false, true, true
		}
		if !strings.HasPrefix(remainder, ":") || len(remainder) == 1 || strings.Contains(remainder[1:], ":") {
			return "", "", false, false, false
		}
		return host, remainder[1:], true, true, true
	}

	if strings.ContainsAny(authority, "[]") {
		return "", "", false, false, false
	}
	switch strings.Count(authority, ":") {
	case 0:
		return authority, "", false, false, authority != ""
	case 1:
		host, port, _ = strings.Cut(authority, ":")
		if host == "" || port == "" {
			return "", "", false, false, false
		}
		return host, port, true, false, true
	default:
		return "", "", false, false, false
	}
}

func canonicalOriginHost(host string, bracketed bool) (string, bool) {
	address, err := netip.ParseAddr(host)
	if err == nil {
		if bracketed != address.Is6() {
			return "", false
		}
		if address.Is6() {
			return "[" + address.String() + "]", true
		}
		return address.String(), true
	}
	if bracketed || looksLikeLegacyIPAddress(host) {
		return "", false
	}

	host = strings.ToLower(host)
	if !validDNSHostname(host) {
		return "", false
	}
	return host, true
}

func canonicalPortNumber(port string) (uint64, bool) {
	if port == "" || (len(port) > 1 && port[0] == '0') {
		return 0, false
	}
	for i := range len(port) {
		if port[i] < '0' || port[i] > '9' {
			return 0, false
		}
	}
	value, err := strconv.ParseUint(port, 10, 16)
	if err != nil || value == 0 {
		return 0, false
	}
	return value, true
}

func validDNSHostname(host string) bool {
	if host == "" || len(host) > 253 || strings.HasSuffix(host, ".") {
		return false
	}
	for i := range len(host) {
		if host[i] >= 0x80 {
			return false
		}
	}

	labels := strings.Split(host, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := range len(label) {
			character := label[i]
			if (character >= 'a' && character <= 'z') ||
				(character >= '0' && character <= '9') || character == '-' {
				continue
			}
			return false
		}
	}

	for i := range len(labels[len(labels)-1]) {
		character := labels[len(labels)-1][i]
		if character >= 'a' && character <= 'z' {
			return true
		}
	}
	return false
}

func looksLikeLegacyIPAddress(host string) bool {
	parts := strings.Split(strings.ToLower(host), ".")
	for _, part := range parts {
		if part == "" {
			return false
		}
		if allDecimalDigits(part) {
			continue
		}
		if strings.HasPrefix(part, "0x") && len(part) > 2 && allHexadecimalDigits(part[2:]) {
			continue
		}
		return false
	}
	return true
}

func allDecimalDigits(value string) bool {
	for i := range len(value) {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return value != ""
}

func allHexadecimalDigits(value string) bool {
	for i := range len(value) {
		character := value[i]
		if (character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') {
			continue
		}
		return false
	}
	return value != ""
}

func prohibitedControlIndex(raw string) (int, bool) {
	for i := range len(raw) {
		if (raw[i] < 0x20 && raw[i] != '\t') || raw[i] == 0x7f {
			return strings.Count(raw[:i], ","), true
		}
	}
	return 0, false
}

func trimSpaceAndTab(value string) string {
	return strings.Trim(value, " \t")
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func corsConfigurationError(index int, reason string) error {
	return fmt.Errorf("%s[%d]: %s", corsAllowedOriginsEnvironment, index, reason)
}
