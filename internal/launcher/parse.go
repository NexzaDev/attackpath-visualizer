package launcher

import (
	"fmt"
	"strings"
)

func ParseLaunchArgs(args []string) (bool, Options, string, error) {
	if len(args) == 0 {
		return true, Options{}, "", nil
	}
	if args[0] != "--ui" && args[0] != "--ask" && args[0] != "--no-prompt" && args[0] != "--db" && !strings.HasPrefix(args[0], "--ui=") {
		return false, Options{}, "", nil
	}
	options := Options{}
	dbPath := ""
	for index := 0; index < len(args); index++ {
		argument := args[index]
		switch argument {
		case "--ask":
			options.Ask = true
		case "--no-prompt":
			options.NoPrompt = true
		case "--ui":
			index++
			if index >= len(args) {
				return false, Options{}, "", fmt.Errorf("--ui requires gui or tui")
			}
			options.UI = args[index]
		case "--db":
			index++
			if index >= len(args) {
				return false, Options{}, "", fmt.Errorf("--db requires a path")
			}
			dbPath = args[index]
		case "--ui=gui":
			options.UI = "gui"
		default:
			if strings.HasPrefix(argument, "--ui=") {
				options.UI = strings.TrimPrefix(argument, "--ui=")
			} else {
				return false, Options{}, "", fmt.Errorf("unexpected launcher argument %q", argument)
			}
		}
	}
	return true, options, dbPath, nil
}
