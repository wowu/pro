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
	head, err := repo.goGitRepository.Head()
	if err != nil {
		// HEAD points at a branch without commits yet
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", ErrNoActiveBranch
		}

		return "", err
	}

	// HEAD is detached, so it doesn't point at a branch
	if !head.Name().IsBranch() {
		return "", ErrNoActiveBranch
	}

	return head.Name().Short(), nil
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
