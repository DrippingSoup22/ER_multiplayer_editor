package app

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
		Description: "The unmodified FromSoftware defaults. Use this to return all invasion settings to exactly how the game shipped.\r\n\r\n" +
			"Target list: 5 | Search interval: 30 s | Request timeout: 20 s | Area count: 5",
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
		Description: "Noticeably faster invasions optimised for same-region play, with server load in mind.\r\n\r\n" +
			"Search interval drops from 30 s to 8 s so the client checks for available hosts nearly four times as often without flooding the server. " +
			"The candidate list grows to 12, giving each search more options before the cycle restarts. " +
			"Timeout holds at 10 s — generous for same-region handshakes (which complete in 2–5 s) while promptly dropping dead sessions. " +
			"Area count widens to 10 for better map coverage, helping in less-populated late-game zones.\r\n\r\n" +
			"Target list: 12 | Search interval: 8 s | Request timeout: 10 s | Area count: 10",
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
		Description: "Maximum invasion frequency — pushed to the practical speed ceiling for same-region play.\r\n\r\n" +
			"Search interval drops to 4 s, near the floor where the server's own round-trip time becomes the bottleneck. " +
			"The full 20-candidate list is requested each time so no eligible host is missed. " +
			"Timeout tightens to 6 s — same-region connections typically complete in 2–4 s, so 6 s remains reliable while aggressively recycling dead slots. " +
			"Area count raised to 20 for the broadest practical geographic coverage.\r\n\r\n" +
			"Target list: 20 | Search interval: 4 s | Request timeout: 6 s | Area count: 20",
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
		Description: "The unmodified FromSoftware defaults. Use this to restore all summon-sign settings to their shipped state.\r\n\r\n" +
			"Summon timeout: 45 s | Refresh: 60 s | Pool: 20 | Upload interval: 30 s | Download span: 30 s | Upload span: 60 s",
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
		Description: "Significantly more responsive sign pool for active co-op, with moderate server impact.\r\n\r\n" +
			"Refresh interval drops from 60 s to 15 s so new signs appear four times as fast. " +
			"The sign pool doubles to 40 for better coverage in busy areas. " +
			"All background sync intervals (download, upload, heartbeat) tighten to 15–20 s. " +
			"Summon timeout cut from 45 s to 20 s — same-region connections complete in 2–8 s so 20 s is still comfortable while making stuck loading screens resolve much faster. " +
			"Advanced pool parameters scale to maintain the invariant: cell ≤ total ≤ buffer.\r\n\r\n" +
			"Refresh: 15 s | Pool: 40 | Upload interval: 15 s | Download: 15 s | Upload span: 20 s | Summon timeout: 20 s",
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
		Description: "Near real-time sign pool — maximum speed, minimum wasted time.\r\n\r\n" +
			"All refresh and sync intervals align at 8 s, below which the server cannot supply meaningfully different data between cycles. " +
			"The pool expands to 64 for the densest co-op areas. " +
			"Summon timeout cut to 8 s — same-region connections always complete well within this window; any attempt still pending after 8 s has almost certainly failed and is just wasting the slot. " +
			"Advanced pool parameters scale to maintain the invariant: cell ≤ total ≤ buffer.\r\n\r\n" +
			"Refresh: 8 s | Pool: 64 | Upload interval: 8 s | Download: 8 s | Upload span: 8 s | Summon timeout: 8 s",
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
		Description: "The unmodified FromSoftware defaults. Use this to restore all Hunter / Blue Cipher Ring settings to their shipped state.\r\n\r\n" +
			"Cooldown: 20 s | List: 5 | Search: 30–180 s (avg 105 s) | All-area: 30 %",
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
		Description: "Noticeably faster Hunter dispatch, mindful of server load.\r\n\r\n" +
			"The main cooldown drops from 20 s to 8 s so the client re-checks for invaded worlds 2.5 times as often. " +
			"The candidate list grows to 12 for more options per cycle. " +
			"The randomised search interval compresses to 10–45 s, cutting the average wait from ~105 s to ~27 s. " +
			"All-area rate rises to 60 % so the majority of searches cover the full map without always going global.\r\n\r\n" +
			"Cooldown: 8 s | List: 12 | Search: 10–45 s (avg 27 s) | All-area: 60 %",
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
		Description: "Maximum Hunter dispatch speed — pushed to the practical ceiling.\r\n\r\n" +
			"Cooldown drops to 3 s, near the floor before the server's own processing time dominates. " +
			"The random search interval compresses to 3–15 s, cutting the average wait from ~105 s to ~9 s — roughly 12x faster than vanilla. " +
			"Candidate list raised to 20 for maximum coverage per cycle. " +
			"All-area rate set to 80 % for near-global coverage without always forcing a full-map search.\r\n\r\n" +
			"Cooldown: 3 s | List: 20 | Search: 3–15 s (avg 9 s) | All-area: 80 %",
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
		Description: "The unmodified FromSoftware defaults. Use this to restore Taunter's Tongue settings to their shipped state.\r\n\r\n" +
			"Visitor list: 10 | Timeout: 60 s | Download span: 60 s",
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
			"The visitor list doubles to 20 for more candidates per cycle. " +
			"Timeout drops from 60 s to 15 s — same-region invaders connect in 2–5 s so 15 s is still comfortable while dropping failed slots much faster than vanilla. " +
			"Download span drops to 15 s so the candidate list stays current and stale entries are replaced quickly.\r\n\r\n" +
			"Visitor list: 20 | Timeout: 15 s | Download span: 15 s",
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
			"Visitor list raised to 30 for the largest candidate pool per cycle. " +
			"Timeout cut to 6 s — same-region connections always establish within this window; any slot still pending at 6 s is almost certainly dead and is recycled immediately. " +
			"Download span drops to 8 s for a near real-time candidate list with minimal staleness.\r\n\r\n" +
			"Visitor list: 30 | Timeout: 6 s | Download span: 8 s",
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
