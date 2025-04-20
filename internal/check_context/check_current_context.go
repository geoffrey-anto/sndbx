package check_current_context

import "os"

func CheckIfDockerfileExists() (bool, string) {
	// Check if Dockerfile exists in the current directory
	// Return true if it exists, false otherwise

	files := []string{
		"Dockerfile",
		"Dockerfile.*",
	}

	filesInCurrentDirectory, err := os.ReadDir(".")

	if err != nil {
		panic("Cannot read current file path! Please check permissions\n")
	}

	for _, file := range files {
		for _, fileInCurrentDirectory := range filesInCurrentDirectory {
			if fileInCurrentDirectory.Name() == file {
				return true, fileInCurrentDirectory.Name()
			}
		}
	}

	return false, ""
}
