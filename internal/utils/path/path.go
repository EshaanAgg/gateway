// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package path

import (
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ValidateOutputPath takes an output file path and returns it as an absolute path.
// It returns an error if the absolute path cannot be determined or if the parent directory does not exist.
func ValidateOutputPath(outputPath string) (string, error) {
	outputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Dir(outputPath)); err != nil {
		return "", err
	}
	return outputPath, nil
}

// ListDirsAndFiles return a list of directories and files from a list of paths recursively.
func ListDirsAndFiles(paths []string) (dirs sets.Set[string], files sets.Set[string]) {
	dirs, files = sets.New[string](), sets.New[string]()
	// Separate paths by whether is a directory or not.
	paths = sets.NewString(paths...).UnsortedList()
	for _, path := range paths {
		var p os.FileInfo
		p, err := os.Lstat(path)
		if err != nil {
			// skip
			continue
		}

		if p.IsDir() {
			dirs.Insert(path)
		} else {
			files.Insert(path)
		}
	}

	// Ignore filepath if its parent directory is also be watched.
	var ignoreFiles []string
	for fp := range files {
		if dirs.Has(filepath.Dir(fp)) {
			ignoreFiles = append(ignoreFiles, fp)
		}
	}
	files.Delete(ignoreFiles...)

	return
}

// Traverses the directory recursively and adds the same to the subDirs.
// It only traverses the non-hidden directories.
func traverseSubDirs(curDir string, subDirs sets.Set[string]) {
	subDirs.Insert(curDir)
	files, err := os.ReadDir(curDir)
	if err != nil {
		// skip
		return
	}

	for _, file := range files {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
			fpath := filepath.Join(curDir, file.Name())
			traverseSubDirs(fpath, subDirs)
		}
	}
}

// GetSubDirs returns all the subdirectories of given directories.
// It only traverses the non-hidden directories recursively.
// The passed directories are also included in the result.
func GetSubDirs(initDirs []string) sets.Set[string] {
	dirs := sets.New[string]()
	for _, dir := range initDirs {
		traverseSubDirs(dir, dirs)
	}
	return dirs
}

// GetParentDirs returns all the parent directories of given files.
func GetParentDirs(files []string) sets.Set[string] {
	parents := sets.New[string]()
	for _, f := range files {
		parents.Insert(filepath.Dir(f))
	}
	return parents
}
