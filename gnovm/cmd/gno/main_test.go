package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func TestMain_Gnodev(t *testing.T) {
	tc := []testMainCase{
		{args: []string{""}, errShouldBe: "flag: help requested"},
	}

	testMainCaseRun(t, tc)
}

type testMainCase struct {
	args                 []string
	testDir              string
	simulateExternalRepo bool

	// for the following FooContain+FooBe expected couples, if both are empty,
	// then the test suite will require that the "got" is not empty.
	errShouldContain     string
	errShouldBe          string
	stderrShouldContain  string
	stdoutShouldBe       string
	stdoutShouldContain  string
	stderrShouldBe       string
	recoverShouldContain string
	recoverShouldBe      string
}

func testMainCaseRun(t *testing.T, tc []testMainCase) {
	t.Helper()

	workingDir, err := os.Getwd()
	require.Nil(t, err)

	for _, test := range tc {
		errShouldBeEmpty := test.errShouldContain == "" && test.errShouldBe == ""
		stdoutShouldBeEmpty := test.stdoutShouldContain == "" && test.stdoutShouldBe == ""
		stderrShouldBeEmpty := test.stderrShouldContain == "" && test.stderrShouldBe == ""
		recoverShouldBeEmpty := test.recoverShouldContain == "" && test.recoverShouldBe == ""

		testName := strings.Join(test.args, " ")
		testName = strings.ReplaceAll(testName+test.testDir, "/", "~")

		t.Run(testName, func(t *testing.T) {
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")

			checkOutputs := func(t *testing.T) {
				t.Helper()

				if stdoutShouldBeEmpty {
					require.Empty(t, mockOut.String(), "stdout should be empty")
				} else {
					t.Log("stdout", mockOut.String())
					if test.stdoutShouldContain != "" {
						require.Contains(t, mockOut.String(), test.stdoutShouldContain, "stdout should contain")
					}
					if test.stdoutShouldBe != "" {
						require.Equal(t, test.stdoutShouldBe, mockOut.String(), "stdout should be")
					}
				}

				if stderrShouldBeEmpty {
					require.Empty(t, mockErr.String(), "stderr should be empty")
				} else {
					t.Log("stderr", mockErr.String())
					if test.stderrShouldContain != "" {
						require.Contains(t, mockErr.String(), test.stderrShouldContain, "stderr should contain")
					}
					if test.stderrShouldBe != "" {
						require.Equal(t, test.stderrShouldBe, mockErr.String(), "stderr should be")
					}
				}
			}

			defer func() {
				if r := recover(); r != nil {
					output := fmt.Sprintf("%v", r)
					t.Log("recover", output)
					require.False(t, recoverShouldBeEmpty, "should panic")
					require.True(t, errShouldBeEmpty, "should not return an error")
					if test.recoverShouldContain != "" {
						require.Regexpf(t, test.recoverShouldContain, output, "recover should contain")
					}
					if test.recoverShouldBe != "" {
						require.Equal(t, test.recoverShouldBe, output, "recover should be")
					}
					checkOutputs(t)
				} else {
					require.True(t, recoverShouldBeEmpty, "should not panic")
				}
			}()

			if test.simulateExternalRepo {
				// create external dir
				tmpDir, cleanUpFn := createTmpDir(t)
				defer cleanUpFn()

				// copy to external dir
				absTestDir, err := filepath.Abs(test.testDir)
				require.Nil(t, err)
				require.Nil(t, copyDir(absTestDir, tmpDir))

				// cd to tmp directory
				os.Chdir(tmpDir)
				defer os.Chdir(workingDir)
			}

			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))
			io.SetErr(commands.WriteNopCloser(mockErr))

			err := newGnocliCmd(io).ParseAndRun(context.Background(), test.args)

			if errShouldBeEmpty {
				require.Nil(t, err, "err should be nil")
			} else {
				t.Log("err", err.Error())
				require.NotNil(t, err, "err shouldn't be nil")
				if test.errShouldContain != "" {
					require.Contains(t, err.Error(), test.errShouldContain, "err should contain")
				}
				if test.errShouldBe != "" {
					require.Equal(t, test.errShouldBe, err.Error(), "err should be")
				}
			}

			checkOutputs(t)
		})
	}
}

func setupTestScript(t *testing.T, txtarDir string) testscript.Params {
	t.Helper()
	// Get root location of github.com/gnolang/gno
	goModPath, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
	require.NoError(t, err)
	rootDir := filepath.Dir(string(goModPath))
	// Build a fresh gno binary in a temp directory
	gnoBin := filepath.Join(t.TempDir(), "gno")
	err = exec.Command("go", "build", "-o", gnoBin, filepath.Join(rootDir, "gnovm", "cmd", "gno")).Run()
	require.NoError(t, err)
	// Define script params
	return testscript.Params{
		Setup: func(env *testscript.Env) error {
			env.Vars = append(env.Vars,
				"GNOROOT="+rootDir, // thx PR 1014 :)
				// by default, $HOME=/no-home, but we need an existing $HOME directory
				// because some commands needs to access $HOME/.cache/go-build
				"HOME="+t.TempDir(),
			)
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// add a custom "gno" command so txtar files can easily execute "gno"
			// without knowing where is the binary or how it is executed.
			"gno": func(ts *testscript.TestScript, neg bool, args []string) {
				err := ts.Exec(gnoBin, args...)
				if err != nil {
					ts.Logf("[%v]\n", err)
					if !neg {
						ts.Fatalf("unexpected gno command failure")
					}
				} else {
					if neg {
						ts.Fatalf("unexpected gno command success")
					}
				}
			},
		},
		Dir: txtarDir,
	}
}
