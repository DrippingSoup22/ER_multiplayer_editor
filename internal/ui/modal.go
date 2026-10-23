package ui

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// modal is a message over a dimmed window. With a confirm label it asks a
// question; without one it only informs. Nothing behind it takes input.
type modal struct {
	title, body string
	confirm     string
	onConfirm   func()

	blocker, ok, cancel widget.Clickable
}

// update reports whether the modal was answered, and how.
func (m *modal) update(gtx C) (done, confirmed bool) {
	if m.ok.Clicked(gtx) {
		return true, true
	}
	return m.cancel.Clicked(gtx), false
}

func (m *modal) layout(gtx C, th *material.Theme) {
	size := gtx.Constraints.Max
	paint.FillShape(gtx.Ops, colorScrim, clip.Rect{Max: size}.Op())
	m.blocker.Layout(gtx, func(gtx C) D { return D{Size: size} })
	layout.Center.Layout(gtx, func(gtx C) D {
		width := gtx.Dp(460)
		gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
		return fill(gtx, colorPanel, func(gtx C) D {
			return layout.UniformInset(unit.Dp(18)).Layout(gtx, func(gtx C) D {
				okLabel, buttons := "OK", []layout.FlexChild{}
				if m.confirm != "" {
					okLabel = m.confirm
					buttons = append(buttons, layout.Rigid(button(th, &m.cancel, "Cancel", true)), layout.Rigid(gap(8)))
				}
				buttons = append(buttons, layout.Rigid(primary(th, &m.ok, okLabel, true)))
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(label(th, 18, m.title, colorAccent).Layout),
					layout.Rigid(gap(10)),
					layout.Rigid(label(th, 14, m.body, colorFg).Layout),
					layout.Rigid(gap(16)),
					layout.Rigid(func(gtx C) D {
						return layout.E.Layout(gtx, func(gtx C) D { return layout.Flex{}.Layout(gtx, buttons...) })
					}),
				)
			})
		})
	})
}

// pt is a size in pixels.
func pt(x, y int) image.Point { return image.Pt(x, y) }
