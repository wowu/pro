package repository

import "errors"

var (
	ErrNoActiveBranch = errors.New("no active branch")
	ErrNoRemoteOrigin = errors.New("no remote named \"origin\" found")
)
