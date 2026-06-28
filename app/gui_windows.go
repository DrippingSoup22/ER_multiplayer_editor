//go:build windows

package app

import (
	"er_pvp_mod/core"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

// ---------------------------------------------------------------------------
// Control IDs
// ---------------------------------------------------------------------------

const (
	idPathEdit    = 1001
	idBrowse      = 1002
	idLoad        = 1003
	idSave        = 1004
	idViewCombo   = 1005
	idPresetCombo = 1006
	idApplyValues = 1007
	idAdvanced    = 1008
	idDocBox      = 1401

	baseRowSliderID = 1200
	baseRowEditID   = 1300
	baseRowLabelID  = 1500 // clickable static labels for parameter rows
	maxVisibleRows  = 12

	// STN_CLICKED is not exported by lxn/win.
	stnClicked = 0

	winClassName = "ERPVPModMainWindow"
)

// Custom WM_APP messages to defer blocking dialogs out of wndProc.
const (
	wmAppBrowse = win.WM_APP + 1
	wmAppSave   = win.WM_APP + 2
)

// Trackbar messages (comctl32).
const (
	tbmGetPos      = win.WM_USER
	tbmSetPos      = win.WM_USER + 5
	tbmSetRangeMin = win.WM_USER + 7
	tbmSetRangeMax = win.WM_USER + 8
	tbmSetTicFreq  = win.WM_USER + 20
)

// ---------------------------------------------------------------------------
// Layout constants — edit here to tweak the entire layout in one place
// ---------------------------------------------------------------------------

const (
	winW = 1060 // wider to accommodate two-column layout
	winH = 660  // shorter — doc panel is on the right, not below params

	margin = int32(14) // outer left/right margin
	gutter = int32(8)  // gap between columns / sections

	// Top bar (file path)
	topBarY = int32(12)
	topBarH = int32(28)

	// Mode + Preset bar — taller to fit the Advanced checkbox
	modeBarY = int32(52)
	modeBarH = int32(78)

	// Left column: parameters
	paramGrpY = int32(142)
	leftColW  = int32(640)
	paramGrpW = leftColW
	paramGrpH = int32(420)

	// Right column: documentation panel
	rightColX = margin + leftColW + gutter // 662
	rightColW = int32(winW) - rightColX - margin // 384

	// Column X positions inside the parameters group box (absolute screen coords)
	colLabelX  = int32(28)  // parameter name
	colLabelW  = int32(140)
	colCurX    = int32(172) // current value
	colCurW    = int32(60)
	colSliderX = int32(236) // trackbar
	colSliderW = int32(220)
	colEditX   = int32(460) // numeric edit
	colEditW   = int32(58)
	colRangeX  = int32(522) // range text
	colRangeW  = int32(118)

	// Column header row
	hdrRowY = paramGrpY + int32(18)
	hdrRowH = int32(16)

	// Parameter rows
	firstRowY = paramGrpY + int32(38)
	rowStride = int32(30)

	// Apply values button (bottom of parameters group box)
	applyValBtnY = paramGrpY + paramGrpH + int32(6)
	applyValBtnW = int32(110)
	applyValBtnH = int32(26)

	// Bottom action bar — directly after the apply button (no doc box below)
	actionBarY = applyValBtnY + applyValBtnH + int32(8)
	actionBarH = int32(32)
)

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

type paramRow struct {
	label     win.HWND
	current   win.HWND
	slider    win.HWND
	edit      win.HWND
	rangeText win.HWND
	meta      ParamMeta
}

var (
	mainHwnd win.HWND
	loaded   *loadedInput

	currentView   ParamView
	advancedMode  bool
	workingValues *core.NetworkParamValues // committed values — shown in Current column
	stagedValues  *core.NetworkParamValues // in-progress edits — shown in New column

	// advancedWarnShown is set the first time the user confirms the Advanced Mode
	// dialog. Subsequent toggles within the same session skip the dialog.
	// Resets to false every time the app starts.
	advancedWarnShown bool

	hPathEdit      win.HWND
	hSaveTypeValue win.HWND
	hViewCombo     win.HWND
	hPresetCombo   win.HWND
	hAdvancedCheck win.HWND
	hDocBox        win.HWND
	hStatusValue   win.HWND

	rows [maxVisibleRows]paramRow

	// Segoe UI 9pt — set once, sent to every control.
	guiFont win.HFONT
)

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

func RunGUI() error {
	// Pin this goroutine to one OS thread — mandatory for Win32 GUI.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	instance := win.GetModuleHandle(nil)
	if instance == 0 {
		return fmt.Errorf("GetModuleHandle failed")
	}

	className := utf16Ptr(winClassName)
	windowTitle := utf16Ptr("ER PvP Mod")

	var wc win.WNDCLASSEX
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.HInstance = instance
	wc.HCursor = win.LoadCursor(0, win.MAKEINTRESOURCE(win.IDC_ARROW))
	wc.HbrBackground = win.HBRUSH(win.COLOR_BTNFACE + 1)
	wc.LpszClassName = className

	if win.RegisterClassEx(&wc) == 0 {
		return fmt.Errorf("RegisterClassEx failed")
	}

	// Total window size: client area is roughly winW×winH; add frame chrome.
	mainHwnd = win.CreateWindowEx(
		0,
		className,
		windowTitle,
		win.WS_OVERLAPPEDWINDOW|win.WS_VISIBLE,
		win.CW_USEDEFAULT, win.CW_USEDEFAULT,
		winW+16, winH+39, // add typical frame/title-bar sizes
		0, 0, instance, nil,
	)
	if mainHwnd == 0 {
		return fmt.Errorf("CreateWindowEx failed")
	}

	var msg win.MSG
	for win.GetMessage(&msg, 0, 0, 0) > 0 {
		win.TranslateMessage(&msg)
		win.DispatchMessage(&msg)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Window procedure
// ---------------------------------------------------------------------------

func wndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	defer func() {
		if r := recover(); r != nil {
			win.MessageBox(0,
				utf16Ptr(fmt.Sprintf("Unexpected error: %v", r)),
				utf16Ptr("ER PvP Mod — Error"),
				win.MB_OK|win.MB_ICONERROR,
			)
		}
	}()
	switch msg {
	case win.WM_CREATE:
		onCreate(hwnd)
		return 0

	case win.WM_HSCROLL:
		onSliderScroll(win.HWND(lParam))
		return 0

	case win.WM_COMMAND:
		id := win.LOWORD(uint32(wParam))
		code := uint16(uint32(wParam) >> 16)

		switch id {
		case idBrowse:
			win.PostMessage(hwnd, wmAppBrowse, 0, 0)
			return 0
		case idLoad:
			loadFromPathEdit()
			return 0
		case idSave:
			win.PostMessage(hwnd, wmAppSave, 0, 0)
			return 0
		case idApplyValues:
			applyCurrentEdits()
			return 0
		case idAdvanced:
			onAdvancedToggle()
			return 0
		case idViewCombo:
			if code == win.CBN_SELCHANGE {
				onViewChanged()
			}
			return 0
		case idPresetCombo:
			if code == win.CBN_SELCHANGE {
				updatePresetDoc()
				if loaded != nil && workingValues != nil {
					previewPreset()
				}
			}
			return 0
		default:
			if id >= baseRowEditID && id < baseRowEditID+maxVisibleRows {
				onRowEditChanged(int(id-baseRowEditID), code)
				return 0
			}
			// Clicking a parameter row label shows its documentation.
			if id >= baseRowLabelID && id < baseRowLabelID+maxVisibleRows {
				if code == stnClicked {
					idx := int(id - baseRowLabelID)
					if rows[idx].meta.Key != "" {
						updateDocForMeta(rows[idx].meta)
					}
				}
				return 0
			}
		}

	case wmAppBrowse:
		browseAndLoad(hwnd)
		return 0

	case wmAppSave:
		savePatched(hwnd)
		return 0

	case win.WM_DESTROY:
		if guiFont != 0 {
			win.DeleteObject(win.HGDIOBJ(guiFont))
		}
		win.PostQuitMessage(0)
		return 0
	}

	return win.DefWindowProc(hwnd, msg, wParam, lParam)
}

// ---------------------------------------------------------------------------
// Font helper
// ---------------------------------------------------------------------------

// createSegoeFont creates a Segoe UI 9pt HFONT.
// Falls back gracefully to the system default if Segoe UI is unavailable.
func createSegoeFont() win.HFONT {
	var lf win.LOGFONT
	lf.LfHeight = -12 // ~9pt at 96 DPI; negative = character height
	lf.LfWeight = win.FW_NORMAL
	lf.LfCharSet = win.DEFAULT_CHARSET
	lf.LfQuality = win.CLEARTYPE_QUALITY
	copy(lf.LfFaceName[:], syscall.StringToUTF16("Segoe UI"))
	return win.CreateFontIndirect(&lf)
}

// setFont sends WM_SETFONT to a control.
func setFont(hwnd win.HWND, font win.HFONT) {
	win.SendMessage(hwnd, win.WM_SETFONT, uintptr(font), 1)
}

// setFontAll is a convenience wrapper for multiple controls.
func setFontAll(font win.HFONT, handles ...win.HWND) {
	for _, h := range handles {
		setFont(h, font)
	}
}

// ---------------------------------------------------------------------------
// onCreate — builds the entire UI
// ---------------------------------------------------------------------------

func onCreate(hwnd win.HWND) {
	instance := win.GetModuleHandle(nil)

	guiFont = createSegoeFont()

	// ── TOP BAR: file path ────────────────────────────────────────────────
	//
	//  [Save file:] [___path edit____________] [Browse] [Load]  [Type: -]
	//

	hFileLbl := createStatic(hwnd, "Save file:", margin, topBarY+5, 65, 18, instance)
	hPathEdit = createEdit(hwnd, "", margin+68, topBarY, 640, topBarH, idPathEdit, instance)
	hBrowseBtn := createButton(hwnd, "Browse…", margin+716, topBarY, 80, topBarH, idBrowse, instance)
	hLoadBtn := createButton(hwnd, "Load", margin+800, topBarY, 56, topBarH, idLoad, instance)

	setFontAll(guiFont, hFileLbl, hPathEdit, hBrowseBtn, hLoadBtn)

	// ── MODE + PRESET BAR ────────────────────────────────────────────────
	//
	//  [ Mode ]                  [ Preset                          ]
	//  View: [combo]             [preset combo          ] [Apply]
	//

	hModeGrp := createGroupBox(hwnd, "Mode", margin, modeBarY, 210, modeBarH, instance)
	hViewLbl := createStatic(hwnd, "View:", margin+14, modeBarY+20, 38, 18, instance)
	hViewCombo = createComboBox(hwnd, margin+56, modeBarY+17, 140, 200, idViewCombo, instance)
	hAdvancedCheck = createCheckbox(hwnd, "Advanced", margin+14, modeBarY+48, 120, 22, idAdvanced, instance)

	hPresetGrp := createGroupBox(hwnd, "Preset", margin+218, modeBarY, 674, modeBarH, instance)
	hPresetLbl := createStatic(hwnd, "Preset:", margin+232, modeBarY+20, 46, 18, instance)
	hPresetCombo = createComboBox(hwnd, margin+282, modeBarY+17, 200, 200, idPresetCombo, instance)

	// Save-type display lives in the freed space to the right of the preset combo.
	hSaveTypeLbl := createStatic(hwnd, "Save type:", margin+500, modeBarY+20, 68, 18, instance)
	hSaveTypeValue = createStatic(hwnd, "—", margin+572, modeBarY+20, 300, 18, instance)

	setFontAll(guiFont,
		hModeGrp, hViewLbl, hViewCombo, hAdvancedCheck,
		hPresetGrp, hPresetLbl, hPresetCombo, hSaveTypeLbl, hSaveTypeValue,
	)

	// ── PARAMETERS GROUP BOX ─────────────────────────────────────────────
	//
	//  Single group box replaces the old "Current values" / "New values" split.
	//  Column headers sit at the top so the user knows what each column is.
	//

	hParamGrp := createGroupBox(hwnd, "Parameters", margin, paramGrpY, paramGrpW, paramGrpH, instance)
	setFont(hParamGrp, guiFont)

	// Column header labels
	hHdrParam := createStatic(hwnd, "Parameter", colLabelX, hdrRowY, colLabelW, hdrRowH, instance)
	hHdrCurrent := createStatic(hwnd, "Current", colCurX, hdrRowY, colCurW, hdrRowH, instance)
	hHdrNew := createStatic(hwnd, "New value", colSliderX, hdrRowY, colSliderW+colEditW+gutter, hdrRowH, instance)
	hHdrRange := createStatic(hwnd, "Range", colRangeX, hdrRowY, colRangeW, hdrRowH, instance)
	setFontAll(guiFont, hHdrParam, hHdrCurrent, hHdrNew, hHdrRange)

	// Parameter rows
	for i := 0; i < maxVisibleRows; i++ {
		y := firstRowY + int32(i)*rowStride
		rows[i].label = createClickableStatic(hwnd, "", colLabelX, y+2, colLabelW, 18, baseRowLabelID+i, instance)
		rows[i].current = createStatic(hwnd, "", colCurX, y+2, colCurW, 18, instance)
		rows[i].slider = createTrackbar(hwnd, colSliderX, y-2, colSliderW, 26, baseRowSliderID+i, instance)
		rows[i].edit = createEdit(hwnd, "", colEditX, y, colEditW, 22, baseRowEditID+i, instance)
		rows[i].rangeText = createStatic(hwnd, "", colRangeX, y+2, colRangeW, 18, instance)
		setFontAll(guiFont,
			rows[i].label, rows[i].current,
			rows[i].slider, rows[i].edit, rows[i].rangeText,
		)
	}

	// "Apply values" sits right-aligned just below the parameters group box
	hApplyVBtn := createButton(
		hwnd, "Apply values",
		margin+paramGrpW-applyValBtnW, applyValBtnY,
		applyValBtnW, applyValBtnH,
		idApplyValues, instance,
	)
	setFont(hApplyVBtn, guiFont)

	// ── RIGHT COLUMN: DOCUMENTATION PANEL ────────────────────────────────
	//
	//  Spans from the top of the parameters group box to the action bar.
	//  Selecting a parameter row (by clicking label or edit) updates this panel.

	docPanelH := actionBarY - paramGrpY - int32(4)
	hDocLbl := createStatic(hwnd, "Documentation", rightColX, paramGrpY, rightColW, 18, instance)
	hDocBox = createReadOnlyMultilineEdit(hwnd, "", rightColX, paramGrpY+20, rightColW, docPanelH-20, idDocBox, instance)
	setFontAll(guiFont, hDocLbl, hDocBox)

	// ── BOTTOM ACTION BAR ─────────────────────────────────────────────────

	hSaveBtn := createButton(hwnd, "Save patched file", margin, actionBarY, 148, actionBarH, idSave, instance)
	hStatusValue = createStatic(hwnd, "Enter a save path or click Browse.", margin+156, actionBarY+7, 600, 18, instance)
	setFontAll(guiFont, hSaveBtn, hStatusValue)

	// ── Initialise combo boxes and default view ───────────────────────────
	initViewCombo()
	switchView(viewByComboIndex(0))
	clearFields()
}

// ---------------------------------------------------------------------------
// Combo box helpers
// ---------------------------------------------------------------------------

func initViewCombo() {
	addComboItem(hViewCombo, "Invader")
	addComboItem(hViewCombo, "Summoning Sign")
	addComboItem(hViewCombo, "Hunter")
	addComboItem(hViewCombo, "Taunter's Tongue")
	win.SendMessage(hViewCombo, win.CB_SETCURSEL, 0, 0)
}

func createComboBox(parent win.HWND, x, y, w, h int32, id int, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		0,
		utf16Ptr("COMBOBOX"),
		nil,
		win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP|win.CBS_DROPDOWNLIST|win.WS_VSCROLL,
		x, y, w, h,
		parent,
		win.HMENU(uintptr(id)),
		instance,
		nil,
	)
}

func createTrackbar(parent win.HWND, x, y, w, h int32, id int, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		0,
		utf16Ptr("msctls_trackbar32"),
		nil,
		win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP,
		x, y, w, h,
		parent,
		win.HMENU(uintptr(id)),
		instance,
		nil,
	)
}

func addComboItem(hwnd win.HWND, s string) {
	win.SendMessage(hwnd, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func comboIndex(hwnd win.HWND) int {
	return int(win.SendMessage(hwnd, win.CB_GETCURSEL, 0, 0))
}

func viewByComboIndex(i int) ParamView {
	switch i {
	case 1:
		return ViewSign
	case 2:
		return ViewHunter
	case 3:
		return ViewTongue
	default:
		return ViewInvader
	}
}

// ---------------------------------------------------------------------------
// View switching
// ---------------------------------------------------------------------------

func onAdvancedToggle() {
	checked := win.SendMessage(hAdvancedCheck, win.BM_GETCHECK, 0, 0) == win.BST_CHECKED
	if checked && !advancedWarnShown {
		const msg = "Advanced mode unlocks hidden parameters and removes the\r\n" +
			"conservative slider limits, allowing values up to the full\r\n" +
			"datatype range.\r\n\r\n" +
			"Extreme combinations can prevent multiplayer sessions from\r\n" +
			"establishing or produce unexpected in-game behaviour.\r\n\r\n" +
			"Only enable this if you know what each parameter does.\r\n\r\n" +
			"Enable Advanced Mode?"
		result := win.MessageBox(mainHwnd,
			utf16Ptr(msg),
			utf16Ptr("Advanced Mode — Read before enabling"),
			win.MB_YESNO|win.MB_ICONWARNING|win.MB_DEFBUTTON2,
		)
		if result != win.IDYES {
			win.SendMessage(hAdvancedCheck, win.BM_SETCHECK, win.BST_UNCHECKED, 0)
			return
		}
	}
	advancedWarnShown = true
	advancedMode = checked
	configureRowsForView(currentView, advancedMode)
	refreshView()
	params := VisibleParamsForView(currentView, advancedMode)
	if len(params) > 0 {
		updateDocForMeta(params[0])
	}
}

func onViewChanged() {
	switchView(viewByComboIndex(comboIndex(hViewCombo)))
}

func switchView(v ParamView) {
	currentView = v
	configureRowsForView(v, advancedMode)
	rebuildPresetCombo()
	refreshView()

	params := VisibleParamsForView(v, advancedMode)
	if len(params) > 0 {
		updateDocForMeta(params[0])
	} else {
		setText(hDocBox, "")
	}
}

func configureRowsForView(v ParamView, advanced bool) {
	params := VisibleParamsForView(v, advanced)
	for i := 0; i < maxVisibleRows; i++ {
		if i < len(params) {
			rows[i].meta = params[i]
			setText(rows[i].label, params[i].Label)
			setText(rows[i].current, "")
			setText(rows[i].edit, "")
			if advanced {
				setText(rows[i].rangeText, params[i].RangeTextAdv())
			} else {
				setText(rows[i].rangeText, params[i].RangeText())
			}
			configureSlider(rows[i].slider, params[i], advanced)
			win.ShowWindow(rows[i].label, win.SW_SHOW)
			win.ShowWindow(rows[i].current, win.SW_SHOW)
			win.ShowWindow(rows[i].slider, win.SW_SHOW)
			win.ShowWindow(rows[i].edit, win.SW_SHOW)
			win.ShowWindow(rows[i].rangeText, win.SW_SHOW)
		} else {
			rows[i].meta = ParamMeta{}
			setText(rows[i].label, "")
			setText(rows[i].current, "")
			setText(rows[i].edit, "")
			setText(rows[i].rangeText, "")
			win.ShowWindow(rows[i].label, win.SW_HIDE)
			win.ShowWindow(rows[i].current, win.SW_HIDE)
			win.ShowWindow(rows[i].slider, win.SW_HIDE)
			win.ShowWindow(rows[i].edit, win.SW_HIDE)
			win.ShowWindow(rows[i].rangeText, win.SW_HIDE)
		}
	}
}

func configureSlider(hwnd win.HWND, meta ParamMeta, advanced bool) {
	lo, hi := meta.MinInt(), meta.MaxInt()
	if advanced {
		lo, hi = meta.AdvMinInt(), meta.AdvMaxInt()
	}
	win.SendMessage(hwnd, tbmSetRangeMin, 1, uintptr(lo))
	win.SendMessage(hwnd, tbmSetRangeMax, 1, uintptr(hi))
	win.SendMessage(hwnd, tbmSetTicFreq, 0, 0)
	win.SendMessage(hwnd, tbmSetPos, 1, uintptr(lo))
}

// ---------------------------------------------------------------------------
// Preset helpers
// ---------------------------------------------------------------------------

func rebuildPresetCombo() {
	win.SendMessage(hPresetCombo, win.CB_RESETCONTENT, 0, 0)
	presets := PresetsForView(currentView)
	for _, p := range presets {
		addComboItem(hPresetCombo, p.Label)
	}
	if len(presets) > 0 {
		win.SendMessage(hPresetCombo, win.CB_SETCURSEL, 0, 0)
		updatePresetDoc()
	} else {
		setText(hDocBox, "No presets available for this view.")
	}
}

func updatePresetDoc() {
	presets := PresetsForView(currentView)
	idx := comboIndex(hPresetCombo)
	if idx < 0 || idx >= len(presets) {
		setText(hDocBox, "")
		return
	}
	p := presets[idx]
	setText(hDocBox, "Preset: "+p.Label+"\r\n\r\nDescription:\r\n"+p.Description)
}

func updateDocForMeta(meta ParamMeta) {
	if meta.Key == "" {
		setText(hDocBox, "")
		return
	}
	setText(hDocBox, meta.DocText(advancedMode))
}

// ---------------------------------------------------------------------------
// File load / save (unchanged logic)
// ---------------------------------------------------------------------------

func browseAndLoad(hwnd win.HWND) {
	path, ok, err := openFileDialog(hwnd)
	if err != nil {
		setStatus("Browse failed: " + err.Error())
		return
	}
	if !ok {
		return
	}
	setText(hPathEdit, path)
	loadPath(path)
}

func loadFromPathEdit() {
	loadPath(getText(hPathEdit))
}

func loadPath(path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		loaded = nil
		workingValues = nil
		clearFields()
		setStatus("Enter a save path or click Browse.")
		return
	}

	in, err := loadInputAuto(path)
	if err != nil {
		loaded = nil
		workingValues = nil
		clearFields()
		setStatus("Load failed: " + err.Error())
		return
	}

	current, err := core.ReadNetworkParams(in.ud11)
	if err != nil {
		loaded = nil
		workingValues = nil
		clearFields()
		setStatus("Read failed: " + err.Error())
		return
	}

	loaded = in
	tmp := *current
	workingValues = &tmp
	staged := *current
	stagedValues = &staged

	switch in.kind {
	case kindPS4Full:
		setText(hSaveTypeValue, "PS4 (memory.dat)")
	case kindPCFull:
		setText(hSaveTypeValue, "PC (ER0000.sl2)")
	default:
		setText(hSaveTypeValue, "Raw USERDATA11")
	}

	refreshView()
	setStatus("File loaded.")
}

// ---------------------------------------------------------------------------
// Field population & apply (unchanged logic)
// ---------------------------------------------------------------------------

// refreshView redraws both columns for the current view using workingValues
// (Current column) and stagedValues (New value column). Call this whenever
// either committed or staged state changes.
func refreshView() {
	if loaded == nil || workingValues == nil || stagedValues == nil {
		return
	}
	params := VisibleParamsForView(currentView, advancedMode)
	for i := 0; i < len(params) && i < maxVisibleRows; i++ {
		setText(rows[i].current, params[i].ValueString(*workingValues))
		staged := params[i].ValueString(*stagedValues)
		setText(rows[i].edit, staged)
		n, _ := strconv.Atoi(staged)
		win.SendMessage(rows[i].slider, tbmSetPos, 1, uintptr(n))
	}
}

// populateNewValues updates only the edit fields and sliders (the "New value"
// column) without touching the "Current value" display column or workingValues.
func populateNewValues(v core.NetworkParamValues) {
	params := VisibleParamsForView(currentView, advancedMode)
	for i := 0; i < len(params) && i < maxVisibleRows; i++ {
		val := params[i].ValueString(v)
		setText(rows[i].edit, val)
		n, _ := strconv.Atoi(val)
		win.SendMessage(rows[i].slider, tbmSetPos, 1, uintptr(n))
	}
}

// previewPreset shows the selected preset's values in the UI (edit fields and
// sliders) without touching workingValues. The actual commit happens when the
// user clicks "Apply values" via applyCurrentEdits.
// previewPreset applies the selected preset directly to stagedValues (all params,
// including hidden/advanced ones) and updates the New value column. workingValues
// and the Current column are untouched until the user clicks Apply Values.
func previewPreset() {
	if loaded == nil || stagedValues == nil {
		return
	}
	presets := PresetsForView(currentView)
	idx := comboIndex(hPresetCombo)
	if idx < 0 || idx >= len(presets) {
		return
	}
	applied, ok := ApplyPreset(currentView, presets[idx].Key, *stagedValues)
	if !ok {
		return
	}
	*stagedValues = applied
	populateNewValues(*stagedValues)
	setStatus("Preset staged. Click 'Apply values' to commit.")
}

func onSliderScroll(slider win.HWND) {
	for i := 0; i < maxVisibleRows; i++ {
		if rows[i].slider == slider && rows[i].meta.Key != "" {
			pos := sliderPos(slider)
			setText(rows[i].edit, strconv.Itoa(pos))
			updateDocForMeta(rows[i].meta)
			if stagedValues != nil {
				_ = rows[i].meta.SetFromInt(stagedValues, pos)
			}
			return
		}
	}
}

func onRowEditChanged(idx int, code uint16) {
	if idx < 0 || idx >= maxVisibleRows {
		return
	}
	if rows[idx].meta.Key == "" {
		return
	}
	if code == win.EN_SETFOCUS {
		updateDocForMeta(rows[idx].meta)
		return
	}
	if code != win.EN_CHANGE && code != win.EN_KILLFOCUS {
		return
	}
	s := strings.TrimSpace(getText(rows[idx].edit))
	if s == "" {
		return
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	meta := rows[idx].meta
	lo, hi := meta.MinInt(), meta.MaxInt()
	if advancedMode {
		lo, hi = meta.AdvMinInt(), meta.AdvMaxInt()
	}
	if v < lo || v > hi {
		return
	}
	win.SendMessage(rows[idx].slider, tbmSetPos, 1, uintptr(v))
	if stagedValues != nil {
		_ = rows[idx].meta.SetFromInt(stagedValues, v)
	}
}

func sliderPos(hwnd win.HWND) int {
	return int(win.SendMessage(hwnd, tbmGetPos, 0, 0))
}

// applyCurrentEdits validates stagedValues and commits them to workingValues,
// updating the Current column across all views at once.
func applyCurrentEdits() {
	if loaded == nil || workingValues == nil || stagedValues == nil {
		setStatus("Load a file first.")
		return
	}
	if err := core.ValidateNetworkParams(*stagedValues); err != nil {
		setStatus("Validation failed: " + err.Error())
		return
	}
	*workingValues = *stagedValues
	refreshView()
	setStatus("All changes applied across all views. Save to write the patched file.")
}

func savePatched(hwnd win.HWND) {
	if loaded == nil || workingValues == nil {
		setStatus("Load a file first.")
		return
	}
	patch := *workingValues
	outPath, ok, err := saveFileDialog(hwnd, defaultEditedFilename(loaded.path))
	if err != nil {
		setStatus("Save dialog failed: " + err.Error())
		return
	}
	if !ok {
		return
	}
	patchedUD11, err := writePatchedOutput(loaded, patch, outPath)
	if err != nil {
		setStatus("Write failed: " + err.Error())
		return
	}
	loaded.ud11 = patchedUD11
	setStatus("Saved: " + outPath)
}

// ---------------------------------------------------------------------------
// Utility
// ---------------------------------------------------------------------------

func clearFields() {
	stagedValues = nil
	setText(hSaveTypeValue, "—")
	for i := 0; i < maxVisibleRows; i++ {
		setText(rows[i].current, "")
		setText(rows[i].edit, "")
		setText(rows[i].rangeText, "")
		win.SendMessage(rows[i].slider, tbmSetPos, 1, 0)
	}
}

func setStatus(msg string) {
	setText(hStatusValue, msg)
}
