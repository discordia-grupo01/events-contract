package admin_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/admin"
	"github.com/discordia-grupo01/events-contract/envelope"
)

func recoveryMeta() envelope.Meta {
	return envelope.Meta{
		EventID:    "evt-1",
		OccurredAt: time.Date(2026, 1, 1, 10, 30, 0, 123456789, time.UTC),
	}
}

func recoveryApproved() admin.AccountRecoveryApproved {
	return admin.AccountRecoveryApproved{Meta: recoveryMeta(), RequestID: "request-1", UserID: "user-1"}
}

func recoveryRejected() admin.AccountRecoveryRejected {
	return admin.AccountRecoveryRejected{Meta: recoveryMeta(), RequestID: "request-1", UserID: "user-1"}
}

func TestSignedRecoveryEventsVerify(t *testing.T) {
	a := recoveryApproved()
	a.Sign(secret)
	if a.Signature == "" || !a.Verify(secret) {
		t.Errorf("AccountRecoveryApproved firmado no verifica: %q", a.Signature)
	}

	r := recoveryRejected()
	r.Sign(secret)
	if r.Signature == "" || !r.Verify(secret) {
		t.Errorf("AccountRecoveryRejected firmado no verifica: %q", r.Signature)
	}
}

func TestTamperedRecoveryEventDoesNotVerify(t *testing.T) {
	tamper := map[string]func(*admin.AccountRecoveryApproved){
		"request_id":  func(e *admin.AccountRecoveryApproved) { e.RequestID = "request-2" },
		"user_id":     func(e *admin.AccountRecoveryApproved) { e.UserID = "user-2" },
		"event_id":    func(e *admin.AccountRecoveryApproved) { e.EventID = "evt-2" },
		"occurred_at": func(e *admin.AccountRecoveryApproved) { e.OccurredAt = e.OccurredAt.Add(time.Nanosecond) },
	}
	for field, change := range tamper {
		t.Run(field, func(t *testing.T) {
			e := recoveryApproved()
			e.Sign(secret)
			change(&e)
			if e.Verify(secret) {
				t.Errorf("cambiar %s no debería verificar", field)
			}
		})
	}
}

func TestRecoveryEventWrongEmptyOrMalformedSignatureDoesNotVerify(t *testing.T) {
	e := recoveryApproved()
	e.Sign(secret)
	if e.Verify("other-secret") {
		t.Error("con otro secreto no debería verificar")
	}
	if e.Verify("") {
		t.Error("con secreto vacío no debería verificar")
	}

	for name, sig := range map[string]string{"vacía": "", "no hex": "zz"} {
		e := recoveryApproved()
		e.Signature = sig
		if e.Verify(secret) {
			t.Errorf("firma %s no debería verificar", name)
		}
	}
}

func TestRejectionCannotBeReplayedAsApproval(t *testing.T) {
	r := recoveryRejected()
	r.Sign(secret)
	a := recoveryApproved()
	a.Signature = r.Signature
	if a.Verify(secret) {
		t.Error("la firma de un rechazo no debería valer como aprobación")
	}

	// Mismo texto firmado salvo la routing key: tampoco sirve como suspensión.
	s := admin.UserSuspended{Meta: recoveryMeta(), UserID: "user-1"}
	s.Signature = a.Signature
	if s.Verify(secret) {
		t.Error("la firma de una recuperación no debería valer como suspensión")
	}
}

func TestRecoveryEventSignatureSurvivesJSONRoundTrip(t *testing.T) {
	e := recoveryApproved()
	e.OccurredAt = time.Date(2026, 1, 1, 7, 30, 0, 123456789, time.FixedZone("ART", -3*60*60))
	e.Sign(secret)

	body, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var got admin.AccountRecoveryApproved
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Verify(secret) {
		t.Error("la firma debería sobrevivir a la serialización")
	}

	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"event_id", "occurred_at", "request_id", "user_id", "signature"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("falta %q en el JSON", key)
		}
	}
}

func TestRecoveryRoutingKeys(t *testing.T) {
	if admin.RoutingKeyAccountRecoveryApproved != "admin.account_recovery_approved" {
		t.Errorf("RoutingKeyAccountRecoveryApproved = %q", admin.RoutingKeyAccountRecoveryApproved)
	}
	if admin.RoutingKeyAccountRecoveryRejected != "admin.account_recovery_rejected" {
		t.Errorf("RoutingKeyAccountRecoveryRejected = %q", admin.RoutingKeyAccountRecoveryRejected)
	}
}
