// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/ignition/config/types"
)

const (
	DefaultDirectoryPermissions os.FileMode = 0755
	DefaultFilePermissions      os.FileMode = 0644
)

type File struct {
	Path     types.Path
	Contents []byte
	Mode     os.FileMode
	Uid      int
	Gid      int
}

func RenderFile(f types.File) *File {
	return &File{
		Path:     f.Path,
		Contents: []byte(f.Contents),
		Mode:     os.FileMode(f.Mode),
		Uid:      f.User.Id,
		Gid:      f.Group.Id,
	}
}

// WriteFile creates and writes the file described by f using the provided context
func (u Util) WriteFile(f *File) error {
	var err error

	path := u.JoinPath(string(f.Path))

	if err := mkdirForFile(path); err != nil {
		return err
	}

	// Create a temporary file in the same directory to ensure it's on the same filesystem
	var tmp *os.File
	if tmp, err = ioutil.TempFile(filepath.Dir(path), "tmp"); err != nil {
		return err
	}
	tmp.Close()
	defer func() {
		if err != nil {
			os.Remove(tmp.Name())
		}
	}()

	if err := ioutil.WriteFile(tmp.Name(), f.Contents, f.Mode); err != nil {
		return err
	}

	// XXX(vc): Note that we assume to be operating on the file we just wrote, this is only guaranteed
	// by using syscall.Fchown() and syscall.Fchmod()

	// Ensure the ownership and mode are as requested (since WriteFile can be affected by sticky bit)
	if err := os.Chown(tmp.Name(), f.Uid, f.Gid); err != nil {
		return err
	}

	if err := os.Chmod(tmp.Name(), f.Mode); err != nil {
		return err
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}

	return nil
}

// mkdirForFile helper creates the directory components of path
func mkdirForFile(path string) error {
	return os.MkdirAll(filepath.Dir(path), DefaultDirectoryPermissions)
}
