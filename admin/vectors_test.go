package admin_test

import (
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/admin"
	"github.com/discordia-grupo01/events-contract/envelope"
)

// Fixed signature vectors. The Elixir ports (admin/events.ex, and the copies in
// messaging and voice) must reproduce these exact values: if the signed text
// ever changes, this test breaks and every port has to change with it.
const vectorSecret = "test-secret"

func TestSignatureVectors(t *testing.T) {
	fraction := envelope.Meta{
		EventID:    "evt-123",
		OccurredAt: time.Date(2026, 10, 14, 12, 30, 45, 123456000, time.UTC),
	}
	wholeSecond := envelope.Meta{
		EventID:    "evt-124",
		OccurredAt: time.Date(2026, 10, 14, 12, 30, 45, 0, time.UTC),
	}

	suspended := admin.UserSuspended{Meta: fraction, UserID: "user-42", Version: 3}
	suspended.Sign(vectorSecret)
	if want := "4e4e1885838edbda6e1964eedc9452fd5825874a6c280dba7f73a5f6400ba710"; suspended.Signature != want {
		t.Errorf("suspended with fraction: got %s, want %s", suspended.Signature, want)
	}

	// Go trims the zeros of the fraction: this one is signed as ...45Z.
	whole := admin.UserSuspended{Meta: wholeSecond, UserID: "user-42", Version: 4}
	whole.Sign(vectorSecret)
	if want := "f96b97d150be7091270411377493ba9ace9b58bfc2f782fd674f02fe639f3f16"; whole.Signature != want {
		t.Errorf("suspended whole second: got %s, want %s", whole.Signature, want)
	}

	reactivated := admin.UserReactivated{Meta: fraction, UserID: "user-42", Version: 3}
	reactivated.Sign(vectorSecret)
	if want := "79d9f8177d1a8f643952e70bdf2fde24dc2c891783eb36cb216c51a1b3f3649a"; reactivated.Signature != want {
		t.Errorf("reactivated: got %s, want %s", reactivated.Signature, want)
	}
}
