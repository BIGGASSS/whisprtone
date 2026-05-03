package utils

import (
	"os/exec"
	"strings"
)

func Copy(content string) error {
    cmd := exec.Command("wl-copy")
    cmd.Stdin = strings.NewReader(content)
    return cmd.Run()
}
