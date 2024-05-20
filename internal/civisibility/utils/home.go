package utils

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func ExpandPath(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return path
	}

	homeFolder := getHomeDir()
	if len(homeFolder) > 0 {
		return filepath.Join(homeFolder, path[1:])
	}

	return path
}

func getHomeDir() string {
	if runtime.GOOS == "windows" {
		if home := os.Getenv("HOME"); home != "" {
			// First prefer the HOME environmental variable
			return home
		}
		if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
			// Prefer standard environment variable USERPROFILE
			return userProfile
		}

		homeDrive := os.Getenv("HOMEDRIVE")
		homePath := os.Getenv("HOMEPATH")
		return homeDrive + homePath
	}

	homeEnv := "HOME"
	if runtime.GOOS == "plan9" {
		// On plan9, env vars are lowercase.
		homeEnv = "home"
	}

	if home := os.Getenv(homeEnv); home != "" {
		// First prefer the HOME environmental variable
		return home
	}

	var stdout bytes.Buffer
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("sh", "-c", `dscl -q . -read /Users/"$(whoami)" NFSHomeDirectory | sed 's/^[^ ]*: //'`)
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil {
			result := strings.TrimSpace(stdout.String())
			if result != "" {
				return result
			}
		}
	} else {
		cmd := exec.Command("getent", "passwd", strconv.Itoa(os.Getuid()))
		cmd.Stdout = &stdout
		if err := cmd.Run(); err == nil {
			if passwd := strings.TrimSpace(stdout.String()); passwd != "" {
				// username:password:uid:gid:gecos:home:shell
				passwdParts := strings.SplitN(passwd, ":", 7)
				if len(passwdParts) > 5 {
					return passwdParts[5]
				}
			}
		}
	}

	// If all else fails, try the shell
	stdout.Reset()
	cmd := exec.Command("sh", "-c", "cd && pwd")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		return strings.TrimSpace(stdout.String())
	}

	return ""
}
