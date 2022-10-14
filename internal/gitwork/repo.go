package gitwork

import (
	"context"
	"go/doc"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/pkg/errors"

	gitconfig "github.com/go-git/go-git/v5/config"
	gitplumbing "github.com/go-git/go-git/v5/plumbing"
	gitobject "github.com/go-git/go-git/v5/plumbing/object"
)

type RepositoryStore func() (billy.Filesystem, error)

// AtDir returns a RepositoryStore that creates a repository at the given
// directory.
func AtDir(path string) RepositoryStore {
	return func() (billy.Filesystem, error) {
		return osfs.New(path), nil
	}
}

// AtTempDir returns a RepositoryStore that creates a repository at a temporary
// directory.
func AtTempDir(dir string) RepositoryStore {
	dir = filepath.Join(os.TempDir(), dir)
	return func() (billy.Filesystem, error) {
		tmp, err := os.MkdirTemp(dir, "gitwork")
		if err != nil {
			return nil, err
		}
		return AtDir(tmp)()
	}
}

func newStorage(fs billy.Filesystem) storage.Storer {
	return filesystem.NewStorageWithOptions(fs, nil, filesystem.Options{
		ExclusiveAccess:    true,
		MaxOpenDescriptors: 4,
	})
}

// CommitHash is the hash of the current commit.
type CommitHash = gitplumbing.Hash

// Author is the author used for making commits.
type Author struct {
	Name  string
	Email string
}

// DefaultAuthor is the default author used for making commits.
var DefaultAuthor = Author{
	Name:  "gitwork",
	Email: "gitwork@localhost", // hostname is used, not localhost
}

func init() {
	hostname, err := os.Hostname()
	if err == nil {
		DefaultAuthor.Email = "gitwork@" + hostname
	}
}

// Repository is a Git workspace. It mimics a regular Git repository.
type Repository struct {
	*git.Repository
	Config struct {
		Author Author
	}
}

// Clone clones a repository into the workspace.
func Clone(ctx context.Context, url string, shallow bool, dst RepositoryStore) (*Repository, error) {
	fs, err := dst()
	if err != nil {
		return nil, err
	}

	opts := &git.CloneOptions{URL: url}
	if shallow {
		opts.Depth = 1
	}

	repo, err := git.CloneContext(ctx, newStorage(fs), fs, opts)
	if err != nil {
		return nil, errors.Wrap(err, "cannot clone repository")
	}

	return newRepository(repo), nil
}

// Open opens an existing repository.
func Open(dst RepositoryStore) (*Repository, error) {
	fs, err := dst()
	if err != nil {
		return nil, err
	}

	repo, err := git.Open(newStorage(fs), fs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open repository")
	}

	return newRepository(repo), nil
}

// Init initializes the workspace with nothing.
func Init(dst RepositoryStore) (*Repository, error) {
	fs, err := dst()
	if err != nil {
		return nil, err
	}

	repo, err := git.Init(newStorage(fs), fs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot init repository")
	}

	return newRepository(repo), nil
}

func newRepository(repo *git.Repository) *Repository {
	r := Repository{
		Repository: repo,
	}
	r.Config.Author = DefaultAuthor
	return &r
}

// Worktree returns the worktree of the repository. It can be used to add and
// checkout files.
func (r *Repository) Worktree() *git.Worktree {
	worktree, err := r.Repository.Worktree()
	if err != nil {
		panic(err)
	}
	return worktree
}

// FS returns the filesystem of the repository.
func (r *Repository) FS() billy.Filesystem {
	return r.Worktree().Filesystem
}

// EditFile edits or creates a file in the repository. f is called on the file
// once it's opened. The file is locked using flock while f is running.
func (r *Repository) OpenFile(path string, flag int) (*RepositoryFile, error) {
	fs := r.FS()

	file, err := fs.OpenFile(path, flag, os.ModePerm)
	if err != nil {
		return nil, err
	}

	repoFile := &RepositoryFile{
		File: file,
	}

	if flag&LockFile != 0 {
		repoFile.locked = 1
		if err := file.Lock(); err != nil {
			return nil, errors.Wrap(err, "cannot lock file")
		}
	}

	return repoFile, nil
}

// Add adds the given files to the index.
func (r *Repository) Add(path string) error {
	_, err := r.Worktree().Add(path)
	return err
}

// AddAll adds all files to the index.
func (r *Repository) AddAll() error {
	return r.Worktree().AddWithOptions(&git.AddOptions{All: true})
}

// Commit commits the changes in the repository. Changes must be added using
// Add.
func (r *Repository) Commit(title, body string) (CommitHash, error) {
	message := title
	if body != "" {
		message += "\n\n" + columnWrap(body, 72)
	}

	tree := r.Worktree()

	return tree.Commit(message, &git.CommitOptions{
		All: false,
		Author: &gitobject.Signature{
			Name:  r.Config.Author.Name,
			Email: r.Config.Author.Email,
			When:  time.Now(),
		},
	})
}

func columnWrap(txt string, col int) string {
	var buf strings.Builder
	doc.ToText(&buf, txt, "", "  ", col)
	return buf.String()
}

// Checkout checks out to the given ref.
func (r *Repository) Checkout(ref string, create bool) error {
	return r.Worktree().Checkout(&git.CheckoutOptions{
		Branch: gitplumbing.ReferenceName(ref),
		Create: create,
		Keep:   true,
	})
}

// Push pushes the current branch to the remote.
func (r *Repository) Push(ctx context.Context, force bool) error {
	head, err := r.Head()
	if err != nil {
		return errors.Wrap(err, "cannot get HEAD")
	}

	err = r.Repository.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []gitconfig.RefSpec{
			refSpecForBranch(head.Name().Short()),
		},
	})

	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		return errors.Wrap(err, "cannot push")
	}

	return nil
}

func refSpecForBranch(branch string) gitconfig.RefSpec {
	return gitconfig.RefSpec("+" +
		gitplumbing.NewBranchReferenceName(branch).String() + ":" +
		gitplumbing.NewRemoteReferenceName("origin", branch).String(),
	)
}
