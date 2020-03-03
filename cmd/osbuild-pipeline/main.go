package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/osbuild/osbuild-composer/internal/common"

	"github.com/osbuild/osbuild-composer/internal/blueprint"
	"github.com/osbuild/osbuild-composer/internal/distro"
	"github.com/osbuild/osbuild-composer/internal/rpmmd"
)

func main() {
	var imageType string
	var blueprintArg string
	var archArg string
	var distroArg string
	flag.StringVar(&imageType, "image-type", "", "image type, e.g. qcow2 or ami")
	flag.StringVar(&blueprintArg, "blueprint", "", "path to a JSON file containing a blueprint to translate")
	flag.StringVar(&archArg, "arch", "", "architecture to create image for, e.g. x86_64")
	flag.StringVar(&distroArg, "distro", "", "distribution to create, e.g. fedora-30")
	flag.Parse()

	// Print help usage if one of the required arguments wasn't provided
	if imageType == "" || blueprintArg == "" || archArg == "" || distroArg == "" {
		flag.Usage()
		return
	}

	// Validate architecture
	if !common.ArchitectureExists(archArg) {
		_, _ = fmt.Fprintf(os.Stderr, "The provided architecture (%s) is not supported. Use one of these:\n", archArg)
		for _, arch := range common.ListArchitectures() {
			_, _ = fmt.Fprintln(os.Stderr, " *", arch)
		}
		return
	}

	blueprint := &blueprint.Blueprint{}
	if blueprintArg != "" {
		file, err := ioutil.ReadFile(blueprintArg)
		if err != nil {
			panic("Could not find blueprint: " + err.Error())
		}
		err = json.Unmarshal([]byte(file), &blueprint)
		if err != nil {
			panic("Could not parse blueprint: " + err.Error())
		}
	}

	distros, err := distro.NewDefaultRegistry([]string{"."})
	if err != nil {
		panic(err)
	}

	d := distros.GetDistro(distroArg)
	if d == nil {
		_, _ = fmt.Fprintf(os.Stderr, "The provided distribution (%s) is not supported. Use one of these:\n", distroArg)
		for _, distro := range distros.List() {
			_, _ = fmt.Fprintln(os.Stderr, " *", distro)
		}
		return
	}

	packages := make([]string, len(blueprint.Packages))
	for i, pkg := range blueprint.Packages {
		packages[i] = pkg.Name
		// If a package has version "*" the package name suffix must be equal to "-*-*.*"
		// Using just "-*" would find any other package containing the package name
		if pkg.Version != "" && pkg.Version != "*" {
			packages[i] += "-" + pkg.Version
		} else if pkg.Version == "*" {
			packages[i] += "-*-*.*"
		}
	}

	pkgs, exclude_pkgs, err := d.BasePackages(imageType, archArg)
	if err != nil {
		panic("could not get base packages: " + err.Error())
	}
	packages = append(pkgs, packages...)

	home, err := os.UserHomeDir()
	if err != nil {
		panic("os.UserHomeDir(): " + err.Error())
	}

	rpmmd := rpmmd.NewRPMMD(path.Join(home, ".cache/osbuild-composer/rpmmd"))
	packageSpecs, checksums, err := rpmmd.Depsolve(packages, exclude_pkgs, d.Repositories(archArg), d.ModulePlatformID(), false)
	if err != nil {
		panic("Could not depsolve: " + err.Error())
	}

	buildPkgs, err := d.BuildPackages(archArg)
	if err != nil {
		panic("Could not get build packages: " + err.Error())
	}
	buildPackageSpecs, _, err := rpmmd.Depsolve(buildPkgs, nil, d.Repositories(archArg), d.ModulePlatformID(), false)
	if err != nil {
		panic("Could not depsolve build packages: " + err.Error())
	}

	size := d.GetSizeForOutputType(imageType, 0)
	pipeline, err := d.Pipeline(blueprint, nil, packageSpecs, buildPackageSpecs, checksums, archArg, imageType, size)
	if err != nil {
		panic(err.Error())
	}

	bytes, err := json.Marshal(pipeline)
	if err != nil {
		panic("could not marshal pipeline into JSON")
	}

	os.Stdout.Write(bytes)
}
