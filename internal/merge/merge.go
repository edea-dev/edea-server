package merge

// SPDX-License-Identifier: EUPL-1.2

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gitlab.com/edea-dev/edea-server/internal/model"
	"gitlab.com/edea-dev/edea-server/internal/repo"
	"gitlab.com/edea-dev/edea-server/internal/util"
	"go.uber.org/zap"
)

// TODO: we need to finish defining the file format so that we can put multiple
// modules into a single repository. we should make it possible to use user-defined
// IDs for sub-modules too and later on allow to reference dependencies.
//
// For now we only support the first module out of a project.

// Merge bench modules together
func Merge(benchName string, modules []model.BenchModule) ([]byte, error) {

	dir, err := os.MkdirTemp("", "edea_merge")
	if err != nil {
		return nil, err
	}

	projectDir := filepath.Join(dir, benchName)

	// clean up after us
	defer os.RemoveAll(dir)

	zap.S().Debugf("created temp directory : %s", dir)

	// processing projects should not take longer than a minute
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var moduleSpec []string

	for _, mod := range modules {
		dir, err := repo.GetModulePath(&mod.Module)
		if err != nil {
			return nil, err
		}
		moduleSpec = append(moduleSpec, dir)
	}

	argv := []string{"-m", "edea", "--output", projectDir}
	argv = append(argv, moduleSpec...)

	mergeCmd := exec.CommandContext(ctx, "python3", argv...)

	// run the merge
	logOutput, err := mergeCmd.CombinedOutput()

	// return the output of the tool and the error for the user to debug issues
	if err != nil {
		return logOutput, util.HintError{
			Hint: "Something went wrong during the merge process, below is the log which should provide more information.",
			Err:  err,
		}
	}

	// now we need to create a zip archive of the merged project

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// TESTING: also put the full bench_module spec into the archive
	spec, err := json.MarshalIndent(modules, "", "\t")

	// Add some files to the archive.
	var files = []struct {
		Name string
		Body []byte
	}{
		{"edea_merge.log", logOutput},
		{"bench.json", spec},
	}
	for _, file := range files {
		f, err := w.Create(filepath.Join(benchName, file.Name))
		if err != nil {
			zap.L().Panic("could not create file in archive", zap.Error(err))
		}
		_, err = f.Write(file.Body)
		if err != nil {
			zap.L().Panic("could not write file in archive", zap.Error(err))
		}
	}

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// convert absolute fs paths to relative archive paths
		f, err := w.Create(filepath.Join(benchName, filepath.Base(file.Name())))
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	// walk the output directory to archive the project files
	if err := filepath.Walk(projectDir, walker); err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Metadata extracts some data from the module
func Metadata(module *model.Module) (map[string]interface{}, error) {

	// processing projects should not take longer than a minute
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	dir, err := repo.GetModulePath(module)
	if err != nil {
		return nil, fmt.Errorf("could not get module path: %w", err)
	}

	mergeCmd := exec.CommandContext(ctx, "python3", "-m", "edea", "--extract-meta", dir)

	// run the merge
	logOutput, err := mergeCmd.CombinedOutput()

	// return the output of the tool and the error for the user to debug issues
	if err != nil {
		zap.L().Warn("metadata extraction failed", zap.String("path", dir))
		return nil, util.HintError{
			Hint: fmt.Sprintf("Something went wrong during the metadata extraction, below is the log which should provide more information:\n%s", logOutput),
			Err:  err,
		}
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(logOutput, &m)
	return m, err
}
