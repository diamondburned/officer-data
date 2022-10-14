package gitwork

import (
	"sync/atomic"

	"github.com/go-git/go-billy/v5"
)

// Additional flags that can be given to Repository.OpenFile.
const (
	// LockFile is a flag that can be given to Repository.OpenFile. It locks the
	// file for writing. The file is unlocked when it is closed.
	LockFile = -1 << iota
)

// RepositoryFile is a file in the repository.
type RepositoryFile struct {
	billy.File // treat the same as *os.File

	locked uint32
}

// Close closes the file.
func (r *RepositoryFile) Close() error {
	var lockedErr error
	if atomic.CompareAndSwapUint32(&r.locked, 1, 0) {
		lockedErr = r.File.Unlock()
	}
	if err := r.File.Close(); err != nil {
		return err
	}
	return lockedErr
}

// Wipe wipes the content of the file.
func (r *RepositoryFile) Wipe() error {
	return r.File.Truncate(0)
}
