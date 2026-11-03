// Package billing has the events published by the billing service, which
// sells Premium and points through Mercado Pago.
//
// billing.premium_updated is signed like the admin events: a forged one would
// give anyone Premium for free, and RabbitMQ alone doesn't stop another client
// with broker credentials from publishing a billing.* routing key. billing and
// every consumer share a secret, and the consumer drops any event whose
// signature doesn't verify.
package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
)

const RoutingKeyPremiumUpdated = "billing.premium_updated"

// PremiumUpdated is a full snapshot of a user's Premium, not a diff. billing
// publishes it when Premium starts, renews, stops renewing, is revoked
// (refund or chargeback) or expires.
//
// Active is what consumers apply. PremiumUntil (nil if the user never paid)
// and Renews are for display: when PremiumUntil passes, billing publishes
// another snapshot with Active false, so consumers never need a clock.
//
// Version grows with every snapshot of the same user; a consumer applies an
// event only if its version is higher than the one it already has, so
// duplicated or out-of-order deliveries are harmless.
type PremiumUpdated struct {
	envelope.Meta
	UserID       string     `json:"user_id"`
	Active       bool       `json:"active"`
	PremiumUntil *time.Time `json:"premium_until"`
	Renews       bool       `json:"renews"`
	Version      int64      `json:"version"`
	Signature    string     `json:"signature"`
}

// Sign sets Signature using secret. billing calls it right before publishing.
// It also moves both timestamps to UTC, so the JSON carries exactly the text
// that was signed.
func (e *PremiumUpdated) Sign(secret string) {
	e.OccurredAt = e.OccurredAt.UTC()
	if e.PremiumUntil != nil {
		until := e.PremiumUntil.UTC()
		e.PremiumUntil = &until
	}
	e.Signature = hex.EncodeToString(e.signature(secret))
}

// Verify reports whether Signature is valid for secret. An empty secret or an
// empty/malformed signature never verifies.
func (e PremiumUpdated) Verify(secret string) bool {
	if secret == "" || e.Signature == "" {
		return false
	}
	got, err := hex.DecodeString(e.Signature)
	if err != nil {
		return false
	}
	return hmac.Equal(got, e.signature(secret))
}

// signature is HMAC-SHA256(secret, routing_key \n event_id \n user_id \n
// active \n premium_until \n renews \n version \n occurred_at). Booleans are
// "true"/"false", a nil premium_until is the empty string, and times are
// RFC 3339 in UTC, so the text is the same after a JSON round trip.
func (e PremiumUpdated) signature(secret string) []byte {
	until := ""
	if e.PremiumUntil != nil {
		until = formatTime(*e.PremiumUntil)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strings.Join([]string{
		RoutingKeyPremiumUpdated,
		e.EventID,
		e.UserID,
		strconv.FormatBool(e.Active),
		until,
		strconv.FormatBool(e.Renews),
		strconv.FormatInt(e.Version, 10),
		formatTime(e.OccurredAt),
	}, "\n")))
	return mac.Sum(nil)
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
