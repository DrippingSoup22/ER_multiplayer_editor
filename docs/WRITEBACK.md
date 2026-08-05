# Save format and write-back safety

ER PvP Mod accepts a decrypted PS4 `memory.dat`. It does not handle the outer
PlayStation encryption or signing layer.

## Data path

The editor uses the fixed PS4 save layout to locate `UserData11`:

```text
memory.dat
  -> UserData11
     -> AES-256-CBC encrypted regulation data
        -> DCX container
           -> Zstandard stream
              -> BND4 archive
                 -> NetworkParam.param, row 0
```

`UserData11` begins after the `0x70`-byte save header, ten `0x280000`-byte slot
regions, and the `0x60000`-byte `UserData10` region. Its expected size is
`0x240010` bytes. The loader also requires the PS4 `" GER"` header and rejects
unrecognized layouts.

## Write pipeline

When saving, the application:

1. Validates every field and the cross-field invariants.
2. Decrypts and decompresses the existing regulation data.
3. Locates `NetworkParam.param` through its BND4 entry instead of assuming a
   fixed archive offset.
4. Changes only the documented bytes in row 0.
5. Recompresses the complete BND4 payload as a Zstandard frame with a 64 KiB
   window, no frame content-size field, and no checksum—the frame shape used by
   SoulsFormats for this PS4 data.
6. Decompresses the new stream and requires byte-for-byte equality with the
   patched BND4 payload.
7. Preserves the original DCX header, updates its compressed and decompressed
   sizes, and AES-encrypts the result into the original fixed-capacity region.
8. Parses the final encrypted `UserData11` and requires every network parameter
   to equal the requested value.
9. Writes a copy of the complete save to the destination selected by the user.

Any failed check returns an error before the destination is written.

## Sequential editor compatibility

ER PvP Mod and ER Merchant Editor operate on different param tables within the
same BND4 archive. Current versions of both tools rebuild the entire compressed
stream correctly, so the output from either editor can be loaded into the other.
Use the first tool's output as the second tool's input and keep the original
decrypted save as a backup.

Do not reuse an ER PvP Mod executable from before the full-stream compression
fix. Those builds could create a Zstandard stream that standard decoders reject.
