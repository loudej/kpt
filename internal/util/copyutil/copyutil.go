package copyutil

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt/internal/types"
	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/sets"

	"github.com/GoogleContainerTools/kpt/internal/migration/io/ioutil"
	"github.com/GoogleContainerTools/kpt/internal/migration/os"
	"github.com/GoogleContainerTools/kpt/internal/migration/path/filepath"
)

// Diff returns a list of files that differ between the source and destination.
//
// Diff is guaranteed to return a non-empty set if any files differ, but
// this set is not guaranteed to contain all differing files.
func Diff(sourceDir, destDir types.FileSystemPath) (sets.String, error) {
	// get set of filenames in the package source
	upstreamFiles := sets.String{}
	err := filepath.Walk(sourceDir, func(path types.FileSystemPath, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip git repo if it exists
		if copyutil.IsDotGitFolder(path.Path) {
			return nil
		}

		upstreamFiles.Insert(strings.TrimPrefix(strings.TrimPrefix(path.Path, sourceDir.Path), string(filepath.Separator)))
		return nil
	})
	if err != nil {
		return sets.String{}, err
	}

	// get set of filenames in the cloned package
	localFiles := sets.String{}
	err = filepath.Walk(destDir, func(path types.FileSystemPath, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip git repo if it exists
		if copyutil.IsDotGitFolder(path.Path) {
			return nil
		}

		localFiles.Insert(strings.TrimPrefix(strings.TrimPrefix(path.Path, destDir.Path), string(filepath.Separator)))
		return nil
	})
	if err != nil {
		return sets.String{}, err
	}

	// verify the source and cloned packages have the same set of filenames
	diff := upstreamFiles.SymmetricDifference(localFiles)

	// verify file contents match
	for _, f := range upstreamFiles.Intersection(localFiles).List() {
		fi, err := os.Stat(filepath.Join(destDir, f))
		if err != nil {
			return diff, err
		}
		if fi.Mode().IsDir() {
			// already checked that this directory exists in the local files
			continue
		}

		// compare upstreamFiles
		b1, err := ioutil.ReadFile(filepath.Join(destDir, f))
		if err != nil {
			return diff, err
		}
		b2, err := ioutil.ReadFile(filepath.Join(sourceDir, f))
		if err != nil {
			return diff, err
		}
		if !bytes.Equal(b1, b2) {
			fmt.Println(copyutil.PrettyFileDiff(string(b1), string(b2)))
			diff.Insert(f)
		}
	}
	// return the differing files
	return diff, nil
}

// SyncFile copies file from src file path to a dst file path by replacement
// deletes dst file if src file doesn't exist
func SyncFile(src, dst types.FileSystemPath) error {
	srcFileInfo, err := os.Stat(src)
	if err != nil {
		// delete dst if source doesn't exist
		if err = deleteFile(dst); err != nil {
			return err
		}
		return nil
	}

	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	var filePerm os.FileMode

	// get the destination file perm if file exists
	dstFileInfo, err := os.Stat(dst)
	if err != nil {
		// get source file perm if destination file doesn't exist
		filePerm = srcFileInfo.Mode().Perm()
	} else {
		filePerm = dstFileInfo.Mode().Perm()
	}

	err = ioutil.WriteFile(dst, input, filePerm)
	if err != nil {
		return err
	}

	return nil
}

// deleteFile deletes file from path, returns no error if file doesn't exist
func deleteFile(path types.FileSystemPath) error {
	_, err := os.Stat(path)
	if err != nil {
		// return nil if file doesn't exist
		return nil
	}
	return os.Remove(path)
}
