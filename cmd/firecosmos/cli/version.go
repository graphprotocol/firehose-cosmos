package cli

import (
	"fmt"
	"strings"
)

func VersionString(version, commit, date string) string {
	var labels []string
	if len(commit) >= 7 {
		labels = append(labels, fmt.Sprintf("Commit %s", commit[0:7]))
	} else if commit != "" {
		labels = append(labels, fmt.Sprintf("Commit %s", commit))
	}

	if date != "" {
		labels = append(labels, fmt.Sprintf("Built %s", date))
	}

	if len(labels) == 0 {
		return version
	}

	return fmt.Sprintf("%s (%s)", version, strings.Join(labels, ", "))
}
