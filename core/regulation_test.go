package core

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

func TestRecompressDCXMatchesPS4FrameShape(t *testing.T) {
	bnd4 := bytes.Repeat([]byte("BND4 multiplayer parameter data\x00"), 8192)
	original := make([]byte, 76)
	copy(original[:4], dcxMagic)
	copy(original[40:44], dcxFormatZSTD)

	dcx, err := recompressDCX(original, bnd4)
	if err != nil {
		t.Fatal(err)
	}
	if got := int(binary.BigEndian.Uint32(dcx[28:32])); got != len(bnd4) {
		t.Fatalf("decompressed size = %d, want %d", got, len(bnd4))
	}
	compSize := int(binary.BigEndian.Uint32(dcx[32:36]))
	if compSize != len(dcx)-76 {
		t.Fatalf("compressed size = %d, payload length = %d", compSize, len(dcx)-76)
	}

	stream := dcx[76:]
	if len(stream) < 6 || !bytes.Equal(stream[:4], []byte{0x28, 0xB5, 0x2F, 0xFD}) {
		t.Fatalf("invalid ZSTD frame header: % X", stream)
	}
	fhd := stream[4]
	if fhd&(1<<2) != 0 {
		t.Error("ZSTD content checksum is enabled")
	}
	if fhd&(1<<5) != 0 {
		t.Error("ZSTD frame is single-segment; expected an explicit 64KB window")
	}
	if fhd>>6 != 0 {
		t.Error("ZSTD frame contains a content-size field")
	}

	windowDescriptor := stream[5]
	exponent := int(windowDescriptor >> 3)
	mantissa := int(windowDescriptor & 7)
	windowBase := 1 << (10 + exponent)
	windowSize := windowBase + (windowBase/8)*mantissa
	if windowSize != soulsFormatsWindowSize {
		t.Errorf("ZSTD window = %d, want %d", windowSize, soulsFormatsWindowSize)
	}

	decoded, err := decompressDCX(dcx, dcxFormatZSTD)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, bnd4) {
		t.Error("recompressed DCX did not round-trip")
	}
}

func TestExtractRegulationRejectsNonPS4Layout(t *testing.T) {
	_, _, _, err := extractRegulation(make([]byte, 64))
	if err == nil {
		t.Fatal("expected non-PS4 UserData11 to be rejected")
	}
}

func TestValidateNetworkParamsRejectsNonFiniteValues(t *testing.T) {
	for _, value := range []float32{
		float32(math.NaN()),
		float32(math.Inf(1)),
		float32(math.Inf(-1)),
	} {
		params := NetworkParamDefaults()
		params.SignDownloadSpan = value
		if err := ValidateNetworkParams(params); err == nil {
			t.Fatalf("ValidateNetworkParams accepted non-finite value %v", value)
		}
	}
}

func TestDecompressDCXRejectsInvalidPayloadSize(t *testing.T) {
	dcx := make([]byte, 76)
	copy(dcx[:4], dcxMagic)
	copy(dcx[40:44], dcxFormatZSTD)
	binary.BigEndian.PutUint32(dcx[32:36], 1)

	_, err := decompressDCX(dcx, dcxFormatZSTD)
	if err == nil || !strings.Contains(err.Error(), "exceeds available payload") {
		t.Fatalf("decompressDCX error = %v, want payload bounds error", err)
	}
}
