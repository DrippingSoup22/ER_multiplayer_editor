package params

import "math"

// ---------------------------------------------------------------------------
// Invader params — NET_BREAKIN_PARAM
// 4 params, all basic. Sorted: scope first (list size, area count), then timing (interval, timeout).
// ---------------------------------------------------------------------------

var invaderParams = []ParamMeta{
	{
		Key: "maxBreakInTargetListCount", View: ViewInvader,
		Label: "Target list size", Internal: "MaxBreakInTargetListCount", Type: "int32",
		Advanced: false, Min: 1, Max: 20, AdvancedMin: 1, AdvancedMax: float64(math.MaxInt32),
		ShortDef: "Maximum candidates returned per server query. The client works through the list attempting connections before restarting the interval and requesting a new one.\r\nRaise → more connection attempts per cycle; fewer wasted intervals when hosts are unavailable.\r\nLower → smaller batch; the interval restarts more often even when valid targets exist.",
		LongDef: "When you use an invasion item — the Bloody Finger, Recusant Finger, or similar — your game does not immediately connect you to someone else's world. It first contacts FromSoftware's matchmaking servers and asks for a list of players it could potentially invade. Those players are online, in an area that allows invasions, and fall within the character level range the game requires for a match.\r\n\r\nThis parameter controls how many names are on that list. Your game then works through the list one by one, trying to establish a connection with each player. If a connection succeeds, you invade. If it fails — because that player just died, already has an invader, moved to a safe area, or has connection problems — the game crosses off that name and tries the next one.\r\n\r\nIf every name on the list fails, the game waits the Search Interval and then asks the server for a completely fresh list. Raising this number means you get more attempts per round before that waiting period kicks in, which directly reduces search time in busy sessions.\r\n\r\nLowering it has the opposite effect: a shorter list is exhausted quickly, forcing the game back to the waiting cycle more often even when plenty of valid targets exist on the server.",
		AdvancedDetails: "PARAMDEF SortID: 200 — 乱入先取得数 (NET_BREAKIN_PARAM)\r\nMemory offset: 0x70  |  Type: int32 (LittleEndian.Uint32)\r\nVanilla value: 5\r\nValidation: 1–20 casual, 1–MaxInt32 advanced\r\nNo cross-field dependencies.",
	},
	{
		Key: "breakInRequestAreaCount", View: ViewInvader,
		Label: "Area search count", Internal: "BreakInRequestAreaCount", Type: "int32",
		Advanced: false, Min: 1, Max: 20, AdvancedMin: 1, AdvancedMax: float64(math.MaxInt32),
		Confidence: CommunityInferred,
		Warning:    "This field is deliberately hidden by FromSoftware in the public PARAMDEF (labelled as unused padding). Its server-side effect is community-inferred. Treat changes conservatively.",
		ShortDef: "Number of geographic zones the server queries when building the candidate list. A hidden field — effect is community-inferred, not officially documented.\r\nRaise → broader map coverage; more candidates in low-density or late-game regions.\r\nLower → narrower search scope; list may thin out in end-game or off-peak areas.",
		LongDef: "Elden Ring's world is divided into geographic regions — Limgrave, Liurnia of the Lakes, Caelid, the Altus Plateau, and so on. When the server builds your invasion candidate list, it does not necessarily search the entire game world at once. This parameter influences how many of those regions it looks across.\r\n\r\nIn early, heavily populated areas like Limgrave, this setting matters little — there are always enough players nearby regardless of how many zones are searched. But in end-game regions like Farum Azula or Miquella's Haligtree, the active player population is much smaller, and a narrow geographic search may return very few or no results at all.\r\n\r\nRaising this value tells the server to cast a wider net, which helps keep the candidate list populated during off-peak hours or in less-active late-game areas.\r\n\r\nAn important note: FromSoftware deliberately obscured this field in their official parameter files, labelling it as empty unused space. Its exact server-side behaviour has been worked out by the community through testing rather than from any official documentation. Changes are generally safe but should be treated with more caution than fully confirmed parameters.",
		AdvancedDetails: "PARAMDEF SortID: 203 — labelled 'dummy8 pad[4]' in the public PARAMDEF to conceal it from standard editors (NET_BREAKIN_PARAM)\r\nMemory offset: 0x7C  |  Type: int32 (read as uint32, treated as signed)\r\nVanilla value: 5\r\nValidation: 1–20 casual, 1–MaxInt32 advanced\r\nExact server-side behavior is community-inferred, not officially documented.",
	},
	{
		Key: "breakInRequestIntervalTimeSec", View: ViewInvader,
		Label: "Search interval", Internal: "BreakInRequestIntervalTimeSec", Type: "float32",
		Advanced: false, Min: 2, Max: 30, AdvancedMin: 1, AdvancedMax: 999,
		ShortDef: "Cooldown between successive candidate list requests. The client waits this long after exhausting a list — or between requests when no targets connect — before querying the server again.\r\nRaise → longer gaps between polls; recently eligible hosts may be missed.\r\nLower → more frequent queries; newly eligible hosts appear in the next batch sooner.",
		LongDef: "Your game does not search for invasion targets in a continuous stream. It works in cycles: request a list of candidates from the server, try each one, then wait a set amount of time before requesting a new list. This parameter is the length of that wait.\r\n\r\nAt the game's default of 30 seconds, up to half a minute can pass between rounds of searching. During that gap, any player who just summoned a cooperator and became invasible will not appear in your results until the next cycle starts. This is why invasions often feel slow even when the server has plenty of eligible targets.\r\n\r\nReducing this to 6–10 seconds means your game checks in with the server much more frequently. A player who just became invasible 5 seconds ago will likely appear in your very next request rather than making you wait through the rest of a 30-second gap.\r\n\r\nGoing below about 4 seconds offers diminishing returns — the server itself needs time to process each request, so querying faster than the round-trip allows does not meaningfully improve results. Values in the 6–10 second range are the practical sweet spot.",
		AdvancedDetails: "PARAMDEF SortID: 201 — 乱入リクエスト間隔[秒] (NET_BREAKIN_PARAM)\r\nMemory offset: 0x74  |  Type: float32\r\nVanilla value: 30.0 s\r\nValidation: 2–30 s casual, 1–999 s advanced\r\nThe server also has a minimum dispatch rate independent of this setting.",
	},
	{
		Key: "breakInRequestTimeOutSec", View: ViewInvader,
		Label: "Request timeout", Internal: "BreakInRequestTimeOutSec", Type: "float32",
		Advanced: false, Min: 3, Max: 20, AdvancedMin: 1, AdvancedMax: 999,
		ShortDef: "Per-candidate handshake deadline. After selecting a host from the list, this is how long the client waits for the session to establish before abandoning that entry and moving to the next.\r\nRaise → slow or cross-regional connections get more time to complete.\r\nLower → unresponsive hosts dropped faster; list cycling accelerates at the risk of cutting valid slow connections.",
		LongDef: "Once the game picks a specific player from its candidate list to invade, it starts a connection process with that player's game. Both sides need to acknowledge each other, agree on the session parameters, and establish a direct link. This typically takes a few seconds when everything is working normally.\r\n\r\nSometimes that process stalls entirely. The target player may have lost their internet connection, their game may have crashed, they may have moved to an area where invasions are disabled, or their network setup may be blocking the connection. When any of those things happen, the connection attempt just hangs.\r\n\r\nThis parameter is how long the game waits in that stalled state before concluding the connection is not going to work and moving on to the next name on the list.\r\n\r\nRaising it gives slow or geographically distant connections more time to succeed — useful if you frequently try to invade players in regions far from your own, where the connection setup naturally takes longer. Lowering it means you stop waiting on clearly failing connections sooner and move to the next candidate faster, though legitimate connections that are just a bit slow may get cut before they finish.",
		AdvancedDetails: "PARAMDEF SortID: 202 — 乱入リクエストタイムアウト時間[秒] (NET_BREAKIN_PARAM)\r\nMemory offset: 0x78  |  Type: float32\r\nVanilla value: 20.0 s\r\nValidation: 3–20 s casual, 1–999 s advanced",
	},
}
