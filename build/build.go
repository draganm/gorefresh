package build

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func BuildBinary(ctx context.Context, dir string) (string, error) {

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "gazer-bin")
	if err != nil {
		return "", err
	}

	err = tmpFile.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	fileName := tmpFile.Name()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", fileName, ".")
	fmt.Println(cmd.Args)

	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output
	cmd.Dir = absDir

	err = cmd.Run()
	if err != nil {

		return "", errors.Join(
			fmt.Errorf("failed to build binary: %w\n%s", err, output.String()),
			os.Remove(fileName),
		)
	}

	return fileName, nil
}
