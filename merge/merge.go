package merge

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/edea-dev/edea/backend/model"
	"gitlab.com/edea-dev/edea/backend/repo"
	"gopkg.in/yaml.v3"
)

// TODO: we need to finish defining the file format so that we can put multiple
// modules into a single repository. we should make it possible to use user-defined
// IDs for sub-modules too and later on allow to reference dependencies.
//
// For now we only support the first module out of a project.

// Project is the top level project configuration
type Project struct {
	Name    string   `yaml:"name"`
	Modules []Module `yaml:"modules"`
}

// Module references the schematic and pcb for this module
type Module struct {
	Name      string `yaml:"name"`
	Schematic string `yaml:"sch"`
	PCB       string `yaml:"pcb"`
	// TODO: add configuration here
}

// Merge bench modules together
func Merge(modules []model.BenchModule) ([]byte, error) {

	pcb, err := os.CreateTemp("", "merge.*.kicad_pcb")
	if err != nil {
		return nil, err
	}
	defer os.Remove(pcb.Name())

	sch, err := os.CreateTemp("", "merge.*.sch")
	if err != nil {
		return nil, err
	}
	defer os.Remove(sch.Name())

	log.Debug().Msgf("creating temp pcb file: %s", pcb.Name(), sch.Name())

	// processing projects should not take longer than a minute
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, mod := range modules {
		g := &repo.Git{URL: mod.Module.RepoURL}
		p := &Project{}

		// read and parse the module configuration out of the repo
		s, err := g.File("edea.yml", false)
		if err != nil {
			return nil, fmt.Errorf("module %s does not contain an edea.yml file", mod.ModuleID)
		}
		if err := yaml.Unmarshal([]byte(s), p); err != nil {
			return nil, err
		}

		pcbFile, err := g.File(p.Modules[0].PCB, true)
		if err != nil {
			return nil, err
		}
		schFile, err := g.File(p.Modules[0].Schematic, true)
		if err != nil {
			return nil, err
		}

		pcbCmd := exec.CommandContext(ctx, "cat", pcb.Name())
		schCmd := exec.CommandContext(ctx, "cat", sch.Name())

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			if err := pcbCmd.Run(); err != nil {
				log.Panic().Err(err).Msg("error while processing pcb file")
			}
		}(wg)

		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			if err := schCmd.Run(); err != nil {
				log.Panic().Err(err).Msg("error while processing sch file")
			}
		}(wg)

		// write the files to merge over stdin to the tool
		pp, err := pcbCmd.StdinPipe()
		pp.Write([]byte(pcbFile))
		pp.Close()

		sp, err := schCmd.StdinPipe()
		sp.Write([]byte(schFile))
		sp.Close()

		// wait for the tools to exit
		wg.Wait()
	}

	// now we need to create a zip archive of the merged project

	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	// read the contents of our newly merged output
	mergedPCB, err := io.ReadAll(pcb)
	mergedSCH, err := io.ReadAll(sch)

	// TESTING: also put the full bench_module spec into the archive
	spec, err := json.MarshalIndent(modules, "", "\t")

	// Add some files to the archive.
	var files = []struct {
		Name string
		Body []byte
	}{
		{"merged.kicad_pcb", mergedPCB},
		{"merged.sch", mergedSCH},
		{"bench.json", spec},
	}
	for _, file := range files {
		f, err := w.Create(file.Name)
		if err != nil {
			log.Panic().Err(err).Msg("could not create file in archive")
		}
		_, err = f.Write(file.Body)
		if err != nil {
			log.Panic().Err(err).Msg("could not write file in archive")
		}
	}

	// Make sure to check the error on Close.

	if err = w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
