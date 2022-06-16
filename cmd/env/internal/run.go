package internal

import (
	"io"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Run will write a TOML environment definition to the output io.Writer with
// the following rules:
// 1. If $PYTHONPYCACHEPREFIX is set, do nothing
// 2. If $PYTHONPYCACHEPREFIX is unset, set it to the expanded value of
//    $HOME/.pycache
func Run(env []string, output io.Writer) error {
	var home string
	for _, v := range env {
		if strings.HasPrefix(v, "HOME=") {
			home = strings.TrimPrefix(v, "HOME=")
		}

		if strings.HasPrefix(v, "PYTHONPYCACHEPREFIX=") {
			return nil
		}
	}

	err := toml.NewEncoder(output).Encode(map[string]string{
		"PYTHONPYCACHEPREFIX": filepath.Join(home, ".pycache"),
	})
	if err != nil {
		return err
	}

	return nil
}
