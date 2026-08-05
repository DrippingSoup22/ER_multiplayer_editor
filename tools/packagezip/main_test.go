package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateArchiveIncludesBundleDirectory(t *testing.T) {
	temp := t.TempDir()
	bundle := filepath.Join(temp, "ER-PvP-Mod-windows-amd64-1.2.3")
	if err := os.Mkdir(bundle, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bundle, "ERPvPMod.exe"), []byte("test executable"), 0755); err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(temp, "release.zip")
	if err := createArchive(bundle, archive); err != nil {
		t.Fatal(err)
	}

	zr, err := zip.OpenReader(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	want := "ER-PvP-Mod-windows-amd64-1.2.3/ERPvPMod.exe"
	for _, entry := range zr.File {
		if entry.Name == want {
			return
		}
	}
	t.Fatalf("archive does not contain %q", want)
}
