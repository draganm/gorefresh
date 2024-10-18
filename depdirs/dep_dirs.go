package depdirs

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"sort"
	"strings"

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
			if err != nil {
				err = errors.Join(err, e)
			}
		}
	})

	if err != nil {
		return nil, fmt.Errorf("while loading packages: %w", err)
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

		packageFiles = append(packageFiles, pkg.OtherFiles...)
		packageFiles = append(packageFiles, pkg.IgnoredFiles...)

		dirs := lo.Map(packageFiles, func(n string, _ int) string {
			d := filepath.Dir(n)
			if !strings.HasSuffix(d, string(filepath.Separator)) {
				d = d + string(filepath.Separator)
			}
			return d
		})

		recurseEmbeddedDirs, err := recurseEmbeddedDirs(pkg.EmbedFiles)
		if err != nil {
			return nil, fmt.Errorf("could not recurse embedded dirs: %w", err)
		}

		dirs = append(dirs, recurseEmbeddedDirs...)

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

func recurseEmbeddedDirs(embeddedFiles []string) ([]string, error) {
	allDirs := []string{}
	for _, embeddedFile := range embeddedFiles {
		embeddedDir := filepath.Dir(embeddedFile)
		allDirs = append(allDirs, embeddedDir)
	}

	allDirs = lo.Uniq(allDirs)

	res := []string{}

	for _, embeddedDir := range allDirs {
		embeddedSubDirs, err := recurseEmbeddedDir(embeddedDir)
		if err != nil {
			return nil, fmt.Errorf("could not recurse embedded dir: %w", err)
		}
		res = append(res, embeddedSubDirs...)
	}

	res = lo.Map(res, func(n string, _ int) string {
		if !strings.HasSuffix(n, string(filepath.Separator)) {
			n = n + string(filepath.Separator)
		}
		return n
	})

	return lo.Uniq(res), nil
}

func recurseEmbeddedDir(embeddedDir string) ([]string, error) {
	allDirs := []string{}
	filepath.WalkDir(embeddedDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("could not walk dir: %w", err)
		}
		if d.IsDir() {
			allDirs = append(allDirs, path)
		}
		return nil
	})
	return allDirs, nil
}
