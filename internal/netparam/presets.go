package netparam

// Presets lists each view's three presets. A nil Values map is the view's
// shipped defaults.
var Presets = map[View][]Preset{
	Invader: {
		{Name: "Vanilla", Description: "The unmodified FromSoftware defaults. Use this to return all invasion settings to exactly how the game shipped.",
			Values: nil}, // the view's shipped defaults
		{Name: "Fast", Description: "A meaningful step up from vanilla, balanced to remain server-friendly while noticeably cutting wait times.\n\n" +
			"The search interval is cut to less than half the vanilla cadence, so the game checks for newly eligible hosts far more often. " +
			"The candidate list and area count are more than doubled, giving each round significantly more options before the cycle has to restart. " +
			"The timeout is tightened from vanilla's generous 20 s — same-region connections complete well within the new window, while genuinely stalled sessions are recycled sooner. " +
			"The interval and timeout are matched so each candidate gets its full window without idle gaps between cycles.",
			Values: map[string]float64{"maxBreakInTargetListCount": 12, "breakInRequestIntervalTimeSec": 12, "breakInRequestTimeOutSec": 10, "breakInRequestAreaCount": 12}},
		{Name: "Aggressive", Description: "The practical ceiling for reliable play, prioritising invasion frequency while keeping cross-region connections viable.\n\n" +
			"The search interval and timeout are aligned so each cycle processes one candidate without idle gaps — the tightest configuration that avoids interrupting in-progress handshakes. " +
			"The maximum candidate list is requested every cycle so eligible hosts are unlikely to be missed in a single pass. " +
			"Geographic coverage is set wide enough to cover low-density and end-game areas without risking stale cross-region matches.",
			Values: map[string]float64{"maxBreakInTargetListCount": 20, "breakInRequestIntervalTimeSec": 7, "breakInRequestTimeOutSec": 7, "breakInRequestAreaCount": 15}},
	},
	FindSigns: {
		{Name: "Vanilla", Description: "The unmodified FromSoftware defaults. Use this to restore all sign-finding settings to their shipped state.",
			Values: nil}, // the view's shipped defaults
		{Name: "Fast", Description: "A substantial improvement for active co-op, designed to feel responsive without generating excessive server traffic.\n\n" +
			"The sign pool is doubled so busy areas show far more options simultaneously. " +
			"The full refresh cycle is significantly shortened so new signs appear quickly after being placed and stale ones vanish promptly. " +
			"The background download interval is tightened to keep the pool current between full refreshes. " +
			"All pool parameters are scaled together to maintain the required invariant.",
			Values: map[string]float64{"reloadSignIntervalTime2": 20, "reloadSignTotalCount": 40, "reloadSignCellCount": 20, "singGetMax": 64, "signDownloadSpan": 15, "reloadSignIntervalTime1": 0, "signCellRangeH": 2, "signCellRangeUp": 2, "signCellRangeDown": 2}},
		{Name: "Aggressive", Description: "Maximum sign responsiveness balanced against P2P mesh stability — timers set at 15 s, the fastest cadence that gives cross-region P2P handshakes room to complete.\n\n" +
			"The pool is expanded to its practical maximum to accommodate the densest co-op areas. " +
			"All pool parameters are scaled together to maintain the required invariant.",
			Values: map[string]float64{"reloadSignIntervalTime2": 15, "reloadSignTotalCount": 64, "reloadSignCellCount": 32, "singGetMax": 64, "signDownloadSpan": 15, "reloadSignIntervalTime1": 0, "signCellRangeH": 2, "signCellRangeUp": 2, "signCellRangeDown": 2}},
	},
	PlaceSign: {
		{Name: "Vanilla", Description: "The unmodified FromSoftware defaults. Use this to restore sign placement settings to their shipped state.",
			Values: nil}, // the view's shipped defaults
		{Name: "Fast", Description: "Significantly improves sign reliability and placement responsiveness for the player waiting to be summoned.\n\n" +
			"The heartbeat is halved from vanilla, substantially reducing the chance of a placed sign silently expiring during a long wait. " +
			"The upload sync interval is tightened to keep the server's picture of your local sign activity current.",
			Values: map[string]float64{"updateSignIntervalTime": 15, "signUpdateSpan": 20}},
		{Name: "Aggressive", Description: "Maximum sign persistence and placement reliability — heartbeat and upload cadence pushed to the practical floor.\n\n" +
			"Sign expiry under normal conditions becomes essentially impossible, and the server is kept as current as possible on local sign activity. " +
			"Suitable for extended co-op sessions in busy areas.",
			Values: map[string]float64{"updateSignIntervalTime": 10, "signUpdateSpan": 10}},
	},
	Hunter: {
		{Name: "Vanilla", Description: "The unmodified FromSoftware defaults. Use this to restore all Hunter and Blue Cipher Ring settings to their shipped state.",
			Values: nil}, // the view's shipped defaults
		{Name: "Fast", Description: "A noticeably more active Hunter experience, balanced to remain mindful of server load.\n\n" +
			"The polling cooldown is reduced so the game checks for available invasions significantly more often. " +
			"The candidate list is enlarged for more options to attempt per cycle. " +
			"The randomised search interval is compressed to a tighter range, cutting the average wait between dispatch attempts to a fraction of the vanilla value. " +
			"The all-area rate is raised so that more searches cover the full map, improving dispatch rates when local invasion activity is low.",
			Values: map[string]float64{"reloadVisitListCoolTime": 8, "maxVisitListCount": 10, "reloadSearchCoopBlueMin": 10, "reloadSearchCoopBlueMax": 40, "allAreaSearchRateCoopBlue": 60}},
		{Name: "Aggressive", Description: "A highly active Hunter configuration, pushing dispatch frequency toward the practical ceiling while keeping connection quality viable.\n\n" +
			"The polling cooldown is significantly reduced from vanilla. " +
			"The randomised search interval is compressed to a tight range, cutting average wait to a small fraction of vanilla. " +
			"The candidate list is enlarged for better coverage per cycle. " +
			"The all-area rate is set to 100%, ensuring every search covers the full map — maximising the available player pool at the cost of potentially higher-latency matches.",
			Values: map[string]float64{"reloadVisitListCoolTime": 5, "maxVisitListCount": 15, "reloadSearchCoopBlueMin": 5, "reloadSearchCoopBlueMax": 20, "allAreaSearchRateCoopBlue": 100}},
	},
}
