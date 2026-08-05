package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unicode/utf16"

	"github.com/klauspost/compress/zstd"
)

// regulationKey is the AES-256-CBC key used to encrypt/decrypt regulation.bin
// inside UserData11 in a PS4 save.
var regulationKey = []byte{
	0x99, 0xBF, 0xFC, 0x36, 0x6A, 0x6B, 0xC8, 0xC6,
	0xF5, 0x82, 0x7D, 0x09, 0x36, 0x02, 0xD6, 0x76,
	0xC4, 0x28, 0x92, 0xA0, 0x1C, 0x20, 0x7F, 0xB0,
	0x24, 0xD3, 0xAF, 0x4E, 0x49, 0x3F, 0xEF, 0x99,
}

const (
	ud11UnkSize = 0x10

	dcxMagic  = "DCX\x00"
	bnd4Magic = "BND4"

	dcxFormatZSTD = "ZSTD"

	// SoulsFormats writes FromSoftware ZSTD streams with windowLog=16.
	soulsFormatsWindowSize = 1 << 16

	networkParamName = "NetworkParam.param"

	// Byte offsets within NETWORK_PARAM_ST (Row 0 data).
	// Calculated sequentially from NetworkParam.xml PARAMDEF.

	// Group: Common (NET_COMMON_PARAM, SortID 0–3)
	offsetSummonTimeoutTime = 0x08 // f32 — SortID 2 — 召喚希望タイムアウト時間[秒]

	// Group: Summon Signs (NET_SUMMON_SIGN_PARAM, SortIDs 102–124)
	// SortID 102 — sparse-area refresh (low sign density); SortID 103 — standard refresh.
	offsetReloadSignIntervalTime1 = 0x18 // f32 — SortID 102 — 召喚サイン再取得待機時間(過疎時)[秒]
	offsetReloadSignIntervalTime2 = 0x1C // f32 — SortID 103 — 召喚サイン再取得待機時間[秒]
	offsetReloadSignTotalCount    = 0x20 // s32 — SortID 105 — 召喚サイン所持可能数上限(全体)
	offsetReloadSignCellCount     = 0x24 // s32 — SortID 106 — 召喚サイン所持可能数上限(セル)
	offsetUpdateSignIntervalTime  = 0x28 // f32 — SortID 107 — 召喚サイン更新待機時間[秒]
	// 0x2C–0x3B: render-only fields (basicExclusiveRange/Height, previewChrWaitingTime, signVisibleRange).
	// Cell-range fields are stored as u32 (4 bytes each) — XML note "実際はu8" means the value fits u8,
	// not that the storage is 1 byte. Layout: 3×u32 (12 B) + 6×f32 keyword-scale (24 B) = 36 B → singGetMax at 0x60 ✓
	offsetSignCellRangeH    = 0x3C // u32 — SortID 113 — 召喚サイン取得セル範囲(水平)
	offsetSignCellRangeUp   = 0x40 // u32 — SortID 114 — 召喚サイン取得セル範囲(上方向)
	offsetSignCellRangeDown = 0x44 // u32 — SortID 115 — 召喚サイン取得セル範囲(下方向)
	offsetSingGetMax        = 0x60 // s32 — SortID 122 — サイン取得上限
	offsetSignDownloadSpan  = 0x64 // f32 — SortID 123 — サインダウンロードスパン
	offsetSignUpdateSpan    = 0x68 // f32 — SortID 124 — サインアップロードスパン

	// Group: Break-In / Invasions (NET_BREAKIN_PARAM, SortIDs 200–203)
	offsetMaxBreakInTargets = 0x70
	offsetBreakInInterval   = 0x74
	offsetBreakInTimeout    = 0x78
	// 0x7C is labelled "dummy8 pad[4]" in PARAMDEF but holds vanilla=5 — it is the
	// undocumented breakInRequestAreaCount field (hidden from editors by FromSoftware).
	offsetBreakInAreaCount = 0x7C

	// Group: Visit / Blue Phantom (NET_VISIT_PARAM, SortIDs 600–610)
	offsetReloadVisitListCoolTime = 0x180
	offsetMaxCoopBlueSummonCount  = 0x184
	offsetMaxVisitListCount       = 0x18C
	offsetReloadSearchCoopBlueMin = 0x190
	offsetReloadSearchCoopBlueMax = 0x194

	// Group: Extra (NET_EXTRA_PARAM, SortID 730)
	offsetAllAreaSearchRateCoopBlue  = 0x1D8 // u8 — all-area search rate, co-op blue  [0-100]
	offsetAllAreaSearchRateVsBlue    = 0x1D9 // u8 — all-area search rate, hunter blue  [0-100]
	offsetAllAreaSearchRateBellGuard = 0x1DA // u8 — all-area search rate, bell-guardian [0-100]

	// Group: Visitor (VISITOR, SortIDs 1500–1530) — not exposed as a UI view;
	// no confirmed link to any specific in-game item or mechanic. Parsed and
	// written transparently to preserve byte-for-byte round-tripping.
	offsetVisitorListMax      = 0x240
	offsetVisitorTimeOutTime  = 0x244
	offsetVisitorDownloadSpan = 0x248
)

// NetworkParamValues holds tunable PvP/multiplayer parameters grouped by player role.
type NetworkParamValues struct {
	// --- Invader role ---
	MaxBreakInTargetListCount     int32   `json:"maxBreakInTargetListCount"`
	BreakInRequestIntervalTimeSec float32 `json:"breakInRequestIntervalTimeSec"`
	BreakInRequestTimeOutSec      float32 `json:"breakInRequestTimeOutSec"`
	BreakInRequestAreaCount       int32   `json:"breakInRequestAreaCount"`

	// --- Cooperator role (summon signs) ---
	SummonTimeoutTime       float32 `json:"summonTimeoutTime"`
	ReloadSignIntervalTime1 float32 `json:"reloadSignIntervalTime1"`
	ReloadSignIntervalTime2 float32 `json:"reloadSignIntervalTime2"`
	ReloadSignTotalCount    int32   `json:"reloadSignTotalCount"`
	ReloadSignCellCount     int32   `json:"reloadSignCellCount"`
	UpdateSignIntervalTime  float32 `json:"updateSignIntervalTime"`
	SignCellRangeH          int32   `json:"signCellRangeH"`
	SignCellRangeUp         int32   `json:"signCellRangeUp"`
	SignCellRangeDown       int32   `json:"signCellRangeDown"`
	SingGetMax              int32   `json:"singGetMax"`
	SignDownloadSpan        float32 `json:"signDownloadSpan"`
	SignUpdateSpan          float32 `json:"signUpdateSpan"`

	// --- Blue role (Blue Cipher Ring) ---
	ReloadVisitListCoolTime    float32 `json:"reloadVisitListCoolTime"`
	MaxCoopBlueSummonCount     int32   `json:"maxCoopBlueSummonCount"`
	MaxVisitListCount          int32   `json:"maxVisitListCount"`
	ReloadSearchCoopBlueMin    float32 `json:"reloadSearchCoopBlueMin"`
	ReloadSearchCoopBlueMax    float32 `json:"reloadSearchCoopBlueMax"`
	AllAreaSearchRateCoopBlue  int32   `json:"allAreaSearchRateCoopBlue"`
	AllAreaSearchRateVsBlue    int32   `json:"allAreaSearchRateVsBlue"`
	AllAreaSearchRateBellGuard int32   `json:"allAreaSearchRateBellGuard"`

	// --- Visitor group (not exposed as a UI view; see comment near offsets above) ---
	VisitorListMax      int32   `json:"visitorListMax"`
	VisitorTimeOutTime  float32 `json:"visitorTimeOutTime"`
	VisitorDownloadSpan float32 `json:"visitorDownloadSpan"`
}

// NetworkParamDefaults returns the vanilla game defaults for all fields.
func NetworkParamDefaults() NetworkParamValues {
	return NetworkParamValues{
		// Invader
		MaxBreakInTargetListCount:     5,
		BreakInRequestIntervalTimeSec: 30.0,
		BreakInRequestTimeOutSec:      20.0,
		BreakInRequestAreaCount:       5,
		// Cooperator
		SummonTimeoutTime:       45.0,
		ReloadSignIntervalTime1: 0.0,
		ReloadSignIntervalTime2: 60.0,
		ReloadSignTotalCount:    20,
		ReloadSignCellCount:     10,
		UpdateSignIntervalTime:  30.0,
		SignCellRangeH:          1,
		SignCellRangeUp:         1,
		SignCellRangeDown:       1,
		SingGetMax:              32,
		SignDownloadSpan:        30.0,
		SignUpdateSpan:          60.0,
		// Blue
		ReloadVisitListCoolTime:    20.0,
		MaxCoopBlueSummonCount:     2,
		MaxVisitListCount:          5,
		ReloadSearchCoopBlueMin:    30.0,
		ReloadSearchCoopBlueMax:    180.0,
		AllAreaSearchRateCoopBlue:  30,
		AllAreaSearchRateVsBlue:    30,
		AllAreaSearchRateBellGuard: 30,
		// Host
		VisitorListMax:      10,
		VisitorTimeOutTime:  60.0,
		VisitorDownloadSpan: 60.0,
	}
}

// NetworkParamFastInvasions returns the "Fast Invasions" preset (Invader role).
func NetworkParamFastInvasions() NetworkParamValues {
	d := NetworkParamDefaults()
	d.MaxBreakInTargetListCount = 10
	d.BreakInRequestIntervalTimeSec = 4.0
	d.BreakInRequestTimeOutSec = 4.0
	return d
}

// NetworkParamLightInvasions returns the "Light / Safer" invasions preset.
// Moderate speed-up with a slightly lower detection surface than Fast Invasions.
func NetworkParamLightInvasions() NetworkParamValues {
	d := NetworkParamDefaults()
	d.MaxBreakInTargetListCount = 8
	d.BreakInRequestIntervalTimeSec = 10.0
	d.BreakInRequestTimeOutSec = 8.0
	return d
}

// --- Functional presets (v0.9 unified system) ---
//
// These three presets are the single source of truth for the active Network UI
// (frontend fetches them via GetNetworkPreset). Each preset touches only its own
// role group; the unconfirmed 0x7C field (BreakInRequestAreaCount) is always kept
// at vanilla 5. They replace the removed global "Aggressive" profile.

// NetworkParamFasterReds returns the "Faster Reds" preset (Invader role).
// Faster red invasion search without the regressions of the removed Aggressive
// preset: the request timeout stays generous (15s) so near-and-far matchmaking
// can complete, and the unconfirmed 0x7C field stays at vanilla 5.
func NetworkParamFasterReds() NetworkParamValues {
	d := NetworkParamDefaults()
	d.MaxBreakInTargetListCount = 8
	d.BreakInRequestIntervalTimeSec = 12.0
	d.BreakInRequestTimeOutSec = 15.0
	// BreakInRequestAreaCount (0x7C) intentionally left at vanilla 5 — semantics unconfirmed.
	return d
}

// NetworkParamFasterSummons returns the "Faster Summons & Pools" preset (Cooperator role).
// Faster sign download/upload and a larger, internally consistent sign buffer
// (cellCount <= totalCount <= singGetMax). Spatial cellGroup ranges are left at
// vanilla — they are an Experimental option, not part of this preset.
func NetworkParamFasterSummons() NetworkParamValues {
	d := NetworkParamDefaults()
	d.ReloadSignIntervalTime2 = 20.0
	d.ReloadSignTotalCount = 40
	d.ReloadSignCellCount = 20
	d.UpdateSignIntervalTime = 15.0
	d.SingGetMax = 64
	d.SignDownloadSpan = 15.0
	d.SignUpdateSpan = 20.0
	return d
}

// NetworkParamFasterBlue returns the "Faster Blue / Hunter" preset (Blue role).
// Faster and wider co-op blue search. MaxCoopBlueSummonCount stays at vanilla 2
// (the server caps actual joins; raising it only inflates client-side search).
// AllAreaSearchRateVsBlue stays at vanilla 30 (retribution blue likely legacy in ER).
func NetworkParamFasterBlue() NetworkParamValues {
	d := NetworkParamDefaults()
	d.ReloadVisitListCoolTime = 8.0
	d.ReloadSearchCoopBlueMin = 10.0
	d.ReloadSearchCoopBlueMax = 40.0
	d.MaxVisitListCount = 10
	d.AllAreaSearchRateCoopBlue = 60
	return d
}

// --- Aggressive functional presets (v0.10) ---
//
// Stronger, clearly observable per-group profiles for PlayStation PvP testing. Like the
// Faster presets they are strictly modular: each touches only its own role group,
// never the unconfirmed 0x7C field (BreakInRequestAreaCount), never Visitor
// fields, never another group. They do NOT restore the removed global "Aggressive"
// profile — there is no cross-group Aggressive and no aggressive-host.

// NetworkParamAggressiveReds returns the "Aggressive Reds" preset (Invader role).
// Maximal red invasion target throughput for Bloody/Recusant Finger testing while
// keeping a safe timeout (12s, not the old broken 3s) and leaving 0x7C at vanilla 5.
// Target-candidate proxy: (60/8)*12 = 90/min vs Faster 40/min vs Vanilla 10/min.
func NetworkParamAggressiveReds() NetworkParamValues {
	d := NetworkParamDefaults()
	d.MaxBreakInTargetListCount = 12
	d.BreakInRequestIntervalTimeSec = 8.0
	d.BreakInRequestTimeOutSec = 12.0
	// BreakInRequestAreaCount (0x7C) intentionally left at vanilla 5 — semantics unconfirmed.
	return d
}

// NetworkParamAggressiveSummons returns the "Aggressive Summon Signs" preset
// (Cooperator role). Fast sign refresh/upload with the largest internally
// consistent sign buffer (cellCount <= totalCount <= singGetMax: 32 <= 64 <= 64).
// Timers set at 15 s — fastest cadence that gives cross-region P2P handshakes
// room to complete. Controls the summon-sign network path only; Summoning Pool
// activation is a separate World/Exploration feature. cellGroup* left untouched.
func NetworkParamAggressiveSummons() NetworkParamValues {
	d := NetworkParamDefaults()
	d.ReloadSignIntervalTime2 = 15.0
	d.ReloadSignTotalCount = 64
	d.ReloadSignCellCount = 32
	d.UpdateSignIntervalTime = 10.0
	d.SingGetMax = 64
	d.SignDownloadSpan = 15.0
	d.SignUpdateSpan = 10.0
	return d
}

// NetworkParamAggressiveBlue returns the "Aggressive Blue / Hunter" preset (Blue
// role). Fastest, widest co-op blue search for hunter-side testing with an active
// Blue Cipher Ring. MaxCoopBlueSummonCount (Blue Search Parallelism) and
// AllAreaSearchRateVsBlue (Retribution) stay vanilla — they are Experimental.
func NetworkParamAggressiveBlue() NetworkParamValues {
	d := NetworkParamDefaults()
	d.ReloadVisitListCoolTime = 5.0
	d.ReloadSearchCoopBlueMin = 5.0
	d.ReloadSearchCoopBlueMax = 20.0
	d.MaxVisitListCount = 15
	d.AllAreaSearchRateCoopBlue = 100
	return d
}

// NetworkParamFastSummons returns the "Fast Summons" preset (Cooperator role). Experimental.
func NetworkParamFastSummons() NetworkParamValues {
	d := NetworkParamDefaults()
	d.ReloadSignIntervalTime2 = 15.0
	d.ReloadSignTotalCount = 40
	d.ReloadSignCellCount = 20
	d.UpdateSignIntervalTime = 10.0
	d.SingGetMax = 64
	d.SignDownloadSpan = 10.0
	d.SignUpdateSpan = 15.0
	return d
}

// NetworkParamFastBlue returns the "Fast Blue" preset (Blue role). Experimental.
func NetworkParamFastBlue() NetworkParamValues {
	d := NetworkParamDefaults()
	d.ReloadVisitListCoolTime = 5.0
	d.MaxCoopBlueSummonCount = 2
	d.MaxVisitListCount = 15
	d.ReloadSearchCoopBlueMin = 10.0
	d.ReloadSearchCoopBlueMax = 30.0
	d.AllAreaSearchRateCoopBlue = 75
	d.AllAreaSearchRateVsBlue = 75
	return d
}

// NetworkParamAggressiveHost returns the "Aggressive Host" preset (Host role). Experimental.
func NetworkParamAggressiveHost() NetworkParamValues {
	d := NetworkParamDefaults()
	d.VisitorListMax = 20
	d.VisitorTimeOutTime = 60.0
	d.VisitorDownloadSpan = 10.0
	return d
}

// NetworkParamFast returns the legacy "Fast Invasions" preset for backward compatibility.
func NetworkParamFast() NetworkParamValues {
	return NetworkParamFastInvasions()
}

// ReadNetworkParams extracts current NetworkParam values from UserData11.
func ReadNetworkParams(ud11 []byte) (*NetworkParamValues, error) {
	paramData, _, _, err := locateNetworkParam(ud11)
	if err != nil {
		return nil, err
	}

	vals := &NetworkParamValues{}

	// Invader
	vals.MaxBreakInTargetListCount = int32(binary.LittleEndian.Uint32(paramData[offsetMaxBreakInTargets:]))
	vals.BreakInRequestIntervalTimeSec = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetBreakInInterval:]))
	vals.BreakInRequestTimeOutSec = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetBreakInTimeout:]))
	vals.BreakInRequestAreaCount = int32(binary.LittleEndian.Uint32(paramData[offsetBreakInAreaCount:]))

	// Cooperator
	vals.SummonTimeoutTime = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetSummonTimeoutTime:]))
	vals.ReloadSignIntervalTime1 = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetReloadSignIntervalTime1:]))
	vals.ReloadSignIntervalTime2 = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetReloadSignIntervalTime2:]))
	vals.ReloadSignTotalCount = int32(binary.LittleEndian.Uint32(paramData[offsetReloadSignTotalCount:]))
	vals.ReloadSignCellCount = int32(binary.LittleEndian.Uint32(paramData[offsetReloadSignCellCount:]))
	vals.UpdateSignIntervalTime = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetUpdateSignIntervalTime:]))
	vals.SignCellRangeH = int32(binary.LittleEndian.Uint32(paramData[offsetSignCellRangeH:]))
	vals.SignCellRangeUp = int32(binary.LittleEndian.Uint32(paramData[offsetSignCellRangeUp:]))
	vals.SignCellRangeDown = int32(binary.LittleEndian.Uint32(paramData[offsetSignCellRangeDown:]))
	vals.SingGetMax = int32(binary.LittleEndian.Uint32(paramData[offsetSingGetMax:]))
	vals.SignDownloadSpan = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetSignDownloadSpan:]))
	vals.SignUpdateSpan = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetSignUpdateSpan:]))

	// Blue
	vals.ReloadVisitListCoolTime = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetReloadVisitListCoolTime:]))
	vals.MaxCoopBlueSummonCount = int32(binary.LittleEndian.Uint32(paramData[offsetMaxCoopBlueSummonCount:]))
	vals.MaxVisitListCount = int32(binary.LittleEndian.Uint32(paramData[offsetMaxVisitListCount:]))
	vals.ReloadSearchCoopBlueMin = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetReloadSearchCoopBlueMin:]))
	vals.ReloadSearchCoopBlueMax = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetReloadSearchCoopBlueMax:]))
	vals.AllAreaSearchRateCoopBlue = int32(paramData[offsetAllAreaSearchRateCoopBlue])
	vals.AllAreaSearchRateVsBlue = int32(paramData[offsetAllAreaSearchRateVsBlue])
	vals.AllAreaSearchRateBellGuard = int32(paramData[offsetAllAreaSearchRateBellGuard])

	// Host
	vals.VisitorListMax = int32(binary.LittleEndian.Uint32(paramData[offsetVisitorListMax:]))
	vals.VisitorTimeOutTime = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetVisitorTimeOutTime:]))
	vals.VisitorDownloadSpan = math.Float32frombits(binary.LittleEndian.Uint32(paramData[offsetVisitorDownloadSpan:]))

	return vals, nil
}

// ValidateNetworkParams checks all field boundaries. Returns nil if valid.
func ValidateNetworkParams(p NetworkParamValues) error {
	// Invader
	if p.MaxBreakInTargetListCount < 1 || p.MaxBreakInTargetListCount > 20 {
		return fmt.Errorf("maxBreakInTargetListCount must be 1-20, got %d", p.MaxBreakInTargetListCount)
	}
	if !finiteInRange(p.BreakInRequestIntervalTimeSec, 4.0, 30.0) {
		return fmt.Errorf("breakInRequestIntervalTimeSec must be 4-30, got %.0f", p.BreakInRequestIntervalTimeSec)
	}
	if !finiteInRange(p.BreakInRequestTimeOutSec, 4.0, 20.0) {
		return fmt.Errorf("breakInRequestTimeOutSec must be 4-20, got %.0f", p.BreakInRequestTimeOutSec)
	}
	if p.BreakInRequestAreaCount < 1 || p.BreakInRequestAreaCount > 20 {
		return fmt.Errorf("breakInRequestAreaCount must be 1-20, got %d", p.BreakInRequestAreaCount)
	}
	// Cooperator
	if !finiteInRange(p.SummonTimeoutTime, 10.0, 120.0) {
		return fmt.Errorf("summonTimeoutTime must be 10-120, got %.0f", p.SummonTimeoutTime)
	}
	if !finiteInRange(p.ReloadSignIntervalTime1, 0.0, 120.0) {
		return fmt.Errorf("reloadSignIntervalTime1 must be 0-120, got %.0f", p.ReloadSignIntervalTime1)
	}
	if !finiteInRange(p.ReloadSignIntervalTime2, 10.0, 60.0) {
		return fmt.Errorf("reloadSignIntervalTime2 must be 10-60, got %.0f", p.ReloadSignIntervalTime2)
	}
	if p.ReloadSignTotalCount < 5 || p.ReloadSignTotalCount > 64 {
		return fmt.Errorf("reloadSignTotalCount must be 5-64, got %d", p.ReloadSignTotalCount)
	}
	if p.ReloadSignCellCount < 1 || p.ReloadSignCellCount > 32 {
		return fmt.Errorf("reloadSignCellCount must be 1-32, got %d", p.ReloadSignCellCount)
	}
	if !finiteInRange(p.UpdateSignIntervalTime, 10.0, 30.0) {
		return fmt.Errorf("updateSignIntervalTime must be 10-30, got %.0f", p.UpdateSignIntervalTime)
	}
	if p.SignCellRangeH < 1 || p.SignCellRangeH > 5 {
		return fmt.Errorf("signCellRangeH must be 1-5, got %d", p.SignCellRangeH)
	}
	if p.SignCellRangeUp < 1 || p.SignCellRangeUp > 5 {
		return fmt.Errorf("signCellRangeUp must be 1-5, got %d", p.SignCellRangeUp)
	}
	if p.SignCellRangeDown < 1 || p.SignCellRangeDown > 5 {
		return fmt.Errorf("signCellRangeDown must be 1-5, got %d", p.SignCellRangeDown)
	}
	if p.SingGetMax < 10 || p.SingGetMax > 64 {
		return fmt.Errorf("singGetMax must be 10-64, got %d", p.SingGetMax)
	}
	if !finiteInRange(p.SignDownloadSpan, 10.0, 30.0) {
		return fmt.Errorf("signDownloadSpan must be 10-30, got %.0f", p.SignDownloadSpan)
	}
	if !finiteInRange(p.SignUpdateSpan, 10.0, 60.0) {
		return fmt.Errorf("signUpdateSpan must be 10-60, got %.0f", p.SignUpdateSpan)
	}
	// Blue
	if !finiteInRange(p.ReloadVisitListCoolTime, 5.0, 20.0) {
		return fmt.Errorf("reloadVisitListCoolTime must be 5-20, got %.0f", p.ReloadVisitListCoolTime)
	}
	if p.MaxCoopBlueSummonCount < 1 || p.MaxCoopBlueSummonCount > 2 {
		return fmt.Errorf("maxCoopBlueSummonCount must be 1-2, got %d", p.MaxCoopBlueSummonCount)
	}
	if p.MaxVisitListCount < 1 || p.MaxVisitListCount > 20 {
		return fmt.Errorf("maxVisitListCount must be 1-20, got %d", p.MaxVisitListCount)
	}
	if !finiteInRange(p.ReloadSearchCoopBlueMin, 5.0, 30.0) {
		return fmt.Errorf("reloadSearchCoopBlueMin must be 5-30, got %.0f", p.ReloadSearchCoopBlueMin)
	}
	if !finiteInRange(p.ReloadSearchCoopBlueMax, 20.0, 180.0) {
		return fmt.Errorf("reloadSearchCoopBlueMax must be 20-180, got %.0f", p.ReloadSearchCoopBlueMax)
	}
	if p.AllAreaSearchRateCoopBlue < 0 || p.AllAreaSearchRateCoopBlue > 100 {
		return fmt.Errorf("allAreaSearchRateCoopBlue must be 0-100, got %d", p.AllAreaSearchRateCoopBlue)
	}
	if p.AllAreaSearchRateVsBlue < 0 || p.AllAreaSearchRateVsBlue > 100 {
		return fmt.Errorf("allAreaSearchRateVsBlue must be 0-100, got %d", p.AllAreaSearchRateVsBlue)
	}
	if p.AllAreaSearchRateBellGuard < 0 || p.AllAreaSearchRateBellGuard > 100 {
		return fmt.Errorf("allAreaSearchRateBellGuard must be 0-100, got %d", p.AllAreaSearchRateBellGuard)
	}
	// Host
	if p.VisitorListMax < 1 || p.VisitorListMax > 100 {
		return fmt.Errorf("visitorListMax must be 1-100, got %d", p.VisitorListMax)
	}
	if !finiteInRange(p.VisitorTimeOutTime, 1.0, 999.0) {
		return fmt.Errorf("visitorTimeOutTime must be 1-999, got %.0f", p.VisitorTimeOutTime)
	}
	if !finiteInRange(p.VisitorDownloadSpan, 1.0, 999.0) {
		return fmt.Errorf("visitorDownloadSpan must be 1-999, got %.0f", p.VisitorDownloadSpan)
	}

	// Cross-field invariants (defensive — the UI also clamps these).
	if err := ValidateCrossFieldConstraints(p); err != nil {
		return err
	}
	return nil
}

func finiteInRange(value, min, max float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0) && value >= min && value <= max
}

// ValidateCrossFieldConstraints checks only the structural invariants that must
// hold independently of per-field bounds. Called by ValidateNetworkParams after
// individual field checks, and available as a standalone check where needed.
func ValidateCrossFieldConstraints(p NetworkParamValues) error {
	if p.ReloadSignCellCount > p.ReloadSignTotalCount {
		return fmt.Errorf("sign pool per-cell cap (%d) must not exceed total pool (%d)", p.ReloadSignCellCount, p.ReloadSignTotalCount)
	}
	if p.ReloadSignTotalCount > p.SingGetMax {
		return fmt.Errorf("sign pool total (%d) must not exceed buffer cap (%d)", p.ReloadSignTotalCount, p.SingGetMax)
	}
	if p.ReloadSearchCoopBlueMin > p.ReloadSearchCoopBlueMax {
		return fmt.Errorf("co-op search min (%.0f) must not exceed max (%.0f)", p.ReloadSearchCoopBlueMin, p.ReloadSearchCoopBlueMax)
	}
	return nil
}

// PatchNetworkParams modifies NetworkParam in UserData11 and returns the patched UserData11.
// It validates both the requested values and the fully encrypted result.
func PatchNetworkParams(ud11 []byte, patch NetworkParamValues) ([]byte, error) {
	if err := ValidateNetworkParams(patch); err != nil {
		return nil, fmt.Errorf("validate network parameters: %w", err)
	}
	regBlob, iv, dcxFormat, err := extractRegulation(ud11)
	if err != nil {
		return nil, fmt.Errorf("extract regulation: %w", err)
	}

	bnd4Data, err := decompressDCX(regBlob, dcxFormat)
	if err != nil {
		return nil, fmt.Errorf("decompress DCX: %w", err)
	}

	paramOffset, paramSize, rowDataOffset, err := findNetworkParamInBND4(bnd4Data)
	if err != nil {
		return nil, fmt.Errorf("find NetworkParam: %w", err)
	}
	const requiredRowBytes = offsetVisitorDownloadSpan + 4
	if rowDataOffset > paramSize || paramSize-rowDataOffset < requiredRowBytes {
		return nil, fmt.Errorf("NetworkParam row is too short: %d bytes (need %d)", paramSize-rowDataOffset, requiredRowBytes)
	}

	d := bnd4Data[paramOffset+rowDataOffset : paramOffset+paramSize]

	// Invader
	binary.LittleEndian.PutUint32(d[offsetMaxBreakInTargets:], uint32(patch.MaxBreakInTargetListCount))
	binary.LittleEndian.PutUint32(d[offsetBreakInInterval:], math.Float32bits(patch.BreakInRequestIntervalTimeSec))
	binary.LittleEndian.PutUint32(d[offsetBreakInTimeout:], math.Float32bits(patch.BreakInRequestTimeOutSec))
	binary.LittleEndian.PutUint32(d[offsetBreakInAreaCount:], uint32(patch.BreakInRequestAreaCount))

	// Cooperator
	binary.LittleEndian.PutUint32(d[offsetSummonTimeoutTime:], math.Float32bits(patch.SummonTimeoutTime))
	binary.LittleEndian.PutUint32(d[offsetReloadSignIntervalTime1:], math.Float32bits(patch.ReloadSignIntervalTime1))
	binary.LittleEndian.PutUint32(d[offsetReloadSignIntervalTime2:], math.Float32bits(patch.ReloadSignIntervalTime2))
	binary.LittleEndian.PutUint32(d[offsetReloadSignTotalCount:], uint32(patch.ReloadSignTotalCount))
	binary.LittleEndian.PutUint32(d[offsetReloadSignCellCount:], uint32(patch.ReloadSignCellCount))
	binary.LittleEndian.PutUint32(d[offsetUpdateSignIntervalTime:], math.Float32bits(patch.UpdateSignIntervalTime))
	binary.LittleEndian.PutUint32(d[offsetSignCellRangeH:], uint32(patch.SignCellRangeH))
	binary.LittleEndian.PutUint32(d[offsetSignCellRangeUp:], uint32(patch.SignCellRangeUp))
	binary.LittleEndian.PutUint32(d[offsetSignCellRangeDown:], uint32(patch.SignCellRangeDown))
	binary.LittleEndian.PutUint32(d[offsetSingGetMax:], uint32(patch.SingGetMax))
	binary.LittleEndian.PutUint32(d[offsetSignDownloadSpan:], math.Float32bits(patch.SignDownloadSpan))
	binary.LittleEndian.PutUint32(d[offsetSignUpdateSpan:], math.Float32bits(patch.SignUpdateSpan))

	// Blue
	binary.LittleEndian.PutUint32(d[offsetReloadVisitListCoolTime:], math.Float32bits(patch.ReloadVisitListCoolTime))
	binary.LittleEndian.PutUint32(d[offsetMaxCoopBlueSummonCount:], uint32(patch.MaxCoopBlueSummonCount))
	binary.LittleEndian.PutUint32(d[offsetMaxVisitListCount:], uint32(patch.MaxVisitListCount))
	binary.LittleEndian.PutUint32(d[offsetReloadSearchCoopBlueMin:], math.Float32bits(patch.ReloadSearchCoopBlueMin))
	binary.LittleEndian.PutUint32(d[offsetReloadSearchCoopBlueMax:], math.Float32bits(patch.ReloadSearchCoopBlueMax))
	d[offsetAllAreaSearchRateCoopBlue] = byte(patch.AllAreaSearchRateCoopBlue)
	d[offsetAllAreaSearchRateVsBlue] = byte(patch.AllAreaSearchRateVsBlue)
	d[offsetAllAreaSearchRateBellGuard] = byte(patch.AllAreaSearchRateBellGuard)

	// Host
	binary.LittleEndian.PutUint32(d[offsetVisitorListMax:], uint32(patch.VisitorListMax))
	binary.LittleEndian.PutUint32(d[offsetVisitorTimeOutTime:], math.Float32bits(patch.VisitorTimeOutTime))
	binary.LittleEndian.PutUint32(d[offsetVisitorDownloadSpan:], math.Float32bits(patch.VisitorDownloadSpan))

	regStart := ud11UnkSize
	originalCiphertextLen := len(ud11) - regStart - 16

	newRegBlob, err := recompressDCX(regBlob, bnd4Data)
	if err != nil {
		return nil, fmt.Errorf("compress DCX: %w", err)
	}

	if len(newRegBlob) > originalCiphertextLen {
		return nil, fmt.Errorf("patched regulation blob (%d bytes) exceeds ciphertext capacity (%d bytes)",
			len(newRegBlob), originalCiphertextLen)
	}

	reencrypted, err := encryptRegulation(newRegBlob, iv, originalCiphertextLen)
	if err != nil {
		return nil, fmt.Errorf("encrypt regulation: %w", err)
	}

	result := make([]byte, len(ud11))
	copy(result, ud11)
	copy(result[regStart:], reencrypted)

	verified, err := ReadNetworkParams(result)
	if err != nil {
		return nil, fmt.Errorf("verify encrypted result: %w", err)
	}
	if *verified != patch {
		return nil, fmt.Errorf("verify encrypted result: parameter values changed during write-back")
	}

	return result, nil
}

// --- internal pipeline ---

func extractRegulation(ud11 []byte) (decrypted []byte, iv []byte, dcxFormat string, err error) {
	regStart := ud11UnkSize
	if len(ud11) <= regStart+16 {
		return nil, nil, "", fmt.Errorf("UserData11 too short: %d bytes", len(ud11))
	}
	if len(ud11) < 4 || string(ud11[:4]) != " GER" {
		return nil, nil, "", fmt.Errorf("UserData11 is not a PS4 save (missing " + `" GER"` + " header)")
	}

	regRaw := ud11[regStart:]
	iv = make([]byte, 16)
	copy(iv, regRaw[:16])
	ciphertext := regRaw[16:]

	block, err := aes.NewCipher(regulationKey)
	if err != nil {
		return nil, nil, "", err
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, nil, "", fmt.Errorf("regulation ciphertext not aligned to block size")
	}

	decrypted = make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, ciphertext)

	if len(decrypted) < 4 || string(decrypted[:4]) != dcxMagic {
		return nil, nil, "", fmt.Errorf("decrypted regulation does not start with DCX header")
	}

	// Detect compression format at offset 40
	if len(decrypted) < 44 {
		return nil, nil, "", fmt.Errorf("decrypted regulation too short for format detection")
	}
	dcxFormat = string(decrypted[40:44])
	if dcxFormat != dcxFormatZSTD {
		return nil, nil, "", fmt.Errorf("unsupported PS4 DCX format: %q", dcxFormat)
	}

	return decrypted, iv, dcxFormat, nil
}

func encryptRegulation(dcxData []byte, iv []byte, originalCiphertextLen int) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("invalid AES IV length: %d", len(iv))
	}
	if originalCiphertextLen < 0 || originalCiphertextLen%aes.BlockSize != 0 {
		return nil, fmt.Errorf("invalid regulation ciphertext capacity: %d", originalCiphertextLen)
	}
	if len(dcxData) > originalCiphertextLen {
		return nil, fmt.Errorf("DCX data (%d bytes) exceeds ciphertext capacity (%d bytes)", len(dcxData), originalCiphertextLen)
	}
	plaintext := make([]byte, originalCiphertextLen)
	copy(plaintext, dcxData)

	block, err := aes.NewCipher(regulationKey)
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, plaintext)

	result := make([]byte, 16+originalCiphertextLen)
	copy(result[:16], iv)
	copy(result[16:], encrypted)
	return result, nil
}

func decompressDCX(dcxData []byte, format string) ([]byte, error) {
	if len(dcxData) < 76 {
		return nil, fmt.Errorf("DCX data too short")
	}

	decompressedSize := int(binary.BigEndian.Uint32(dcxData[28:32]))
	compressedSize := int(binary.BigEndian.Uint32(dcxData[32:36]))

	available := len(dcxData) - 76
	if compressedSize < 0 || compressedSize > available {
		return nil, fmt.Errorf("DCX compressed size %d exceeds available payload %d", compressedSize, available)
	}

	switch format {
	case dcxFormatZSTD:
		compressed := dcxData[76 : 76+compressedSize]
		decoder, err := zstd.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("zstd reader: %w", err)
		}
		defer decoder.Close()
		buf := bytes.NewBuffer(make([]byte, 0, decompressedSize))
		if _, err := io.Copy(buf, decoder); err != nil {
			return nil, fmt.Errorf("zstd decompress: %w", err)
		}
		if buf.Len() != decompressedSize {
			return nil, fmt.Errorf("DCX decompressed size %d does not match header %d", buf.Len(), decompressedSize)
		}
		return buf.Bytes(), nil

	default:
		return nil, fmt.Errorf("unsupported DCX format: %s", format)
	}
}

// recompressDCX follows SoulsFormats' ZstdHelper.WriteZstd frame shape: no
// content-size field, a 64KB window, and no checksum. The PS4 game accepts this
// shape; generic EncodeAll output and the former raw-block splice do not.
func recompressDCX(originalDCX, bnd4Data []byte) ([]byte, error) {
	if len(originalDCX) < 76 {
		return nil, fmt.Errorf("DCX data too short")
	}
	if string(originalDCX[:4]) != dcxMagic || string(originalDCX[40:44]) != dcxFormatZSTD {
		return nil, fmt.Errorf("invalid PS4 ZSTD DCX header")
	}

	var stream bytes.Buffer
	encoder, err := zstd.NewWriter(&stream,
		zstd.WithEncoderLevel(zstd.SpeedBestCompression),
		zstd.WithWindowSize(soulsFormatsWindowSize),
		zstd.WithEncoderCRC(false),
	)
	if err != nil {
		return nil, fmt.Errorf("zstd encoder: %w", err)
	}
	if _, err := encoder.Write(bnd4Data); err != nil {
		encoder.Close()
		return nil, fmt.Errorf("zstd encode: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("zstd encode close: %w", err)
	}

	compressed := stream.Bytes()
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("zstd self-check decoder: %w", err)
	}
	roundTrip, err := decoder.DecodeAll(compressed, nil)
	decoder.Close()
	if err != nil {
		return nil, fmt.Errorf("zstd self-check decompress: %w", err)
	}
	if !bytes.Equal(roundTrip, bnd4Data) {
		return nil, fmt.Errorf("zstd self-check produced different decompressed data")
	}

	dcx := make([]byte, 76+len(compressed))
	copy(dcx, originalDCX[:76])
	binary.BigEndian.PutUint32(dcx[28:32], uint32(len(bnd4Data)))
	binary.BigEndian.PutUint32(dcx[32:36], uint32(len(compressed)))
	copy(dcx[76:], compressed)
	return dcx, nil
}

// locateNetworkParam does the full read pipeline and returns the row data slice.
func locateNetworkParam(ud11 []byte) (rowData []byte, paramOffset int, rowDataOffset int, err error) {
	regBlob, _, _, err := extractRegulation(ud11)
	if err != nil {
		return nil, 0, 0, err
	}

	bnd4Data, err := decompressDCX(regBlob, string(regBlob[40:44]))
	if err != nil {
		return nil, 0, 0, err
	}

	paramOff, paramSize, rowOff, err := findNetworkParamInBND4(bnd4Data)
	if err != nil {
		return nil, 0, 0, err
	}

	const requiredRowBytes = offsetVisitorDownloadSpan + 4
	if rowOff > paramSize || paramSize-rowOff < requiredRowBytes {
		return nil, 0, 0, fmt.Errorf("NetworkParam row is too short: %d bytes (need %d)", paramSize-rowOff, requiredRowBytes)
	}

	return bnd4Data[paramOff+rowOff : paramOff+paramSize], paramOff, rowOff, nil
}

// findNetworkParamInBND4 scans BND4 file entries for NetworkParam.param.
// Returns: paramDataOffset, paramSize, rowDataOffset within the .param file.
func findNetworkParamInBND4(bnd4 []byte) (paramOffset int, paramSize int, rowDataOffset int, err error) {
	if len(bnd4) < 0x40 || string(bnd4[:4]) != bnd4Magic {
		return 0, 0, 0, fmt.Errorf("not a BND4 file")
	}

	fileCount := int(binary.LittleEndian.Uint32(bnd4[0x0C:0x10]))
	const entrySize = 0x24

	targetUTF16 := encodeUTF16LE(networkParamName)

	for i := 0; i < fileCount; i++ {
		entryOff := 0x40 + i*entrySize
		if entryOff+entrySize > len(bnd4) {
			break
		}

		nameOff := int(binary.LittleEndian.Uint32(bnd4[entryOff+32 : entryOff+36]))
		if nameOff <= 0 || nameOff >= len(bnd4) {
			continue
		}

		if matchesUTF16Name(bnd4, nameOff, targetUTF16) {
			compSize64 := binary.LittleEndian.Uint64(bnd4[entryOff+8 : entryOff+16])
			dataOff := int(binary.LittleEndian.Uint32(bnd4[entryOff+24 : entryOff+28]))

			if dataOff > len(bnd4) || compSize64 > uint64(len(bnd4)-dataOff) {
				return 0, 0, 0, fmt.Errorf("NetworkParam data exceeds BND4 bounds")
			}
			compSize := int(compSize64)

			// Parse PARAM to find Row 0 data offset
			paramData := bnd4[dataOff : dataOff+compSize]
			rowOff, err := parseParamRowDataOffset(paramData)
			if err != nil {
				return 0, 0, 0, fmt.Errorf("parse PARAM: %w", err)
			}

			return dataOff, compSize, rowOff, nil
		}
	}

	return 0, 0, 0, fmt.Errorf("NetworkParam.param not found in BND4 (%d files scanned)", fileCount)
}

// parseParamRowDataOffset reads just enough of the PARAM header to locate Row 0's data.
func parseParamRowDataOffset(paramData []byte) (int, error) {
	if len(paramData) < 0x58 {
		return 0, fmt.Errorf("param file too small (%d bytes)", len(paramData))
	}

	formatFlags := paramData[0x2D]
	longDataOffset := (formatFlags & 0x04) != 0

	if !longDataOffset {
		return 0, fmt.Errorf("expected LongDataOffset format flag, got 0x%02X", formatFlags)
	}

	// Row entries start at 0x40 for OffsetParamType + LongDataOffset format
	// Row format: id(4) + pad(4) + data_offset(8) + name_offset(8) = 24 bytes
	rowDataOff64 := binary.LittleEndian.Uint64(paramData[0x48:0x50])
	if rowDataOff64 == 0 || rowDataOff64 >= uint64(len(paramData)) {
		return 0, fmt.Errorf("invalid row data offset: 0x%X", rowDataOff64)
	}

	return int(rowDataOff64), nil
}

func encodeUTF16LE(s string) []byte {
	runes := utf16.Encode([]rune(s))
	b := make([]byte, len(runes)*2)
	for i, r := range runes {
		binary.LittleEndian.PutUint16(b[i*2:], r)
	}
	return b
}

func matchesUTF16Name(data []byte, offset int, target []byte) bool {
	// BND4 names include a full path. We match the suffix (filename).
	// Find the end of the UTF-16LE string (null terminator = 0x0000)
	end := offset
	for end+1 < len(data) {
		if data[end] == 0 && data[end+1] == 0 {
			break
		}
		end += 2
	}

	nameBytes := data[offset:end]
	if len(nameBytes) < len(target) {
		return false
	}

	// Check if the name ends with our target (after last backslash)
	suffix := nameBytes[len(nameBytes)-len(target):]
	return bytes.Equal(suffix, target)
}
