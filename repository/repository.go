package repository

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/helper/chroot"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

type Repository struct {
	workingDirectory string

	// Root git directory, usually workingDirectory/.git
	gitDirectory string

	// Will be different from gitDirectory if the repository is an external worktree
	// e.g. sample-repo/.git/worktrees/sample-repo-external
	// See: https://git-scm.com/docs/git-worktree
	worktreeGitDirectory string
}

// Return git repository in given directory or parent directories.
func FindInParents(path string) (Repository, error) {
	windowsRootPath := filepath.VolumeName(path) + "\\"

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return Repository{}, err
	}

	_, err = git.PlainOpen(absolutePath)
	// Found valid git repository
	if err == nil {
		return makeRepository(absolutePath)
	}

	if errors.Is(err, git.ErrRepositoryNotExists) {
		// Base case - we've reached the root of the filesystem
		if absolutePath == "/" || absolutePath == windowsRootPath {
			return Repository{}, errors.New("no git repository found")
		}

		// Recurse to parent directory
		return FindInParents(filepath.Dir(absolutePath))
	}

	return Repository{}, err
}

func makeRepository(workingDirectory string) (Repository, error) {
	goGitRepository, err := git.PlainOpen(workingDirectory)
	if err != nil {
		return Repository{}, err
	}

	// Get the resolved git directory for this working tree. For a linked
	// worktree this is <common>/.git/worktrees/<id>; otherwise it's the
	// repository's .git directory.
	storage, ok := goGitRepository.Storer.(*filesystem.Storage)
	if !ok {
		return Repository{}, errors.New("storage is not filesystem")
	}
	filesystem, ok := storage.Filesystem().(*chroot.ChrootHelper)
	if !ok {
		return Repository{}, errors.New("filesystem is not ChrootHelper")
	}
	gitDirectory := filesystem.Root()

	// The per-worktree git directory is where this working tree's HEAD lives.
	worktreeGitDirectory := gitDirectory

	// If a "commondir" file is present, we are inside a linked worktree and it
	// points to the shared git directory holding config and remotes.
	commonDir, err := readWorktreeCommonDir(gitDirectory)
	if err != nil {
		return Repository{}, err
	}
	if commonDir != "" {
		gitDirectory = commonDir
	}

	return Repository{
		workingDirectory:     workingDirectory,
		gitDirectory:         gitDirectory,
		worktreeGitDirectory: worktreeGitDirectory,
	}, nil
}

// readWorktreeCommonDir returns the resolved common (shared) git directory when
// gitDir belongs to a linked worktree, or "" when it does not. Presence of a
// "commondir" file in gitDir is git's own signal that this is a linked worktree.
func readWorktreeCommonDir(gitDir string) (string, error) {
	contents, err := os.ReadFile(filepath.Join(gitDir, "commondir"))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	commonDir := strings.TrimSpace(string(contents))
	if commonDir == "" {
		return "", nil
	}

	// commondir may be absolute or relative to the worktree git directory.
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(gitDir, commonDir)
	}

	return filepath.Clean(commonDir), nil
}

func (repo *Repository) CurrentBranchName() (string, error) {
	// Get HEAD file contents from git directory
	// We can't use go-git to get the current branch because it doesn't support worktrees
	headFile, err := os.ReadFile(filepath.Join(repo.worktreeGitDirectory, "HEAD"))
	if err != nil {
		return "", errors.New("unable to read HEAD")
	}

	// Parse HEAD file to get branch name
	headFileSplit := strings.Split(string(headFile), "ref: refs/heads/")
	if len(headFileSplit) != 2 {
		// if the HEAD file doesn't include "ref: refs/heads/", we are not on a branch
		return "", ErrNoActiveBranch
	}
	branch := strings.TrimSpace(headFileSplit[1])

	return branch, nil
}

func (repo *Repository) OriginUrl() (string, error) {
	goGitRepo, err := git.PlainOpen(repo.gitDirectory)
	if err != nil {
		return "", err
	}

	// check if there is a remote named origin
	origin, err := goGitRepo.Remote("origin")
	if err != nil {
		return "", ErrNoRemoteOrigin
	}

	originURL := origin.Config().URLs[0]
	return originURL, nil
}
