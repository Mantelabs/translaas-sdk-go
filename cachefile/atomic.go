package cachefile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func readJSONFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

func parseJSONFile[T any](path string) (*T, error) {
	data, err := readJSONFile(path)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("decode %s: %w", filepath.Base(path), err)
	}
	return &value, nil
}

func writeJSONAtomic(ctx context.Context, path string, value any) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, payload, 0o644); err != nil {
		return fmt.Errorf("write temp file %s: %w", tmpPath, err)
	}

	if err := ctx.Err(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %s: %w", path, err)
	}
	return nil
}
