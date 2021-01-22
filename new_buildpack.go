package pack

import (
	"context"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"

	"github.com/buildpacks/pack/internal/build"
	"github.com/buildpacks/pack/internal/dist"
	"github.com/buildpacks/pack/internal/style"
)

var (
	bashBinBuild = `
#!/usr/bin/env bash

set -euo pipefail

layers_dir="$1"
env_dir="$2/env"
plan_path="$3"

exit 0
`
	bashBinDetect = `
#!/usr/bin/env bash

exit 0
`
)

type NewBuildpackOptions struct {
	// The base directory to generate assets
	Path string

	// The ID of the output buildpack artifact.
	ID string

	// The stacks this buildpack will work with
	Stacks []dist.Stack
}

func (c *Client) NewBuildpack(ctx context.Context, opts NewBuildpackOptions) error {
	buildpackTOML := dist.BuildpackDescriptor{
		API:    build.SupportedPlatformAPIVersions.Latest(),
		Stacks: opts.Stacks,
		Info: dist.BuildpackInfo{
			ID:      opts.ID,
			Version: "0.0.0",
		},
	}

	f, err := os.Create(filepath.Join(opts.Path, "buildpack.toml"))
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(buildpackTOML); err != nil {
		return err
	}
	defer f.Close()
	c.logger.Infof("    %s  buildpack.toml", style.Key("create"))

	if err := os.MkdirAll(filepath.Join(opts.Path, "bin"), 0755); err != nil {
		return err
	}

	return createBashBuildpack(opts.Path, c)
}

func createBashBuildpack(path string, c *Client) error {
	if err := createBinScript(path, "build", bashBinBuild); err != nil {
		return err
	}
	c.logger.Infof("    %s  bin/build", style.Key("create"))

	if err := createBinScript(path, "detect", bashBinDetect); err != nil {
		return err
	}
	c.logger.Infof("    %s  bin/build", style.Key("create"))

	return nil
}

func createBinScript(path, name, contents string) error {
	bin := filepath.Join(path, "bin", name)
	f, err := os.Create(bin)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.WriteString(contents); err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		if err = os.Chmod(bin, 0755); err != nil {
			return err
		}
	}
	return nil
}