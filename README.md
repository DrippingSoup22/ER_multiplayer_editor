# ER PvP Mod — Elden Ring Multiplayer Parameter Editor

A Windows GUI tool that edits the network parameters stored inside an Elden Ring save file, letting you tune invasion frequency, summon sign responsiveness, hunter dispatch speed, and Taunter's Tongue behaviour without touching the game files.

> **Back up your save before making any changes.** The tool writes directly to your save file on "Save patched file". A corrupt save cannot be recovered without a backup.

---

## What it does

Elden Ring stores a set of timing and capacity values in `NETWORK_PARAM_ST` inside your save file. These values control how aggressively the game searches for multiplayer sessions — how often it polls the server, how many candidates it fetches, how long it waits for a connection, and so on.

This tool reads those values, lets you change them through a guided interface, and writes the result back.

---

## Supported save formats

| Format | File | Platform |
|--------|------|----------|
| PC save | `ER0000.sl2` | Windows / Steam |
| PS4 save | `memory.dat` | PlayStation 4 |

---

## Features

- **4 parameter views** — Invader, Summoning Sign, Hunter, Taunter's Tongue
- **27 tunable parameters** with plain-English documentation and confidence indicators for every field
- **Scrollable parameter list** — mouse wheel supported; scrollbar appears automatically when a view has more parameters than fit at once
- **Advanced mode** — reveals additional hidden parameters and shows technical documentation (memory offsets, PARAMDEF IDs, cross-field constraints)
- **Unlock ranges** — available inside Advanced mode; disables all sliders and removes every numerical limit, allowing values to be entered freely via the edit fields
- **3 presets per view** — Vanilla (exact game defaults), Fast, Aggressive; in basic mode, presets only affect visible parameters
- **Non-destructive editing** — changes are staged in memory and only written on explicit save
- **Current vs new** display — the left column always shows what is on disk; the right column shows your pending edits
- **Save warning** — if you attempt to save a parameter whose in-game effect is unconfirmed, a confirmation prompt appears before the file picker

---

## How to use

1. **Open your save file**
   - Click **Browse…** and select your `ER0000.sl2` (PC) or `memory.dat` (PS4)
   - The *Save type* indicator in the top bar confirms the format was recognised

2. **Choose a view**
   - **Invader** — controls how often and how broadly the game searches for invasion targets
   - **Summoning Sign** — controls sign pool size, refresh rate, and summon timeout
   - **Hunter** — controls Blue Cipher Ring dispatch frequency and search coverage
   - **Taunter's Tongue** — controls how fast and often invaders arrive when the item is active

3. **Edit values**
   - Drag a slider, type in an edit field, or pick a preset from the dropdown
   - Changes are previewed in the *New value* column without touching the file
   - Click a parameter name or edit field to read its full documentation in the right panel
   - Switch views freely — your edits persist across all views until you save

4. **Commit and save**
   - Click **Apply values** to lock your edits as the new current values (shown in the *Current* column)
   - Click **Save patched file** to write the modified save to disk (a file picker will appear)

---

## Advanced mode

Click the **Advanced** checkbox. Advanced mode:

- Reveals parameters not exposed in standard view
- Shows technical documentation for each field (PARAMDEF SortID, memory offset, vanilla value, validation range, cross-field constraints)

> Advanced combinations can prevent multiplayer sessions from establishing. Keep a backup and test changes before distributing a patched file.

### Unlock ranges

When Advanced mode is on, an **Unlock ranges** checkbox appears. Enabling it:

- Disables all sliders and removes every numerical limit
- Values must be entered by typing directly into the edit fields
- No upper or lower boundary applies

> Values outside the tested ranges can produce undefined matchmaking behaviour or corrupt the parameter block. Apply the Vanilla preset at any time to restore safe defaults.

---

## Presets

Each view has three presets:

| Preset | Intent |
|--------|--------|
| **Vanilla** | Exact values shipped with the game |
| **Fast** | Meaningful improvement, server-friendly cadence |
| **Aggressive** | Near the practical speed ceiling for same-region play |

Selecting a preset updates the *New value* column immediately. Click **Apply values** to commit. In basic mode, presets only affect visible parameters — enable Advanced mode first to have presets also apply hidden parameters.

---

## Parameter confidence indicators

Each parameter's documentation panel shows how well its effect is understood:

- **No badge** — effect confirmed; consistent with community findings
- **[ COMMUNITY-INFERRED ]** — effect understood through community testing, not officially documented by FromSoftware
- **[ UNCONFIRMED ]** — in-game effect uncertain or untested in Elden Ring; parameter may be vestigial

Parameters with known constraints or limitations also display a **⚠ WARNING** notice. The tool will prompt for confirmation before saving if any unconfirmed parameter has been modified.

---

## Notes and caveats

- `allAreaSearchRateBellGuard` is a legacy field from Dark Souls 2's Bell Keeper system. Its in-game effect in Elden Ring is unconfirmed.
- `BreakInRequestAreaCount` is deliberately hidden by FromSoftware in the public PARAMDEF (labelled as unused padding). Its server-side behaviour is community-inferred.
- The vanilla value for `summonTimeoutTime` (45 s) was read from an unmodified PS4 save.

---

## Building from source

Requires Go 1.23+ and the MinGW cross-compiler (`x86_64-w64-mingw32-gcc`) for Windows targets.

```bash
git clone https://github.com/DrippingSoup22/ER_multiplayer_editor
cd ER_multiplayer_editor
make gui          # produces bin/er_pvp_mod_gui.exe
```

---

## Roadmap

- [ ] Linux version (TUI or web interface)

---

## Windows security note

When you download and run the executable for the first time, Windows may show a *"Windows protected your PC"* SmartScreen warning because the file is not digitally signed. This is normal for open-source tools distributed without a paid code-signing certificate.

To proceed: click **More info**, then **Run anyway**.

---

## License

GPL v3 — see [LICENSE](LICENSE) for details.

The regulation parsing core (`core/regulation.go`) is adapted from [EldenRing-SaveForge](https://github.com/oisis/EldenRing-SaveForge) by oisis, which is also GPL v3. This project's GPL v3 license is a direct consequence of that dependency.

---

## Credits

- Regulation parsing core adapted from [EldenRing-SaveForge](https://github.com/oisis/EldenRing-SaveForge) by oisis
- PC save format (`.sl2` BND4 layout) reverse-engineered with reference to [ER-Save-Lib](https://github.com/ClayAmore/ER-Save-Lib) and [SoulsFormats](https://github.com/JKAnderson/SoulsFormats)
