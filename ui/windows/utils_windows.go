//go:build windows

package windows

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/lxn/win"
)

func utf16Ptr(s string) *uint16 {
	p, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return p
}

func utf16FromMultiString(parts ...string) []uint16 {
	var out []uint16
	for _, p := range parts {
		out = append(out, syscall.StringToUTF16(p)...)
	}
	out = append(out, 0)
	return out
}

func setText(hwnd win.HWND, s string) {
	win.SendMessage(hwnd, win.WM_SETTEXT, 0, uintptr(unsafe.Pointer(utf16Ptr(s))))
}

func getText(hwnd win.HWND) string {
	n := win.SendMessage(hwnd, win.WM_GETTEXTLENGTH, 0, 0)
	buf := make([]uint16, n+1)
	win.SendMessage(hwnd, win.WM_GETTEXT, n+1, uintptr(unsafe.Pointer(&buf[0])))
	return syscall.UTF16ToString(buf)
}

func createStatic(parent win.HWND, text string, x, y, w, h int32, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		0,
		utf16Ptr("STATIC"),
		utf16Ptr(text),
		win.WS_CHILD|win.WS_VISIBLE,
		x, y, w, h,
		parent,
		0,
		instance,
		nil,
	)
}

func createEdit(parent win.HWND, text string, x, y, w, h int32, id int, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		win.WS_EX_CLIENTEDGE,
		utf16Ptr("EDIT"),
		utf16Ptr(text),
		win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP|win.ES_AUTOHSCROLL,
		x, y, w, h,
		parent,
		win.HMENU(uintptr(id)),
		instance,
		nil,
	)
}

func createReadOnlyMultilineEdit(parent win.HWND, text string, x, y, w, h int32, id int, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		win.WS_EX_CLIENTEDGE,
		utf16Ptr("EDIT"),
		utf16Ptr(text),
		win.WS_CHILD|win.WS_VISIBLE|win.WS_VSCROLL|
			win.ES_MULTILINE|win.ES_AUTOVSCROLL|win.ES_READONLY,
		x, y, w, h,
		parent,
		win.HMENU(uintptr(id)),
		instance,
		nil,
	)
}

// createClickableStatic creates a STATIC control with SS_NOTIFY so it generates
// STN_CLICKED via WM_COMMAND when the user clicks on it.
func createClickableStatic(parent win.HWND, text string, x, y, w, h int32, id int, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		0,
		utf16Ptr("STATIC"),
		utf16Ptr(text),
		win.WS_CHILD|win.WS_VISIBLE|win.SS_NOTIFY,
		x, y, w, h,
		parent,
		win.HMENU(uintptr(id)),
		instance,
		nil,
	)
}

func createCheckbox(parent win.HWND, text string, x, y, w, h int32, id int, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		0,
		utf16Ptr("BUTTON"),
		utf16Ptr(text),
		win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP|win.BS_AUTOCHECKBOX,
		x, y, w, h,
		parent,
		win.HMENU(uintptr(id)),
		instance,
		nil,
	)
}

func createButton(parent win.HWND, text string, x, y, w, h int32, id int, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		0,
		utf16Ptr("BUTTON"),
		utf16Ptr(text),
		win.WS_CHILD|win.WS_VISIBLE|win.WS_TABSTOP|win.BS_PUSHBUTTON,
		x, y, w, h,
		parent,
		win.HMENU(uintptr(id)),
		instance,
		nil,
	)
}

func createGroupBox(parent win.HWND, text string, x, y, w, h int32, instance win.HINSTANCE) win.HWND {
	return win.CreateWindowEx(
		0,
		utf16Ptr("BUTTON"),
		utf16Ptr(text),
		win.WS_CHILD|win.WS_VISIBLE|win.BS_GROUPBOX,
		x, y, w, h,
		parent,
		0,
		instance,
		nil,
	)
}

func openFileDialog(owner win.HWND) (string, bool, error) {
	buf := make([]uint16, 4096)
	filter := utf16FromMultiString(
		"Save files (*.dat;*.sl2)", "*.dat;*.sl2",
		"All files (*.*)", "*.*",
	)

	title, err := syscall.UTF16PtrFromString("Open save file")
	if err != nil {
		return "", false, fmt.Errorf("build open title: %w", err)
	}

	ofn := win.OPENFILENAME{}
	ofn.LStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.HwndOwner = owner
	ofn.LpstrFilter = &filter[0]
	ofn.LpstrFile = &buf[0]
	ofn.NMaxFile = uint32(len(buf))
	ofn.LpstrTitle = title
	ofn.Flags = win.OFN_EXPLORER | win.OFN_FILEMUSTEXIST | win.OFN_PATHMUSTEXIST

	if win.GetOpenFileName(&ofn) {
		return syscall.UTF16ToString(buf), true, nil
	}

	code := win.CommDlgExtendedError()
	if code == 0 {
		return "", false, nil
	}
	return "", false, fmt.Errorf("GetOpenFileName failed, CommDlgExtendedError=0x%X", code)
}

func saveFileDialog(owner win.HWND, defaultName string) (string, bool, error) {
	buf := make([]uint16, 4096)
	copy(buf, syscall.StringToUTF16(defaultName))

	filter := utf16FromMultiString(
		"Save files (*.dat;*.sl2)", "*.dat;*.sl2",
		"All files (*.*)", "*.*",
	)

	title, err := syscall.UTF16PtrFromString("Save patched file")
	if err != nil {
		return "", false, fmt.Errorf("build save title: %w", err)
	}

	ofn := win.OPENFILENAME{}
	ofn.LStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.HwndOwner = owner
	ofn.LpstrFilter = &filter[0]
	ofn.LpstrFile = &buf[0]
	ofn.NMaxFile = uint32(len(buf))
	ofn.LpstrTitle = title
	ofn.Flags = win.OFN_EXPLORER | win.OFN_OVERWRITEPROMPT | win.OFN_PATHMUSTEXIST

	if win.GetSaveFileName(&ofn) {
		return syscall.UTF16ToString(buf), true, nil
	}

	code := win.CommDlgExtendedError()
	if code == 0 {
		return "", false, nil
	}
	return "", false, fmt.Errorf("GetSaveFileName failed, CommDlgExtendedError=0x%X", code)
}
