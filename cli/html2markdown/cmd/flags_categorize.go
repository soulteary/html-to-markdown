package cmd

import (
	"fmt"
	"strings"
)

const flagProvidedButNotDefinedErr = "flag provided but not defined: -"

func formatFlag(name string) string {
	if len(name) == 1 {
		return "-" + name
	}
	return "--" + name
}
func (cli *CLI) categorizeFlagError(err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()

	if strings.HasPrefix(message, flagProvidedButNotDefinedErr) {
		flagName := strings.TrimPrefix(message, flagProvidedButNotDefinedErr)

		return NewCLIError(fmt.Errorf("unknown flag: %s", formatFlag(flagName)))
	}

	return err
}
