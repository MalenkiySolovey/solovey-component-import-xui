package preview

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MalenkiySolovey/solovey-component-import-xui/mapping"
	"github.com/MalenkiySolovey/solovey-component-import-xui/source"
)

func TestPreviewBlocksIncompleteSourceAndDependentClient(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x-ui.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, statement := range []string{
		`CREATE TABLE inbounds (id INTEGER, port INTEGER, protocol TEXT, tag TEXT, settings TEXT, stream_settings TEXT)`,
		`CREATE TABLE client_traffics (id INTEGER, inbound_id INTEGER, email TEXT)`,
		`INSERT INTO inbounds VALUES (1, 1234, 'vless', 'fixture', '{"clients":[{"email":"alice"}],"unknownField":"fixture-secret"}', '{"network":"tcp","security":"none"}')`,
		`INSERT INTO client_traffics VALUES (1, 1, 'alice')`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	before, err := source.Hash(path)
	if err != nil {
		t.Fatal(err)
	}
	src, err := source.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	plan := Plan[struct{}]{
		Items: []Item[struct{}]{
			{Kind: KindInbound, SrcID: int64(1), Action: ActionCreate, PreviewJSON: json.RawMessage(`{}`)},
			{Kind: KindClient, SrcID: "alice", SrcTag: "alice", Action: ActionCreate, PreviewJSON: json.RawMessage(`{}`)},
		},
	}
	if err := Classify(t.Context(), src, &plan); err != nil {
		t.Fatal(err)
	}
	for _, item := range plan.Items {
		if item.Action != ActionSkip || item.Classification != mapping.UnsupportedBlocked || !mapping.Blocking(item.Fields) || string(item.PreviewJSON) != "null" {
			t.Fatalf("%s escaped classification: %+v", item.Kind, item)
		}
	}
	encoded, err := json.Marshal(plan)
	if err != nil || strings.Contains(string(encoded), "fixture-secret") {
		t.Fatal("classification leaked a source value", err)
	}
	after, err := source.Hash(path)
	if err != nil || after != before {
		t.Fatal("classifier modified staged source", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := Classify(ctx, src, &plan); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled classification", err)
	}
}

func TestPlanRevisionBindsProposalButNotStagingPath(t *testing.T) {
	plan := Plan[struct{}]{
		Source: PlanSource{Path: "first/path.db", Hash: "source-content"},
		Items:  []Item[struct{}]{{Kind: KindInbound, SrcID: int64(123), Action: ActionCreate, PreviewJSON: json.RawMessage(`{}`)}},
	}
	first, err := plan.Revision()
	if err != nil || len(first) != 64 {
		t.Fatalf("invalid digest %q: %v", first, err)
	}
	plan.Source.Path = "second/path.db"
	second, err := plan.Revision()
	if err != nil || first != second {
		t.Fatal("staging path changed the revision", err)
	}
	plan.Items[0].Action = ActionSkip
	third, err := plan.Revision()
	if err != nil || third == second {
		t.Fatal("revision did not bind approved action", err)
	}
}
