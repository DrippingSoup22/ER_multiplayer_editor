package params

import (
	"er_pvp_mod/core"
	"fmt"
	"strconv"
	"strings"
)

// ParamView identifies which player-role tab is active.
type ParamView string

const (
	ViewInvader ParamView = "invader"
	ViewSign    ParamView = "sign"
	ViewHunter  ParamView = "hunter"
	ViewTongue  ParamView = "tongue"
)

// Confidence describes how well-established a parameter's in-game effect is.
// Confirmed is the zero value so it need not be set explicitly in param literals.
type Confidence int

const (
	Confirmed        Confidence = iota // effect well-understood; consistent with community findings
	CommunityInferred                  // effect inferred from community testing; not officially documented
	Unconfirmed                        // effect uncertain; parameter may be vestigial or untested in ER
)

// ParamMeta describes a single tunable field: its labels, documentation, casual and advanced
// slider bounds, and the key used to dispatch ValueString / SetFromInt.
type ParamMeta struct {
	Key      string
	View     ParamView
	Label    string
	Internal string
	Type     string

	// Advanced marks parameters that are hidden in casual mode.
	Advanced bool

	// Casual-mode slider bounds.
	Min float64
	Max float64

	// Unlock-ranges slider bounds — datatype ceiling/floor, no conservative limits.
	UnlockMin float64
	UnlockMax float64

	// Confidence describes how well-established this parameter's effect is.
	// Defaults to Confirmed (zero value) — only set explicitly for non-confirmed params.
	Confidence Confidence

	// Warning is an optional operational caveat shown prominently in the doc panel.
	// Use for cross-field constraints, server-side overrides, or known limitations.
	Warning string

	ShortDef        string // shown always; covers the essence + raise/lower summary
	LongDef         string // shown always; full plain-English explanation
	AdvancedDetails string // shown only in Advanced Mode; technical/code details
}

// ---------------------------------------------------------------------------
// Range helpers
// ---------------------------------------------------------------------------

func (p ParamMeta) MinInt() int { return int(p.Min) }
func (p ParamMeta) MaxInt() int { return int(p.Max) }

func (p ParamMeta) UnlockMinInt() int { return int(p.UnlockMin) }
func (p ParamMeta) UnlockMaxInt() int { return int(p.UnlockMax) }

func (p ParamMeta) RangeText() string {
	return fmt.Sprintf("%.0f – %.0f", p.Min, p.Max)
}

func (p ParamMeta) RangeTextUnlock() string {
	return fmt.Sprintf("%.0f – %.0f", p.UnlockMin, p.UnlockMax)
}

// ---------------------------------------------------------------------------
// Documentation
// ---------------------------------------------------------------------------

// DocText returns the full documentation string for this parameter.
// Advanced details are appended only when advanced=true.
func (p ParamMeta) DocText(advanced bool) string {
	if p.Key == "" {
		return ""
	}
	const sep = "· · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · · ·\r\n\r\n"

	// Confidence / warning block — metadata displayed immediately after the name.
	var meta string
	switch p.Confidence {
	case CommunityInferred:
		meta = "[ COMMUNITY-INFERRED ]\r\nEffect inferred from community testing — not officially documented.\r\n\r\n"
	case Unconfirmed:
		meta = "[ UNCONFIRMED ]\r\nIn-game effect uncertain or untested in Elden Ring.\r\n\r\n"
	}
	if p.Warning != "" {
		meta += "⚠  WARNING\r\n" + p.Warning + "\r\n\r\n"
	}

	// Separate the raise/lower direction lines from the description sentence.
	shortDef := strings.ReplaceAll(p.ShortDef, "\r\nRaise →", "\r\n\r\nRaise →")
	shortDef = strings.ReplaceAll(shortDef, "\r\nSet to 0 →", "\r\n\r\nSet to 0 →")

	s := strings.ToUpper(p.Label) + "\r\n\r\n" +
		meta +
		sep + "[SHORT DEFINITION]\r\n\r\n" + shortDef + "\r\n\r\n" +
		sep + "[LONG DEFINITION]\r\n\r\n" + p.LongDef

	if advanced && p.AdvancedDetails != "" {
		s += "\r\n\r\n" + sep + "[ADVANCED DETAILS]\r\n\r\n" + p.AdvancedDetails
	}
	return s
}

// ---------------------------------------------------------------------------
// Value access — switch-based dispatch keeps regulation.go untouched
// ---------------------------------------------------------------------------

func (p ParamMeta) ValueString(v core.NetworkParamValues) string {
	switch p.Key {
	// Invader
	case "maxBreakInTargetListCount":
		return strconv.Itoa(int(v.MaxBreakInTargetListCount))
	case "breakInRequestIntervalTimeSec":
		return fmt.Sprintf("%.0f", v.BreakInRequestIntervalTimeSec)
	case "breakInRequestTimeOutSec":
		return fmt.Sprintf("%.0f", v.BreakInRequestTimeOutSec)
	case "breakInRequestAreaCount":
		return strconv.Itoa(int(v.BreakInRequestAreaCount))
	// Sign
	case "summonTimeoutTime":
		return fmt.Sprintf("%.0f", v.SummonTimeoutTime)
	case "reloadSignIntervalTime2":
		return fmt.Sprintf("%.0f", v.ReloadSignIntervalTime2)
	case "reloadSignTotalCount":
		return strconv.Itoa(int(v.ReloadSignTotalCount))
	case "updateSignIntervalTime":
		return fmt.Sprintf("%.0f", v.UpdateSignIntervalTime)
	case "signDownloadSpan":
		return fmt.Sprintf("%.0f", v.SignDownloadSpan)
	case "signUpdateSpan":
		return fmt.Sprintf("%.0f", v.SignUpdateSpan)
	case "reloadSignIntervalTime1":
		return fmt.Sprintf("%.0f", v.ReloadSignIntervalTime1)
	case "reloadSignCellCount":
		return strconv.Itoa(int(v.ReloadSignCellCount))
	case "singGetMax":
		return strconv.Itoa(int(v.SingGetMax))
	case "signCellRangeH":
		return strconv.Itoa(int(v.SignCellRangeH))
	case "signCellRangeUp":
		return strconv.Itoa(int(v.SignCellRangeUp))
	case "signCellRangeDown":
		return strconv.Itoa(int(v.SignCellRangeDown))
	// Hunter
	case "reloadVisitListCoolTime":
		return fmt.Sprintf("%.0f", v.ReloadVisitListCoolTime)
	case "maxVisitListCount":
		return strconv.Itoa(int(v.MaxVisitListCount))
	case "reloadSearchCoopBlueMin":
		return fmt.Sprintf("%.0f", v.ReloadSearchCoopBlueMin)
	case "reloadSearchCoopBlueMax":
		return fmt.Sprintf("%.0f", v.ReloadSearchCoopBlueMax)
	case "allAreaSearchRateCoopBlue":
		return strconv.Itoa(int(v.AllAreaSearchRateCoopBlue))
	case "maxCoopBlueSummonCount":
		return strconv.Itoa(int(v.MaxCoopBlueSummonCount))
	case "allAreaSearchRateVsBlue":
		return strconv.Itoa(int(v.AllAreaSearchRateVsBlue))
	case "allAreaSearchRateBellGuard":
		return strconv.Itoa(int(v.AllAreaSearchRateBellGuard))
	// Tongue
	case "visitorListMax":
		return strconv.Itoa(int(v.VisitorListMax))
	case "visitorTimeOutTime":
		return fmt.Sprintf("%.0f", v.VisitorTimeOutTime)
	case "visitorDownloadSpan":
		return fmt.Sprintf("%.0f", v.VisitorDownloadSpan)
	default:
		return ""
	}
}

func (p ParamMeta) SetFromInt(dst *core.NetworkParamValues, n int) error {
	switch p.Key {
	// Invader
	case "maxBreakInTargetListCount":
		dst.MaxBreakInTargetListCount = int32(n)
	case "breakInRequestIntervalTimeSec":
		dst.BreakInRequestIntervalTimeSec = float32(n)
	case "breakInRequestTimeOutSec":
		dst.BreakInRequestTimeOutSec = float32(n)
	case "breakInRequestAreaCount":
		dst.BreakInRequestAreaCount = int32(n)
	// Sign
	case "summonTimeoutTime":
		dst.SummonTimeoutTime = float32(n)
	case "reloadSignIntervalTime2":
		dst.ReloadSignIntervalTime2 = float32(n)
	case "reloadSignTotalCount":
		dst.ReloadSignTotalCount = int32(n)
	case "updateSignIntervalTime":
		dst.UpdateSignIntervalTime = float32(n)
	case "signDownloadSpan":
		dst.SignDownloadSpan = float32(n)
	case "signUpdateSpan":
		dst.SignUpdateSpan = float32(n)
	case "reloadSignIntervalTime1":
		dst.ReloadSignIntervalTime1 = float32(n)
	case "reloadSignCellCount":
		dst.ReloadSignCellCount = int32(n)
	case "singGetMax":
		dst.SingGetMax = int32(n)
	case "signCellRangeH":
		dst.SignCellRangeH = int32(n)
	case "signCellRangeUp":
		dst.SignCellRangeUp = int32(n)
	case "signCellRangeDown":
		dst.SignCellRangeDown = int32(n)
	// Hunter
	case "reloadVisitListCoolTime":
		dst.ReloadVisitListCoolTime = float32(n)
	case "maxVisitListCount":
		dst.MaxVisitListCount = int32(n)
	case "reloadSearchCoopBlueMin":
		dst.ReloadSearchCoopBlueMin = float32(n)
	case "reloadSearchCoopBlueMax":
		dst.ReloadSearchCoopBlueMax = float32(n)
	case "allAreaSearchRateCoopBlue":
		dst.AllAreaSearchRateCoopBlue = int32(n)
	case "maxCoopBlueSummonCount":
		dst.MaxCoopBlueSummonCount = int32(n)
	case "allAreaSearchRateVsBlue":
		dst.AllAreaSearchRateVsBlue = int32(n)
	case "allAreaSearchRateBellGuard":
		dst.AllAreaSearchRateBellGuard = int32(n)
	// Tongue
	case "visitorListMax":
		dst.VisitorListMax = int32(n)
	case "visitorTimeOutTime":
		dst.VisitorTimeOutTime = float32(n)
	case "visitorDownloadSpan":
		dst.VisitorDownloadSpan = float32(n)
	default:
		return fmt.Errorf("unsupported parameter key: %s", p.Key)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

func ParamsForView(v ParamView) []ParamMeta {
	switch v {
	case ViewSign:
		return signParams
	case ViewHunter:
		return hunterParams
	case ViewTongue:
		return tongueParams
	default:
		return invaderParams
	}
}

func VisibleParamsForView(v ParamView, advanced bool) []ParamMeta {
	all := ParamsForView(v)
	if advanced {
		return all
	}
	out := make([]ParamMeta, 0, len(all))
	for _, p := range all {
		if !p.Advanced {
			out = append(out, p)
		}
	}
	return out
}
