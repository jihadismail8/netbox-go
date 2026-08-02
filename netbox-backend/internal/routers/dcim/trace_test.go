package dcim

import (
	"net/netip"
	"testing"
)

// ---- splitNode ----

func TestSplitNode(t *testing.T) {
	cases := []struct {
		in      string
		ct, obj uint64
		ok      bool
	}{
		{"42:7", 42, 7, true},
		{"1:100", 1, 100, true},
		{"99:", 0, 0, false}, // missing obj
		{":5", 0, 0, false},  // missing ct
		{"garbage", 0, 0, false},
		{"42:7:extra", 0, 0, false}, // too many parts -> atoi fails on "7:extra"... actually splits to 42, "7:extra"
		{"0:0", 0, 0, true},
	}
	for _, c := range cases {
		ct, obj, ok := splitNode(c.in)
		if ok != c.ok {
			t.Errorf("splitNode(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && (ct != c.ct || obj != c.obj) {
			t.Errorf("splitNode(%q) = (%d, %d), want (%d, %d)", c.in, ct, obj, c.ct, c.obj)
		}
	}
}

// ---- decodePath ----

func TestDecodePath(t *testing.T) {
	raw := `[["42:7"],["55:1"],["43:8"]]`
	steps, err := decodePath(raw)
	if err != nil {
		t.Fatalf("decodePath error: %v", err)
	}
	if len(steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(steps))
	}
	wantStep0 := []string{"42:7"}
	if len(steps[0]) != 1 || steps[0][0] != wantStep0[0] {
		t.Errorf("steps[0] = %v, want %v", steps[0], wantStep0)
	}
}

func TestDecodePath_Nil(t *testing.T) {
	steps, err := decodePath(nil)
	if err != nil || steps != nil {
		t.Errorf("decodePath(nil) = (%v, %v), want (nil, nil)", steps, err)
	}
}

func TestDecodePath_EmptyArray(t *testing.T) {
	steps, err := decodePath(`[]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(steps) != 0 {
		t.Errorf("expected 0 steps, got %d", len(steps))
	}
}

// ---- tableForCT ----

func TestTableForCT(t *testing.T) {
	cases := []struct {
		ct        string
		wantTable string
		wantCable bool
	}{
		{"dcim.interface", "dcim_interface", false},
		{"dcim.consoleport", "dcim_consoleport", false},
		{"dcim.cable", "dcim_cable", true},
		{"dcim.frontport", "dcim_frontport", false},
		{"dcim.rearport", "dcim_rearport", false},
		{"dcim.powerfeed", "dcim_powerfeed", false},
		{"unknown.model", "", false},
	}
	for _, c := range cases {
		table, _, isCable := tableForCT(c.ct)
		if table != c.wantTable || isCable != c.wantCable {
			t.Errorf("tableForCT(%q) = (%q, %v), want (%q, %v)",
				c.ct, table, isCable, c.wantTable, c.wantCable)
		}
	}
}

// ---- atoi / toFloat / toBool ----

func TestAtoi(t *testing.T) {
	cases := map[string]struct {
		val uint64
		ok  bool
	}{
		"42":  {42, true},
		"0":   {0, true},
		"":    {0, false},
		"abc": {0, false},
		" 7 ": {7, true},
	}
	for in, want := range cases {
		got, err := atoi(in)
		if (err == nil) != want.ok {
			t.Errorf("atoi(%q) ok mismatch: got err=%v, want ok=%v", in, err, want.ok)
			continue
		}
		if want.ok && got != want.val {
			t.Errorf("atoi(%q) = %d, want %d", in, got, want.val)
		}
	}
}

func TestToFloat(t *testing.T) {
	cases := []struct {
		in   interface{}
		want float64
	}{
		{float64(3), 3},
		{int64(5), 5},
		{int(7), 7},
		{"3.14", 3.14},
		{[]byte("2.5"), 2.5},
		{nil, 0},
	}
	for _, c := range cases {
		if got := toFloat(c.in); got != c.want {
			t.Errorf("toFloat(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestToBool(t *testing.T) {
	cases := []struct {
		in   interface{}
		want bool
	}{
		{true, true},
		{false, false},
		{int64(1), true},
		{int64(0), false},
		{float64(1), true},
		{nil, false},
	}
	for _, c := range cases {
		if got := toBool(c.in); got != c.want {
			t.Errorf("toBool(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

// ---- modelToTable completeness ----

func TestPathEndpointModelsHaveTableMapping(t *testing.T) {
	for _, m := range PathEndpointModels {
		if _, ok := modelToTable[m]; !ok {
			t.Errorf("PathEndpointModel %q missing from modelToTable", m)
		}
	}
}

func TestPassThroughModelsHaveTableMapping(t *testing.T) {
	for _, m := range PassThroughModels {
		if _, ok := modelToTable[m]; !ok {
			t.Errorf("PassThroughModel %q missing from modelToTable", m)
		}
	}
}

// ---- keep netip referenced so the import is stable across Go versions ----
var _ = netip.Addr{}
