package email

import (
	"errors"
	"testing"
)

func TestParseSendError(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"with message", `{"message":"invalid api key"}`, "email: failed to send: invalid api key"},
		{"empty message", `{"message":""}`, "email: failed to send"},
		{"malformed json", `not json`, "email: failed to send"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := parseSendError([]byte(c.body))
			if !errors.Is(err, ErrSendFailed) {
				t.Fatalf("expected ErrSendFailed, got %v", err)
			}
			if err.Error() != c.want {
				t.Errorf("got %q, want %q", err.Error(), c.want)
			}
		})
	}
}
