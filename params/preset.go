package params

import "er_pvp_mod/core"

// PresetMeta describes a named preset for one view.
// Description uses \r\n for line breaks (displayed in a Win32 EDIT control).
type PresetMeta struct {
	Key         string
	View        ParamView
	Label       string
	Description string
	Apply       func(core.NetworkParamValues) core.NetworkParamValues
}

// ---------------------------------------------------------------------------
// Invader presets
// ---------------------------------------------------------------------------

var invaderPresets = []PresetMeta{
	{
		Key: "vanilla", View: ViewInvader, Label: "Vanilla",
		Description: "The unmodified FromSoftware defaults. Use this to return all invasion settings to exactly how the game shipped.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			d := core.NetworkParamDefaults()
			v.MaxBreakInTargetListCount = d.MaxBreakInTargetListCount
			v.BreakInRequestIntervalTimeSec = d.BreakInRequestIntervalTimeSec
			v.BreakInRequestTimeOutSec = d.BreakInRequestTimeOutSec
			v.BreakInRequestAreaCount = d.BreakInRequestAreaCount
			return v
		},
	},
	{
		Key: "fast", View: ViewInvader, Label: "Fast",
		Description: "A meaningful step up from vanilla, balanced to remain server-friendly while noticeably cutting wait times.\r\n\r\n" +
			"The search interval is cut to less than half the vanilla cadence, so the game checks for newly eligible hosts far more often. " +
			"The candidate list and area count are more than doubled, giving each round significantly more options before the cycle has to restart. " +
			"The timeout is tightened from vanilla's generous 20 s — same-region connections complete well within the new window, while genuinely stalled sessions are recycled sooner. " +
			"The interval and timeout are matched so each candidate gets its full window without idle gaps between cycles.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.MaxBreakInTargetListCount = 12
			v.BreakInRequestIntervalTimeSec = 12.0
			v.BreakInRequestTimeOutSec = 10.0
			v.BreakInRequestAreaCount = 12
			return v
		},
	},
	{
		Key: "aggressive", View: ViewInvader, Label: "Aggressive",
		Description: "The practical ceiling for reliable play, prioritising invasion frequency while keeping cross-region connections viable.\r\n\r\n" +
			"The search interval and timeout are aligned so each cycle processes one candidate without idle gaps — the tightest configuration that avoids interrupting in-progress handshakes. " +
			"The maximum candidate list is requested every cycle so eligible hosts are unlikely to be missed in a single pass. " +
			"Geographic coverage is set wide enough to cover low-density and end-game areas without risking stale cross-region matches.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.MaxBreakInTargetListCount = 20
			v.BreakInRequestIntervalTimeSec = 7.0
			v.BreakInRequestTimeOutSec = 7.0
			v.BreakInRequestAreaCount = 15
			return v
		},
	},
}

// ---------------------------------------------------------------------------
// Find Signs presets (ViewSign)
// ---------------------------------------------------------------------------

var signPresets = []PresetMeta{
	{
		Key: "vanilla", View: ViewSign, Label: "Vanilla",
		Description: "The unmodified FromSoftware defaults. Use this to restore all sign-finding settings to their shipped state.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			d := core.NetworkParamDefaults()
			v.SummonTimeoutTime = d.SummonTimeoutTime
			v.ReloadSignIntervalTime1 = d.ReloadSignIntervalTime1
			v.ReloadSignIntervalTime2 = d.ReloadSignIntervalTime2
			v.ReloadSignTotalCount = d.ReloadSignTotalCount
			v.ReloadSignCellCount = d.ReloadSignCellCount
			v.SingGetMax = d.SingGetMax
			v.SignDownloadSpan = d.SignDownloadSpan
			v.SignCellRangeH = d.SignCellRangeH
			v.SignCellRangeUp = d.SignCellRangeUp
			v.SignCellRangeDown = d.SignCellRangeDown
			return v
		},
	},
	{
		Key: "fast", View: ViewSign, Label: "Fast",
		Description: "A substantial improvement for active co-op, designed to feel responsive without generating excessive server traffic.\r\n\r\n" +
			"The sign pool is doubled so busy areas show far more options simultaneously. " +
			"The full refresh cycle is significantly shortened so new signs appear quickly after being placed and stale ones vanish promptly. " +
			"The background download interval is tightened to keep the pool current between full refreshes. " +
			"All pool parameters are scaled together to maintain the required invariant.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.ReloadSignIntervalTime2 = 20.0
			v.ReloadSignTotalCount = 40
			v.ReloadSignCellCount = 20
			v.SingGetMax = 64
			v.SignDownloadSpan = 15.0
			v.ReloadSignIntervalTime1 = 0.0
			v.SignCellRangeH = 2
			v.SignCellRangeUp = 2
			v.SignCellRangeDown = 2
			return v
		},
	},
	{
		Key: "aggressive", View: ViewSign, Label: "Aggressive",
		Description: "Near real-time sign responsiveness — all timers aligned at the practical floor below which the server cannot supply meaningfully different data between cycles.\r\n\r\n" +
			"The pool is expanded to its practical maximum to accommodate the densest co-op areas. " +
			"All pool parameters are scaled together to maintain the required invariant.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.ReloadSignIntervalTime2 = 10.0
			v.ReloadSignTotalCount = 64
			v.ReloadSignCellCount = 32
			v.SingGetMax = 96
			v.SignDownloadSpan = 10.0
			v.ReloadSignIntervalTime1 = 0.0
			v.SignCellRangeH = 2
			v.SignCellRangeUp = 2
			v.SignCellRangeDown = 2
			return v
		},
	},
}

// ---------------------------------------------------------------------------
// Place Sign presets (ViewSignPlace)
// ---------------------------------------------------------------------------

var signPlacePresets = []PresetMeta{
	{
		Key: "vanilla", View: ViewSignPlace, Label: "Vanilla",
		Description: "The unmodified FromSoftware defaults. Use this to restore sign placement settings to their shipped state.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			d := core.NetworkParamDefaults()
			v.UpdateSignIntervalTime = d.UpdateSignIntervalTime
			v.SignUpdateSpan = d.SignUpdateSpan
			return v
		},
	},
	{
		Key: "fast", View: ViewSignPlace, Label: "Fast",
		Description: "Significantly improves sign reliability and placement responsiveness for the player waiting to be summoned.\r\n\r\n" +
			"The heartbeat is halved from vanilla, substantially reducing the chance of a placed sign silently expiring during a long wait. " +
			"The upload sync interval is tightened to keep the server's picture of your local sign activity current.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.UpdateSignIntervalTime = 15.0
			v.SignUpdateSpan = 20.0
			return v
		},
	},
	{
		Key: "aggressive", View: ViewSignPlace, Label: "Aggressive",
		Description: "Maximum sign persistence and placement reliability — heartbeat and upload cadence pushed to the practical floor.\r\n\r\n" +
			"Sign expiry under normal conditions becomes essentially impossible, and the server is kept as current as possible on local sign activity. " +
			"Suitable for extended co-op sessions in busy areas.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.UpdateSignIntervalTime = 10.0
			v.SignUpdateSpan = 10.0
			return v
		},
	},
}

// ---------------------------------------------------------------------------
// Hunter presets
// ---------------------------------------------------------------------------

var hunterPresets = []PresetMeta{
	{
		Key: "vanilla", View: ViewHunter, Label: "Vanilla",
		Description: "The unmodified FromSoftware defaults. Use this to restore all Hunter and Blue Cipher Ring settings to their shipped state.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			d := core.NetworkParamDefaults()
			v.ReloadVisitListCoolTime = d.ReloadVisitListCoolTime
			v.MaxVisitListCount = d.MaxVisitListCount
			v.ReloadSearchCoopBlueMin = d.ReloadSearchCoopBlueMin
			v.ReloadSearchCoopBlueMax = d.ReloadSearchCoopBlueMax
			v.AllAreaSearchRateCoopBlue = d.AllAreaSearchRateCoopBlue
			v.MaxCoopBlueSummonCount = d.MaxCoopBlueSummonCount
			v.AllAreaSearchRateVsBlue = d.AllAreaSearchRateVsBlue
			v.AllAreaSearchRateBellGuard = d.AllAreaSearchRateBellGuard
			return v
		},
	},
	{
		Key: "fast", View: ViewHunter, Label: "Fast",
		Description: "A noticeably more active Hunter experience, balanced to remain mindful of server load.\r\n\r\n" +
			"The polling cooldown is reduced so the game checks for available invasions significantly more often. " +
			"The candidate list is enlarged for more options to attempt per cycle. " +
			"The randomised search interval is compressed to a tighter range, cutting the average wait between dispatch attempts to a fraction of the vanilla value. " +
			"The all-area rate is raised so that more searches cover the full map, improving dispatch rates when local invasion activity is low.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.ReloadVisitListCoolTime = 8.0
			v.MaxVisitListCount = 10
			v.ReloadSearchCoopBlueMin = 10.0
			v.ReloadSearchCoopBlueMax = 40.0
			v.AllAreaSearchRateCoopBlue = 60
			return v
		},
	},
	{
		Key: "aggressive", View: ViewHunter, Label: "Aggressive",
		Description: "A highly active Hunter configuration, pushing dispatch frequency toward the practical ceiling while keeping connection quality viable.\r\n\r\n" +
			"The polling cooldown is significantly reduced from vanilla. " +
			"The randomised search interval is compressed to a tight range, cutting average wait to a small fraction of vanilla. " +
			"The candidate list is enlarged for better coverage per cycle. " +
			"The all-area rate is set to 100%, ensuring every search covers the full map — maximising the available player pool at the cost of potentially higher-latency matches.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.ReloadVisitListCoolTime = 5.0
			v.MaxVisitListCount = 15
			v.ReloadSearchCoopBlueMin = 5.0
			v.ReloadSearchCoopBlueMax = 20.0
			v.AllAreaSearchRateCoopBlue = 100
			return v
		},
	},
}

// ---------------------------------------------------------------------------
// Lookup helpers
// ---------------------------------------------------------------------------

func PresetsForView(v ParamView) []PresetMeta {
	switch v {
	case ViewSign:
		return signPresets
	case ViewSignPlace:
		return signPlacePresets
	case ViewHunter:
		return hunterPresets
	default:
		return invaderPresets
	}
}

func FindPreset(v ParamView, key string) (PresetMeta, bool) {
	for _, p := range PresetsForView(v) {
		if p.Key == key {
			return p, true
		}
	}
	return PresetMeta{}, false
}

func ApplyPreset(v ParamView, key string, current core.NetworkParamValues) (core.NetworkParamValues, bool) {
	p, ok := FindPreset(v, key)
	if !ok {
		return current, false
	}
	return p.Apply(current), true
}
