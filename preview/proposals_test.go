package preview

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MalenkiySolovey/solovey-component-import-xui/source"
)

func TestRoutingProposalKeepsDNSOnlyAndSkipsEmpty(t *testing.T) {
	for _, tc := range []struct {
		input, action string
	}{
		{"", ActionSkip},
		{`{"outbounds":[{"tag":"direct","protocol":"freedom"}],"dns":{"servers":["1.1.1.1"]}}`, ActionCreate},
	} {
		candidate, err := ProposeRouting[struct{}](tc.input)
		if err != nil || candidate.Item.Action != tc.action || candidate.Item.Kind != KindRouting {
			t.Fatalf("proposal for %q = %+v, %v", tc.input, candidate, err)
		}
	}
}

func TestRoutingDisabledNoticeProjectsOnlyExcludedOutbounds(t *testing.T) {
	if item, ok := RoutingDisabledNotice[struct{}](`{"outbounds":[{"tag":"direct","protocol":"freedom"}]}`); ok {
		t.Fatalf("system outbound returned a notice: %+v", item)
	}
	proxy := `{"outbounds":[{"tag":"chain","protocol":"vless","settings":{"vnext":[{"address":"proxy.example","port":443,"users":[{"id":"u"}]}]}}]}`
	item, ok := RoutingDisabledNotice[struct{}](proxy)
	if !ok || item.Kind != KindRouting || item.Action != ActionSkip || string(item.PreviewJSON) != "null" ||
		len(item.Warnings) != 1 || !strings.Contains(item.Warnings[0], "1 proxy outbound(s)") {
		t.Fatalf("want a warning-only routing item, got %+v", item)
	}
}

func TestHistoricalProposalUsesOnlyAggregatedTraffic(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, statement := range []string{
		`CREATE TABLE inbounds (id INTEGER, port INTEGER, protocol TEXT, tag TEXT, settings TEXT, stream_settings TEXT)`,
		`CREATE TABLE client_traffics (id INTEGER, inbound_id INTEGER, email TEXT, up INTEGER, down INTEGER)`,
		`INSERT INTO client_traffics VALUES (1, 1, 'a-private-client-name', 10, 0)`,
		`CREATE TABLE outbound_traffics (id INTEGER, tag TEXT, up INTEGER, down INTEGER)`,
		`INSERT INTO outbound_traffics VALUES (1, 'a-private-outbound-name', 0, 20)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	src, err := source.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	plan := Plan[struct{}]{}
	if err := AddHistorical(ctx, src, &plan); !errors.Is(err, context.Canceled) || len(plan.Items) != 0 {
		t.Fatalf("canceled history changed the plan: %+v, %v", plan, err)
	}
	if err := AddHistorical(t.Context(), src, &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.Items) != 1 || !plan.Defaults.IncludeHistory {
		t.Fatalf("unexpected historical proposal %+v", plan)
	}
	item := plan.Items[0]
	if item.Kind != KindHistory || item.Action != ActionCreate ||
		string(item.PreviewJSON) != "{\n  \"client_traffics\": 1,\n  \"mode\": \"aggregated_only\",\n  \"outbound_traffics\": 1\n}" {
		t.Fatalf("unexpected historical projection %+v", item)
	}
}
