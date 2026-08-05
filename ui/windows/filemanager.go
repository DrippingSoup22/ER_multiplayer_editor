package windows

import (
	"er_pvp_mod/core"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ps4HeaderSize  = 0x70
	slotSize       = 0x280000
	userData10Size = 0x60000
	userData11Size = 0x240010
)

type loadedInput struct {
	data    []byte
	ud11    []byte
	ud11Off int
	ud11End int
	path    string
}

// tryReadNetworkParams calls core.ReadNetworkParams and converts any panic
// (e.g. a slice-bounds error on garbage data during format probing) into an
// ordinary error so callers can handle it gracefully.
func tryReadNetworkParams(ud11 []byte) (vals *core.NetworkParamValues, err error) {
	defer func() {
		if r := recover(); r != nil {
			vals = nil
			err = fmt.Errorf("parse error: %v", r)
		}
	}()
	return core.ReadNetworkParams(ud11)
}

func loadInputAuto(path string) (*loadedInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// PS4 save (memory.dat): USERDATA11 at a fixed offset.
	ud11Off := ps4HeaderSize + 10*slotSize + userData10Size
	ud11End := ud11Off + userData11Size
	if len(data) >= ud11End {
		ud11 := data[ud11Off:ud11End]
		if vals, err := tryReadNetworkParams(ud11); err == nil && vals != nil {
			return &loadedInput{
				data: data, ud11: ud11,
				ud11Off: ud11Off, ud11End: ud11End, path: path,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported or unrecognized save format: expected a PS4 memory.dat")
}

func defaultEditedFilename(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)) +
		"_edited" + filepath.Ext(path)
}

func writePatchedOutput(loaded *loadedInput, patch core.NetworkParamValues, outPath string) ([]byte, error) {
	patchedUD11, err := core.PatchNetworkParams(loaded.ud11, patch)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(loaded.data))
	copy(out, loaded.data)
	copy(out[loaded.ud11Off:loaded.ud11End], patchedUD11)

	if err := os.WriteFile(outPath, out, 0644); err != nil {
		return nil, err
	}
	return patchedUD11, nil
}
