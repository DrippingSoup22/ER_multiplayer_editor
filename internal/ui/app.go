// Package ui is the editor's window: load a save, edit the four views'
// values, save a new file. Loading, saving and the native file dialogs run off
// the frame goroutine and hand their results back through State.results.
package ui

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/ncruces/zenity"

	"github.com/DrippingSoup22/ER_multiplayer_editor/internal/netparam"
)

// State is the whole window. Only the frame goroutine touches it.
type State struct {
	window *app.Window
	th     *material.Theme

	save              *netparam.Save
	values            netparam.Values // what Save writes; the file's own are save.Values
	view              netparam.View
	advanced          bool
	explainedAdvanced bool
	doc               docTarget
	status            string
	failed            bool // the status reports an error
	busy              bool // a dialog, load or save is running

	path                                  widget.Editor
	browse, load, reset, revert, saveFile widget.Clickable
	tabs                                  [4]widget.Clickable
	presets                               [3]widget.Clickable
	advancedBox                           widget.Bool
	rows                                  map[string]*row
	table, docList                        widget.List
	modal                                 *modal

	results chan func(*State)
}

// docTarget is what the documentation panel explains: a preset of the current
// view when preset >= 0, else a field.
type docTarget struct {
	field  string
	preset int
}

// New builds the window state.
func New(window *app.Window) *State {
	s := &State{
		window:  window,
		th:      newTheme(),
		rows:    map[string]*row{},
		doc:     docTarget{preset: -1},
		status:  "Open a decrypted PS4 memory.dat to begin.",
		results: make(chan func(*State), 4),
	}
	s.path.SingleLine, s.path.Submit = true, true
	s.table.Axis, s.docList.Axis = layout.Vertical, layout.Vertical
	for _, f := range netparam.Fields {
		r := &row{field: f}
		r.edit.SingleLine, r.edit.Submit = true, true
		s.rows[f.Key] = r
	}
	return s
}

// Load opens path, as the Load button does.
func (s *State) Load(path string) {
	s.path.SetText(path)
	s.startLoad()
}

// Layout draws one frame.
func (s *State) Layout(gtx C) D {
	for drained := false; !drained; {
		select {
		case apply := <-s.results:
			apply(s)
		default:
			drained = true
		}
	}
	s.update(gtx)
	dims := fill(gtx, colorBg, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(s.layoutFileBar),
			layout.Rigid(rule),
			layout.Rigid(s.layoutControls),
			layout.Rigid(rule),
			layout.Flexed(1, func(gtx C) D {
				return layout.Flex{}.Layout(gtx,
					layout.Flexed(0.58, s.layoutTable),
					layout.Rigid(func(gtx C) D {
						return D{Size: pt(gtx.Dp(1), gtx.Constraints.Max.Y)}
					}),
					layout.Flexed(0.42, s.layoutDoc),
				)
			}),
			layout.Rigid(rule),
			layout.Rigid(s.layoutFooter),
		)
	})
	if s.modal != nil {
		s.modal.layout(gtx, s.th)
	}
	return dims
}

// update handles this frame's clicks before anything is drawn.
func (s *State) update(gtx C) {
	if s.modal != nil {
		if done, confirmed := s.modal.update(gtx); done {
			m := s.modal
			s.modal = nil
			if confirmed && m.onConfirm != nil {
				m.onConfirm()
			}
		}
		return
	}
	for {
		ev, ok := s.path.Update(gtx)
		if !ok {
			break
		}
		if _, submit := ev.(widget.SubmitEvent); submit {
			s.startLoad()
		}
	}
	if s.browse.Clicked(gtx) {
		s.startBrowse()
	}
	if s.load.Clicked(gtx) {
		s.startLoad()
	}
	for i := range s.tabs {
		if s.tabs[i].Clicked(gtx) {
			s.view = netparam.Views[i]
			s.doc = docTarget{preset: -1}
		}
	}
	for i := range s.presets {
		if s.presets[i].Clicked(gtx) && s.save != nil {
			s.applyPreset(i)
		}
	}
	if s.advancedBox.Update(gtx) {
		s.advanced = s.advancedBox.Value
		if s.advanced && !s.explainedAdvanced {
			s.explainedAdvanced = true
			s.modal = &modal{title: "Advanced mode", body: "Advanced mode shows the fields hidden in the standard view, and each " +
				"field's technical notes.\n\nSome combinations can stop multiplayer from connecting. Nothing is written " +
				"until you save, and the input file is never changed: keep a backup anyway, and change one group at a time."}
		}
	}
	if s.reset.Clicked(gtx) && s.save != nil {
		s.modal = &modal{title: "Reset to vanilla", confirm: "Reset",
			body: "Set every field in every view to the game's shipped value? Your edits will be lost; the file is changed only when you save.",
			onConfirm: func() {
				s.values = netparam.Defaults()
				s.setStatus("Every field is back to vanilla. Save to write it.", false)
			}}
	}
	if s.revert.Clicked(gtx) && s.save != nil {
		s.values = s.save.Values.Clone()
		s.setStatus("Edits discarded: the values are the file's again.", false)
	}
	if s.saveFile.Clicked(gtx) && s.save != nil && !s.busy {
		s.startSave()
	}
}

func (s *State) applyPreset(i int) {
	preset := netparam.Presets[s.view][i]
	before := s.values
	s.values = preset.Apply(s.view, s.values, s.advanced)
	s.doc = docTarget{preset: i}
	note := ""
	if !s.advanced && !preset.Apply(s.view, before, true).Equal(s.values) {
		note = " Its advanced fields are unchanged: turn on Advanced to set them too."
	}
	s.setStatus(fmt.Sprintf("%s preset set for %s.%s", preset.Name, s.view, note), false)
}

func (s *State) setStatus(message string, failed bool) {
	s.status, s.failed = message, failed
}

// post hands a result to the frame goroutine.
func (s *State) post(apply func(*State)) {
	s.results <- apply
	s.window.Invalidate()
}

func (s *State) startBrowse() {
	if s.busy {
		return
	}
	s.busy = true
	go func() {
		path, err := zenity.SelectFile(zenity.Title("Open a decrypted PS4 save"), zenity.FileFilters{
			{Name: "PS4 saves", Patterns: []string{"*.dat"}, CaseFold: true},
			{Name: "All files", Patterns: []string{"*"}},
		})
		s.post(func(s *State) {
			s.busy = false
			if err == nil {
				s.Load(path)
			} else if !errors.Is(err, zenity.ErrCanceled) {
				s.setStatus("The file dialog failed: "+err.Error(), true)
			}
		})
	}()
}

func (s *State) startLoad() {
	path := strings.Trim(strings.TrimSpace(s.path.Text()), `"`)
	if path == "" || s.busy {
		return
	}
	s.busy = true
	s.setStatus("Loading "+filepath.Base(path)+"…", false)
	go func() {
		save, err := netparam.Open(path)
		s.post(func(s *State) {
			s.busy = false
			if err != nil {
				s.setStatus("Could not open "+filepath.Base(path)+": "+err.Error(), true)
				return
			}
			s.save, s.values = save, save.Values.Clone()
			s.setStatus("Loaded "+filepath.Base(path)+". Edit the values, then save a new file.", false)
		})
	}()
}

func (s *State) startSave() {
	if err := netparam.Validate(s.values); err != nil {
		s.setStatus("Not saved: "+err.Error(), true)
		return
	}
	write := func() {
		s.busy = true
		save, values := s.save, s.values.Clone()
		suggested := strings.TrimSuffix(filepath.Base(save.Path), filepath.Ext(save.Path)) + "_edited" + filepath.Ext(save.Path)
		go func() {
			out, err := zenity.SelectFileSave(zenity.Title("Save the edited file as"), zenity.Filename(suggested), zenity.ConfirmOverwrite())
			if err == nil {
				err = save.Write(out, values)
			}
			var written *netparam.Save
			if err == nil {
				written, err = netparam.Open(out)
			}
			s.post(func(s *State) {
				s.busy = false
				switch {
				case errors.Is(err, zenity.ErrCanceled):
				case err != nil:
					s.setStatus("Not saved: "+err.Error(), true)
				default:
					s.save, s.values = written, written.Values.Clone()
					s.path.SetText(out)
					s.setStatus("Saved "+out+". It is now the open file.", false)
				}
			})
		}()
	}
	if n := s.changedUnconfirmed(); n > 0 {
		s.modal = &modal{title: "Unconfirmed fields", confirm: "Save anyway", onConfirm: write,
			body: fmt.Sprintf("%d changed field(s) have an unconfirmed effect in Elden Ring: the game may ignore them, or "+
				"behave differently from the description. Keep a backup of the original save.", n)}
		return
	}
	write()
}

// changedUnconfirmed counts unconfirmed fields set away from vanilla.
func (s *State) changedUnconfirmed() int {
	n := 0
	for _, f := range netparam.Fields {
		if f.Confidence == netparam.Unconfirmed && s.values[f.Key] != f.Default {
			n++
		}
	}
	return n
}

func (s *State) layoutFileBar(gtx C) D {
	return panel(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(label(s.th, 15, "Save file", colorMuted).Layout),
			layout.Rigid(gap(10)),
			layout.Flexed(1, func(gtx C) D { return input(gtx, s.th, &s.path, "Path to a decrypted PS4 memory.dat") }),
			layout.Rigid(gap(10)),
			layout.Rigid(button(s.th, &s.browse, "Browse…", !s.busy)),
			layout.Rigid(gap(6)),
			layout.Rigid(primary(s.th, &s.load, "Load", !s.busy)),
		)
	})
}

func (s *State) layoutControls(gtx C) D {
	loaded := s.save != nil
	children := []layout.FlexChild{}
	for i, view := range netparam.Views {
		children = append(children, layout.Rigid(toggle(s.th, &s.tabs[i], view.String(), view == s.view, true)), layout.Rigid(gap(4)))
	}
	children = append(children, layout.Rigid(gap(18)), layout.Rigid(label(s.th, 14, "Preset", colorMuted).Layout), layout.Rigid(gap(8)))
	for i, preset := range netparam.Presets[s.view] {
		selected := s.doc.preset == i
		children = append(children, layout.Rigid(toggle(s.th, &s.presets[i], preset.Name, selected, loaded)), layout.Rigid(gap(4)))
	}
	children = append(children,
		layout.Flexed(1, func(gtx C) D { return D{Size: pt(gtx.Constraints.Min.X, 0)} }),
		layout.Rigid(func(gtx C) D {
			c := material.CheckBox(s.th, &s.advancedBox, "Advanced")
			c.Color, c.IconColor = colorFg, colorAccent
			return c.Layout(gtx)
		}),
		layout.Rigid(gap(12)),
		layout.Rigid(button(s.th, &s.reset, "Reset all to vanilla", loaded)),
	)
	return panel(gtx, func(gtx C) D {
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
	})
}

func (s *State) layoutFooter(gtx C) D {
	changed := s.save != nil && !s.save.Values.Equal(s.values)
	return panel(gtx, func(gtx C) D {
		c := colorFg
		if s.failed {
			c = colorError
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, label(s.th, 14, s.status, c).Layout),
			layout.Rigid(gap(10)),
			layout.Rigid(button(s.th, &s.revert, "Discard edits", changed && !s.busy)),
			layout.Rigid(gap(6)),
			layout.Rigid(primary(s.th, &s.saveFile, "Save patched file…", s.save != nil && !s.busy)),
		)
	})
}

// input is a bordered single-line text field.
func input(gtx C, th *material.Theme, ed *widget.Editor, hint string) D {
	return fill(gtx, colorInput, func(gtx C) D {
		return layout.Inset{Top: unit.Dp(7), Bottom: unit.Dp(7), Left: unit.Dp(8), Right: unit.Dp(8)}.Layout(gtx, func(gtx C) D {
			e := material.Editor(th, ed, hint)
			e.Color, e.HintColor = colorFg, colorMuted
			e.TextSize = unit.Sp(14)
			return e.Layout(gtx)
		})
	})
}
