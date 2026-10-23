package ui

import (
	"fmt"
	"image/color"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"

	"github.com/DrippingSoup22/ER_multiplayer_editor/internal/netparam"
)

// layoutDoc explains the selected preset or field.
func (s *State) layoutDoc(gtx C) D {
	var parts []layout.Widget
	heading := func(txt string) {
		parts = append(parts, gap(10), label(s.th, 13, strings.ToUpper(txt), colorMuted).Layout, gap(4))
	}
	paragraphs := func(txt string, c color.NRGBA) {
		// A summary's raise and lower lines read better apart.
		for _, p := range strings.Split(strings.ReplaceAll(txt, "\nRaise", "\n\nRaise"), "\n\n") {
			parts = append(parts, label(s.th, 14, strings.TrimSpace(p), c).Layout, gap(8))
		}
	}
	body := func(txt string) { paragraphs(txt, colorFg) }
	warn := func(txt string) { paragraphs(txt, colorWarn) }

	switch {
	case s.doc.preset >= 0:
		preset := netparam.Presets[s.view][s.doc.preset]
		parts = append(parts, label(s.th, 20, preset.Name+" preset — "+s.view.String(), colorAccent).Layout, gap(6))
		body(preset.Description)
		if !s.advanced {
			body("In the standard view a preset sets only the fields shown; turn on Advanced to set the hidden ones too.")
		}
	case s.doc.field != "":
		f := s.rows[s.doc.field].field
		parts = append(parts, label(s.th, 20, f.Label, colorAccent).Layout)
		switch f.Confidence {
		case netparam.CommunityInferred:
			heading("◇ Community-inferred")
			warn("The effect comes from community testing, not from FromSoftware's documentation.")
		case netparam.Unconfirmed:
			heading("? Unconfirmed")
			warn("The effect in Elden Ring is uncertain: the field may be unused.")
		}
		if f.Warning != "" {
			heading("Warning")
			warn(f.Warning)
		}
		heading("In short")
		body(f.Short)
		heading("In full")
		body(f.Long)
		if s.advanced {
			heading("Technical notes")
			body(f.Details)
			body(fmt.Sprintf("Offset 0x%X · vanilla %s · editable %s – %s", f.Offset, f.Format(f.Default), f.Format(f.Min), f.Format(f.Max)))
		}
	default:
		parts = append(parts, label(s.th, 20, "ER Multiplayer Editor", colorAccent).Layout, gap(6))
		body("Tune how a PS4 save searches for invasions, summon signs and Blue Cipher Ring hunts. " +
			"Open a decrypted memory.dat, pick a view, set values or a preset, and save a new file: the input is never changed.\n\n" +
			"Click a field's name, or use its slider, to read what it does here. ◇ marks a community-inferred field, ? an unconfirmed one.\n\n" +
			"Online on PlayStation the server restores these values the first time a character loads. Load the character, " +
			"quit to the title screen without closing the game, and load it again: the edited values then apply.")
	}
	return layout.UniformInset(unit.Dp(14)).Layout(gtx, func(gtx C) D {
		return material.List(s.th, &s.docList).Layout(gtx, len(parts), func(gtx C, i int) D {
			return parts[i](gtx)
		})
	})
}
