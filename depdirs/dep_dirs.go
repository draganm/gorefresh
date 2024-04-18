package depdirs

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"sort"

	"github.com/samber/lo"
	"golang.org/x/tools/go/packages"
)

func DependencyDirs(
	pkgDir string,
) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedDeps |
			packages.NeedImports |
			packages.NeedName |
			packages.NeedEmbedFiles |
			packages.NeedFiles,
		Dir: pkgDir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, fmt.Errorf("could not open packages: %w", err)
	}

	packageDirs := []string{}

	err = nil
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		for _, e := range p.Errors {
			err = errors.Join(err, e)
		}
	})

	if err != nil {
		return nil, fmt.Errorf("while loading packages:\n%w", err)
	}

	allPackages := map[string]*packages.Package{}

	for _, pkg := range pkgs {
		visitDeps(pkg, allPackages)
	}

	packageNames := lo.Keys(allPackages)
	slices.Sort(packageNames)

	for _, packageName := range packageNames {

		pkg := allPackages[packageName]

		packageFiles := []string{}
		packageFiles = append(packageFiles, pkg.GoFiles...)
		packageFiles = append(packageFiles, pkg.EmbedFiles...)
		packageFiles = append(packageFiles, pkg.OtherFiles...)
		packageFiles = append(packageFiles, pkg.IgnoredFiles...)

		dirs := lo.Map(packageFiles, func(n string, _ int) string {
			return filepath.Dir(n)
		})

		packageDirs = append(packageDirs, lo.Uniq(dirs)...)

	}

	packageDirs = lo.Uniq(packageDirs)
	sort.Strings(packageDirs)

	return packageDirs, nil

}

func visitDeps(pkg *packages.Package, visit map[string]*packages.Package) {
	_, alreadyVisited := visit[pkg.PkgPath]
	if alreadyVisited {
		return
	}
	visit[pkg.PkgPath] = pkg

	for _, dep := range pkg.Imports {
		visitDeps(dep, visit)
	}

}
