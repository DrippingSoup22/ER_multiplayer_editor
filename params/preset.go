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
			"The search interval is substantially shorter, so the game checks for newly eligible hosts far more often. " +
			"The candidate list is larger, giving each round more options before the cycle has to restart. " +
			"The timeout is tightened: same-region connections complete well within the new window, while genuinely stalled sessions are recycled sooner. " +
			"Geographic coverage is widened to maintain reasonable candidate lists in less-populated areas.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.MaxBreakInTargetListCount = 12
			v.BreakInRequestIntervalTimeSec = 8.0
			v.BreakInRequestTimeOutSec = 10.0
			v.BreakInRequestAreaCount = 10
			return v
		},
	},
	{
		Key: "aggressive", View: ViewInvader, Label: "Aggressive",
		Description: "The practical ceiling for same-region play, prioritising invasion frequency above all else.\r\n\r\n" +
			"The search interval approaches the floor where server response time becomes the bottleneck rather than the wait. " +
			"The maximum candidate list is requested every cycle so no eligible host is ever missed in a single pass. " +
			"The timeout is tight enough to recycle dead sessions quickly while still allowing typical same-region connections to complete. " +
			"Geographic coverage is set to the widest practical scope.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.MaxBreakInTargetListCount = 20
			v.BreakInRequestIntervalTimeSec = 4.0
			v.BreakInRequestTimeOutSec = 6.0
			v.BreakInRequestAreaCount = 20
			return v
		},
	},
}

// ---------------------------------------------------------------------------
// Summoning Sign presets
// ---------------------------------------------------------------------------

var signPresets = []PresetMeta{
	{
		Key: "vanilla", View: ViewSign, Label: "Vanilla",
		Description: "The unmodified FromSoftware defaults. Use this to restore all summon-sign settings to their shipped state.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			d := core.NetworkParamDefaults()
			v.SummonTimeoutTime = d.SummonTimeoutTime
			v.ReloadSignIntervalTime1 = 0.0
			v.ReloadSignIntervalTime2 = d.ReloadSignIntervalTime2
			v.ReloadSignTotalCount = d.ReloadSignTotalCount
			v.UpdateSignIntervalTime = d.UpdateSignIntervalTime
			v.SignDownloadSpan = d.SignDownloadSpan
			v.SignUpdateSpan = d.SignUpdateSpan
			v.ReloadSignCellCount = d.ReloadSignCellCount
			v.SingGetMax = d.SingGetMax
			return v
		},
	},
	{
		Key: "fast", View: ViewSign, Label: "Fast",
		Description: "A substantial improvement for active co-op, designed to feel responsive without generating excessive server traffic.\r\n\r\n" +
			"The sign pool is doubled so busy areas show far more options simultaneously. " +
			"The full refresh cycle is significantly shortened so new signs appear quickly after being placed and stale ones vanish promptly. " +
			"Background sync intervals and the sign heartbeat are tightened to keep the pool current and your own placed sign reliably registered. " +
			"The summon timeout is reduced so stuck loading screens resolve noticeably faster while same-region connections still have ample time to complete. " +
			"All pool parameters are scaled together to maintain the required invariant.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.SummonTimeoutTime = 20.0
			v.ReloadSignIntervalTime1 = 15.0
			v.ReloadSignIntervalTime2 = 15.0
			v.ReloadSignTotalCount = 40
			v.UpdateSignIntervalTime = 15.0
			v.SignDownloadSpan = 15.0
			v.SignUpdateSpan = 20.0
			v.ReloadSignCellCount = 20
			v.SingGetMax = 64
			return v
		},
	},
	{
		Key: "aggressive", View: ViewSign, Label: "Aggressive",
		Description: "Near real-time sign responsiveness — all timers aligned at the practical floor below which the server cannot supply meaningfully different data between cycles.\r\n\r\n" +
			"The pool is expanded to its practical maximum to accommodate the densest co-op areas. " +
			"The summon timeout is cut to a short window — same-region connections always complete well within it, and anything still pending beyond that point has almost certainly failed and is just occupying a slot. " +
			"All pool parameters are scaled together to maintain the required invariant.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.SummonTimeoutTime = 8.0
			v.ReloadSignIntervalTime1 = 8.0
			v.ReloadSignIntervalTime2 = 8.0
			v.ReloadSignTotalCount = 64
			v.UpdateSignIntervalTime = 8.0
			v.SignDownloadSpan = 8.0
			v.SignUpdateSpan = 8.0
			v.ReloadSignCellCount = 32
			v.SingGetMax = 96
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
			v.MaxVisitListCount = 12
			v.ReloadSearchCoopBlueMin = 10.0
			v.ReloadSearchCoopBlueMax = 45.0
			v.AllAreaSearchRateCoopBlue = 60
			return v
		},
	},
	{
		Key: "aggressive", View: ViewHunter, Label: "Aggressive",
		Description: "Maximum dispatch frequency — pushed to the practical ceiling for most configurations.\r\n\r\n" +
			"The polling cooldown approaches the point where server processing time becomes the limiting factor rather than the wait itself. " +
			"The randomised search interval is compressed to a very tight range, cutting the average wait between attempts to a small fraction of the vanilla value. " +
			"The candidate list is set to its practical maximum for the best coverage per cycle. " +
			"The all-area rate is raised high enough to ensure near-global coverage on most searches without forcing every single cycle to go fully global.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.ReloadVisitListCoolTime = 3.0
			v.MaxVisitListCount = 20
			v.ReloadSearchCoopBlueMin = 3.0
			v.ReloadSearchCoopBlueMax = 15.0
			v.AllAreaSearchRateCoopBlue = 80
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
			"The connection timeout is significantly reduced — same-region invaders establish well within the new window, while failed slots are recycled much faster than vanilla. " +
			"The list refresh interval is shortened to keep the candidate pool current and minimise wasted attempts on stale entries.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.VisitorListMax = 20
			v.VisitorTimeOutTime = 15.0
			v.VisitorDownloadSpan = 15.0
			return v
		},
	},
	{
		Key: "aggressive", View: ViewTongue, Label: "Aggressive",
		Description: "Maximum invasion rate — pushed to the practical speed ceiling for same-region play.\r\n\r\n" +
			"The candidate pool is raised to its largest practical size for the best coverage per cycle. " +
			"The connection timeout is cut to a short window — same-region invaders always establish within it, and any slot still pending beyond that point is almost certainly dead. " +
			"The list refreshes frequently enough that staleness is rarely a factor even during the most active PvP sessions.",
		Apply: func(v core.NetworkParamValues) core.NetworkParamValues {
			v.VisitorListMax = 30
			v.VisitorTimeOutTime = 6.0
			v.VisitorDownloadSpan = 8.0
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
