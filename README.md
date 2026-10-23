# ER Multiplayer Editor

ER Multiplayer Editor is a free desktop save editor for **Elden Ring on PS4**.
It edits the network parameters embedded in a decrypted `memory.dat`, so you
can tune invasion searches, summon-sign refreshes, sign uploads and Blue Cipher
Ring hunts without touching the installed game. Earlier releases were called
ER PvP Mod.

The application is available for **Windows and Linux**.

**[Download the latest release →](../../releases/latest)**

## Read this first

- **Always keep an untouched backup.** Never experiment on your only save.
- The input must already be decrypted. PlayStation's outer save encryption and
  signing are separate from this editor; use your usual save tool before and
  after editing.
- Only PS4 `memory.dat` files are supported. PC saves are refused on purpose:
  the PC game downloads vanilla network parameters at every login.
- Edits are written to a new file only when you select **Save patched file**.
  The input file is never overwritten.
- Multiplayer behaviour is partly decided by the server and only partly
  understood by the community. No modified save can be guaranteed risk-free
  online.

### Important: applying changes while playing online

When Elden Ring connects online on PlayStation, the server restores these
parameters the first time a character is loaded. To make the edited values
take effect:

1. Start Elden Ring and load the character.
2. Exit that character back to the title menu **without closing the game**.
3. Load the character again in the same game session.

The edited values are used after the second load. Do not close Elden Ring
between steps 2 and 3.

## Quick start

1. Download the ZIP for your system from the latest release, extract it and
   run `ERMultiplayerEditor`.
2. Back up and decrypt your PS4 save with your usual save tool.
3. Select **Browse…**, choose the decrypted `memory.dat`, and select **Load**.
4. Pick a view and set values with the sliders or boxes, or select a preset.
   Changed values show in gold.
5. Select **Save patched file…** and choose a new destination.
6. Prepare the output for your console with the same save tool, and follow the
   two-load procedure above if you play online.

On Linux, file dialogs need `zenity`, `matedialog` or `qarma`; install one with
your distribution's package manager if Browse or Save does not open.

Before a file is written, the editor checks every value and the rules between
them, rebuilds the embedded regulation, writes it beside the output, reopens it
and checks that every value reads back. A failed check produces an error
instead of a damaged save.

## Features

### Four views

- **Invader**: candidate-list size, request cadence, timeout and search area.
- **Find Signs**: the host's sign pool, refresh timing, summon timeout and
  search-cell coverage.
- **Place Sign**: how often a placed sign is kept alive and uploaded.
- **Hunter**: Blue Cipher Ring polling, candidate capacity, timing and
  all-area search rates.

The editor covers 24 parameters. Click a parameter's name to read what it does,
its range and any warning.

### Presets

Every view has three presets:

| Preset | Intent |
|---|---|
| **Vanilla** | the values shipped with the game |
| **Fast** | a noticeable improvement at a moderate request rate |
| **Aggressive** | the practical ceiling within the editor's safe ranges |

A preset sets only the fields shown. **Reset all to vanilla** restores every
view; **Discard edits** returns to the values in the loaded file.

### Advanced mode and confidence marks

**Advanced** shows the hidden fields and each field's technical notes: its
offset in the row, its vanilla value and its editable range.

- No mark: the effect is reasonably established.
- **◇** community-inferred: the meaning comes from community testing, not from
  official documentation.
- **?** unconfirmed: the field may be unused in Elden Ring.

The editor asks before saving a file where an unconfirmed field differs from
vanilla. Change one group at a time: some combinations can stop a multiplayer
session from forming.

## Compatibility with ER Merchant Editor

ER Multiplayer Editor and
[ER Merchant Editor](https://github.com/DrippingSoup22/ER_merchant_editor) can
edit the same decrypted save in either order: they change different tables, and
both rebuild and verify the whole regulation. Pass the output of the first into
the second. Do not use ER PvP Mod builds older than 2.7: they compressed the
regulation in a way the game rejects.

## Limitations

- It does not open PC `.sl2` files, and it does not decrypt, encrypt or sign the
  PlayStation save container.
- The server's matchmaking rules and the player population still apply: faster
  local polling cannot create sessions or widen matchmaking ranges.
- Area search count is labelled padding in the public param definitions; its
  meaning is community-inferred.
- The Bell Guard all-area search rate is inherited from an older game's Bell
  Keeper system; its effect in Elden Ring is unconfirmed.
- No modified save can be guaranteed safe from online enforcement.

## Building from source

The required Go version is declared in `go.mod`.

```sh
go run ./cmd/ermultiplayereditor  # run the GUI
scripts/build-windows.sh          # the Windows release ZIP
```

Windows cross-compiles without cgo. [The design](docs/DESIGN.md) covers the
Linux build and how releases are made.

## Windows security note

Windows SmartScreen may show **Windows protected your PC** the first time the
unsigned program runs. To proceed, select **More info**, then **Run anyway**.

## License and attribution

GPLv3 — see [LICENSE](LICENSE). The save format code is
[ER_shared](https://github.com/DrippingSoup22/ER_shared), whose regulation core
is adapted from the GPLv3
[EldenRing-SaveForge](https://github.com/oisis/EldenRing-SaveForge) by oisis.
The PS4-compatible Zstandard frame follows
[SoulsFormats](https://github.com/JKAnderson/SoulsFormats); param definitions
come from [soulsmods/Paramdex](https://github.com/soulsmods/Paramdex). The Lora
font is under the SIL Open Font License.

ER Multiplayer Editor is not affiliated with FromSoftware or Bandai Namco.
