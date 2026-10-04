package main

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yasunori0418/outturn/go/conformance"

	"github.com/yasunori0418/layat/internal/engine"
)

// TestListGenerationsJSONInfoGenerations pins list-generations' --json shape: {number, date,
// current} rows in result.info.generations, empty items, no generation slot, conformant envelope.
func TestListGenerationsJSONInfoGenerations(t *testing.T) {
	checker, err := conformance.NewDefaultChecker()
	if err != nil {
		t.Fatalf("conformance.NewDefaultChecker: %v", err)
	}

	gens := []engine.Generation{
		{Number: 1, Date: "2026-07-18 09:00:00"},
		{Number: 2, Date: "2026-07-19 12:00:00", Current: true},
	}
	r, buf := newListGenerationsTestRun()
	r.beginSubject("home").setPayload(&outturnPayload[*generationsInfo]{info: &generationsInfo{Generations: generationRows(gens)}})
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if findings := checker.Check(buf.Bytes()); len(findings) > 0 {
		t.Fatalf("conformance findings: %v\ndocument: %s", findings, buf.String())
	}

	doc := decodeEnvelope(t, buf)
	if doc["dryRun"] != false {
		t.Errorf("dryRun = %v, want false (a read-only command reflects only the flag)", doc["dryRun"])
	}
	sr := doc["results"].([]any)[0].(map[string]any)
	if _, ok := sr["generation"]; ok {
		t.Errorf("generation slot present, want absent (info.generations is the single encoding): %v", sr["generation"])
	}
	res := sr["result"].(map[string]any)
	if items := res["items"].([]any); len(items) != 0 {
		t.Errorf("items = %v, want empty (the enumeration is info, not items)", items)
	}
	rows := res["info"].(map[string]any)["generations"].([]any)
	if len(rows) != 2 {
		t.Fatalf("info.generations = %v, want 2 rows", rows)
	}
	current := rows[1].(map[string]any)
	if current["number"] != json.Number("2") || current["date"] != "2026-07-19 12:00:00" || current["current"] != true {
		t.Errorf("rows[1] = %v, want number=2 date verbatim current=true", current)
	}
	if first := rows[0].(map[string]any); first["current"] != false {
		t.Errorf("rows[0].current = %v, want false (the key is always present)", first["current"])
	}
}

// TestListGenerationsJSONInfoAbsentWithoutListing: when list-generations fails after registering
// the subject but before any listing exists, result.info stays absent.
func TestListGenerationsJSONInfoAbsentWithoutListing(t *testing.T) {
	r, buf := newListGenerationsTestRun()
	r.beginSubject("home")
	if err := r.emit(errors.New("layat: list-generations is home mode only")); err != nil {
		t.Fatalf("emit: %v", err)
	}
	assertNoInfoKeys(t, decodeEnvelope(t, buf))
}

// TestListGenerationsJSONEmptyStaysArray: an empty profile emits "generations": [] rather than
// dropping the key.
func TestListGenerationsJSONEmptyStaysArray(t *testing.T) {
	r, buf := newListGenerationsTestRun()
	r.beginSubject("empty").setPayload(&outturnPayload[*generationsInfo]{info: &generationsInfo{Generations: generationRows(nil)}})
	if err := r.emit(nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	info := decodeEnvelope(t, buf)["results"].([]any)[0].(map[string]any)["result"].(map[string]any)["info"].(map[string]any)
	rows, ok := info["generations"].([]any)
	if !ok {
		t.Fatalf("info = %v, want the generations key present as an array", info)
	}
	if len(rows) != 0 {
		t.Errorf("info.generations = %v, want []", rows)
	}
}
