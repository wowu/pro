package repository

import (
	"errors"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type Repository struct {
	goGitRepository *git.Repository
}

// Return git repository in given directory or parent directories.
func FindInParents(path string) (Repository, error) {
	windowsRootPath := filepath.VolumeName(path) + "\\"

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return Repository{}, err
	}

	goGitRepository, err := git.PlainOpen(absolutePath)

	if err == nil {
		return Repository{goGitRepository: goGitRepository}, nil
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

func (repo *Repository) CurrentBranchName() (string, error) {
	// get current head
	head, err := repo.goGitRepository.Head()
	if err != nil {
		return "", err
	}

	if !head.Name().IsBranch() {
		return "", ErrNoActiveBranch
	}

	// current branch name
	return head.Name().Short(), nil
}

func (repo *Repository) OriginUrl() (string, error) {
	// check if there is a remote named origin
	origin, err := repo.goGitRepository.Remote("origin")
	if err != nil {
		return "", ErrNoRemoteOrigin
	}

	originURL := origin.Config().URLs[0]
	return originURL, nil
}
