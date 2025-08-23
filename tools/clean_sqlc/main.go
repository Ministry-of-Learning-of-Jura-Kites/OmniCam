package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SQL []struct {
		Codegen []struct {
			Out string `yaml:"out"`
		} `yaml:"codegen"`
	} `yaml:"sql"`
}

func main() {
	// Read sqlc.yaml
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalln("Unable to get caller information")
		return
	}

	// get omnicam root path
	// root/tools/clean_sql.go
	rootpath := filepath.Dir(filepath.Dir(filename))

	data, err := os.ReadFile(filepath.Join(rootpath, "sqlc.yaml"))
	if err != nil {
		log.Fatalf("failed to read sqlc.yaml: %v", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse sqlc.yaml: %v", err)
	}

	// Extract output paths
	for _, sql := range cfg.SQL {
		for _, cg := range sql.Codegen {
			if cg.Out != "" {
				fmt.Printf("Cleaning directory: %s\n", cg.Out)
				if err := deleteContent(cg.Out); err != nil {
					log.Fatalf("failed cleaning %s: %v", cg.Out, err)
				}
			}
		}
	}
}

func deleteContent(path string) error {
	// return filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	os.Remov
	// 	return nil
	// })
	return os.RemoveAll(path)
}
