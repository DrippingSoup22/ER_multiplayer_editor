package params

import "math"

// ---------------------------------------------------------------------------
// Taunter's Tongue params — VISITOR
// 3 params, all basic. Advanced mode expands slider limits only.
// ---------------------------------------------------------------------------

var tongueParams = []ParamMeta{
	{
		Key: "visitorListMax", View: ViewTongue,
		Label: "Visitor list size", Internal: "VisitorListMax", Type: "int32",
		Advanced: false, Min: 1, Max: 30, AdvancedMin: 1, AdvancedMax: float64(math.MaxInt32),
		ShortDef: "Candidate pool size per download cycle while Taunter's Tongue is active — how many invasion-ready players are fetched per batch. Mirrors Target List Size but for inbound invaders.\r\nRaise → more candidates per cycle; fewer wasted downloads when many invaders are online but most are out of range.\r\nLower → smaller batch; higher risk of exhausting the list without finding a compatible connection.",
		LongDef: "Taunter's Tongue is an item that makes your world invadable even without a cooperator present. While it is active, your game periodically downloads a list of players who are currently looking to invade someone, then tries to connect with them one by one.\r\n\r\nThis parameter controls how many players are on that downloaded list. At the default of 10, your game receives ten names per download and works through them in sequence. If all ten fail to connect — they are already invading someone else, they are outside your matchmaking level range, or their connection could not be established — your game waits for the next download cycle and tries a fresh batch.\r\n\r\nThe list works the same way as the candidate list for invaders, just from the other direction. Instead of you reaching out to potential hosts, you are pulling potential invaders toward your world. A larger list means more connection attempts per cycle, which reduces how often an entire download is wasted because all its entries were unavailable by the time your game tried them.",
		AdvancedDetails: "PARAMDEF SortID: 1500 — 訪問対象者リスト取得最大値 (VISITOR param group)\r\nMemory offset: 0x240  |  Type: int32 (PARAMDEF notes 'actually u8' for value range; stored as 32-bit)\r\nVanilla value: 10\r\nValidation: 1–100 (all modes)\r\nThe VISITOR group (0x240–0x248) is structurally independent from NET_VISIT_PARAM (Hunter params at 0x180–0x194) despite both relating to multiplayer player traffic.",
	},
	{
		Key: "visitorTimeOutTime", View: ViewTongue,
		Label: "Visitor timeout", Internal: "VisitorTimeOutTime", Type: "float32",
		Advanced: false, Min: 5, Max: 180, AdvancedMin: 1, AdvancedMax: 999,
		ShortDef: "Per-candidate connection deadline for inbound invaders — how long the client waits for the session handshake to complete before abandoning that entry and trying the next on the list.\r\nRaise → slow or cross-regional invaders get more time to connect.\r\nLower → unresponsive candidates dropped faster; faster list cycling at the cost of filtering out slow but valid connections.",
		LongDef: "When your game tries to connect to an invader from its candidate list, both your game and the invader's game need to negotiate and establish a connection. This normally takes a few seconds when things are working correctly. Sometimes it stalls — the invader may have just invaded someone else between when the list was downloaded and when your game tried to reach them, their internet connection may have dropped, or network conditions may be preventing the link from forming.\r\n\r\nThis parameter is how long your game waits for that connection process to complete before giving up on that specific invader and moving to the next one on the list.\r\n\r\nAt the default of 60 seconds with 10 candidates, if every single candidate fails to connect your game could spend up to ten minutes working through the list before downloading a fresh one. This is the frustrating 'Taunter's Tongue spinner' experience many players encounter — the invasion wheel spins for a very long time with nothing happening.\r\n\r\nReducing this significantly — to 10–20 seconds — means unresponsive candidates are abandoned much faster and your game moves on to the next option sooner. The trade-off is that invaders with higher latency connections, particularly those in distant countries, may not be given enough time to finish setting up the connection, so you will tend to match with players who have faster connections relative to you.",
		AdvancedDetails: "PARAMDEF SortID: 1520 — 訪問待ちタイムアウト (VISITOR)\r\nMemory offset: 0x244  |  Type: float32\r\nVanilla value: 60 s\r\nValidation: 1–600 s\r\nDistinct from SummonTimeoutTime (0x08, NET_COMMON_PARAM), which governs the co-op summon handshake. This parameter applies specifically to the Taunter's Tongue visitor (invader) connection context.",
	},
	{
		Key: "visitorDownloadSpan", View: ViewTongue,
		Label: "Visitor download span", Internal: "VisitorDownloadSpan", Type: "float32",
		Advanced: false, Min: 5, Max: 180, AdvancedMin: 1, AdvancedMax: 999,
		ShortDef: "Candidate list refresh interval while Taunter's Tongue is active — how often the client discards its current invader pool and downloads a fresh one from the server.\r\nRaise → invader pool grows stale; more wasted attempts on candidates no longer available.\r\nLower → pool stays current; less time spent on outdated entries during active PvP sessions.",
		LongDef: "The list of potential invaders your game downloads has a limited lifespan. From the moment it was created, the players on the list may have begun invading someone else, logged off, changed character builds, or moved out of your matchmaking range. The longer your game spends working through a list, the more of its entries point to players who are no longer actually available.\r\n\r\nThis parameter controls how often your game replaces the entire list with a fresh download, regardless of whether it has finished working through the current one.\r\n\r\nAt the default of 60 seconds, your invader pool can be up to a minute out of date. During quiet periods when the same invaders are online for long stretches and invasion activity is low, this creates few problems. During active PvP sessions where invasions are starting and ending rapidly and invaders cycle between sessions quickly, a 60-second-old list is often mostly stale by the time your game tries the later entries on it.\r\n\r\nReducing this value keeps the pool more current, meaning more of your connection attempts are directed at players who are actually still looking for a target. The improvement is most noticeable during peak activity periods when the available invader pool is changing quickly.",
		AdvancedDetails: "PARAMDEF SortID: 1530 — 訪問者リストダウンロード間隔[秒] (VISITOR)\r\nMemory offset: 0x248  |  Type: float32\r\nVanilla value: 60 s\r\nValidation: 1–600 s\r\nConceptual counterpart to SignDownloadSpan (0x64, NET_SUMMON_SIGN_PARAM) but operates in the Taunter's Tongue matchmaking path, not the co-op sign path.",
	},
}
