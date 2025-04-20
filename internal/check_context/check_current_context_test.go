package check_current_context

import (
	"os"
	"testing"
)

// helper function to create a temporary file
func createTempFile(t *testing.T, filename string) {
	t.Helper()
	file, err := os.Create(filename)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	file.Close()
}

// helper function to clean up file
func removeFile(t *testing.T, filename string) {
	t.Helper()
	err := os.Remove(filename)
	if err != nil {
		t.Fatalf("failed to remove temp file: %v", err)
	}
}

func TestCheckIfDockerfileExists_FileExists(t *testing.T) {
	filename := "Dockerfile"
	createTempFile(t, filename)
	defer removeFile(t, filename)

	found, fname := CheckIfDockerfileExists()

	if !found {
		t.Errorf("expected Dockerfile to be found, got found = false")
	}
	if fname != filename {
		t.Errorf("expected filename to be %s, got %s", filename, fname)
	}
}

func TestCheckIfDockerfileExists_FileDoesNotExist(t *testing.T) {
	found, fname := CheckIfDockerfileExists()

	if found {
		t.Errorf("expected Dockerfile not to be found, but got found = true")
	}
	if fname != "" {
		t.Errorf("expected filename to be empty, got %s", fname)
	}
}

func TestCheckIfDockerfileExists_WildcardPatternIgnored(t *testing.T) {
	// This test proves that "Dockerfile.*" is not handled as a wildcard
	createTempFile(t, "Dockerfile.dev")
	defer removeFile(t, "Dockerfile.dev")

	found, fname := CheckIfDockerfileExists()

	if found {
		t.Errorf("expected Dockerfile.dev to be ignored, but got found = true with %s", fname)
	}
}
