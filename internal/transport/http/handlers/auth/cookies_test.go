// Cookie-scope invariant tests for the auth handler. The refresh
// cookie must ride at Path=/ so Next.js middleware (web/src/proxy.ts)
// can see it on every inbound dashboard request and drive SSR silent
// refresh. Narrowing the path would silently break that flow — this
// test pins the contract.
package auth

import (
	"net/http"
	"testing"
)

// TestAuthCookies_AllEmitAtPathRoot — buildAuthCookies must emit
// access_token AND refresh_token AND csrf_token with Path=/. The
// SSR silent-refresh path in web/src/proxy.ts depends on every
// session cookie being attached to every same-origin request.
func TestAuthCookies_AllEmitAtPathRoot(t *testing.T) {
	cookies := buildAuthCookies("access-v", "refresh-v", "csrf-v", "")
	assertAllPathRoot(t, cookies, "buildAuthCookies")
}

// TestClearAuthCookies_AllEmitAtPathRoot — same invariant on the
// clear path. RFC 6265 §4.1.2: a clear Set-Cookie must match the
// set Set-Cookie's attributes (Path in particular) for the browser
// to drop the entry reliably.
func TestClearAuthCookies_AllEmitAtPathRoot(t *testing.T) {
	cookies := buildClearAuthCookies("")
	assertAllPathRoot(t, cookies, "buildClearAuthCookies")
}

// TestRefreshCookiePath_IsRoot — constant-level assertion. If this
// changes, SSR silent refresh silently breaks because the browser
// stops sending refresh_token to the Next.js middleware boundary.
func TestRefreshCookiePath_IsRoot(t *testing.T) {
	if refreshCookiePath != "/" {
		t.Fatalf("refreshCookiePath must be \"/\" for SSR silent refresh; got %q", refreshCookiePath)
	}
}

func assertAllPathRoot(t *testing.T, cookies []http.Cookie, caller string) {
	t.Helper()
	want := map[string]bool{
		cookieNameAccess:  false,
		cookieNameRefresh: false,
		cookieNameCSRF:    false,
	}
	for _, c := range cookies {
		if _, tracked := want[c.Name]; !tracked {
			continue
		}
		want[c.Name] = true
		if c.Path != "/" {
			t.Errorf("%s: cookie %q has Path=%q, want \"/\"", caller, c.Name, c.Path)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("%s: missing cookie %q in output (expected 3 session cookies)", caller, name)
		}
	}
}
