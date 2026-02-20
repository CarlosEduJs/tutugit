package diff

import (
	"strings"
	"testing"
)

func TestParseDiff(t *testing.T) {
	rawDiff := `diff --git a/file1.go b/file1.go
index 1234567..89abcde 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package main
 
+import "fmt"
+
 func main() {
@@ -10,3 +11,3 @@
 func hello() {
-	println("hello")
+	fmt.Println("hello world")
 }
diff --git a/file2.txt b/file2.txt
new file mode 100644
index 0000000..9876543
--- /dev/null
+++ b/file2.txt
@@ -0,0 +1 @@
+hello from file 2
`

	files := ParseDiff(rawDiff)

	if len(files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(files))
	}

	// File 1 Checks
	f1 := files[0]
	if f1.Path != "file1.go" {
		t.Errorf("Expected path file1.go, got %s", f1.Path)
	}
	if len(f1.Hunks) != 2 {
		t.Errorf("Expected 2 hunks for file1.go, got %d", len(f1.Hunks))
	}

	h1 := f1.Hunks[0]
	if h1.Header != "@@ -1,3 +1,4 @@" {
		t.Errorf("Unexpected hunk header: %s", h1.Header)
	}
	if !strings.Contains(h1.Content, "+import \"fmt\"") {
		t.Errorf("Expected hunk content to contain '+import \"fmt\"', got:\n%s", h1.Content)
	}

	// File 2 Checks
	f2 := files[1]
	if f2.Path != "file2.txt" {
		t.Errorf("Expected path file2.txt, got %s", f2.Path)
	}
	if len(f2.Hunks) != 1 {
		t.Errorf("Expected 1 hunk for file2.txt, got %d", len(f2.Hunks))
	}
}

func TestParseDiff_NoChanges(t *testing.T) {
	files := ParseDiff("")
	if len(files) != 0 {
		t.Errorf("Expected 0 files for empty diff, got %d", len(files))
	}
}
