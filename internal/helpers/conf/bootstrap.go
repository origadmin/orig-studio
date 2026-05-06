/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package conf

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const defaultConfig = `server:
  http:
    addr: ":8080"
    timeout: "10s"

data:
  databases:
    default:
      name: origcms
      dialect: sqlite3
      source: "data/origcms.db?_fk=1"

security:
  authn:
    configs:
      - type: jwt
        jwt:
          signing_key: "change-me-in-production"
          signing_method: "HS256"
          access_token_ttl: "3600s"
          refresh_token_ttl: "720h"
`

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

// FindConfPath finds or generates the bootstrap config file path.
// Search order:
//  1. Flag-specified path (-conf)
//  2. ./resources/configs/bootstrap.yaml (dev mode, cwd)
//  3. <exec_dir>/resources/configs/bootstrap.yaml (deployed binary)
//  4. configs/bootstrap.yaml (cwd fallback)
//  5. Auto-generate configs/bootstrap.yaml with defaults
func FindConfPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}

	candidates := []string{
		"./resources/configs/bootstrap.yaml",
	}

	if exec, err := os.Executable(); err == nil {
		execDir := filepath.Dir(exec)
		candidates = append(candidates, filepath.Join(execDir, "resources", "configs", "bootstrap.yaml"))
	}

	candidates = append(candidates, "configs/bootstrap.yaml")

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	genPath := "configs/bootstrap.yaml"
	if err := generateDefaultConfig(genPath); err == nil {
		return genPath
	}

	return genPath
}

func generateDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(defaultConfig), 0644)
}
