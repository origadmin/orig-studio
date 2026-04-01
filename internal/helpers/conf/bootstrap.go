/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// InitEnvAndConf initializes environment variables and finds the configuration file.
func InitEnvAndConf(moduleEnvSuffix string, flagConfPath string) string {
	envFiles := []string{}
	baseEnvPath := findEnvPath(".env")
	if baseEnvPath != "" {
		envFiles = append(envFiles, baseEnvPath)
	}

	if moduleEnvSuffix != "" {
		moduleEnvPath := findEnvPath(".env" + moduleEnvSuffix)
		if moduleEnvPath != "" {
			envFiles = append(envFiles, moduleEnvPath)
		}
	}

	if len(envFiles) > 0 {
		_ = godotenv.Overload(envFiles...)
	}

	return FindConfPath(flagConfPath)
}

func findEnvPath(envName string) string {
	exec, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(exec)
		p := filepath.Join(execDir, "resources", envName)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	devPath := filepath.Join("resources", envName)
	if _, err := os.Stat(devPath); err == nil {
		return devPath
	}
	return ""
}

func FindConfPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}
	devPath := "./resources/configs/bootstrap.yaml"
	if _, err := os.Stat(devPath); err == nil {
		return devPath
	}
	return "configs/bootstrap.yaml"
}
