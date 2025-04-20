package utils

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"golang.org/x/term"
)

// streamTerminal handles interactive terminal sessions with a container
func StreamTerminal(resp types.HijackedResponse) error {
	// Set up raw mode for terminal
	oldState, err := term.MakeRaw(0)
	if err != nil {
		return err
	}
	defer term.Restore(0, oldState)

	// Create a channel to signal when to exit
	done := make(chan error)

	// Copy data from container to terminal
	go func() {
		_, err := io.Copy(os.Stdout, resp.Reader)
		done <- err
	}()

	// Copy data from terminal to container, handle EOF (Ctrl+D)
	go func() {
		_, err := io.Copy(resp.Conn, os.Stdin)
		// Close the connection when EOF is received
		resp.CloseWrite()
		done <- err
	}()

	// Wait for either goroutine to finish
	return <-done
}

// Helper to create a tar stream from a directory
func CreateTarContext(dir string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = relPath

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		_, err = io.Copy(tw, file)
		return err
	})

	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf, nil
}
