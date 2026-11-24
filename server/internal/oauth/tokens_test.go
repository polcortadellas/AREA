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

func TestRefreshDueRefreshesOnlyTokensExpiringSoon(t *testing.T) {
	e := setup(t)
	other, _ := e.newUser(t, "other@example.com")
	e.fake.allowRefresh("rt-soon")
	e.fake.allowRefresh("rt-later")
	e.seed(t, e.user, "github", "soon", "rt-soon", at(5*time.Minute))
	e.seed(t, other, "github", "later", "rt-later", at(time.Hour))
	e.seed(t, other, "google", "no-expiry", "", nil)

	n, err := e.svc.RefreshDue(context.Background(), time.Now().Add(refreshWindow))

	if err != nil || n != 1 {
		t.Fatalf("refreshed %d, %v; want 1", n, err)
	}
	if string(e.row(t, e.user, "github").AccessTokenEnc) != "at-1" || string(e.row(t, other, "github").AccessTokenEnc) != "later" {
		t.Error("wrong tokens refreshed")
	}
}

func TestRefreshDueSweepsInParallelRefreshEachTokenOnce(t *testing.T) {
	e := setup(t)
	e.fake.allowRefresh("rt-seed")
	e.seed(t, e.user, "github", "soon", "rt-seed", at(time.Minute))

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := e.svc.RefreshDue(context.Background(), time.Now().Add(refreshWindow)); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if n := e.fake.refreshes(); n != 1 {
		t.Errorf("%d refreshes, want 1", n)
	}
}

func TestRefreshDueKeepsGoingAfterAFailure(t *testing.T) {
	e := setup(t)
	other, _ := e.newUser(t, "other@example.com")
	e.fake.allowRefresh("rt-good")
	e.seed(t, e.user, "github", "revoked", "rt-revoked", at(time.Minute))
	e.seed(t, other, "github", "good", "rt-good", at(2*time.Minute))

	n, err := e.svc.RefreshDue(context.Background(), time.Now().Add(refreshWindow))

	if err != nil || n != 1 {
		t.Fatalf("refreshed %d, %v; want 1", n, err)
	}
	if r := e.row(t, e.user, "github"); r.RefreshTokenEnc != nil {
		t.Errorf("revoked refresh token kept: %q", r.RefreshTokenEnc)
	}
}

func TestRefreshDueKeepsTheRefreshTokenWhenTheProviderIsDown(t *testing.T) {
	e := setup(t)
	e.fake.allowRefresh("rt-seed")
	e.fake.setUnavailable(true)
	e.seed(t, e.user, "github", "soon", "rt-seed", at(time.Minute))

	n, err := e.svc.RefreshDue(context.Background(), time.Now().Add(refreshWindow))

	if err != nil || n != 0 {
		t.Fatalf("refreshed %d, %v; want 0", n, err)
	}
	if r := e.row(t, e.user, "github"); string(r.RefreshTokenEnc) != "rt-seed" {
		t.Errorf("refresh token = %q, want rt-seed kept for the next sweep", r.RefreshTokenEnc)
	}

	e.fake.setUnavailable(false)
	if n, _ := e.svc.RefreshDue(context.Background(), time.Now().Add(refreshWindow)); n != 1 {
		t.Errorf("after recovery refreshed %d, want 1", n)
	}
}

func TestRunRefresherSweepsAtStartAndStopsWithContext(t *testing.T) {
	e := setup(t)
	e.fake.allowRefresh("rt-seed")
	e.seed(t, e.user, "github", "soon", "rt-seed", at(time.Minute))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		e.svc.RunRefresher(ctx)
		close(done)
	}()

	deadline := time.Now().Add(5 * time.Second)
	for e.fake.refreshes() == 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("RunRefresher did not stop")
	}
	if n := e.fake.refreshes(); n != 1 {
		t.Errorf("%d refreshes, want 1", n)
	}
}

func TestRefreshDueSkipsServicesWithoutProvider(t *testing.T) {
	e := setup(t)
	e.seed(t, e.user, "spotify", "soon", "rt-x", at(time.Minute))

	n, err := e.svc.RefreshDue(context.Background(), time.Now().Add(refreshWindow))

	if err != nil || n != 0 {
		t.Errorf("refreshed %d, %v; want 0", n, err)
	}
}
