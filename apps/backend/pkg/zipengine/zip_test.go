package zipengine_test

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/lopor-ai/lopor/pkg/zipengine"
)

func TestCreateZipArchive(t *testing.T) {
	engine := zipengine.NewZipEngine()

	req := zipengine.ZipArchiveRequest{
		ArchiveName:   "project.zip",
		IncludeReadme: true,
		ReadmeTitle:   "Lopor Sample Project",
		Files: []zipengine.FileItem{
			{Path: "main.go", Content: "package main\n\nfunc main() {}"},
			{Path: "go.mod", Content: "module example.com/sample\n\ngo 1.22"},
		},
	}

	zipBytes, err := engine.CreateZipArchive(req)
	if err != nil {
		t.Fatalf("unexpected error creating zip archive: %v", err)
	}

	if len(zipBytes) == 0 {
		t.Fatalf("expected non-empty zip bytes")
	}

	// Verify zip contents unzipping in-memory
	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("failed to read generated zip archive: %v", err)
	}

	if len(zipReader.File) != 3 { // README.md + main.go + go.mod
		t.Errorf("expected 3 files in zip, got %d", len(zipReader.File))
	}

	foundReadme := false
	for _, f := range zipReader.File {
		if f.Name == "README.md" {
			foundReadme = true
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open README inside zip: %v", err)
			}
			content, _ := io.ReadAll(rc)
			rc.Close()
			if !bytes.Contains(content, []byte("Lopor Sample Project")) {
				t.Errorf("expected README title in content")
			}
		}
	}

	if !foundReadme {
		t.Errorf("README.md file was not found in generated zip")
	}
}
