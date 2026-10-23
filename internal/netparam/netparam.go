// Package netparam reads and writes the multiplayer timings and capacities in
// NetworkParam row 0 of a decrypted PS4 save. The save format itself is
// ER_shared's.
package netparam

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"

	reg "github.com/DrippingSoup22/ER_shared/regulation"
)

// View is the player role a field belongs to.
type View int

const (
	Invader View = iota
	FindSigns
	PlaceSign
	Hunter
)

// Views is every view, in display order.
var Views = []View{Invader, FindSigns, PlaceSign, Hunter}

func (v View) String() string {
	return [...]string{"Invader", "Find Signs", "Place Sign", "Hunter"}[v]
}

// Kind is a field's storage type in the row.
type Kind int

const (
	S32 Kind = iota
	F32
	U8
)

// Confidence is how well a field's in-game effect is established.
type Confidence int

const (
	Confirmed         Confidence = iota // consistent with game data and community findings
	CommunityInferred                   // inferred from community testing, not documented
	Unconfirmed                         // possibly unused in Elden Ring
)

// Field is one editable value of the row.
type Field struct {
	Key, Label        string
	View              View
	Offset            int
	Kind              Kind
	Min, Max, Default float64
	Advanced          bool // shown only in Advanced mode
	Confidence        Confidence
	Warning           string // an operational caveat
	Short, Long       string // the summary and the full explanation
	Details           string // technical notes for Advanced mode
}

// Format shows a value without a trailing ".0", and a float to two decimals:
// a shipped float can hold a value like 7e-45 that only reads as 0.
func (f Field) Format(v float64) string {
	if f.Kind == F32 {
		v = math.Round(v*100) / 100
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// Parse reads a typed value: a whole number, or a decimal for a float field.
func (f Field) Parse(s string) (float64, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || (f.Kind != F32 && v != math.Trunc(v)) {
		return 0, fmt.Errorf("%s needs a whole number", f.Label)
	}
	if f.Kind == F32 {
		v = float64(float32(v))
	}
	return v, nil
}

// Values maps a field's key to its value. float64 holds every s32, u8 and f32
// exactly, so a value read and written back unchanged keeps its bytes.
type Values map[string]float64

// Clone copies values.
func (v Values) Clone() Values {
	out := make(Values, len(v))
	for key, value := range v {
		out[key] = value
	}
	return out
}

// Equal reports whether both hold the same value for every field.
func (v Values) Equal(other Values) bool {
	for _, f := range Fields {
		if v[f.Key] != other[f.Key] {
			return false
		}
	}
	return true
}

// Preset is one of a view's named value sets. Nil Values means the view's
// shipped defaults.
type Preset struct {
	Name, Description string
	Values            Values
}

// FieldsOf lists a view's fields; standard mode leaves out the advanced ones.
func FieldsOf(view View, advanced bool) []Field {
	var out []Field
	for _, f := range Fields {
		if f.View == view && (advanced || !f.Advanced) {
			out = append(out, f)
		}
	}
	return out
}

// Defaults is every field's shipped value.
func Defaults() Values {
	out := make(Values, len(Fields))
	for _, f := range Fields {
		out[f.Key] = f.Default
	}
	return out
}

// Apply returns values with the preset's settings for the fields of view that
// the mode shows; the rest are left as they are.
func (p Preset) Apply(view View, values Values, advanced bool) Values {
	out := values.Clone()
	for _, f := range FieldsOf(view, advanced) {
		switch {
		case p.Values == nil:
			out[f.Key] = f.Default
		default:
			if value, ok := p.Values[f.Key]; ok {
				out[f.Key] = value
			}
		}
	}
	return out
}

// Validate checks every field's range and the three cross-field rules the
// game's sign pool and Hunter search depend on.
func Validate(values Values) error {
	for _, f := range Fields {
		v, ok := values[f.Key]
		if !ok || math.IsNaN(v) || math.IsInf(v, 0) || v < f.Min || v > f.Max {
			return fmt.Errorf("%s must be %s-%s, got %s", f.Label, f.Format(f.Min), f.Format(f.Max), f.Format(v))
		}
	}
	switch {
	case values["reloadSignCellCount"] > values["reloadSignTotalCount"]:
		return errors.New("the per-cell sign cap must not exceed the total sign pool")
	case values["reloadSignTotalCount"] > values["singGetMax"]:
		return errors.New("the total sign pool must not exceed the sign buffer cap")
	case values["reloadSearchCoopBlueMin"] > values["reloadSearchCoopBlueMax"]:
		return errors.New("the co-op search minimum must not exceed its maximum")
	}
	return nil
}

// ErrPCSave refuses a PC save: the PC game downloads vanilla network
// parameters at every login, so an edit there would never take effect.
var ErrPCSave = errors.New("this is a PC save; the PC game replaces these parameters at every login, so only PS4 saves are edited")

// Save is an opened PS4 save.
type Save struct {
	Path       string
	Values     Values // as the save holds them
	regulation *reg.Regulation
}

// Open reads a decrypted PS4 save.
func Open(path string) (*Save, error) {
	regulation, err := reg.Load(path)
	if err != nil {
		return nil, err
	}
	if regulation.Platform != reg.PS4 {
		return nil, ErrPCSave
	}
	values, err := read(regulation.BND4)
	if err != nil {
		return nil, err
	}
	return &Save{Path: path, Values: values, regulation: regulation}, nil
}

// Write validates values and writes a new save at outPath, which may replace a
// file the user chose but never the input. The written file is reopened and
// every value must read back before it takes the output's name.
func (s *Save) Write(outPath string, values Values) error {
	if err := Validate(values); err != nil {
		return err
	}
	edited := *s.regulation
	edited.BND4 = bytes.Clone(s.regulation.BND4)
	data, err := row(edited.BND4)
	if err != nil {
		return err
	}
	for _, f := range Fields {
		put(data[f.Offset:], f.Kind, values[f.Key])
	}
	output, _, err := reg.Build(&edited)
	if err != nil {
		return err
	}
	return reg.ReplaceAtomicValidated(s.Path, outPath, output, func(candidate string) error {
		written, err := reg.Load(candidate)
		if err != nil {
			return err
		}
		got, err := read(written.BND4)
		if err != nil {
			return err
		}
		for _, f := range Fields {
			if get(data[f.Offset:], f.Kind) != got[f.Key] {
				return fmt.Errorf("%s reads back as %s", f.Label, f.Format(got[f.Key]))
			}
		}
		return nil
	})
}

// row is NetworkParam's single row, checked against the measured layout so a
// regulation whose table changed is refused rather than written at guessed
// offsets.
func row(bnd4 []byte) ([]byte, error) {
	layout := reg.MustLayout("NetworkParam")
	data, rows, err := reg.LoadRowsByID(bnd4, layout.Entry, layout.ParamType, layout.DataVersion)
	if err != nil {
		return nil, err
	}
	r, ok := rows[0]
	if !ok || len(rows) != 1 {
		return nil, fmt.Errorf("%s holds %d rows, expected row 0 alone: %w", layout.Entry, len(rows), reg.ErrUnsupportedRegulation)
	}
	if err := reg.RequireRowSize(data, r, layout.RowSize, layout.Entry); err != nil {
		return nil, err
	}
	return data[r.DataOffset : r.DataOffset+layout.RowSize], nil
}

func read(bnd4 []byte) (Values, error) {
	data, err := row(bnd4)
	if err != nil {
		return nil, err
	}
	values := make(Values, len(Fields))
	for _, f := range Fields {
		values[f.Key] = get(data[f.Offset:], f.Kind)
	}
	return values, nil
}

func get(b []byte, kind Kind) float64 {
	switch kind {
	case F32:
		return float64(math.Float32frombits(binary.LittleEndian.Uint32(b)))
	case U8:
		return float64(b[0])
	default:
		return float64(int32(binary.LittleEndian.Uint32(b)))
	}
}

func put(b []byte, kind Kind, v float64) {
	switch kind {
	case F32:
		binary.LittleEndian.PutUint32(b, math.Float32bits(float32(v)))
	case U8:
		b[0] = byte(v)
	default:
		binary.LittleEndian.PutUint32(b, uint32(int32(v)))
	}
}
