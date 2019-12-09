// Copyright 2019 Gitea. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

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
	// Get all prs associated with that sha and repo

	// Check if there is a scheduled pr in the db
	// Check if all checks succeeded
	// Merge if all checks succeeded
}
