package oauth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCredentialsReturnsAValidTokenWithoutRefreshing(t *testing.T) {
	e := setup(t)
	e.seed(t, e.user, "github", "valid", "rt-unused", at(time.Hour))
	e.seed(t, e.user, "google", "never-expires", "", nil)

	gh, err1 := e.svc.Credentials(context.Background(), e.user, "github")
	gg, err2 := e.svc.Credentials(context.Background(), e.user, "google")

	if err1 != nil || err2 != nil || gh.AccessToken != "valid" || gg.AccessToken != "never-expires" {
		t.Errorf("got %q (%v), %q (%v)", gh.AccessToken, err1, gg.AccessToken, err2)
	}
	if n := e.fake.refreshes(); n != 0 {
		t.Errorf("%d refreshes, want 0", n)
	}
}

func TestCredentialsRefreshesAnExpiredTokenOnce(t *testing.T) {
	e := setup(t)
	e.fake.allowRefresh("rt-seed")
	e.seed(t, e.user, "github", "expired", "rt-seed", at(-time.Minute))

	// The provider accepts each refresh token once, so concurrent callers must
	// not refresh in parallel.
	const callers = 5
	got := make([]string, callers)
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, err := e.svc.Credentials(context.Background(), e.user, "github")
			got[i], errs[i] = c.AccessToken, err
		}()
	}
	wg.Wait()

	for i := range callers {
		if errs[i] != nil || got[i] != "at-1" {
			t.Errorf("caller %d: %q, %v; want at-1", i, got[i], errs[i])
		}
	}
	if n := e.fake.refreshes(); n != 1 {
		t.Errorf("%d refreshes, want 1", n)
	}
	r := e.row(t, e.user, "github")
	if string(r.RefreshTokenEnc) != "rt-1" || r.ExpiresAt == nil || time.Until(*r.ExpiresAt) < 59*time.Minute {
		t.Errorf("stored refresh token %q, expires_at %v", r.RefreshTokenEnc, r.ExpiresAt)
	}
}

func TestCredentialsErrors(t *testing.T) {
	e := setup(t)
	e.seed(t, e.user, "github", "expired", "rt-revoked", at(-time.Minute))
	e.seed(t, e.user, "google", "expired", "", at(-time.Minute))

	_, notLinked := e.svc.Credentials(context.Background(), e.user, "spotify")
	_, revoked := e.svc.Credentials(context.Background(), e.user, "github")
	_, noRefresh := e.svc.Credentials(context.Background(), e.user, "google")

	if !errors.Is(notLinked, ErrNotLinked) {
		t.Errorf("unlinked service: %v, want ErrNotLinked", notLinked)
	}
	if !errors.Is(revoked, ErrRelinkRequired) || !errors.Is(noRefresh, ErrRelinkRequired) {
		t.Errorf("revoked: %v, no refresh token: %v; want ErrRelinkRequired", revoked, noRefresh)
	}
	if r := e.row(t, e.user, "github"); r.RefreshTokenEnc != nil {
		t.Errorf("revoked refresh token kept: %q", r.RefreshTokenEnc)
	}
}
