package netparam

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	reg "github.com/DrippingSoup22/ER_shared/regulation"
)

// save is a private save, named by an environment variable; the default is
// where the workspace keeps one. Tests that need it skip without it.
func save(t *testing.T, env, fallback string) string {
	t.Helper()
	path := os.Getenv(env)
	if path == "" {
		path = fallback
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("%s is absent: %v", path, err)
	}
	return path
}

func ps4Save(t *testing.T) string {
	return save(t, "ER_SAVE", filepath.Join("..", "..", "..", "ER_research", "save_file", "memory-1.17.1.dat"))
}

// The table is data, so its consistency is checked rather than trusted: every
// field fits the row without overlap, the defaults validate, and every preset
// names only its own view's fields and validates once applied.
func TestFieldsAndPresetsAreConsistent(t *testing.T) {
	width := map[Kind]int{S32: 4, F32: 4, U8: 1}
	used := map[int]string{}
	for _, f := range Fields {
		if f.Offset+width[f.Kind] > reg.MustLayout("NetworkParam").RowSize || f.Min > f.Default || f.Default > f.Max {
			t.Errorf("%s: offset %#x, default %v outside %v-%v", f.Key, f.Offset, f.Default, f.Min, f.Max)
		}
		for b := f.Offset; b < f.Offset+width[f.Kind]; b++ {
			if other, clash := used[b]; clash {
				t.Errorf("%s overlaps %s at %#x", f.Key, other, b)
			}
			used[b] = f.Key
		}
	}
	if err := Validate(Defaults()); err != nil {
		t.Fatal(err)
	}
	for _, view := range Views {
		own := map[string]bool{}
		for _, f := range FieldsOf(view, true) {
			own[f.Key] = true
		}
		for _, preset := range Presets[view] {
			for key := range preset.Values {
				if !own[key] {
					t.Errorf("%s %s sets %s, not one of its fields", view, preset.Name, key)
				}
			}
			for _, advanced := range []bool{false, true} {
				if err := Validate(preset.Apply(view, Defaults(), advanced)); err != nil {
					t.Errorf("%s %s: %v", view, preset.Name, err)
				}
			}
		}
	}
}

// Standard mode leaves advanced fields alone: a preset applied there changes
// only what the view shows.
func TestPresetsSparesAdvancedFieldsInStandardMode(t *testing.T) {
	values := Defaults()
	aggressive := Presets[FindSigns][2].Apply(FindSigns, values, false)
	if aggressive["signCellRangeH"] != values["signCellRangeH"] || aggressive["reloadSignTotalCount"] == values["reloadSignTotalCount"] {
		t.Fatalf("standard-mode preset: cell range %v, sign pool %v", aggressive["signCellRangeH"], aggressive["reloadSignTotalCount"])
	}
}

// A write lands every value, changes nothing outside the regulation, and is
// refused - writing nothing - for an invalid value or the input path.
func TestWriteRoundTripsAndRefuses(t *testing.T) {
	in := ps4Save(t)
	s, err := Open(in)
	if err != nil {
		t.Fatal(err)
	}
	values := s.Values
	for _, view := range Views {
		values = Presets[view][2].Apply(view, values, true)
	}
	out := filepath.Join(t.TempDir(), "out.dat")
	if err := s.Write(out, values); err != nil {
		t.Fatal(err)
	}
	written, err := Open(out)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range Fields {
		if written.Values[f.Key] != values[f.Key] {
			t.Errorf("%s: wrote %v, read %v", f.Key, values[f.Key], written.Values[f.Key])
		}
	}
	original, _ := os.ReadFile(in)
	edited, _ := os.ReadFile(out)
	if ud11 := reg.UserData11Offset(reg.PS4); !bytes.Equal(original[:ud11], edited[:ud11]) {
		t.Error("bytes before the regulation changed")
	}

	bad := values.Clone()
	bad["reloadSignCellCount"] = bad["reloadSignTotalCount"] + 1
	refused := filepath.Join(t.TempDir(), "refused.dat")
	if err := s.Write(refused, bad); err == nil {
		t.Error("a sign cap above the pool was accepted")
	}
	if _, err := os.Stat(refused); !os.IsNotExist(err) {
		t.Error("a refused write left a file")
	}
	if err := s.Write(in, values); !errors.Is(err, reg.ErrOutputExists) {
		t.Errorf("writing over the input returned %v", err)
	}
}

// PC is refused: the game replaces these parameters there at every login.
func TestOpenRefusesAPCSave(t *testing.T) {
	pc := save(t, "ER_PC_SAVE", filepath.Join("..", "..", "..", "er_bak", "ER0000.sl2"))
	if _, err := Open(pc); !errors.Is(err, ErrPCSave) {
		t.Fatalf("Open(PC save) = %v", err)
	}
}
