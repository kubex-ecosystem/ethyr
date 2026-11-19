package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// Watcheri18nCmd is a function that registers the i18n watcher command.
func Watcheri18nCmd() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "uso: %s <caminho-do-projeto>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
}
