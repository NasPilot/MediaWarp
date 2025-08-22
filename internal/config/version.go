package config

import (
	"MediaWarp/constants"
	"time"
)

var (
	appVersion string = "v0.1.14"
	commitHash string = "Unknown"
	buildDate  string = "Unknown"
)

func parseBuildTime(s string) string {
	if t, err := time.Parse(time.RFC3339, s); err != nil {
		return "Unkown"
	} else {
		return t.Local().Format(constants.FORMATE_TIME + " -07:00")
	}
}
