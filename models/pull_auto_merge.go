// Copyright 2019 Gitea. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import (
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	pullservice "code.gitea.io/gitea/services/pull"
)

type ScheduledPullRequestMerge struct {
	ID         int64      `xorm:"pk autoincr"`
	PullID     int64      `xorm:"BIGINT"`
	UserID     int64      `xorm:"BIGINT"`
	MergeStyle MergeStyle `xorm:"varchar(50)"`
	Message    string     `xorm:"TEXT"`
}

// ScheduleAutoMerge schedules a pull request to be merged when all checks succeed
func ScheduleAutoMerge(opts *ScheduledPullRequestMerge) (err error) {
	// Check if we already have a merge scheduled for that pull request
	exists, err := x.Exist(&ScheduledPullRequestMerge{PullID: opts.PullID})
	if err != nil {
		return
	}
	if exists {
		// Maybe FIXME: Should we return a custom error here?
		return nil
	}

	_, err = x.Insert(opts)
	return err
}

// MergeScheduledPullRequest merges a previously scheduled pull request when all checks succeeded
// Maybe FIXME: Move the whole check this function does into a seperate go routine and just ping from here?
func MergeScheduledPullRequest(SHA string, repo *Repository) (err error) {
	// First, get the branch associated with that commit sha
	commit := &git.Commit{ID: git.MustIDFromString(SHA)}
	branch, err := commit.GetBranchName()
	if err != nil {
		return err
	}
	// Then get all prs for that branch
	pr := &PullRequest{}
	_, err = x.Where("head_branch = ? AND head_repo_id = ?", branch, repo.ID).Get(pr)
	if err != nil {
		return err
	}

	// Check if there is a scheduled pr in the db
	scheduledPRM := &ScheduledPullRequestMerge{}
	_, err = x.Where("pull_id", pr.ID).Get(scheduledPRM)
	if err != nil {
		return err
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

	sha, err := headGitRepo.GetBranchCommitID(pr.HeadBranch)
	if err != nil {
		return
	}

	commitStatuses, err := GetLatestCommitStatus(repo, sha, 0)
	if err != nil {
		return
	}

	// Check if all checks succeeded
	for _, status := range commitStatuses {
		if status.State != CommitStatusSuccess {
			log.Info("Scheduled auto merge pr has unsuccessful status checks [PRID: %d, State: %s, Commit: %s, Context: %s]", pr.ID, status.State, sha, status.Context)
			return nil
		}
	}

	// Merge if all checks succeeded
	doer, err := GetUserByID(scheduledPRM.UserID)
	if err != nil {
		return
	}
	// FIXME: Is headGitRepo the right thing to use here? Maybe we should get the git repo based on scheduledPRM.RepoID?
	err = pullservice.Merge(pr, doer, headGitRepo, scheduledPRM.MergeStyle, scheduledPRM.Message)
	return
}
