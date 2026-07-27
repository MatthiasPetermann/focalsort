package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsImageFile(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		path string
		want bool
	}{
		{name: "jpg lower", path: "a.jpg", want: true},
		{name: "jpg upper", path: "a.JPG", want: true},
		{name: "jpeg", path: "a.jpeg", want: true},
		{name: "png", path: "a.png", want: false},
		{name: "no ext", path: "a", want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isImageFile(tc.path)
			if got != tc.want {
				t.Fatalf("isImageFile(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestCollectImageFilesSorted(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", nested, err)
	}

	mustWrite(t, filepath.Join(root, "b.jpg"))
	mustWrite(t, filepath.Join(root, "a.jpeg"))
	mustWrite(t, filepath.Join(root, "z.txt"))
	mustWrite(t, filepath.Join(nested, "c.JPG"))

	recursiveFiles, err := collectImageFiles(root, true)
	if err != nil {
		t.Fatalf("collectImageFiles recursive: %v", err)
	}

	if len(recursiveFiles) != 3 {
		t.Fatalf("recursive file count = %d, want 3", len(recursiveFiles))
	}

	for i := 1; i < len(recursiveFiles); i++ {
		if recursiveFiles[i-1] > recursiveFiles[i] {
			t.Fatalf("recursive files not sorted: %v", recursiveFiles)
		}
	}

	flatFiles, err := collectImageFiles(root, false)
	if err != nil {
		t.Fatalf("collectImageFiles non-recursive: %v", err)
	}

	if len(flatFiles) != 2 {
		t.Fatalf("non-recursive file count = %d, want 2", len(flatFiles))
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
