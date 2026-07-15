package repository

import "errors"

var (
	ErrNoRepository   = errors.New("no git repository found")
	ErrNoActiveBranch = errors.New("no active branch")
	ErrNoRemoteOrigin = errors.New("no remote named \"origin\" found")
)
