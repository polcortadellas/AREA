package mail

import (
	"net"
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

func TestSendTimesOutOnHungServer(t *testing.T) {
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			<-t.Context().Done()
			_ = conn.Close()
		}
	}()

	s := SMTP{Addr: ln.Addr().String(), From: "a@b.c", Timeout: 200 * time.Millisecond}
	start := time.Now()
	if err := s.Send(t.Context(), "user@example.com", "s", "b"); err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("Send blocked for %v", elapsed)
	}
}

func TestBuildRejectsHeaderInjection(t *testing.T) {
	if _, err := build("a@b.c", "victim@example.com\r\nBcc: x@evil.com", "s", "b", time.Now()); err == nil {
		t.Fatal("expected error")
	}
}
