package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadVar(name string) string {
	variable := os.Getenv(name)
	if variable != "" {
		return variable
	}
	return loadVarFromPath(name)
}

func loadVarFromPath(name string) string {
	filename := fmt.Sprintf("/run/secrets/%s", strings.ToLower(name))

	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	variable, _ := reader.ReadString('\n')
	if variable != "" {
		return strings.Trim(variable, "\n")
	}

	return ""
}
