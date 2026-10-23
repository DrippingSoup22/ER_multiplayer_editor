package ui

import (
	"math"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/DrippingSoup22/ER_multiplayer_editor/internal/netparam"
)

// row is one field's line in the table: its name (click for its
// documentation), the file's value, a slider and a box for the new value.
type row struct {
	field  netparam.Field
	name   widget.Clickable
	slider widget.Float
	edit   widget.Editor
}

const (
	currentWidth = unit.Dp(64)
	editWidth    = unit.Dp(70)
	rangeWidth   = unit.Dp(74)
)

func (s *State) layoutTable(gtx C) D {
	fields := netparam.FieldsOf(s.view, s.advanced)
	header := func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				return s.columns(gtx,
					label(s.th, 13, "Field", colorMuted).Layout, label(s.th, 13, "Current", colorMuted).Layout,
					label(s.th, 13, "New value", colorMuted).Layout, func(gtx C) D { return D{} },
					label(s.th, 13, "Range", colorMuted).Layout)
			}),
			layout.Rigid(gap(6)),
			layout.Rigid(rule),
		)
	}
	// The header is the list's first line, so it shares the rows' width.
	return layout.UniformInset(unit.Dp(10)).Layout(gtx, func(gtx C) D {
		return material.List(s.th, &s.table).Layout(gtx, len(fields)+1, func(gtx C, i int) D {
			if i == 0 {
				return header(gtx)
			}
			return s.layoutRow(gtx, s.rows[fields[i-1].Key])
		})
	})
}

// columns lays out one table line: name, current, slider, new value, range.
func (s *State) columns(gtx C, name, current, slider, edit, rng layout.Widget) D {
	fixed := func(width unit.Dp, w layout.Widget) layout.FlexChild {
		return layout.Rigid(func(gtx C) D {
			gtx.Constraints.Min.X, gtx.Constraints.Max.X = gtx.Dp(width), gtx.Dp(width)
			return w(gtx)
		})
	}
	flexed := func(weight float32, w layout.Widget) layout.FlexChild {
		return layout.Flexed(weight, func(gtx C) D {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			dims := w(gtx)
			dims.Size.X = gtx.Constraints.Max.X
			return dims
		})
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		flexed(0.45, name),
		fixed(currentWidth, current),
		flexed(0.55, slider),
		layout.Rigid(gap(8)),
		fixed(editWidth, edit),
		layout.Rigid(gap(10)),
		fixed(rangeWidth, rng),
	)
}

func (s *State) layoutRow(gtx C, r *row) D {
	f := r.field
	loaded := s.save != nil
	s.updateRow(gtx, r, loaded)
	return layout.Inset{Top: unit.Dp(6), Bottom: unit.Dp(6)}.Layout(gtx, func(gtx C) D {
		name := func(gtx C) D {
			c := colorFg
			if s.doc.preset < 0 && s.doc.field == f.Key {
				c = colorAccent
			}
			text := f.Label
			switch f.Confidence {
			case netparam.CommunityInferred:
				text += "  ◇"
			case netparam.Unconfirmed:
				text += "  ?"
			}
			return material.Clickable(gtx, &r.name, label(s.th, 15, text, c).Layout)
		}
		current, changed := "—", false
		if loaded {
			current = f.Format(s.save.Values[f.Key])
			changed = s.values[f.Key] != s.save.Values[f.Key]
		}
		slider := func(gtx C) D {
			if !loaded {
				gtx = gtx.Disabled()
			}
			sl := material.Slider(s.th, &r.slider)
			sl.Color = colorAccent
			return sl.Layout(gtx)
		}
		edit := func(gtx C) D {
			if !loaded {
				gtx = gtx.Disabled()
			}
			return fill(gtx, colorInput, func(gtx C) D {
				return layout.UniformInset(unit.Dp(5)).Layout(gtx, func(gtx C) D {
					e := material.Editor(s.th, &r.edit, "")
					e.TextSize = unit.Sp(14)
					e.Color = colorFg
					if changed {
						e.Color = colorWarn
					}
					return e.Layout(gtx)
				})
			})
		}
		return s.columns(gtx, name, label(s.th, 14, current, colorMuted).Layout, slider, edit,
			label(s.th, 13, f.Format(f.Min)+" – "+f.Format(f.Max), colorMuted).Layout)
	})
}

// updateRow moves a change made with the slider or the box into s.values,
// and keeps both showing s.values otherwise (after a preset, say).
func (s *State) updateRow(gtx C, r *row, loaded bool) {
	f := r.field
	if r.name.Clicked(gtx) {
		s.doc = docTarget{field: f.Key, preset: -1}
	}
	if !loaded {
		r.edit.SetText("")
		return
	}
	span := f.Max - f.Min
	if r.slider.Update(gtx) {
		s.values[f.Key] = f.Min + math.Round(float64(r.slider.Value)*span)
		s.doc = docTarget{field: f.Key, preset: -1}
	}
	// Only typing changes a value: the box is also rewritten below to follow
	// presets, and that must not turn a value the user never touched into
	// its rounded display.
	focused := gtx.Focused(&r.edit)
	for {
		ev, ok := r.edit.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent, widget.SubmitEvent:
			if v, err := f.Parse(r.edit.Text()); focused && err == nil && v >= f.Min && v <= f.Max && f.Format(v) != f.Format(s.values[f.Key]) {
				s.values[f.Key] = v
			}
		}
	}
	if focused {
		s.doc = docTarget{field: f.Key, preset: -1}
	}
	text := f.Format(s.values[f.Key])
	if !focused && r.edit.Text() != text {
		r.edit.SetText(text)
	}
	if !r.slider.Dragging() {
		r.slider.Value = float32((s.values[f.Key] - f.Min) / span)
	}
}
