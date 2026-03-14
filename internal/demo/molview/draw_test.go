package molview

import "testing"

func TestEmbeddedMoleculesLoaded(t *testing.T) {
	if got := MoleculeCount(); got == 0 {
		t.Fatal("expected embedded molecules to load")
	}
	if molecules[0].Name == "" {
		t.Fatal("expected first molecule to have a name")
	}
	if len(molecules[0].Atoms) == 0 {
		t.Fatal("expected first molecule to contain atoms")
	}
}

func TestStateClamp(t *testing.T) {
	st := State{
		MoleculeIdx: 9999,
		Thickness:   -1,
		TextSize:    99,
		Scale:       0,
	}
	st.Clamp()
	if st.MoleculeIdx != MoleculeCount()-1 {
		t.Fatalf("unexpected molecule index: got %d", st.MoleculeIdx)
	}
	if st.Thickness != 0.1 {
		t.Fatalf("unexpected thickness clamp: got %v", st.Thickness)
	}
	if st.TextSize != 5.0 {
		t.Fatalf("unexpected text size clamp: got %v", st.TextSize)
	}
	if st.Scale != 0.05 {
		t.Fatalf("unexpected scale clamp: got %v", st.Scale)
	}
}
