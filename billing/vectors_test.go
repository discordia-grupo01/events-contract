package billing_test

import (
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/billing"
	"github.com/discordia-grupo01/events-contract/envelope"
)

// Fixed signature vectors. The Elixir port (billing/events.ex, and the copy
// in messaging) must reproduce these exact values: if the signed text ever
// changes, this test breaks and every port has to change with it.
const vectorSecret = "test-secret"

func TestSignatureVectors(t *testing.T) {
	until := time.Date(2026, 12, 3, 12, 30, 45, 500000000, time.UTC)
	tests := map[string]struct {
		event billing.PremiumUpdated
		want  string
	}{
		"active with fraction": {
			event: billing.PremiumUpdated{
				Meta:   envelope.Meta{EventID: "evt-123", OccurredAt: time.Date(2026, 11, 3, 12, 30, 45, 123456000, time.UTC)},
				UserID: "user-42", Active: true, PremiumUntil: &until, Renews: true, Version: 3,
			},
			want: "845a3876e1443f855ac3aeae4dbccf92adee5748e42b07888d60f614cf4222a2",
		},
		// Go trims the zeros of the fraction: occurred_at is signed as ...45Z.
		"never paid, whole second": {
			event: billing.PremiumUpdated{
				Meta:   envelope.Meta{EventID: "evt-124", OccurredAt: time.Date(2026, 11, 3, 12, 30, 45, 0, time.UTC)},
				UserID: "user-42", Active: false, PremiumUntil: nil, Renews: false, Version: 1,
			},
			want: "c30c148b5d9adf70dcf95381f6bbd8302630d4b25a8eac1a06a2e84e3d03ba22",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.event.Sign(vectorSecret)
			if tc.event.Signature != tc.want {
				t.Errorf("got %s, want %s", tc.event.Signature, tc.want)
			}
		})
	}
}
