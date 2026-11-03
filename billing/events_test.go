package billing_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/billing"
	"github.com/discordia-grupo01/events-contract/envelope"
)

const secret = "shared-secret"

func premiumUpdated() billing.PremiumUpdated {
	until := time.Date(2026, 12, 3, 12, 0, 0, 0, time.UTC)
	return billing.PremiumUpdated{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 11, 3, 12, 0, 0, 123456789, time.UTC),
		},
		UserID:       "user-1",
		Active:       true,
		PremiumUntil: &until,
		Renews:       true,
		Version:      3,
	}
}

func TestSignedEventVerifies(t *testing.T) {
	e := premiumUpdated()
	e.Sign(secret)
	if e.Signature == "" || !e.Verify(secret) {
		t.Errorf("signed PremiumUpdated does not verify: %q", e.Signature)
	}

	never := premiumUpdated()
	never.Active, never.PremiumUntil, never.Renews = false, nil, false
	never.Sign(secret)
	if !never.Verify(secret) {
		t.Error("a snapshot without premium_until should verify")
	}
}

func TestSignatureIsDeterministic(t *testing.T) {
	a, b := premiumUpdated(), premiumUpdated()
	a.Sign(secret)
	b.Sign(secret)
	if a.Signature != b.Signature {
		t.Error("the same event must give the same signature")
	}
}

func TestTamperedEventDoesNotVerify(t *testing.T) {
	tamper := map[string]func(*billing.PremiumUpdated){
		"event_id":    func(e *billing.PremiumUpdated) { e.EventID = "evt-2" },
		"user_id":     func(e *billing.PremiumUpdated) { e.UserID = "user-2" },
		"active":      func(e *billing.PremiumUpdated) { e.Active = !e.Active },
		"renews":      func(e *billing.PremiumUpdated) { e.Renews = !e.Renews },
		"version":     func(e *billing.PremiumUpdated) { e.Version++ },
		"occurred_at": func(e *billing.PremiumUpdated) { e.OccurredAt = e.OccurredAt.Add(time.Nanosecond) },
		"premium_until later": func(e *billing.PremiumUpdated) {
			later := e.PremiumUntil.AddDate(1, 0, 0)
			e.PremiumUntil = &later
		},
		"premium_until removed": func(e *billing.PremiumUpdated) { e.PremiumUntil = nil },
		"signature": func(e *billing.PremiumUpdated) {
			// Always changes the first hex digit, whatever it is.
			first := byte('0')
			if e.Signature[0] == '0' {
				first = '1'
			}
			e.Signature = string(first) + e.Signature[1:]
		},
	}
	for name, change := range tamper {
		t.Run(name, func(t *testing.T) {
			e := premiumUpdated()
			e.Sign(secret)
			change(&e)
			if e.Verify(secret) {
				t.Errorf("changing %s must not verify", name)
			}
		})
	}
}

func TestWrongOrEmptySecretDoesNotVerify(t *testing.T) {
	e := premiumUpdated()
	e.Sign(secret)
	if e.Verify("other-secret") {
		t.Error("a different secret must not verify")
	}
	if e.Verify("") {
		t.Error("an empty secret must never verify")
	}

	empty := premiumUpdated()
	empty.Sign("")
	if empty.Verify("") {
		t.Error("signing and verifying with an empty secret must not accept the event")
	}
}

func TestMissingOrMalformedSignatureDoesNotVerify(t *testing.T) {
	for _, sig := range []string{"", "not-hex", "zz", "abcd"} {
		e := premiumUpdated()
		e.Signature = sig
		if e.Verify(secret) {
			t.Errorf("signature %q must not verify", sig)
		}
	}
}

// Times in another zone are signed and sent in UTC, so the signature still
// verifies after the JSON round trip and the Elixir port can use the
// strings as received.
func TestSignatureSurvivesJSONRoundTrip(t *testing.T) {
	art := time.FixedZone("ART", -3*60*60)
	e := premiumUpdated()
	e.OccurredAt = e.OccurredAt.In(art)
	until := e.PremiumUntil.In(art)
	e.PremiumUntil = &until
	e.Sign(secret)

	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "-03:00") {
		t.Errorf("timestamps must travel in UTC: %s", raw)
	}
	var got billing.PremiumUpdated
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !got.Verify(secret) {
		t.Errorf("event does not verify after a JSON round trip: %s", raw)
	}
}

func TestJSONShape(t *testing.T) {
	e := premiumUpdated()
	e.PremiumUntil = nil
	e.Sign(secret)
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, k := range []string{"event_id", "occurred_at", "user_id", "active", "premium_until", "renews", "version", "signature"} {
		if _, ok := fields[k]; !ok {
			t.Errorf("missing field %q in %s", k, raw)
		}
	}
	if fields["premium_until"] != nil {
		t.Errorf("premium_until should be null when the user never paid: %s", raw)
	}
}
