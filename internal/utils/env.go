package utils

import (
	"bufio"
	"os"
	"strings"
)

func isFile(filename string) bool {
	return filename != "" && !strings.Contains(filename, ":") && !strings.Contains(filename, "=")
}

func isFileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func ParseEnv(env string) []string {
	env_vars := []string{}

	if isFile(env) {
		if isFileExists(env) {
			file, err := os.Open(env)
			if err != nil {
				panic("failed to open file")
			}
			defer file.Close()

			// Read the file and parse the environment variables

			var envVars []string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.TrimSpace(line) != "" {
					envVars = append(envVars, line)
				}
			}
			if err := scanner.Err(); err != nil {
				panic("failed to read file")
			}
			for _, envVar := range envVars {
				keyValue := strings.Split(envVar, "=")

				if len(keyValue) == 2 {
					env_vars = append(env_vars, envVar)
				} else {
					panic("invalid environment variable format")
				}
			}
		} else {
			panic("file not found")
		}
	} else {
		envVars := strings.Split(env, ":")
		for _, envVar := range envVars {
			keyValue := strings.Split(envVar, "=")
			if len(keyValue) == 2 {
				env_vars = append(env_vars, envVar)
			} else {
				panic("invalid environment variable format")
			}
		}
	}

	return env_vars
}
