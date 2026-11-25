package node

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeSource struct {
	calls []string
	err   error
}

func (f *fakeSource) Credentials(_ context.Context, userID uuid.UUID, service string) (Credentials, error) {
	f.calls = append(f.calls, service)
	if f.err != nil {
		return Credentials{}, f.err
	}
	return Credentials{AccessToken: service + "-token-" + userID.String()}, nil
}

func TestCredentialsFor(t *testing.T) {
	r := NewRegistry()
	r.RegisterService(Service{Name: "linked", Description: "needs an account", OAuth2: true})
	r.RegisterService(Service{Name: "open", Description: "no account"})
	linked := fakeReaction{meta: Metadata{Service: "linked", Name: "act"}}
	open := fakeReaction{meta: Metadata{Service: "open", Name: "act"}}
	r.RegisterReaction(linked)
	r.RegisterReaction(open)
	user := uuid.New()
	ctx := context.Background()

	src := &fakeSource{}
	creds, err := r.CredentialsFor(ctx, src, user, linked)
	if err != nil || creds.AccessToken != "linked-token-"+user.String() {
		t.Errorf("OAuth2 service: got %+v, %v", creds, err)
	}
	if creds, err := r.CredentialsFor(ctx, src, user, open); err != nil || creds.AccessToken != "" {
		t.Errorf("service without OAuth2: got %+v, %v", creds, err)
	}
	if len(src.calls) != 1 {
		t.Errorf("source asked %d times, want only for the OAuth2 service", len(src.calls))
	}

	if creds, err := r.CredentialsFor(ctx, nil, user, linked); err != nil || creds.AccessToken != "" {
		t.Errorf("nil source: got %+v, %v", creds, err)
	}

	notLinked := errors.New("service not linked")
	_, err = r.CredentialsFor(ctx, &fakeSource{err: notLinked}, user, linked)
	if !errors.Is(err, notLinked) || err.Error() != "linked: service not linked" {
		t.Errorf("got %v, want the source error with the service name", err)
	}
}
