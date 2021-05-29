package merge

// SPDX-License-Identifier: EUPL-1.2

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/config"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/repo"
	"gitlab.com/edea-dev/edea/backend/util"
	"gopkg.in/yaml.v3"
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

	log.Debug().Msgf("created temp directory : %s", dir)

	// processing projects should not take longer than a minute
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var moduleSpec []string

	for _, mod := range modules {
		g := &repo.Git{URL: mod.Module.RepoURL}
		p := &repo.Project{}

		// read and parse the module configuration out of the repo
		s, err := g.File("edea.yml", false)
		if err != nil {
			// assuming old format, i.e. no sub-modules
			log.Info().Msgf("module %s does not contain an edea.yml file, assuming project files are in top-level dir", mod.ModuleID)

			repoDir, _ := g.Dir()
			moduleSpec = append(moduleSpec, repoDir)
			continue
		}
		if err := yaml.Unmarshal([]byte(s), p); err != nil {
			return nil, util.HintError{
				Hint: fmt.Sprintf("Could not parse edea.yml for \"%s\"\nTry checking if the syntax is correct.", mod.Name),
				Err:  err,
			}
		}

		v, ok := p.Modules[mod.Module.Sub]
		if !ok {
			log.Panic().Err(errors.New("sub-module specified but does not exist")).Msg("the sub-module key in the database does not exist in the repo edea.yml")
		}

		repoDir, _ := g.Dir() // at this point we already know the it's cached
		dir := strings.ReplaceAll(v.Directory, "../", "")
		dir = strings.TrimPrefix(dir, "/")
		dir = filepath.Join(repoDir, dir)
		moduleSpec = append(moduleSpec, dir)
	}

	argv := []string{"edea_merge_tool", "--output", projectDir, "--module"}
	argv = append(argv, moduleSpec...)

	mergeCmd := exec.CommandContext(ctx, "/usr/bin/python3", argv...)

	mergeCmd.Dir = config.Cfg.MergeTool

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
			log.Panic().Err(err).Msg("could not create file in archive")
		}
		_, err = f.Write(file.Body)
		if err != nil {
			log.Panic().Err(err).Msg("could not write file in archive")
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
