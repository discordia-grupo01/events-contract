package reports_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/discordia-grupo01/events-contract/envelope"
	"github.com/discordia-grupo01/events-contract/reports"
)

func TestReportUpdatedUserReportJSONShape(t *testing.T) {
	reported := "user-2"
	resolvedAt := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	event := reports.ReportUpdated{
		Meta: envelope.Meta{
			EventID:    "evt-1",
			OccurredAt: time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
		},
		ReportID:       "report-1",
		ServerID:       "server-1",
		Kind:           reports.KindUser,
		ReportedUserID: &reported,
		Reason:         "spam",
		Status:         reports.StatusResolved,
		CreatedAt:      time.Date(2026, 1, 1, 9, 30, 0, 0, time.UTC),
		ResolvedAt:     &resolvedAt,
	}

	assertJSONShape(t, event, map[string]any{
		"event_id":         "evt-1",
		"occurred_at":      "2026-01-02T10:00:00Z",
		"report_id":        "report-1",
		"server_id":        "server-1",
		"kind":             "user",
		"reported_user_id": "user-2",
		"channel_id":       nil,
		"message_id":       nil,
		"reason":           "spam",
		"status":           "resolved",
		"created_at":       "2026-01-01T09:30:00Z",
		"resolved_at":      "2026-01-02T10:00:00Z",
	})
}

func TestReportUpdatedPendingMessageReportJSONShape(t *testing.T) {
	channel := "channel-1"
	message := "message-1"
	event := reports.ReportUpdated{
		Meta: envelope.Meta{
			EventID:    "evt-2",
			OccurredAt: time.Date(2026, 1, 1, 9, 30, 0, 0, time.UTC),
		},
		ReportID:  "report-2",
		ServerID:  "server-1",
		Kind:      reports.KindMessage,
		ChannelID: &channel,
		MessageID: &message,
		Reason:    "insultos",
		Status:    reports.StatusPending,
		CreatedAt: time.Date(2026, 1, 1, 9, 30, 0, 0, time.UTC),
	}

	assertJSONShape(t, event, map[string]any{
		"event_id":         "evt-2",
		"occurred_at":      "2026-01-01T09:30:00Z",
		"report_id":        "report-2",
		"server_id":        "server-1",
		"kind":             "message",
		"reported_user_id": nil,
		"channel_id":       "channel-1",
		"message_id":       "message-1",
		"reason":           "insultos",
		"status":           "pending",
		"created_at":       "2026-01-01T09:30:00Z",
		"resolved_at":      nil,
	})
}

func TestConstants(t *testing.T) {
	if reports.RoutingKeyReportUpdated != "servers.report_updated" {
		t.Errorf("RoutingKeyReportUpdated = %q", reports.RoutingKeyReportUpdated)
	}
	for name, tc := range map[string]struct{ got, want string }{
		"KindUser":        {reports.KindUser, "user"},
		"KindMessage":     {reports.KindMessage, "message"},
		"StatusPending":   {reports.StatusPending, "pending"},
		"StatusResolved":  {reports.StatusResolved, "resolved"},
		"StatusDismissed": {reports.StatusDismissed, "dismissed"},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %q, want %q", name, tc.got, tc.want)
		}
	}
}

// assertJSONShape checks that event marshals to exactly the keys in want --
// no missing field and no unexpected one (e.g. the reporter leaking into the
// event). A nil value in want means the key must be present as JSON null.
func assertJSONShape(t *testing.T, event any, want map[string]any) {
	t.Helper()

	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Errorf("missing key %q in %s", key, body)
			continue
		}
		if !reflect.DeepEqual(gotValue, wantValue) {
			t.Errorf("key %q = %v, want %v", key, gotValue, wantValue)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d -- unexpected field in %s", len(got), len(want), body)
	}
}
