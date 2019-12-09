// Copyright 2019 Gitea. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package automerge

import (
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	pullservice "code.gitea.io/gitea/services/pull"
)

// This package merges a previously scheduled pull request on successful status check.
// It is a separate package to avoid cyclic dependencies.

// MergeScheduledPullRequest merges a previously scheduled pull request when all checks succeeded
// Maybe FIXME: Move the whole check this function does into a separate go routine and just ping from here?
func MergeScheduledPullRequest(sha string, repo *models.Repository) (err error) {
	// First, get the branch associated with that commit sha
	r, err := git.OpenRepository(repo.RepoPath())
	if err != nil {
		return err
	}
	defer r.Close()
	commitID := git.MustIDFromString(sha)
	tree := git.NewTree(r, commitID)
	commit := &git.Commit{
		Tree: *tree,
		ID:   commitID,
	}
	branch, err := commit.GetBranchName()
	if err != nil {
		return err
	}
	// We get the branch name with a \n at the end which is not in the db so we strip it out
	branch = strings.Trim(branch, "\n")
	// Then get all prs for that branch
	pr, err := models.GetPullRequestByHeadBranch(branch, repo)
	if err != nil {
		return err
	}

	// Check if there is a scheduled pr in the db
	exists, scheduledPRM, err := models.GetScheduledMergeRequestByPullID(pr.ID)
	if err != nil {
		return err
	}
	if !exists {
		log.Info("No scheduled pull request merge exists for this pr [PRID: %d]", pr.ID)
		return nil
	}

	// Get all checks for this pr
	// We get the latest sha commit hash again to handle the case where the check of a previous push
	// did not succeed or was not finished yet.

	err = pr.GetHeadRepo()
	if err != nil {
		return
	}

	headGitRepo, err := git.OpenRepository(pr.HeadRepo.RepoPath())
	if err != nil {
		return
	}
	defer headGitRepo.Close()

	headBranchExist := headGitRepo.IsBranchExist(pr.HeadBranch)

	if pr.HeadRepo == nil || !headBranchExist {
		log.Info("Head branch of auto merge pr does not exist [HeadRepoID: %d, Branch: %s, PRID: %d]", pr.HeadRepoID, pr.HeadBranch, pr.ID)
		return nil
	}

	newestSha, err := headGitRepo.GetBranchCommitID(pr.HeadBranch)
	if err != nil {
		return
	}

	commitStatuses, err := models.GetLatestCommitStatus(repo, newestSha, 0)
	if err != nil {
		return
	}

	// Check if all checks succeeded
	for _, status := range commitStatuses {
		if status.State != models.CommitStatusSuccess {
			log.Info("Scheduled auto merge pr has unsuccessful status checks [PRID: %d, State: %s, Commit: %s, Context: %s]", pr.ID, status.State, sha, status.Context)
			return nil
		}
	}

	// Merge if all checks succeeded
	doer, err := models.GetUserByID(scheduledPRM.UserID)
	if err != nil {
		return
	}
	// FIXME: Is headGitRepo the right thing to use here? Maybe we should get the git repo based on scheduledPRM.RepoID?
	err = pullservice.Merge(pr, doer, headGitRepo, scheduledPRM.MergeStyle, scheduledPRM.Message)
	return
}
