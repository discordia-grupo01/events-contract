package admin_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/admin"
	"github.com/discordia-grupo01/events-contract/envelope"
)

const secret = "shared-secret"

func suspended() admin.UserSuspended {
	return admin.UserSuspended{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 1, 1, 10, 30, 0, 123456789, time.UTC),
		},
		UserID:  "user-1",
		Version: 3,
	}
}

func reactivated() admin.UserReactivated {
	e := suspended()
	return admin.UserReactivated{Meta: e.Meta, UserID: e.UserID, Version: e.Version}
}

func TestSignedEventVerifies(t *testing.T) {
	s := suspended()
	s.Sign(secret)
	if s.Signature == "" || !s.Verify(secret) {
		t.Errorf("UserSuspended firmado no verifica: %q", s.Signature)
	}

	r := reactivated()
	r.Sign(secret)
	if r.Signature == "" || !r.Verify(secret) {
		t.Errorf("UserReactivated firmado no verifica: %q", r.Signature)
	}
}

func TestSignatureIsDeterministic(t *testing.T) {
	a, b := suspended(), suspended()
	a.Sign(secret)
	b.Sign(secret)
	if a.Signature != b.Signature {
		t.Error("la misma entrada debe dar la misma firma")
	}
}

func TestTamperedEventDoesNotVerify(t *testing.T) {
	tamper := map[string]func(*admin.UserSuspended){
		"user_id":     func(e *admin.UserSuspended) { e.UserID = "user-2" },
		"version":     func(e *admin.UserSuspended) { e.Version++ },
		"event_id":    func(e *admin.UserSuspended) { e.EventID = "evt-2" },
		"occurred_at": func(e *admin.UserSuspended) { e.OccurredAt = e.OccurredAt.Add(time.Nanosecond) },
		"firma": func(e *admin.UserSuspended) {
			// Cambia siempre el primer dígito hex, sea cual sea.
			first := byte('0')
			if e.Signature[0] == '0' {
				first = '1'
			}
			e.Signature = string(first) + e.Signature[1:]
		},
	}
	for name, change := range tamper {
		t.Run(name, func(t *testing.T) {
			e := suspended()
			e.Sign(secret)
			change(&e)
			if e.Verify(secret) {
				t.Errorf("modificar %s no debería verificar", name)
			}
		})
	}
}

func TestWrongOrEmptySecretDoesNotVerify(t *testing.T) {
	e := suspended()
	e.Sign(secret)
	if e.Verify("otro-secreto") {
		t.Error("un secreto distinto no debe verificar")
	}
	if e.Verify("") {
		t.Error("un secreto vacío nunca debe verificar")
	}

	// Firmar con secreto vacío tampoco habilita nada: ni siquiera con el mismo secreto vacío.
	empty := suspended()
	empty.Sign("")
	if empty.Verify("") {
		t.Error("firmar y verificar con secreto vacío no debe dar por válido el evento")
	}
}

func TestMissingOrMalformedSignatureDoesNotVerify(t *testing.T) {
	for _, sig := range []string{"", "no-es-hex", "zz", "abcd"} {
		e := suspended()
		e.Signature = sig
		if e.Verify(secret) {
			t.Errorf("la firma %q no debería verificar", sig)
		}
	}
}

// La firma incluye la routing key: un UserSuspended válido no sirve como
// UserReactivated aunque tenga los mismos datos.
func TestSignatureCannotBeReplayedAsTheOtherEvent(t *testing.T) {
	s := suspended()
	s.Sign(secret)

	r := reactivated()
	r.Signature = s.Signature
	if r.Verify(secret) {
		t.Error("la firma de una suspensión no debe valer para una reactivación")
	}
}

// Lo que viaja por RabbitMQ es JSON: tras serializar y deserializar el evento
// tiene que seguir verificando (la fecha se compara como instante en UTC).
func TestSignatureSurvivesJSONRoundTrip(t *testing.T) {
	in := suspended()
	in.OccurredAt = time.Date(2026, 1, 1, 7, 30, 0, 123456789, time.FixedZone("ART", -3*3600))
	in.Sign(secret)

	body, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out admin.UserSuspended
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	if !out.Verify(secret) {
		t.Errorf("el evento no verifica después del viaje JSON: %s", body)
	}
}

func TestUserSuspendedJSONShape(t *testing.T) {
	e := suspended()
	e.Sign(secret)
	assertKeys(t, e, []string{"event_id", "occurred_at", "user_id", "version", "signature"})
}

func TestUserReactivatedJSONShape(t *testing.T) {
	e := reactivated()
	e.Sign(secret)
	assertKeys(t, e, []string{"event_id", "occurred_at", "user_id", "version", "signature"})
}

// El evento no puede llevar nada más: ni motivo de la suspensión ni email.
func assertKeys(t *testing.T, event any, want []string) {
	t.Helper()
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range want {
		if _, ok := got[key]; !ok {
			t.Errorf("falta la clave %q en %s", key, body)
		}
	}
	if len(got) != len(want) {
		t.Errorf("hay %d claves, quería %d -- campo inesperado en %s", len(got), len(want), body)
	}
}

func TestRoutingKeys(t *testing.T) {
	if admin.RoutingKeyUserSuspended != "admin.user_suspended" {
		t.Errorf("RoutingKeyUserSuspended = %q", admin.RoutingKeyUserSuspended)
	}
	if admin.RoutingKeyUserReactivated != "admin.user_reactivated" {
		t.Errorf("RoutingKeyUserReactivated = %q", admin.RoutingKeyUserReactivated)
	}
}

func serverMeta() envelope.Meta {
	return envelope.Meta{
		EventID:    "evt-1",
		OccurredAt: time.Date(2026, 1, 1, 10, 30, 0, 123456789, time.UTC),
	}
}

func serverSuspended() admin.ServerSuspended {
	return admin.ServerSuspended{Meta: serverMeta(), ServerID: "server-1", Version: 3}
}

func serverReactivated() admin.ServerReactivated {
	return admin.ServerReactivated{Meta: serverMeta(), ServerID: "server-1", Version: 3}
}

func serverDeletionRequested() admin.ServerDeletionRequested {
	return admin.ServerDeletionRequested{Meta: serverMeta(), ServerID: "server-1", Version: 3}
}

func TestSignedServerEventsVerify(t *testing.T) {
	s := serverSuspended()
	s.Sign(secret)
	if s.Signature == "" || !s.Verify(secret) {
		t.Errorf("ServerSuspended firmado no verifica: %q", s.Signature)
	}

	r := serverReactivated()
	r.Sign(secret)
	if r.Signature == "" || !r.Verify(secret) {
		t.Errorf("ServerReactivated firmado no verifica: %q", r.Signature)
	}

	d := serverDeletionRequested()
	d.Sign(secret)
	if d.Signature == "" || !d.Verify(secret) {
		t.Errorf("ServerDeletionRequested firmado no verifica: %q", d.Signature)
	}
}

func TestTamperedServerEventDoesNotVerify(t *testing.T) {
	tamper := map[string]func(*admin.ServerDeletionRequested){
		"server_id":   func(e *admin.ServerDeletionRequested) { e.ServerID = "server-2" },
		"version":     func(e *admin.ServerDeletionRequested) { e.Version++ },
		"event_id":    func(e *admin.ServerDeletionRequested) { e.EventID = "evt-2" },
		"occurred_at": func(e *admin.ServerDeletionRequested) { e.OccurredAt = e.OccurredAt.Add(time.Nanosecond) },
		"firma": func(e *admin.ServerDeletionRequested) {
			first := byte('0')
			if e.Signature[0] == '0' {
				first = '1'
			}
			e.Signature = string(first) + e.Signature[1:]
		},
	}
	for name, change := range tamper {
		t.Run(name, func(t *testing.T) {
			e := serverDeletionRequested()
			e.Sign(secret)
			change(&e)
			if e.Verify(secret) {
				t.Errorf("modificar %s no debería verificar", name)
			}
		})
	}
}

func TestServerEventWrongOrEmptySecretDoesNotVerify(t *testing.T) {
	e := serverSuspended()
	e.Sign(secret)
	if e.Verify("otro-secreto") {
		t.Error("un secreto distinto no debe verificar")
	}
	if e.Verify("") {
		t.Error("un secreto vacío nunca debe verificar")
	}

	empty := serverSuspended()
	empty.Sign("")
	if empty.Verify("") {
		t.Error("firmar y verificar con secreto vacío no debe dar por válido el evento")
	}
}

func TestServerEventMissingOrMalformedSignatureDoesNotVerify(t *testing.T) {
	for _, sig := range []string{"", "no-es-hex", "zz", "abcd"} {
		e := serverReactivated()
		e.Signature = sig
		if e.Verify(secret) {
			t.Errorf("la firma %q no debería verificar", sig)
		}
	}
}

// Los cinco eventos de admin comparten secreto, id de sujeto, versión y fecha:
// lo único que los distingue es la routing key, que va dentro del texto
// firmado. La firma de cada uno no tiene que valer para ninguno de los otros;
// en especial, una suspensión no puede pasar por una eliminación.
func TestSignatureCannotBeReplayedAsAnotherAdminEvent(t *testing.T) {
	meta := envelope.Meta{
		EventID:    "evt-1",
		OccurredAt: time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC),
	}
	const id = "same-id"
	const version = 3

	us := admin.UserSuspended{Meta: meta, UserID: id, Version: version}
	ur := admin.UserReactivated{Meta: meta, UserID: id, Version: version}
	ss := admin.ServerSuspended{Meta: meta, ServerID: id, Version: version}
	sr := admin.ServerReactivated{Meta: meta, ServerID: id, Version: version}
	sd := admin.ServerDeletionRequested{Meta: meta, ServerID: id, Version: version}
	us.Sign(secret)
	ur.Sign(secret)
	ss.Sign(secret)
	sr.Sign(secret)
	sd.Sign(secret)

	events := []struct {
		name      string
		signature string
		verifies  func(sig string) bool
	}{
		{"user_suspended", us.Signature, func(sig string) bool { e := us; e.Signature = sig; return e.Verify(secret) }},
		{"user_reactivated", ur.Signature, func(sig string) bool { e := ur; e.Signature = sig; return e.Verify(secret) }},
		{"server_suspended", ss.Signature, func(sig string) bool { e := ss; e.Signature = sig; return e.Verify(secret) }},
		{"server_reactivated", sr.Signature, func(sig string) bool { e := sr; e.Signature = sig; return e.Verify(secret) }},
		{"server_deletion_requested", sd.Signature, func(sig string) bool { e := sd; e.Signature = sig; return e.Verify(secret) }},
	}

	for _, from := range events {
		for _, to := range events {
			got := to.verifies(from.signature)
			if from.name == to.name && !got {
				t.Errorf("la firma de %s debe verificar en sí misma", from.name)
			}
			if from.name != to.name && got {
				t.Errorf("la firma de %s no debe valer para %s", from.name, to.name)
			}
		}
	}
}

func TestServerEventSignatureSurvivesJSONRoundTrip(t *testing.T) {
	at := time.Date(2026, 1, 1, 7, 30, 0, 123456789, time.FixedZone("ART", -3*3600))

	s := serverSuspended()
	s.OccurredAt = at
	s.Sign(secret)
	r := serverReactivated()
	r.OccurredAt = at
	r.Sign(secret)
	d := serverDeletionRequested()
	d.OccurredAt = at
	d.Sign(secret)

	for name, in := range map[string]any{"suspended": s, "reactivated": r, "deletion_requested": d} {
		body, err := json.Marshal(in)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(body, &fields); err != nil {
			t.Fatal(err)
		}
		if _, ok := fields["server_id"]; !ok {
			t.Errorf("%s: falta server_id en %s", name, body)
		}
		if _, ok := fields["user_id"]; ok {
			t.Errorf("%s: no debe llevar user_id: %s", name, body)
		}
	}

	body, _ := json.Marshal(s)
	var outS admin.ServerSuspended
	if err := json.Unmarshal(body, &outS); err != nil || !outS.Verify(secret) {
		t.Errorf("ServerSuspended no verifica después del viaje JSON: %v %s", err, body)
	}
	body, _ = json.Marshal(r)
	var outR admin.ServerReactivated
	if err := json.Unmarshal(body, &outR); err != nil || !outR.Verify(secret) {
		t.Errorf("ServerReactivated no verifica después del viaje JSON: %v %s", err, body)
	}
	body, _ = json.Marshal(d)
	var outD admin.ServerDeletionRequested
	if err := json.Unmarshal(body, &outD); err != nil || !outD.Verify(secret) {
		t.Errorf("ServerDeletionRequested no verifica después del viaje JSON: %v %s", err, body)
	}
}
