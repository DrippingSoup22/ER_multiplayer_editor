//go:build windows

package windows

import (
	"er_pvp_mod/core"
	"er_pvp_mod/params"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/win"
)

// ---------------------------------------------------------------------------
// Control IDs
// ---------------------------------------------------------------------------

const (
	idPathEdit     = 1001
	idBrowse       = 1002
	idLoad         = 1003
	idSave         = 1004
	idViewCombo    = 1005
	idPresetCombo  = 1006
	idApplyValues  = 1007
	idAdvanced     = 1008
	idUnlockRanges = 1009
	idApplyView    = 1010
	idResetAll     = 1011
	idDocBox       = 1401

	baseRowSliderID = 1200
	baseRowEditID   = 1300
	baseRowLabelID  = 1500 // clickable static labels for parameter rows
	maxVisibleRows  = 8    // rows visible at once; scroll for more

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

// Scrollbar control messages and scroll codes (not all exported by lxn/win).
const (
	sbsVert = 0x0001 // SBS_VERT style for a vertical scrollbar control

	sbmSetRange = 0x00E2 // SBM_SETRANGE: wParam=min, lParam=max
	sbmSetPos   = 0x00E0 // SBM_SETPOS:   wParam=pos, lParam=redraw

	sbLineUp        = 0
	sbLineDown      = 1
	sbPageUp        = 2
	sbPageDown      = 3
	sbThumbPosition = 4
	sbThumbTrack    = 5
	sbTop           = 6
	sbBottom        = 7

	wmMouseWheel = 0x020A // WM_MOUSEWHEEL (not exported by lxn/win)
)

// ---------------------------------------------------------------------------
// Layout constants — edit here to tweak the entire layout in one place
// ---------------------------------------------------------------------------

const (
	winW = 1060
	winH = 480

	margin     = int32(14) // outer margin
	gutter     = int32(8)  // gap between sections
	scrollBarW = int32(18) // vertical scrollbar between param box and doc panel

	// Top bar — file path, browse/load buttons, and save-type indicator in one line
	topBarY = int32(12)
	topBarH = int32(28)

	// Controls bar — view + preset (left) and advanced options (right), no group boxes
	ctrlBarY = int32(48)

	// Parameter group box (left) and documentation group box (right)
	paramGrpY = int32(86)
	leftColW  = int32(640)
	paramGrpW = leftColW
	paramGrpH = int32(290)

	// Documentation group box (right column, offset to leave room for the scrollbar)
	rightColX = margin + leftColW + gutter + scrollBarW + 4 // 684
	rightColW = int32(winW) - rightColX - margin             // 362

	// Column X positions inside the parameter group box
	colLabelX  = int32(28)
	colLabelW  = int32(140)
	colCurX    = int32(172)
	colCurW    = int32(60)
	colSliderX = int32(236)
	colSliderW = int32(220)
	colEditX   = int32(460)
	colEditW   = int32(58)
	colRangeX  = int32(522)
	colRangeW  = int32(118)

	// Column header row
	hdrRowY = paramGrpY + int32(18)
	hdrRowH = int32(16)

	// Parameter rows
	firstRowY = paramGrpY + int32(46)
	rowStride  = int32(30)

	// Action rows — Apply (current view) + Apply All on row 1; Save on row 2
	actionRow1Y = paramGrpY + paramGrpH + int32(8)
	actionRow1H = int32(32)
	actionRow2Y = actionRow1Y + actionRow1H + int32(6)
	actionRow2H = int32(32)
	applyBtnW   = int32(90)
	saveBtnW    = int32(188) // matches combined width of two Apply buttons + gap
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
	meta      params.ParamMeta
}

var (
	mainHwnd win.HWND
	loaded   *loadedInput

	currentView   params.ParamView
	advancedMode  bool
	workingValues *core.NetworkParamValues // committed values — shown in Current column
	stagedValues  *core.NetworkParamValues // in-progress edits — shown in New column

	// advancedWarnShown / unlockRangesWarnShown track whether the one-time
	// informational message for each mode has already been shown this session.
	advancedWarnShown     bool
	unlockRanges          bool
	unlockRangesWarnShown bool

	// selectedPresetKeys remembers the last chosen preset key per view so that
	// switching views and returning does not reset the preset combo to Vanilla.
	selectedPresetKeys = map[params.ParamView]string{
		params.ViewInvader:   "vanilla",
		params.ViewSign:      "vanilla",
		params.ViewSignPlace: "vanilla",
		params.ViewHunter:    "vanilla",
		params.ViewTongue:    "vanilla",
	}

	rowScrollOffset  int     // index of first visible param row
	lastScrollTime   int64   // Unix milliseconds of last mouse-wheel scroll event
	origTrackbarProc uintptr // original WndProc for msctls_trackbar32 (shared by all instances)

	hPathEdit          win.HWND
	hSaveTypeValue     win.HWND
	hViewCombo         win.HWND
	hPresetCombo       win.HWND
	hAdvancedCheck     win.HWND
	hUnlockRangesCheck win.HWND
	hParamScrollBar    win.HWND
	hDocBox            win.HWND
	hStatusValue       win.HWND

	rows [maxVisibleRows]paramRow

	// Segoe UI 9pt — set once, sent to every control.
	guiFont win.HFONT
)

// ---------------------------------------------------------------------------
// Trackbar subclassing — redirect WM_MOUSEWHEEL to the main window
// ---------------------------------------------------------------------------
//
// By default, Win32 routes WM_MOUSEWHEEL to whatever child control is under
// the cursor. Trackbars consume the message and change their own value.
// We subclass every trackbar with a replacement WndProc that intercepts
// WM_MOUSEWHEEL and forwards it to mainHwnd for parameter-list scrolling.

var (
	procSetWindowLongPtrW = syscall.NewLazyDLL("user32.dll").NewProc("SetWindowLongPtrW")
	procCallWindowProcW   = syscall.NewLazyDLL("user32.dll").NewProc("CallWindowProcW")
	procLockWindowUpdate  = syscall.NewLazyDLL("user32.dll").NewProc("LockWindowUpdate")
	trackbarProcPtr       = syscall.NewCallback(trackbarWndProcImpl)
)

const gwlpWndProc = ^uintptr(3) // GWLP_WNDPROC = -4 expressed as uintptr

func subclassTrackbar(hwnd win.HWND) {
	prev, _, _ := procSetWindowLongPtrW.Call(uintptr(hwnd), gwlpWndProc, trackbarProcPtr)
	if origTrackbarProc == 0 && prev != 0 {
		origTrackbarProc = prev
	}
}

func callWindowProc(proc uintptr, hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	r, _, _ := procCallWindowProcW.Call(proc, uintptr(hwnd), uintptr(msg), wParam, lParam)
	return r
}

func lockWindowUpdate(hwnd win.HWND) {
	procLockWindowUpdate.Call(uintptr(hwnd))
}

// trackbarWndProcImpl is the replacement WndProc for all trackbar controls.
// It intercepts WM_MOUSEWHEEL and redirects it to the main window so that
// scrolling the mouse wheel scrolls the parameter list, not the slider value.
func trackbarWndProcImpl(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	if msg == wmMouseWheel {
		win.SendMessage(mainHwnd, wmMouseWheel, wParam, lParam)
		return 0
	}
	return callWindowProc(origTrackbarProc, hwnd, msg, wParam, lParam)
}

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
		win.WS_OVERLAPPED|win.WS_CAPTION|win.WS_SYSMENU|win.WS_MINIMIZEBOX|win.WS_VISIBLE,
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

	case win.WM_VSCROLL:
		if win.HWND(lParam) == hParamScrollBar {
			onParamScroll(uint16(wParam), uint16(wParam>>16))
		}
		return 0

	case wmMouseWheel:
		// Rate-limit to one row per 500 ms so the list scrolls at a comfortable pace.
		now := time.Now().UnixMilli()
		if now-lastScrollTime < 150 {
			return 0
		}
		lastScrollTime = now

		delta := int16(win.HIWORD(uint32(wParam)))
		ps := params.VisibleParamsForView(currentView, advancedMode)
		if len(ps) <= maxVisibleRows {
			return 0
		}
		maxOffset := len(ps) - maxVisibleRows
		// Positive delta = wheel forward (up) = show earlier params = decrease offset.
		if delta > 0 {
			rowScrollOffset--
		} else {
			rowScrollOffset++
		}
		if rowScrollOffset < 0 {
			rowScrollOffset = 0
		}
		if rowScrollOffset > maxOffset {
			rowScrollOffset = maxOffset
		}
		configureRowsForView(currentView, advancedMode)
		refreshView()
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
		case idApplyView:
			applyCurrentViewEdits()
			return 0
		case idResetAll:
			resetToVanilla()
			return 0
		case idAdvanced:
			onAdvancedToggle()
			return 0
		case idUnlockRanges:
			onUnlockRangesToggle()
			return 0
		case idViewCombo:
			if code == win.CBN_SELCHANGE {
				onViewChanged()
			}
			return 0
		case idPresetCombo:
			if code == win.CBN_SELCHANGE {
				updatePresetDoc()
				presets := params.PresetsForView(currentView)
				if idx := comboIndex(hPresetCombo); idx >= 0 && idx < len(presets) {
					selectedPresetKeys[currentView] = presets[idx].Key
				}
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

	// ── TOP BAR ──────────────────────────────────────────────────────────
	// File path, browse/load, and save-type indicator — all file info in one line.

	hFileLbl := createStatic(hwnd, "Save file:", margin, topBarY+5, 65, 18, instance)
	hPathEdit = createEdit(hwnd, "", margin+68, topBarY, 540, topBarH, idPathEdit, instance)
	hBrowseBtn := createButton(hwnd, "Browse…", margin+616, topBarY, 72, topBarH, idBrowse, instance)
	hLoadBtn := createButton(hwnd, "Load", margin+696, topBarY, 52, topBarH, idLoad, instance)
	hSaveTypeLbl := createStatic(hwnd, "Save type:", int32(770), topBarY+5, 70, 18, instance)
	hSaveTypeValue = createStatic(hwnd, "—", int32(844), topBarY+5, 202, 18, instance)

	setFontAll(guiFont, hFileLbl, hPathEdit, hBrowseBtn, hLoadBtn, hSaveTypeLbl, hSaveTypeValue)

	// ── CONTROLS BAR ─────────────────────────────────────────────────────
	// View + Preset on the left. Advanced + Unlock ranges right-aligned.
	// No group boxes — single clean horizontal row.

	hViewLbl := createStatic(hwnd, "View:", margin, ctrlBarY+6, 38, 18, instance)
	hViewCombo = createComboBox(hwnd, margin+42, ctrlBarY+2, 140, 200, idViewCombo, instance)
	hPresetLbl := createStatic(hwnd, "Preset:", margin+196, ctrlBarY+6, 46, 18, instance)
	hPresetCombo = createComboBox(hwnd, margin+246, ctrlBarY+2, 200, 200, idPresetCombo, instance)
	hAdvancedCheck = createCheckbox(hwnd, "Advanced", int32(806), ctrlBarY+6, 90, 22, idAdvanced, instance)
	hUnlockRangesCheck = createCheckbox(hwnd, "Unlock ranges", int32(906), ctrlBarY+6, 136, 22, idUnlockRanges, instance)
	win.ShowWindow(hUnlockRangesCheck, win.SW_HIDE)

	hResetBtn := createButton(hwnd, "Reset to vanilla", margin+246+200+16, ctrlBarY+2, 120, topBarH, idResetAll, instance)
	setFontAll(guiFont, hViewLbl, hViewCombo, hPresetLbl, hPresetCombo, hAdvancedCheck, hUnlockRangesCheck, hResetBtn)

	// ── PARAMETERS GROUP BOX ─────────────────────────────────────────────
	//
	//  Single group box replaces the old "Current values" / "New values" split.
	//  Column headers sit at the top so the user knows what each column is.
	//

	hParamGrp := createGroupBox(hwnd, "Parameters", margin, paramGrpY, paramGrpW, paramGrpH, instance)
	setFont(hParamGrp, guiFont)

	// Column header labels
	hHdrParam := createStatic(hwnd, "Name", colLabelX, hdrRowY, colLabelW, hdrRowH, instance)
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

	// ── PARAM LIST SCROLLBAR ─────────────────────────────────────────────
	// Sits between the param group box and the doc panel. Auto-shown when
	// the current view has more params than fit in the visible rows.

	hParamScrollBar = win.CreateWindowEx(
		0, utf16Ptr("SCROLLBAR"), nil,
		win.WS_CHILD|sbsVert,
		margin+leftColW+gutter, firstRowY-2,
		scrollBarW, int32(maxVisibleRows)*rowStride+4,
		hwnd, 0, instance, nil,
	)

	// ── DOCUMENTATION GROUP BOX ───────────────────────────────────────────
	// Spans the same vertical range as the param box + action bar gap so it
	// aligns flush with the bottom of the param group box.

	docGrpH := actionRow1Y - paramGrpY
	hDocGrp := createGroupBox(hwnd, "Documentation", rightColX, paramGrpY, rightColW, docGrpH, instance)
	hDocBox = createReadOnlyMultilineEdit(hwnd, "", rightColX+4, paramGrpY+20, rightColW-8, docGrpH-26, idDocBox, instance)
	setFontAll(guiFont, hDocGrp, hDocBox)

	// ── ACTION ROWS ───────────────────────────────────────────────────────
	// Row 1: Apply (current view) and Apply All on the left; status to the right.
	// Row 2: Save patched file, aligned directly below the Apply buttons.

	hApplyViewBtn := createButton(hwnd, "Apply", margin, actionRow1Y, applyBtnW, actionRow1H, idApplyView, instance)
	hApplyAllBtn := createButton(hwnd, "Apply All", margin+applyBtnW+gutter, actionRow1Y, applyBtnW, actionRow1H, idApplyValues, instance)
	hStatusValue = createStatic(hwnd, "Enter a save path or click Browse.", margin+2*applyBtnW+2*gutter, actionRow1Y+7, 440, 18, instance)
	hSaveBtn := createButton(hwnd, "Save patched file", margin, actionRow2Y, saveBtnW, actionRow2H, idSave, instance)
	setFontAll(guiFont, hApplyViewBtn, hApplyAllBtn, hSaveBtn, hStatusValue)

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
	addComboItem(hViewCombo, "Find Signs")
	addComboItem(hViewCombo, "Place Sign")
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
	hwnd := win.CreateWindowEx(
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
	// Subclass so WM_MOUSEWHEEL scrolls the parameter list, not the slider value.
	subclassTrackbar(hwnd)
	return hwnd
}

func addComboItem(hwnd win.HWND, s string) {
	win.SendMessage(hwnd, win.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func comboIndex(hwnd win.HWND) int {
	return int(win.SendMessage(hwnd, win.CB_GETCURSEL, 0, 0))
}

func viewByComboIndex(i int) params.ParamView {
	switch i {
	case 1:
		return params.ViewSign
	case 2:
		return params.ViewSignPlace
	case 3:
		return params.ViewHunter
	case 4:
		return params.ViewTongue
	default:
		return params.ViewInvader
	}
}

// ---------------------------------------------------------------------------
// View switching
// ---------------------------------------------------------------------------

func onAdvancedToggle() {
	checked := win.SendMessage(hAdvancedCheck, win.BM_GETCHECK, 0, 0) == win.BST_CHECKED
	if checked && !advancedWarnShown {
		win.MessageBox(mainHwnd,
			utf16Ptr("ADVANCED MODE\r\n\r\n"+
				"Reveals additional parameters hidden in standard view and\r\n"+
				"adds technical documentation to the panel for each field.\r\n\r\n"+
				"Risk: some value combinations can prevent multiplayer from\r\n"+
				"working. Changes are staged in memory — nothing is written\r\n"+
				"to disk until you save. Keep a backup of your save file\r\n"+
				"before applying unusual values."),
			utf16Ptr("Advanced Mode"),
			win.MB_OK|win.MB_ICONINFORMATION,
		)
		advancedWarnShown = true
	}
	advancedMode = checked
	rowScrollOffset = 0
	if !advancedMode {
		// Turning off Advanced also resets Unlock ranges.
		unlockRanges = false
		win.SendMessage(hUnlockRangesCheck, win.BM_SETCHECK, win.BST_UNCHECKED, 0)
		win.ShowWindow(hUnlockRangesCheck, win.SW_HIDE)
	} else {
		win.ShowWindow(hUnlockRangesCheck, win.SW_SHOW)
	}
	configureRowsForView(currentView, advancedMode)
	refreshControlStates()
	refreshView()
	ps := params.VisibleParamsForView(currentView, advancedMode)
	if len(ps) > 0 {
		updateDocForMeta(ps[0])
	}
}

func onUnlockRangesToggle() {
	checked := win.SendMessage(hUnlockRangesCheck, win.BM_GETCHECK, 0, 0) == win.BST_CHECKED
	if checked && !unlockRangesWarnShown {
		win.MessageBox(mainHwnd,
			utf16Ptr("UNLOCK RANGES\r\n\r\n"+
				"Disables all sliders and removes every numerical limit.\r\n"+
				"Values can only be entered by typing in the edit fields\r\n"+
				"— no boundaries apply.\r\n\r\n"+
				"Risk: values outside tested ranges can break matchmaking\r\n"+
				"or corrupt the save data. Apply the Vanilla preset at any\r\n"+
				"time to restore safe defaults. Keep a backup of your\r\n"+
				"save file before proceeding."),
			utf16Ptr("Unlock Ranges"),
			win.MB_OK|win.MB_ICONWARNING,
		)
		unlockRangesWarnShown = true
	}
	unlockRanges = checked
	configureRowsForView(currentView, advancedMode) // update slider ranges and range text
	refreshView()                                   // repopulate values cleared by configureRowsForView
	refreshControlStates()
}

func onViewChanged() {
	switchView(viewByComboIndex(comboIndex(hViewCombo)))
}

func switchView(v params.ParamView) {
	currentView = v
	rowScrollOffset = 0
	configureRowsForView(v, advancedMode)
	rebuildPresetCombo()
	refreshView()

	ps := params.VisibleParamsForView(v, advancedMode)
	if len(ps) > 0 {
		updateDocForMeta(ps[0])
	} else {
		setText(hDocBox, "")
	}
}

func configureRowsForView(v params.ParamView, advanced bool) {
	ps := params.VisibleParamsForView(v, advanced)

	// Clamp scroll offset to a valid position.
	maxOffset := len(ps) - maxVisibleRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if rowScrollOffset > maxOffset {
		rowScrollOffset = maxOffset
	}
	if rowScrollOffset < 0 {
		rowScrollOffset = 0
	}

	// Slice of params visible in the current scroll window.
	end := rowScrollOffset + maxVisibleRows
	if end > len(ps) {
		end = len(ps)
	}
	visible := ps[rowScrollOffset:end]

	for i := 0; i < maxVisibleRows; i++ {
		if i < len(visible) {
			p := visible[i]
			rows[i].meta = p
			setText(rows[i].label, p.Label)
			setText(rows[i].current, "")
			setText(rows[i].edit, "")
			if unlockRanges {
				setText(rows[i].rangeText, p.RangeTextUnlock())
			} else {
				setText(rows[i].rangeText, p.RangeText())
			}
			configureSlider(rows[i].slider, p)
			win.ShowWindow(rows[i].label, win.SW_SHOW)
			win.ShowWindow(rows[i].current, win.SW_SHOW)
			win.ShowWindow(rows[i].slider, win.SW_SHOW)
			win.ShowWindow(rows[i].edit, win.SW_SHOW)
			win.ShowWindow(rows[i].rangeText, win.SW_SHOW)
		} else {
			rows[i].meta = params.ParamMeta{}
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

	updateParamScrollBar(len(ps))
}

// configureSlider sets the trackbar range based on unlockRanges (not advancedMode).
// Advanced mode only controls which parameters are visible and what docs are shown;
// the slider range is exclusively controlled by the Unlock ranges button.
func configureSlider(hwnd win.HWND, meta params.ParamMeta) {
	lo, hi := meta.MinInt(), meta.MaxInt()
	if unlockRanges {
		lo, hi = meta.UnlockMinInt(), meta.UnlockMaxInt()
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
	presets := params.PresetsForView(currentView)
	for _, p := range presets {
		addComboItem(hPresetCombo, p.Label)
	}
	if len(presets) == 0 {
		setText(hDocBox, "No presets available for this view.")
		return
	}
	// Restore the previously selected preset for this view (defaults to "vanilla").
	selIdx := 0
	for i, p := range presets {
		if p.Key == selectedPresetKeys[currentView] {
			selIdx = i
			break
		}
	}
	win.SendMessage(hPresetCombo, win.CB_SETCURSEL, uintptr(selIdx), 0)
	updatePresetDoc()
}

func updatePresetDoc() {
	presets := params.PresetsForView(currentView)
	idx := comboIndex(hPresetCombo)
	if idx < 0 || idx >= len(presets) {
		setText(hDocBox, "")
		return
	}
	p := presets[idx]
	setText(hDocBox, "Preset: "+p.Label+"\r\n\r\nDescription:\r\n"+p.Description)
}

func updateDocForMeta(meta params.ParamMeta) {
	if meta.Key == "" {
		setText(hDocBox, "")
		return
	}
	setText(hDocBox, meta.DocText(advancedMode))
}

// ---------------------------------------------------------------------------
// File load / save
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
	rowScrollOffset = 0
	refreshControlStates()

	switch in.kind {
	case kindPS4Full:
		setText(hSaveTypeValue, "PS4 (memory.dat)")
	case kindPCFull:
		setText(hSaveTypeValue, "PC (ER0000.sl2)")
	}

	refreshView()
	setStatus("File loaded.")
}

// ---------------------------------------------------------------------------
// Field population & apply
// ---------------------------------------------------------------------------

// refreshView redraws all three display columns (Current, New value, Range)
// for every visible row. It uses rows[i].meta rather than re-deriving the
// param list so the scroll offset is already baked in.
func refreshView() {
	if loaded == nil || workingValues == nil || stagedValues == nil {
		return
	}
	for i := 0; i < maxVisibleRows; i++ {
		if rows[i].meta.Key == "" {
			continue // hidden row
		}
		setText(rows[i].current, rows[i].meta.ValueString(*workingValues))
		staged := rows[i].meta.ValueString(*stagedValues)
		setText(rows[i].edit, staged)
		n, _ := strconv.Atoi(staged)
		win.SendMessage(rows[i].slider, tbmSetPos, 1, uintptr(n))
		// Restore range text — cleared by clearFields but not reset by loadPath.
		if unlockRanges {
			setText(rows[i].rangeText, rows[i].meta.RangeTextUnlock())
		} else {
			setText(rows[i].rangeText, rows[i].meta.RangeText())
		}
	}
}

// populateNewValues updates only the edit fields and sliders (the "New value"
// column) without touching the "Current value" display column or workingValues.
func populateNewValues(v core.NetworkParamValues) {
	for i := 0; i < maxVisibleRows; i++ {
		if rows[i].meta.Key == "" {
			continue
		}
		val := rows[i].meta.ValueString(v)
		setText(rows[i].edit, val)
		n, _ := strconv.Atoi(val)
		win.SendMessage(rows[i].slider, tbmSetPos, 1, uintptr(n))
	}
}

// previewPreset applies the selected preset directly to stagedValues (all params,
// including hidden/advanced ones) and updates the New value column. workingValues
// and the Current column are untouched until the user clicks Apply Values.
func previewPreset() {
	if loaded == nil || stagedValues == nil {
		return
	}
	presets := params.PresetsForView(currentView)
	idx := comboIndex(hPresetCombo)
	if idx < 0 || idx >= len(presets) {
		return
	}
	applied, ok := params.ApplyPreset(currentView, presets[idx].Key, *stagedValues)
	if !ok {
		return
	}

	// In basic mode, revert advanced-only parameter changes back to the current
	// staged values — presets only affect parameters the user can actually see.
	filtered := false
	if !advancedMode {
		for _, p := range params.ParamsForView(currentView) {
			if !p.Advanced {
				continue
			}
			if s := p.ValueString(*stagedValues); s != "" {
				if n, err := strconv.Atoi(s); err == nil {
					// Only mark as filtered if the preset actually changed this param.
					if p.ValueString(applied) != s {
						filtered = true
					}
					_ = p.SetFromInt(&applied, n)
				}
			}
		}
	}

	*stagedValues = applied
	populateNewValues(*stagedValues)
	if filtered {
		setStatus("Preset staged — advanced parameters unchanged (enable Advanced mode to apply them).")
	} else {
		setStatus("Preset staged. Click 'Apply values' to commit.")
	}
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
	if unlockRanges {
		lo, hi = meta.UnlockMinInt(), meta.UnlockMaxInt()
	}
	if v < lo || v > hi {
		// Snap the edit field back to the current staged value so it doesn't
		// retain an out-of-range number after focus leaves.
		if code == win.EN_KILLFOCUS && stagedValues != nil {
			setText(rows[idx].edit, rows[idx].meta.ValueString(*stagedValues))
		}
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

func updateParamScrollBar(totalParams int) {
	if hParamScrollBar == 0 {
		return
	}
	if totalParams <= maxVisibleRows {
		win.ShowWindow(hParamScrollBar, win.SW_HIDE)
		return
	}
	win.ShowWindow(hParamScrollBar, win.SW_SHOW)
	maxOffset := totalParams - maxVisibleRows
	win.SendMessage(hParamScrollBar, sbmSetRange, 0, uintptr(maxOffset))
	win.SendMessage(hParamScrollBar, sbmSetPos, uintptr(rowScrollOffset), 1)
}

func onParamScroll(code, thumbPos uint16) {
	ps := params.VisibleParamsForView(currentView, advancedMode)
	total := len(ps)
	if total <= maxVisibleRows {
		return
	}
	maxOffset := total - maxVisibleRows
	switch int(code) {
	case sbLineUp:
		rowScrollOffset--
	case sbLineDown:
		rowScrollOffset++
	case sbPageUp:
		rowScrollOffset -= maxVisibleRows
	case sbPageDown:
		rowScrollOffset += maxVisibleRows
	case sbThumbTrack, sbThumbPosition:
		rowScrollOffset = int(thumbPos)
	case sbTop:
		rowScrollOffset = 0
	case sbBottom:
		rowScrollOffset = maxOffset
	}
	if rowScrollOffset < 0 {
		rowScrollOffset = 0
	}
	if rowScrollOffset > maxOffset {
		rowScrollOffset = maxOffset
	}
	configureRowsForView(currentView, advancedMode)
	refreshView()
}

// applyCurrentEdits validates stagedValues and commits them to workingValues
// for all views. In basic mode, advanced parameters are not committed.
func applyCurrentEdits() {
	if loaded == nil || workingValues == nil || stagedValues == nil {
		setStatus("Load a file first.")
		return
	}
	var valErr error
	if unlockRanges {
		valErr = core.ValidateCrossFieldConstraints(*stagedValues)
	} else {
		valErr = core.ValidateNetworkParams(*stagedValues)
	}
	if valErr != nil {
		setStatus("Validation failed: " + valErr.Error())
		return
	}
	allViews := []params.ParamView{
		params.ViewInvader, params.ViewSign, params.ViewSignPlace, params.ViewHunter, params.ViewTongue,
	}
	for _, v := range allViews {
		for _, p := range params.ParamsForView(v) {
			if !advancedMode && p.Advanced {
				continue
			}
			if s := p.ValueString(*stagedValues); s != "" {
				if n, err := strconv.Atoi(s); err == nil {
					_ = p.SetFromInt(workingValues, n)
				}
			}
		}
	}
	refreshView()
	setStatus("All views applied. Save to write the patched file.")
}

// applyCurrentViewEdits commits staged values for the current view only.
// In basic mode, advanced parameters are not committed.
func applyCurrentViewEdits() {
	if loaded == nil || workingValues == nil || stagedValues == nil {
		setStatus("Load a file first.")
		return
	}
	var valErr error
	if unlockRanges {
		valErr = core.ValidateCrossFieldConstraints(*stagedValues)
	} else {
		valErr = core.ValidateNetworkParams(*stagedValues)
	}
	if valErr != nil {
		setStatus("Validation failed: " + valErr.Error())
		return
	}
	for _, p := range params.ParamsForView(currentView) {
		if !advancedMode && p.Advanced {
			continue
		}
		if s := p.ValueString(*stagedValues); s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				_ = p.SetFromInt(workingValues, n)
			}
		}
	}
	refreshView()
	setStatus(viewDisplayName(currentView) + " values applied.")
}

// resetToVanilla resets all parameters in all views to the shipped defaults.
func resetToVanilla() {
	if loaded == nil {
		setStatus("Load a file first.")
		return
	}
	result := win.MessageBox(mainHwnd,
		utf16Ptr("This will reset ALL parameters in ALL views to the game's\r\nshipped defaults. Your staged and applied edits will be lost.\r\n\r\nContinue?"),
		utf16Ptr("Reset to Vanilla"),
		win.MB_YESNO|win.MB_ICONWARNING|win.MB_DEFBUTTON2,
	)
	if result != win.IDYES {
		return
	}
	defaults := core.NetworkParamDefaults()
	*workingValues = defaults
	*stagedValues = defaults
	refreshView()
	setStatus("All parameters reset to vanilla.")
}

// viewDisplayName returns the human-readable label for a ParamView.
func viewDisplayName(v params.ParamView) string {
	switch v {
	case params.ViewSign:
		return "Find Signs"
	case params.ViewSignPlace:
		return "Place Sign"
	case params.ViewHunter:
		return "Hunter"
	case params.ViewTongue:
		return "Taunter's Tongue"
	default:
		return "Invader"
	}
}

func savePatched(hwnd win.HWND) {
	if loaded == nil || workingValues == nil {
		setStatus("Load a file first.")
		return
	}

	// Warn before showing the file picker if any uncertain parameters were modified.
	if warning := buildSaveWarning(*workingValues); warning != "" {
		result := win.MessageBox(hwnd,
			utf16Ptr(warning),
			utf16Ptr("Save — Uncertain Values Detected"),
			win.MB_YESNO|win.MB_ICONWARNING|win.MB_DEFBUTTON2,
		)
		if result != win.IDYES {
			return
		}
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

// buildSaveWarning checks workingValues against vanilla defaults and returns
// a warning message if any community-inferred or unconfirmed parameters were
// modified. Returns an empty string if everything is safe to save silently.
func buildSaveWarning(vals core.NetworkParamValues) string {
	defaults := core.NetworkParamDefaults()
	unconfirmedCount := 0

	allViews := []params.ParamView{
		params.ViewInvader, params.ViewSign, params.ViewSignPlace, params.ViewHunter, params.ViewTongue,
	}
	for _, v := range allViews {
		for _, p := range params.ParamsForView(v) {
			if p.Confidence != params.Unconfirmed {
				continue
			}
			if p.ValueString(vals) != p.ValueString(defaults) {
				unconfirmedCount++
			}
		}
	}

	if unconfirmedCount == 0 {
		return ""
	}

	noun := func(n int, word string) string {
		if n == 1 {
			return fmt.Sprintf("%d %s", n, word)
		}
		return fmt.Sprintf("%d %ss", n, word)
	}

	msg := "SAVE — UNCONFIRMED VALUES DETECTED\r\n\r\n" +
		"Some of your modified parameters have unconfirmed effects\r\n" +
		"in Elden Ring — their in-game behaviour is unknown or\r\n" +
		"possibly vestigial.\r\n\r\n" +
		"  • " + noun(unconfirmedCount, "parameter") + " with unconfirmed effects\r\n"

	msg += "\r\nResults may differ from what the documentation describes.\r\n" +
		"A backup of your original save is strongly recommended.\r\n\r\n" +
		"Save the file?"

	return msg
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
	if hParamScrollBar != 0 {
		win.ShowWindow(hParamScrollBar, win.SW_HIDE)
	}
	refreshControlStates()
}

// refreshControlStates applies the correct enabled/disabled state to every row
// based on whether a file is loaded and whether Unlock ranges is active.
//
// slider: enabled when a file is loaded AND ranges are not unlocked.
// edit:   enabled whenever a file is loaded.
func refreshControlStates() {
	sliderEnabled := loaded != nil && !unlockRanges
	editEnabled := loaded != nil
	for i := 0; i < maxVisibleRows; i++ {
		win.EnableWindow(rows[i].slider, sliderEnabled)
		win.EnableWindow(rows[i].edit, editEnabled)
	}
}

func setStatus(msg string) {
	setText(hStatusValue, msg)
}
