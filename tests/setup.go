package tests

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func RunTest(t *testing.T, dir string) error {
	startDir := filepath.Join(dir, "start")
	workDir := filepath.Join(dir, "work")

	err := copyDir(startDir, workDir)
	require.NoError(t, err)
	defer func() { require.NoError(t, os.RemoveAll(workDir)) }()

	cmd := exec.Command("go", "generate", "./...")
	cmd.Dir = workDir
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=$PATH:%s", filepath.Dir(dir)))
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", os.Getenv("HOME")))

	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr

	t.Log("go generate ./...")
	err = cmd.Run()
	for _, s := range strings.Split(strings.TrimSpace(stdOut.String()), "\n") {
		t.Log(s)
	}
	if err != nil {
		return errors.Join(err, errors.New(stdOut.String()), errors.New(stdErr.String()))
	}

	compareDirs(t, workDir, filepath.Join(dir, "expected"))

	return nil
}

func compareDirs(t *testing.T, dir1, dir2 string) {
	_ = filepath.Walk(dir1, func(f1Path string, f1Info fs.FileInfo, err error) error {
		require.NoError(t, err)

		f2Path := filepath.Join(dir2, strings.TrimPrefix(f1Path, dir1))
		if strings.HasSuffix(f2Path, ".go") {
			f2Path += ".txt"
		}

		f2, err := os.Open(f2Path)
		require.NoError(t, err)

		f2Info, err := f2.Stat()
		require.NoError(t, err)

		if f1Info.IsDir() && f2Info.IsDir() {
			return nil
		}

		if f1Info.IsDir() && !f2Info.IsDir() {
			t.Error(fmt.Errorf("%s not dir", f2Path))
		}

		if !f1Info.IsDir() && f2Info.IsDir() {
			t.Error(fmt.Errorf("%s not file", f2Path))
		}

		f1Bytes, err := os.ReadFile(f1Path)
		require.NoError(t, err)

		f2Bytes, err := os.ReadFile(f2Path)
		require.NoError(t, err)

		require.Equal(t, trimLines(string(f1Bytes)), trimLines(string(f2Bytes)))

		return nil
	})
}

func copyDir(from string, to string) error {
	return filepath.Walk(from, func(path string, info fs.FileInfo, err error) error {
		dest := strings.Replace(path, from, to, 1)
		if info.IsDir() {
			err := os.MkdirAll(dest, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			err = os.WriteFile(strings.TrimSuffix(dest, ".txt"), src, fs.ModePerm)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func trimLines(ll string) string {
	lines := strings.Split(ll, "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}
