package ui

import (
	_ "embed"
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/font/opentype"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// The Elden Ring palette of ER Merchant Editor: a near-black warm ground,
// bronze dividers, parchment-gold text.
var (
	colorBg       = color.NRGBA{R: 0x0E, G: 0x0C, B: 0x0A, A: 0xFF}
	colorPanel    = color.NRGBA{R: 0x16, G: 0x13, B: 0x0F, A: 0xFF}
	colorInput    = color.NRGBA{R: 0x1C, G: 0x18, B: 0x12, A: 0xFF}
	colorFg       = color.NRGBA{R: 0xD8, G: 0xCD, B: 0xB4, A: 0xFF}
	colorMuted    = color.NRGBA{R: 0x8C, G: 0x7E, B: 0x63, A: 0xFF}
	colorDivider  = color.NRGBA{R: 0x3A, G: 0x31, B: 0x22, A: 0xFF}
	colorAccent   = color.NRGBA{R: 0xB2, G: 0x93, B: 0x53, A: 0xFF}
	colorWarn     = color.NRGBA{R: 0xC9, G: 0xA4, B: 0x5C, A: 0xFF}
	colorError    = color.NRGBA{R: 0xC0, G: 0x3A, B: 0x2E, A: 0xFF}
	colorDisabled = color.NRGBA{R: 0x35, G: 0x2F, B: 0x25, A: 0xFF}
	colorContrast = color.NRGBA{R: 0x12, G: 0x0F, B: 0x0B, A: 0xFF}
	colorScrim    = color.NRGBA{A: 0xA0}
)

// Lora, SIL Open Font License (fonts/Lora-OFL.txt).
//
//go:embed fonts/Lora.ttf
var loraTTF []byte

func newTheme() *material.Theme {
	face, err := opentype.Parse(loraTTF)
	if err != nil {
		panic(err)
	}
	th := material.NewTheme()
	// The embedded face is added to the system fonts, which still cover the
	// Japanese field names in the technical notes.
	th.Shaper = text.NewShaper(text.WithCollection([]font.FontFace{{Font: face.Font(), Face: face}}))
	th.Face = "Lora"
	th.Palette = material.Palette{Bg: colorBg, Fg: colorFg, ContrastBg: colorAccent, ContrastFg: colorContrast}
	return th
}

type (
	C = layout.Context
	D = layout.Dimensions
)

// fill paints the area w takes with a background color.
func fill(gtx C, bg color.NRGBA, w layout.Widget) D {
	macro := op.Record(gtx.Ops)
	dims := w(gtx)
	call := macro.Stop()
	paint.FillShape(gtx.Ops, bg, clip.Rect{Max: dims.Size}.Op())
	call.Add(gtx.Ops)
	return dims
}

// panel is a padded, filled section of the window.
func panel(gtx C, w layout.Widget) D {
	return fill(gtx, colorPanel, func(gtx C) D {
		return layout.UniformInset(unit.Dp(10)).Layout(gtx, w)
	})
}

// rule is a one-pixel divider across the available width.
func rule(gtx C) D {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
	paint.FillShape(gtx.Ops, colorDivider, clip.Rect{Max: size}.Op())
	return D{Size: size}
}

// button is a flat secondary button; a disabled one draws muted and ignores
// input.
func button(th *material.Theme, clk *widget.Clickable, label string, enabled bool) layout.Widget {
	return styledButton(th, clk, label, enabled, false)
}

// primary is the button for the main action.
func primary(th *material.Theme, clk *widget.Clickable, label string, enabled bool) layout.Widget {
	return styledButton(th, clk, label, enabled, true)
}

func styledButton(th *material.Theme, clk *widget.Clickable, label string, enabled, main bool) layout.Widget {
	return func(gtx C) D {
		b := material.Button(th, clk, label)
		b.TextSize = unit.Sp(14)
		b.Inset = layout.Inset{Top: unit.Dp(7), Bottom: unit.Dp(7), Left: unit.Dp(12), Right: unit.Dp(12)}
		b.Background, b.Color = colorInput, colorFg
		if main {
			b.Background, b.Color = colorAccent, colorContrast
		}
		if !enabled {
			gtx = gtx.Disabled()
			b.Background, b.Color = colorPanel, colorDisabled
		}
		return b.Layout(gtx)
	}
}

// toggle is a button that shows whether it is the selected one of a group.
func toggle(th *material.Theme, clk *widget.Clickable, label string, selected, enabled bool) layout.Widget {
	return func(gtx C) D {
		b := material.Button(th, clk, label)
		b.TextSize = unit.Sp(14)
		b.Inset = layout.Inset{Top: unit.Dp(7), Bottom: unit.Dp(7), Left: unit.Dp(12), Right: unit.Dp(12)}
		b.Background, b.Color = colorInput, colorFg
		if selected {
			b.Background, b.Color = colorAccent, colorContrast
		}
		if !enabled {
			gtx = gtx.Disabled()
			b.Color = colorMuted
		}
		return b.Layout(gtx)
	}
}

func label(th *material.Theme, size float32, txt string, c color.NRGBA) material.LabelStyle {
	l := material.Label(th, unit.Sp(size), txt)
	l.Color = c
	return l
}

// gap is empty space of the given size.
func gap(dp unit.Dp) layout.Widget {
	return func(gtx C) D { return D{Size: image.Pt(gtx.Dp(dp), gtx.Dp(dp))} }
}
