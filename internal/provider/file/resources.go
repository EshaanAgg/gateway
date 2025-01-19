// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

// loadFromFilesAndDirs loads resources from specific files and directories.
// The directories are traversed recursively to load resources from all files.
func loadFromFilesAndDirs(files, dirs []string) ([]*resource.Resources, error) {
	var rs []*resource.Resources

	for _, file := range files {
		r, err := loadFromFile(file)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}

	for _, dir := range dirs {
		r, err := loadFromDir(dir)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r...)
	}

	return rs, nil
}

// loadFromFile loads resources from a specific file.
func loadFromFile(path string) (*resource.Resources, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s is not exist", path)
		}
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return resource.LoadResourcesFromYAMLBytes(bytes, false)
}

func loadFromDir(path string) ([]*resource.Resources, error) {
	rs := make([]*resource.Resources, 0)

	// This function modifies the `rs` slice directly.
	err := traverseDirectory(path, &rs)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// traverseDirectory is a helper function that recursively traverses the directory
// and loads resources from all files while skipping hidden files and directories.
func traverseDirectory(dirPath string, rs *[]*resource.Resources) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// Skip hidden files and directories.
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			// Recursively process subdirectories.
			if err := traverseDirectory(fullPath, rs); err != nil {
				return err
			}
		} else {
			// Load resources from files.
			r, err := loadFromFile(fullPath)
			if err != nil {
				return err
			}
			*rs = append(*rs, r)
		}
	}

	return nil
}
