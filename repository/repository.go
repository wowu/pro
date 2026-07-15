package repository

import (
	"errors"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
)

type Repository struct {
	goGitRepository *git.Repository
}

// Return git repository in given directory or parent directories.
func FindInParents(path string) (Repository, error) {
	goGitRepository, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return Repository{}, ErrNoRepository
		}

		return Repository{}, err
	}

	return Repository{goGitRepository: goGitRepository}, nil
}

func (repo *Repository) CurrentBranchName() (string, error) {
	// Read HEAD without resolving it, only the branch name is needed
	head, err := repo.goGitRepository.Reference(plumbing.HEAD, false)
	if err != nil {
		return "", err
	}

	// HEAD holds a hash instead of pointing at a branch, so it is detached
	if head.Type() != plumbing.SymbolicReference || !head.Target().IsBranch() {
		return "", ErrNoActiveBranch
	}

	return head.Target().Short(), nil
}

func (repo *Repository) OriginUrl() (string, error) {
	origin, err := repo.goGitRepository.Remote("origin")
	if err != nil {
		if errors.Is(err, git.ErrRemoteNotFound) {
			return "", ErrNoRemoteOrigin
		}

		return "", err
	}

	urls := origin.Config().URLs
	if len(urls) == 0 {
		return "", ErrNoRemoteOrigin
	}

	return urls[0], nil
}
