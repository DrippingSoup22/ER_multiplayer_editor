# ER PvP Mod — Elden Ring Multiplayer Parameter Editor

A Windows GUI tool that edits the network parameters stored inside an Elden Ring save file, letting you tune invasion frequency, summon sign responsiveness, and hunter dispatch speed without touching the game files.

**Edits only take effect on PS4.** The PS4 game reads these parameters from the save file on the second launch after modification. On PC the game re-fetches all network parameters from FromSoftware's servers at every login, overwriting the save — PC saves can be loaded and modified but changes will have no in-game effect.

> **Back up your save before making any changes.** The tool writes directly to your save file on "Save patched file". A corrupt save cannot be recovered without a backup.

---

## How it works

Elden Ring stores a set of timing and capacity values in `NETWORK_PARAM_ST` inside your save file. These control how aggressively the game searches for multiplayer sessions — how often it polls the server, how many candidates it fetches, how long it waits for a connection, and so on.

**On PS4**, the game reads these values from the save file on the second launch after it has been modified. Edits made with this tool take effect in-game.

**On PC**, the game re-fetches all network parameters from FromSoftware's servers at every login, overwriting whatever is in the save file. Edits made at the save level have no runtime effect on PC.

This tool reads the parameter block, lets you change values through a guided interface, and writes the result back to the file.

---

## Supported save formats

| Format | File | Edits take effect |
|--------|------|-------------------|
| PS4 save | `memory.dat` | Yes — on the second launch after modifying |
| PC save | `ER0000.sl2` | No — server parameters overwrite save on login |

---

## Features

- **4 parameter views** — Invader, Find Signs, Place Sign, Hunter
- **24 tunable parameters** with plain-English documentation and confidence indicators for every field
- **Scrollable parameter list** — mouse wheel supported; scrollbar appears automatically when a view has more parameters than fit at once
- **Advanced mode** — reveals additional hidden parameters and shows technical documentation (memory offsets, PARAMDEF IDs, vanilla values, cross-field constraints)
- **3 presets per view** — Vanilla (exact game defaults), Fast, Aggressive; in basic mode presets only affect visible parameters, in Advanced mode they also apply hidden ones
- **Apply / Apply All** — commit edits for the current view only, or for all views at once
- **Reset to vanilla** — one-click reset of every parameter in every view to the game's shipped defaults
- **Non-destructive editing** — changes are staged in memory and only written on explicit save
- **Current vs new** display — the left column always shows what is on disk; the right column shows your pending edits
- **Save warning** — if any parameter with an unconfirmed effect has been modified, a confirmation prompt appears before the file picker

---

## How to use

1. **Open your save file**
   - Click **Browse…** and select your `memory.dat` (PS4) or `ER0000.sl2` (PC)
   - The *Save type* indicator in the top bar confirms the format was recognised

2. **Choose a view**
   - **Invader** — how often and how broadly the game searches for invasion targets
   - **Find Signs** — sign pool size, refresh rate, and summon timeout (host side)
   - **Place Sign** — heartbeat and upload cadence for keeping your own sign alive (phantom side)
   - **Hunter** — Blue Cipher Ring dispatch frequency and search coverage

3. **Edit values**
   - Drag a slider, type in an edit field, or pick a preset from the dropdown
   - Changes are previewed in the *New value* column without touching the file
   - Click a parameter name or edit field to read its full documentation in the right panel
   - Switch views freely — edits persist across all views until you save

4. **Commit and save**
   - Click **Apply** to lock edits for the current view, or **Apply All** to commit every view at once — locked values appear in the *Current* column
   - Use **Reset to vanilla** (controls bar, next to the preset dropdown) to discard all staged and committed edits and return every parameter to the game's shipped defaults
   - Click **Save patched file** to write the modified save to disk (a file picker will appear)

5. **Make it take effect (PS4 only)**
   - Copy the patched `memory.dat` back to your PS4
   - **Launch the game → quit to the main menu or close it → launch again**
   - On the second launch the game reads the network parameters from the save file

---

## Advanced mode

Click the **Advanced** checkbox. Advanced mode:

- Reveals parameters not exposed in the standard view
- Shows technical documentation for each field (PARAMDEF SortID, memory offset, vanilla value, validation range, cross-field constraints)

> Advanced combinations can prevent multiplayer sessions from establishing. Keep a backup and test changes before distributing a patched file.

## Presets

Each view has three presets:

| Preset | Intent |
|--------|--------|
| **Vanilla** | Exact values shipped with the game |
| **Fast** | Meaningful improvement, server-friendly cadence |
| **Aggressive** | Near the practical speed ceiling while keeping cross-region connections viable |

Selecting a preset updates the *New value* column immediately. Click **Apply** or **Apply All** to commit. In basic mode, presets only affect visible parameters — enable Advanced mode first to have presets also apply hidden parameters.

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

---

## Building from source

Requires Go 1.23+ and the MinGW cross-compiler (`x86_64-w64-mingw32-gcc`) for Windows targets.

```bash
git clone https://github.com/DrippingSoup22/ER_multiplayer_editor
cd ER_multiplayer_editor
make gui          # produces bin/er_pvp_mod_gui.exe
```

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
