// fsops.go — عمليات ملفات ذرية صغيرة مشتركة (tmp + rename).
package nexus

import (
	"encoding/json"
	"errors"
	"os"
)

func writeFileAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func readJSONInto(path string, dest any) error {
	b, err := readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return json.Unmarshal(b, dest)
}
