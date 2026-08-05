# ER PvP Mod — Elden Ring Multiplayer Parameter Editor

ER PvP Mod is a free Windows save editor for **Elden Ring on PS4**. It edits
the network parameters embedded in a decrypted `memory.dat`, allowing you to
tune invasion searches, summon-sign refreshes, sign uploads, and Blue Cipher
Ring dispatch behavior without changing the installed game files.

**[Download the latest release →](../../releases/latest)**

## Read this first

- **Always keep an untouched backup.** Never experiment on your only save.
- The input must already be decrypted. PlayStation's outer save encryption and
  signing are separate from this editor; use your normal save-management tool
  before and after editing.
- Only PS4 `memory.dat` files are supported. PC saves are intentionally rejected
  because the PC game downloads vanilla network parameters at every login.
- Changes are staged in memory and written only after you select **Save patched
  file**. The suggested output name ends in `_edited.dat`.
- Multiplayer behavior is partly server-controlled and community-understood.
  No modified save can be guaranteed risk-free for online play.

### Important: applying changes while playing online

When Elden Ring connects online on PlayStation, the server restores these
network parameters when a character is loaded for the first time. To make the
edited values take effect:

1. Start Elden Ring and load the character.
2. Exit that character back to the title menu **without closing the game**.
3. Load the character again in the same game session.

The edited parameters are used after the second character load. Do not close
Elden Ring between steps 2 and 3.

## Quick start

1. Download the Windows ZIP from the latest release, extract it, and run
   `ERPvPMod.exe`.
2. Back up and decrypt your PS4 save with your usual save tool.
3. Select **Browse...**, choose the decrypted `memory.dat`, and select **Load**.
4. Choose a view, adjust values directly or select a preset, then select
   **Apply** for that view or **Apply All** for every view.
5. Select **Save patched file** and choose a new destination.
6. Prepare the output for your console with the same external save workflow.
7. If you play online, follow the two-load procedure above.

Before an output file is written, the editor validates every value, rebuilds
the embedded regulation data with the PS4-compatible Zstandard frame layout,
decompresses it for an exact data check, re-encrypts it, and parses the final
result back to confirm every requested value.

## Features

### Four focused views

- **Invader** — candidate-list size, request cadence, timeout, and search area.
- **Find Signs** — host-side sign-pool capacity, refresh timing, summon timeout,
  and search-cell coverage.
- **Place Sign** — phantom-side heartbeat and upload cadence for a placed sign.
- **Hunter** — Blue Cipher Ring polling, candidate capacity, timing, and
  all-area search rates.

The editor exposes 24 parameters. Each field includes a plain-English
explanation, its valid range, and—where applicable—a warning or confidence
label.

### Presets and editing controls

Every view provides three presets:

| Preset | Intent |
|---|---|
| **Vanilla** | Restore the values shipped with the game |
| **Fast** | A noticeable improvement with a moderate request cadence |
| **Aggressive** | The practical speed ceiling within the editor's safety bounds |

- **Current** shows the applied in-memory value; **New value** shows the pending
  edit.
- **Apply** commits only the current view; **Apply All** commits every view.
- **Reset to vanilla** restores all four views to the shipped defaults.
- Presets affect visible parameters in standard mode and include hidden fields
  when Advanced mode is enabled.
- Cross-field constraints are enforced, including sign-pool capacity ordering
  and Hunter minimum/maximum search timing.

### Advanced mode and confidence labels

Advanced mode reveals additional fields and technical details such as PARAMDEF
SortIDs, row offsets, vanilla values, and cross-field constraints.

- No badge means the effect is reasonably established through available game
  data and community findings.
- **COMMUNITY-INFERRED** means the interpretation is not officially documented.
- **UNCONFIRMED** means the field may be unused or its Elden Ring behavior has
  not been established.

The editor asks for confirmation before saving when an unconfirmed field differs
from vanilla. Advanced combinations can still prevent a multiplayer session
from establishing, so change one group at a time and retain a backup.

## Compatibility with ER Merchant Editor

The latest ER PvP Mod and
[ER Merchant Editor](https://github.com/DrippingSoup22/ER_merchant_editor) can
be used on the same decrypted save in either order. They patch different param
tables and both fully decompress, rebuild, verify, and recompress the embedded
regulation data. Always use the newest releases and pass the output of the first
editor into the second; older ER PvP Mod builds used a faulty compression
strategy and must not be used.

## Limitations

- The editor supports decrypted PS4 `memory.dat` files only. It does not open PC
  `.sl2` files or decrypt, encrypt, or sign the outer PlayStation save container.
- Server-side matchmaking rules and player population still apply; faster local
  polling cannot create eligible sessions or bypass normal matchmaking ranges.
- `BreakInRequestAreaCount` is hidden as padding in the public PARAMDEF; its
  server-side meaning is community-inferred.
- `allAreaSearchRateBellGuard` is inherited from an older Bell Keeper system and
  its effect in Elden Ring is unconfirmed.
- No modified save can be guaranteed safe from online enforcement.

See [Save format and write-back safety](docs/WRITEBACK.md) for the technical
pipeline and file-size assumptions.

## Building from source

The required Go version is declared in `go.mod`. The Windows executable
cross-compiles without cgo, so a C compiler is not required.

```sh
git clone https://github.com/DrippingSoup22/ER_multiplayer_editor
cd ER_multiplayer_editor
make gui
```

The executable is written to `bin/er_pvp_mod_gui.exe`.

To create the same versioned ZIP used for a public release:

```sh
VERSION=2.7.0 make release
```

The archive is written to `dist/ER-PvP-Mod-windows-amd64-2.7.0.zip`.

## Windows security note

Windows SmartScreen may show **Windows protected your PC** the first time the
unsigned executable is launched. This project does not currently use a paid
code-signing certificate. To proceed, select **More info**, then **Run anyway**.

## License and attribution

GPLv3 — see [LICENSE](LICENSE). The regulation parsing core is adapted from the
GPLv3-licensed
[EldenRing-SaveForge](https://github.com/oisis/EldenRing-SaveForge) project by
oisis. The PS4-compatible Zstandard frame settings follow the behavior documented
by [SoulsFormats](https://github.com/JKAnderson/SoulsFormats).

ER PvP Mod is not affiliated with FromSoftware or Bandai Namco.
