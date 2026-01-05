/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package writer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

type IfExistsAction string

const (
	Override IfExistsAction = "Override"
	Skip     IfExistsAction = "Skip"
)

type ChartFile struct {
	// Path is the relative path to the chart directory
	Path string

	// Content is the textual content of the file
	Content string

	// IfExistsAction defines what to do with the existing file. By default, will keep the file.
	IfExistsAction
}

// ChartWriter writes the files to the disk
type ChartWriter struct {
	Directory  string
	FileSystem machinery.Filesystem
}

// WriteFile persists one file on the disk
func (c ChartWriter) WriteFile(file ChartFile) error {
	content := c.updateEOF(file.Content)
	path := filepath.Join(c.Directory, file.Path)

	// Check if the file already exists and should be skipped
	if _, err := os.Stat(path); !os.IsNotExist(err) && file.IfExistsAction == Skip {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := c.FileSystem.FS.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Use afero to write directly through the filesystem
	if err := afero.WriteFile(c.FileSystem.FS, path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// WriteFiles persists the files on the disk
func (c ChartWriter) WriteFiles(files []ChartFile) error {
	var errors []error
	for _, file := range files {
		err := c.WriteFile(file)
		if err != nil {
			errors = append(errors, fmt.Errorf("unable to write file %s: %w", file.Path, err))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("errors writing files: %v", errors)
	}
	return nil
}

func (c ChartWriter) updateEOF(content string) string {
	if content != "" && content[len(content)-1] != '\n' {
		content += "\n"
	}
	return content
}
