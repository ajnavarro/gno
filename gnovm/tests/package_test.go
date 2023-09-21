package tests

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	// "go/build"

	"github.com/jaekwon/testify/require"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestPackages(t *testing.T) {
	// find all packages with *_test.gno files.
	rootDirs := []string{
		filepath.Join("..", "stdlibs"),
		filepath.Join("..", "..", "examples"),
	}
	testDirs := map[string]string{} // aggregate here, pkgPath -> dir
	pkgPaths := []string{}
	for _, rootDir := range rootDirs {
		fileSystem := os.DirFS(rootDir)
		fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_test.gno") {
				dirPath := filepath.Dir(path)
				if _, exists := testDirs[dirPath]; exists {
					// already exists.
				} else {
					testDirs[dirPath] = filepath.Join(rootDir, dirPath)
					pkgPaths = append(pkgPaths, dirPath)
				}
			}
			return nil
		})
	}
	// Sort pkgPaths for determinism.
	sort.Strings(pkgPaths)
	// For each package with testfiles (in testDirs), call Machine.TestMemPackage.
	for _, pkgPath := range pkgPaths {
		testDir := testDirs[pkgPath]
		t.Run(pkgPath, func(t *testing.T) {
			if pkgPath == "gno.land/p/demo/avl" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/p/demo/flow" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/p/demo/grc/exts/vault" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/p/demo/grc/grc1155" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/p/demo/grc/grc20" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/p/demo/grc/grc721" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/p/demo/microblog" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/p/demo/rand" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/demo/art/gnoface" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/demo/foo1155" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/demo/foo20" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/demo/keystore" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/demo/tests" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/gnoland/blog" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/gnoland/faucet" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/gnoland/pages" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/x/manfred_outfmt" {
				t.Skip("TODO: fix test")
			}
			if pkgPath == "gno.land/r/x/nir1218_evaluation_proposal" {
				t.Skip("TODO: fix test")
			}

			runPackageTest(t, testDir, pkgPath)
		})
	}
}

func runPackageTest(t *testing.T, dir string, path string) {
	t.Helper()

	memPkg := gno.ReadMemPackage(dir, path)
	require.False(t, memPkg.IsEmpty())

	stdin := new(bytes.Buffer)
	// stdout := new(bytes.Buffer)
	stdout := os.Stdout
	stderr := new(bytes.Buffer)
	rootDir := filepath.Join("..", "..")
	store := TestStore(rootDir, path, stdin, stdout, stderr, ImportModeStdlibsOnly)
	store.SetLogStoreOps(true)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "test",
		Output:  stdout,
		Store:   store,
		Context: nil,
	})
	m.TestMemPackage(t, memPkg)

	// Check that machine is empty.
	err := m.CheckEmpty()
	if err != nil {
		t.Log("last state: \n", m.String())
		panic(fmt.Sprintf("machine not empty after main: %v", err))
	}
}
