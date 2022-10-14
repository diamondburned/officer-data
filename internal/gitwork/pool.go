package gitwork

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/sync/singleflight"
)

// Pool helps manage a pool of git repositories.
type Pool struct {
	Author    Author
	RootPath  string
	RemoteURL string

	repoMu       sync.Mutex
	repoFlight   singleflight.Group
	repositories map[string]*PooledRepository
}

// NewPool creates a new pool with the given store.
func NewPool(rootPath, remoteURL string) (*Pool, error) {
	if rootPath == "" {
		rootPath = os.TempDir()
	}

	if err := os.MkdirAll(rootPath, os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "failed to create pool root directory")
	}

	return &Pool{
		Author:    DefaultAuthor,
		RootPath:  rootPath,
		RemoteURL: remoteURL,
	}, nil
}

// Clone clones the repository at the given URL into the pool. If dstDir is
// empty, then a random directory is created. If dstDir already exists, then
// the repository is opened instead.
func (p *Pool) Clone(ctx context.Context, shallow bool, dstDir string) (*PooledRepository, error) {
	if dstDir == "" {
		var err error

		dstDir, err = os.MkdirTemp(p.RootPath, "repo-")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create temporary directory")
		}
	}

	return p.lock(dstDir, func() (*Repository, error) {
		return Clone(ctx, p.RemoteURL, shallow, AtDir(filepath.Join(p.RootPath, dstDir)))
	})
}

// Open opens a repository in the pool. The given path is relative to the pool's
// root path.
func (p *Pool) Open(dir string) (*PooledRepository, error) {
	return p.lock(dir, func() (*Repository, error) {
		return Open(AtDir(filepath.Join(p.RootPath, dir)))
	})
}

func (p *Pool) lock(path string, f func() (*Repository, error)) (*PooledRepository, error) {
	p.repoMu.Lock()

	repo, ok := p.repositories[path]
	if ok {
		p.repoMu.Unlock()
		return repo, nil
	}

	p.repoMu.Unlock()

	v, err, _ := p.repoFlight.Do(path, func() (interface{}, error) {
		repo, err := f()
		if err != nil {
			return nil, err
		}

		pooled := &PooledRepository{
			Repository: repo,
			pool:       p,
			path:       path,
		}

		p.repoMu.Lock()
		p.repositories[path] = pooled
		p.repoMu.Unlock()

		return pooled, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(*PooledRepository), nil
}

// Delete deletes all repositories that belong to the pool. Beware when calling
// this method.
func (p *Pool) Delete() error {
	return os.RemoveAll(p.RootPath)
}

// PooledRepository is a repository that belongs to a pool.
type PooledRepository struct {
	*Repository
	pool *Pool
	path string
}

// Path returns the path to the repository in the pool.
func (r *PooledRepository) Path() string {
	return r.path
}

// AbsPath returns the absolute path to the repository in the pool.
func (r *PooledRepository) AbsPath() string {
	return filepath.Join(r.pool.RootPath, r.path)
}

// Pool returns the pool that the repository belongs to.
func (r *PooledRepository) Pool() *Pool {
	return r.pool
}
