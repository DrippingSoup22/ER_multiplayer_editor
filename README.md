# ER PvP Mod — Elden Ring Multiplayer Parameter Editor

A Windows GUI tool that edits the network parameters stored inside an Elden Ring save file, letting you tune invasion frequency, summon sign responsiveness, hunter dispatch speed, and Taunter's Tongue behaviour without touching the game files.

> **Back up your save before making any changes.** The tool writes directly to your save file on "Save patched file". A corrupt save cannot be recovered without a backup.

---

## What it does

Elden Ring stores a set of timing and capacity values in `NETWORK_PARAM_ST` inside your save file. These values control how aggressively the game searches for multiplayer sessions — how often it polls the server, how many candidates it fetches, how long it waits for a connection, and so on.

This tool reads those values, lets you change them through a guided interface, and writes the result back. The regulation parsing core is derived from [EldenRing-SaveForge](https://github.com/oisis/EldenRing-SaveForge).

---

## Supported save formats

| Format | File | Platform |
|--------|------|----------|
| PC save | `ER0000.sl2` | Windows / Steam |
| PS4 save | `memory.dat` | PlayStation 4 |
| Raw USERDATA11 | any | Extracted slot (advanced) |

---

## Features

- **4 parameter views** — Invader, Summoning Sign, Hunter, Taunter's Tongue
- **27 tunable parameters** with plain-English documentation for every field
- **Advanced mode** — unlocks additional parameters and removes the conservative slider limits; shows technical details (memory offsets, PARAMDEF IDs, cross-field constraints)
- **3 presets per view** — Vanilla (exact game defaults), Fast, Aggressive
- **Non-destructive editing** — changes are staged in memory and only written on explicit save
- **Current vs new** display — the left column always shows what is on disk; the right column shows your pending edits

---

## How to use

1. **Open your save file**
   - Click **Browse…** and select your `ER0000.sl2` (PC), `memory.dat` (PS4), or a raw USERDATA11 extract
   - The *Save type* indicator confirms the format was recognised

2. **Choose a view**
   - **Invader** — controls how often and how broadly the game searches for invasion targets
   - **Summoning Sign** — controls sign pool size, refresh rate, and summon timeout
   - **Hunter** — controls Blue Cipher Ring dispatch frequency and search coverage
   - **Taunter's Tongue** — controls how fast and often invaders arrive when the item is active

3. **Edit values**
   - Drag a slider, type in an edit field, or pick a preset from the dropdown
   - Changes are previewed in the *New value* column without touching the file
   - Click a parameter label or edit field to read its full documentation in the right panel
   - Switch views freely — your edits persist across all views until you save

4. **Commit and save**
   - Click **Apply values** to lock your edits as the new current values (shown in the *Current* column)
   - Click **Save patched file** to write the modified save to disk (a file picker will appear)

---

## Parameter views at a glance

### Invader
Controls the invasion matchmaking loop — how many candidate hosts are fetched, how often the game retries, and how long each connection attempt waits before being abandoned.

### Summoning Sign
Controls sign pool size and refresh cadence. Raising the pool (total + per-cell cap) and shortening the refresh intervals makes co-op signs appear faster and more reliably. Advanced mode exposes spatial cell-range parameters.

### Hunter
Controls Blue Cipher Ring dispatch. The key levers are the cooldown between search cycles and the randomised interval that spreads out server requests. Raising the all-area search rate ensures the game looks across the full map, not just your local region.

### Taunter's Tongue
Controls how quickly the game fetches fresh invader candidates and how long it waits for each connection attempt. Shorter values mean fewer wasted slots and faster cycling through the candidate list.

---

## Advanced mode

Click the **Advanced** checkbox (you will be shown a warning on first activation per session). Advanced mode:

- Reveals parameters not exposed by standard editors (spatial cell ranges, secondary refresh timers, retribution-blue search rates)
- Expands all slider limits to the datatype ceiling
- Shows technical documentation (PARAMDEF SortID, memory offset, vanilla value, validation range, cross-field constraints)

> Advanced combinations can prevent multiplayer sessions from establishing. Test changes before distributing a patched file.

---

## Presets

Each view has three presets:

| Preset | Intent |
|--------|--------|
| **Vanilla** | Exact values shipped with the game |
| **Fast** | Meaningful improvement, server-friendly cadence |
| **Aggressive** | Near the practical speed ceiling for same-region play |

Selecting a preset updates the *New value* column immediately. Click **Apply values** to commit.

---

## Notes and caveats

- `BreakInRequestAreaCount` (Area search count) is stored in a field FromSoftware deliberately labelled as unused padding in the public PARAMDEF. Its exact server-side behaviour is community-inferred, not officially documented.
- `allAreaSearchRateBellGuard` exists in the data as a legacy field from earlier FromSoftware titles. Its in-game effect in Elden Ring is unconfirmed.
- The vanilla value for `summonTimeoutTime` (45 s) was read from an unmodified PS4 save and confirmed against the SaveForge specification.
- All other vanilla values match the SaveForge specification exactly.

---

## Building from source

Requires Go 1.23+ and the MinGW cross-compiler (`x86_64-w64-mingw32-gcc`) for Windows targets.

```bash
git clone https://github.com/YOUR_USERNAME/er_pvp_mod
cd er_pvp_mod
make gui          # produces bin/er_pvp_mod_gui.exe
```

---

## Roadmap

- [ ] Linux version (TUI or web interface)
- [ ] Verify `summonTimeoutTime` vanilla value across game versions
- [ ] Confirm `allAreaSearchRateBellGuard` behaviour in ER

---

## License

MIT — see [LICENSE](LICENSE) for details.

---

## Credits

- Regulation parsing core adapted from [EldenRing-SaveForge](https://github.com/oisis/EldenRing-SaveForge) by oisis
- PC save format (`.sl2` BND4 layout) reverse-engineered with reference to [ER-Save-Lib](https://github.com/ClayAmore/ER-Save-Lib) and [SoulsFormats](https://github.com/JKAnderson/SoulsFormats)
