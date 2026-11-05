package boosts_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/boosts"
	"github.com/discordia-grupo01/events-contract/envelope"
)

const secret = "test-secret"

func event() boosts.SagaEvent {
	return boosts.SagaEvent{
		Meta:     envelope.Meta{EventID: "evt-1", OccurredAt: time.Date(2026, 11, 5, 12, 30, 45, 500000000, time.UTC)},
		BoostID:  "boost-1",
		ServerID: "server-1",
		UserID:   "user-1",
	}
}

func signed(key string) boosts.SagaEvent {
	e := event()
	e.Sign(key, secret)
	return e
}

// Fixed vectors: if the signed text ever changes, billing and servers stop
// understanding each other's in-flight events, so a change must be deliberate.
func TestSignatureVectors(t *testing.T) {
	rejected := event()
	rejected.Reason = boosts.ReasonNotAMember
	rejected.Sign(boosts.RoutingKeyRejected, secret)
	if want := "c2dd20275d7a1b1b59642e08eb9b45ae51d779742788eec57547877494a32862"; rejected.Signature != want {
		t.Errorf("rejected = %s, want %s", rejected.Signature, want)
	}
	if got, want := signed(boosts.RoutingKeyRequested).Signature, "0be33630734d4cc5fe05b4ae4f3206ee8090fa7151523b818993976f167b67bb"; got != want {
		t.Errorf("requested = %s, want %s", got, want)
	}
}

func TestVerifyAcceptsEveryStep(t *testing.T) {
	for _, key := range []string{
		boosts.RoutingKeyRequested, boosts.RoutingKeyCancelled, boosts.RoutingKeyRemoved,
		boosts.RoutingKeyApplied, boosts.RoutingKeyRejected,
	} {
		if !signed(key).Verify(key, secret) {
			t.Errorf("%s doesn't verify", key)
		}
	}
}

func TestVerifySurvivesAJSONRoundTrip(t *testing.T) {
	sent := event()
	sent.OccurredAt = time.Date(2026, 11, 5, 9, 30, 45, 123456000, time.FixedZone("ART", -3*3600))
	sent.Reason = boosts.ReasonServerSuspended
	sent.Sign(boosts.RoutingKeyRejected, secret)

	body, err := json.Marshal(sent)
	if err != nil {
		t.Fatal(err)
	}
	var got boosts.SagaEvent
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Verify(boosts.RoutingKeyRejected, secret) {
		t.Error("the event stopped verifying after the round trip")
	}
}

func TestVerifyRejects(t *testing.T) {
	key := boosts.RoutingKeyApplied
	tampered := map[string]func(*boosts.SagaEvent){
		"event id":  func(e *boosts.SagaEvent) { e.EventID = "other" },
		"boost id":  func(e *boosts.SagaEvent) { e.BoostID = "other" },
		"server id": func(e *boosts.SagaEvent) { e.ServerID = "other" },
		"user id":   func(e *boosts.SagaEvent) { e.UserID = "other" },
		"reason":    func(e *boosts.SagaEvent) { e.Reason = "other" },
		"time":      func(e *boosts.SagaEvent) { e.OccurredAt = e.OccurredAt.Add(time.Second) },
		"not hex":   func(e *boosts.SagaEvent) { e.Signature = "nope" },
		"empty":     func(e *boosts.SagaEvent) { e.Signature = "" },
		"truncated": func(e *boosts.SagaEvent) { e.Signature = e.Signature[:10] },
	}
	for name, change := range tampered {
		t.Run(name, func(t *testing.T) {
			e := signed(key)
			change(&e)
			if e.Verify(key, secret) {
				t.Error("verified")
			}
		})
	}
	if signed(key).Verify(boosts.RoutingKeyRemoved, secret) {
		t.Error("an applied was accepted as a removed")
	}
	if signed(key).Verify(key, "other-secret") {
		t.Error("verified with another secret")
	}

	empty := event()
	empty.Sign(key, "")
	if empty.Verify(key, "") {
		t.Error("an empty secret verified")
	}
}

func TestSagaEventJSONShape(t *testing.T) {
	body, err := json.Marshal(signed(boosts.RoutingKeyApplied))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	signature, _ := got["signature"].(string)
	if len(signature) != 64 || strings.ToLower(signature) != signature {
		t.Errorf("signature = %q", signature)
	}
	delete(got, "signature")
	want := map[string]any{
		"event_id":    "evt-1",
		"occurred_at": "2026-11-05T12:30:45.5Z",
		"boost_id":    "boost-1",
		"server_id":   "server-1",
		"user_id":     "user-1",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRejectedCarriesTheReason(t *testing.T) {
	e := event()
	e.Reason = boosts.ReasonCancelled
	body, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"reason":"cancelled"`) {
		t.Errorf("body = %s", body)
	}
}

func TestServerBoostUpdatedJSONShape(t *testing.T) {
	body, err := json.Marshal(boosts.ServerBoostUpdated{
		Meta:       envelope.Meta{EventID: "evt-1", OccurredAt: time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)},
		ServerID:   "server-1",
		BoostCount: 15,
		Level:      2,
		Version:    7,
	})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"event_id":    "evt-1",
		"occurred_at": "2026-11-05T12:00:00Z",
		"server_id":   "server-1",
		"boost_count": float64(15),
		"level":       float64(2),
		"version":     float64(7),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestRoutingKeys(t *testing.T) {
	keys := map[string]string{
		boosts.RoutingKeyRequested: "billing.server_boost_requested",
		boosts.RoutingKeyCancelled: "billing.server_boost_cancelled",
		boosts.RoutingKeyRemoved:   "billing.server_boost_removed",
		boosts.RoutingKeyApplied:   "servers.server_boost_applied",
		boosts.RoutingKeyRejected:  "servers.server_boost_rejected",
		boosts.RoutingKeyUpdated:   "servers.server_boost_updated",
	}
	for got, want := range keys {
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}
