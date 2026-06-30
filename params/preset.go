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
			"The interval and timeout are matched so each candidate gets its full window without idle gaps between cycles.\r\n\r\n" +
			"Interval: 12 s | Timeout: 10 s | List: 12 | Area: 12",
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
			"Geographic coverage is set wide enough to cover low-density and end-game areas without risking stale cross-region matches.\r\n\r\n" +
			"Interval: 7 s | Timeout: 7 s | List: 20 | Area: 15",
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
			v.ReloadSignIntervalTime1 = 0.0
			v.ReloadSignIntervalTime2 = d.ReloadSignIntervalTime2
			v.ReloadSignTotalCount = d.ReloadSignTotalCount
			v.ReloadSignCellCount = d.ReloadSignCellCount
			v.SingGetMax = d.SingGetMax
			v.SignDownloadSpan = d.SignDownloadSpan
			return v
		},
	},
	{
		Key: "fast", View: ViewSign, Label: "Fast",
		Description: "A substantial improvement for active co-op, designed to feel responsive without generating excessive server traffic.\r\n\r\n" +
			"The sign pool is doubled so busy areas show far more options simultaneously. " +
			"The full refresh cycle is significantly shortened so new signs appear quickly after being placed and stale ones vanish promptly. " +
			"The background download interval is tightened to keep the pool current between full refreshes. " +
			"All pool parameters are scaled together to maintain the required invariant.\r\n\r\n" +
			"Refresh: 20 s | Pool: 40 | Per cell: 20 | Buffer cap: 64 | Download span: 15 s",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.ReloadSignIntervalTime2 = 20.0
			v.ReloadSignTotalCount = 40
			v.ReloadSignCellCount = 20
			v.SingGetMax = 64
			v.SignDownloadSpan = 15.0
			return v
		},
	},
	{
		Key: "aggressive", View: ViewSign, Label: "Aggressive",
		Description: "Near real-time sign responsiveness — all timers aligned at the practical floor below which the server cannot supply meaningfully different data between cycles.\r\n\r\n" +
			"The pool is expanded to its practical maximum to accommodate the densest co-op areas. " +
			"All pool parameters are scaled together to maintain the required invariant.\r\n\r\n" +
			"Refresh: 10 s | Pool: 64 | Per cell: 32 | Buffer cap: 96 | Download span: 10 s",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.ReloadSignIntervalTime2 = 10.0
			v.ReloadSignTotalCount = 64
			v.ReloadSignCellCount = 32
			v.SingGetMax = 96
			v.SignDownloadSpan = 10.0
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
			"The upload sync interval is tightened to keep the server's picture of your local sign activity current.\r\n\r\n" +
			"Heartbeat: 15 s | Upload span: 20 s",
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
			"Suitable for extended co-op sessions in busy areas.\r\n\r\n" +
			"Heartbeat: 10 s | Upload span: 10 s",
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
			"The all-area rate is raised so that more searches cover the full map, improving dispatch rates when local invasion activity is low.\r\n\r\n" +
			"Cooldown: 8 s | List: 10 | Search: 10–40 s (avg 25 s) | All-area: 60 %",
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
			"The all-area rate is set to 100%, ensuring every search covers the full map — maximising the available player pool at the cost of potentially higher-latency matches.\r\n\r\n" +
			"Cooldown: 5 s | List: 15 | Search: 5–20 s (avg 12 s) | All-area: 100 %",
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
// Taunter's Tongue presets
// ---------------------------------------------------------------------------

var tonguePresets = []PresetMeta{
	{
		Key: "vanilla", View: ViewTongue, Label: "Vanilla",
		Description: "The unmodified FromSoftware defaults. Use this to restore Taunter's Tongue settings to their shipped state.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			d := core.NetworkParamDefaults()
			v.VisitorListMax = d.VisitorListMax
			v.VisitorTimeOutTime = d.VisitorTimeOutTime
			v.VisitorDownloadSpan = d.VisitorDownloadSpan
			return v
		},
	},
	{
		Key: "fast", View: ViewTongue, Label: "Fast",
		Description: "More frequent invasions while remaining server-friendly.\r\n\r\n" +
			"The candidate pool is doubled for more connection attempts per cycle, reducing the chance of exhausting the list without finding a match. " +
			"The connection timeout is meaningfully reduced from vanilla's 60 s — same-region invaders establish well within the new window, while failed slots cycle faster. " +
			"Cross-regional invaders retain enough time to complete their handshake. " +
			"The list refresh interval is shortened to keep the candidate pool current.\r\n\r\n" +
			"List: 20 | Timeout: 25 s | Refresh: 20 s",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.VisitorListMax = 20
			v.VisitorTimeOutTime = 25.0
			v.VisitorDownloadSpan = 20.0
			return v
		},
	},
	{
		Key: "aggressive", View: ViewTongue, Label: "Aggressive",
		Description: "Maximum invasion rate while preserving cross-regional connections — pushed to the practical speed ceiling without sacrificing connection breadth.\r\n\r\n" +
			"The candidate pool is raised to its largest practical size for the best coverage per cycle. " +
			"The connection timeout is cut to a third of the vanilla value while still giving cross-regional invaders enough time to establish. " +
			"The list refreshes at its fastest viable cadence, keeping the pool current even during the most active PvP sessions.\r\n\r\n" +
			"List: 30 | Timeout: 20 s | Refresh: 10 s",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.VisitorListMax = 30
			v.VisitorTimeOutTime = 20.0
			v.VisitorDownloadSpan = 10.0
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
	case ViewTongue:
		return tonguePresets
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
