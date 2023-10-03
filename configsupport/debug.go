package configsupport

import "os"

var isDebug bool

func init() {
	switch os.Getenv("DEBUG") {
	case "1", "true":
		isDebug = true
	}
}

// IsDebug returns true when DEBUG environment variable is "1" or "true".
func IsDebug() bool {
	return isDebug
}
