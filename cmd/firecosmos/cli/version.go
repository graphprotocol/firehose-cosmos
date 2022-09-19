package main

import (
	"fmt"
	"strings"
)

var (
	Version     = "0.5.0"
	BuildTime   = ""
	BuildCommit = ""
)

func versionString() string {
	var labels []string

	if len(BuildCommit) >= 7 {
		labels = append(labels, fmt.Sprintf("Commit %s", BuildCommit[0:7]))
	} else if BuildCommit != "" {
		labels = append(labels, fmt.Sprintf("Commit %s", BuildCommit))
	}

	if BuildTime != "" {
		labels = append(labels, fmt.Sprintf("Built %s", BuildTime))
	}

	if len(labels) == 0 {
		return Version
	}

	return fmt.Sprintf("%s (%s)", Version, strings.Join(labels, ", "))
}
