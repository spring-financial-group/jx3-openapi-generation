package utils

import "runtime"

func color(ansi string) string {
	if runtime.GOOS == "windows" {
		return ""
	}

	return ansi
}

var (
	Reset  = color("\033[0m")
	Red    = color("\033[31m")
	Green  = color("\033[32m")
	Yellow = color("\033[33m")
	Blue   = color("\033[34m")
	Purple = color("\033[35m")
	Cyan   = color("\033[36m")
	Gray   = color("\033[37m")
	White  = color("\033[97m")
)
