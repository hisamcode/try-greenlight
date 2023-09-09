package vcs

import (
	"fmt"
	"runtime/debug"
)

func Version() string {
	var (
		time     string
		revision string
		modified bool
	)

	bi, ok := debug.ReadBuildInfo()
	if ok {
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.time":
				time = s.Value
			case "vcs.revision":
				revision = s.Value
			case "vcs.modified":
				if s.Value == "true" {
					modified = true
				}
			}
		}
	}

	// Its modified, ada perubahan di versinya, jadi ada tambahan -dirty suffix.
	// Commit dulu, lalu build ulang, baru version suffix -dirty nya ilang.
	if modified {
		return fmt.Sprintf("%s-%s-dirty", time, revision)
	}

	return fmt.Sprintf("%s-%s", time, revision)
}
