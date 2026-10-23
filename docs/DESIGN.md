# Design

How ER Multiplayer Editor is built and the rules it keeps. The save format (the
PS4 container, the regulation's encryption and compression, BND4 and PARAM) is
[ER_shared](https://github.com/DrippingSoup22/ER_shared)'s; this project owns
the edits to `NetworkParam`.

## Packages

```text
cmd/ermultiplayereditor -> internal/ui -> internal/netparam -> ER_shared/regulation
```

| package | owns |
|---|---|
| `internal/netparam` | the 24 fields, their ranges and the rules between them, the presets, reading and writing the save |
| `internal/ui` | the Gio window: layout, widget state, dialogs; it never touches the save's bytes |

`fields.go` and `presets.go` hold the fields' documentation and the presets as
data. `netparam_test.go` checks them against each other: every preset stays in
range and keeps the cross-field rules.

## The row

`NetworkParam` has a single row, 632 bytes. Its layout is measured and pinned in
ER_shared's table registry; a regulation whose table differs is refused rather
than written at guessed offsets. A field is an `s32`, an `f32` or a `u8` at a
fixed offset. `breakInRequestAreaCount` lives in the pad at `0x7C`, which holds
5 in vanilla. Every other byte of the row is preserved.

## The rules it keeps

- Each field has a range; values outside it are refused.
- Sign pool: per cell ≤ total ≤ buffer cap.
- Hunter: co-op search min ≤ co-op search max.
- PC saves are refused: the PC game replaces these values at every login.

## Saving

A save validates every value, clones the BND4, writes the fields, rebuilds the
regulation (ER_shared checks that the new stream decompresses to the same BND4
and fits), writes the file beside the output, reopens it, checks every field
reads back, and only then gives it the output's name. The output may replace a
file the user chose in the dialog, never the input. After a save the output
becomes the open file.

There is one level of edits: the values on screen. **Discard edits** returns to
the file's values, **Reset all to vanilla** to the shipped ones. Saving a file
where an unconfirmed field differs from vanilla asks first.

## The GUI

Gio widget state belongs to the frame goroutine. Dialogs, loading and saving run
outside it and hand their result back through a channel that the next frame
drains. Only typing in a value's box changes the value: the box also follows
presets, and a rewrite must not turn an untouched value into its rounded
display.

## Building and releases

```sh
go test ./... && go vet ./...
go run ./cmd/ermultiplayereditor  # ER_EDITOR_SAVE=<path> opens a save at start
scripts/build-windows.sh          # dist/ER-Multiplayer-Editor-windows-amd64-<version>.zip
scripts/build-linux.sh            # on Linux only
```

Windows cross-compiles without cgo. Linux builds natively on Ubuntu 22.04 (for
an old glibc baseline) and needs gcc, pkg-config and the Wayland, X11,
xkbcommon, EGL/GLES, libffi, Xcursor, Xfixes and Vulkan development packages;
its file dialogs need `zenity`, `matedialog` or `qarma` at run time.

`.github/workflows/release.yml` tests and vets on both targets on a `v*` tag,
packs each archive with ER_shared's `releasezip`, and attaches them to a draft
release.

## Tests

The tests that need a real save read `ER_SAVE` (a decrypted PS4 save whose
regulation matches the game) and `ER_PC_SAVE` (any PC save), default to where
the author's workspace keeps them, and skip when they are absent. They read the
saves only and write to temporary directories.
