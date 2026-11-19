package mail

import (
	"strings"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	msg, err := build("AREA <no-reply@area.local>", "user@example.com", "Hello", "line 1\nline 2", time.Unix(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	s := string(msg)
	for _, want := range []string{"To: user@example.com\r\n", "Subject: Hello\r\n", "\r\n\r\nline 1\r\nline 2"} {
		if !strings.Contains(s, want) {
			t.Errorf("message missing %q:\n%s", want, s)
		}
	}
}

func TestBuildRejectsHeaderInjection(t *testing.T) {
	if _, err := build("a@b.c", "victim@example.com\r\nBcc: x@evil.com", "s", "b", time.Now()); err == nil {
		t.Fatal("expected error")
	}
}
