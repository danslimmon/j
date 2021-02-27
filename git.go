package main

import (
	"github.com/go-git/go-git/v5"
)

// gitPull pulls the latest changes into the j workspace.
//
// workspace is the path to the repository.
//
// gitPull always pulls HEAD from origin. If HEAD is detached, or origin is not available, it will
// error.
func gitPull(workspace string) error {
	r, err := git.PlainOpen(workspace)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}

// gitCommit commits any changes to the specified files in the j workspace.
//
// workspace is the path to the repository. files are specified relative to the workspace path. msg
// is the commit message.
func gitCommit(workspace string, files []string, msg string) error {
	r, err := git.PlainOpen(workspace)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	for _, f := range files {
		_, err = w.Add(f)
		if err != nil {
			return err
		}
	}

	_, err = w.Commit(msg, &git.CommitOptions{All: true})
	if err != nil {
		return err
	}

	return nil
}

// gitPush pushes to origin.
//
// workspace is the path to the repository.
func gitPush(workspace string) error {
	r, err := git.PlainOpen(workspace)
	if err != nil {
		return err
	}

	return r.Push(&git.PushOptions{})
}
