package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Image struct {
	URI string `json:"uri"`
}

type GLTF struct {
	Images []Image `json:"images"`
}

func main() {
	// CLI flags
	gltfPath := flag.String("gltf", "", "Path to the .gltf file (required)")
	baseURL := flag.String("base", "", "Base URL for textures (required, e.g. https://example.com/assets/)")
	outputDir := flag.String("out", "textures", "Output directory for downloaded textures")
	flag.Parse()

	if *gltfPath == "" || *baseURL == "" {
		fmt.Println("Usage: go run ./tools -gltf model.gltf -base https://example.com/assets/ [-out textures]")
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, os.ModePerm); err != nil {
		fmt.Printf("❌ Failed to create output dir %s: %v\n", *outputDir, err)
		os.Exit(1)
	}

	file, err := os.Open(*gltfPath)
	if err != nil {
		fmt.Printf("❌ Error opening GLTF file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var gltf GLTF
	if err := json.NewDecoder(file).Decode(&gltf); err != nil {
		fmt.Printf("❌ Error decoding GLTF JSON: %v\n", err)
		os.Exit(1)
	}

	for _, img := range gltf.Images {
		if img.URI == "" {
			continue
		}

		url := fmt.Sprintf("%s%s", *baseURL, img.URI)
		localPath := filepath.Join(*outputDir, filepath.Base(img.URI))

		fmt.Printf("⬇️  Downloading %s → %s\n", img.URI, localPath)

		if err := downloadFile(url, localPath); err != nil {
			fmt.Printf("   ⚠️  Failed: %v\n", err)
		} else {
			fmt.Printf("   ✅ Saved: %s\n", localPath)
		}
	}
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

/*
Usage:
go run ./tools/download_textures_from_gltf/main.go -gltf <gltf path> \
  -base <Base url> \
  -out textures
** Base url for poly.cam is https://storage.polycam.io/captures/<UUID>/
** Paths are relative to terminal cwd
*/
