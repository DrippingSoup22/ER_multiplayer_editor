package windows

import (
	"encoding/binary"
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

type inputKind string

const (
	kindPS4Full inputKind = "ps4-memory-dat"
	kindPCFull  inputKind = "pc-sl2"
)

type loadedInput struct {
	kind    inputKind
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

// loadPCSave parses an Elden Ring PC save file (.sl2), which is a BND4 container
// holding USERDATA slot blobs. It iterates entries and returns the first one that
// can be successfully parsed as USERDATA11 (the network params slot).
//
// Layout verified from real ER0000.sl2 hex dump:
//
//	0x0C       i32  file count (12 for ER)
//	0x20       i64  entry size (0x20 = 32 bytes for ER)
//	0x31       u8   format byte — bit 0x10 = LongOffsets (64-bit data offsets)
//	entries start at 0x40, each entry:
//	  +0x08    i64  data size
//	  +0x10    u32  data offset (or i64 if LongOffsets bit is set)
//
// Entry data is stored raw (no additional encryption at the BND4 container level).
func loadPCSave(path string, data []byte) (*loadedInput, error) {
	const containerHeaderSize = 0x40
	if len(data) < containerHeaderSize || string(data[:4]) != "BND4" {
		return nil, fmt.Errorf("not a BND4 container")
	}

	fileCount := int(int32(binary.LittleEndian.Uint32(data[0x0C:])))
	entrySize := int(binary.LittleEndian.Uint64(data[0x20:]))
	longOffsets := (data[0x31] & 0x10) != 0

	if fileCount <= 0 || fileCount > 128 || entrySize < 0x14 || entrySize > 0x100 {
		return nil, fmt.Errorf("invalid BND4 header: fileCount=%d entrySize=0x%X", fileCount, entrySize)
	}

	for i := 0; i < fileCount; i++ {
		e := containerHeaderSize + i*entrySize
		if e+entrySize > len(data) {
			break
		}

		rawSize := int64(binary.LittleEndian.Uint64(data[e+0x08:]))
		var rawOff int64
		if longOffsets {
			rawOff = int64(binary.LittleEndian.Uint64(data[e+0x10:]))
		} else {
			rawOff = int64(binary.LittleEndian.Uint32(data[e+0x10:]))
		}

		if rawSize <= 0 || rawOff <= 0 || rawOff+rawSize > int64(len(data)) {
			continue
		}

		off, end := int(rawOff), int(rawOff+rawSize)
		blob := data[off:end]
		if _, err := tryReadNetworkParams(blob); err == nil {
			return &loadedInput{
				kind: kindPCFull, data: data,
				ud11: blob, ud11Off: off, ud11End: end, path: path,
			}, nil
		}
	}
	return nil, fmt.Errorf("no parseable USERDATA11 slot found in PC save")
}

func loadInputAuto(path string) (*loadedInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// PC save (.sl2): BND4 container with USERDATA slot entries.
	if len(data) >= 4 && string(data[:4]) == "BND4" {
		if in, err := loadPCSave(path, data); err == nil {
			return in, nil
		}
	}

	// PS4 save (memory.dat): USERDATA11 at a fixed offset.
	ud11Off := ps4HeaderSize + 10*slotSize + userData10Size
	ud11End := ud11Off + userData11Size
	if len(data) >= ud11End {
		ud11 := data[ud11Off:ud11End]
		if vals, err := tryReadNetworkParams(ud11); err == nil && vals != nil {
			return &loadedInput{
				kind: kindPS4Full, data: data, ud11: ud11,
				ud11Off: ud11Off, ud11End: ud11End, path: path,
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported or unrecognized save format: expected a PC save (ER0000.sl2) or PS4 save (memory.dat)")
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
