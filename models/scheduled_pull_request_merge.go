// Copyright 2019 Gitea. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

// ScheduledPullRequestMerge represents a pull request scheduled for merging when checks succeed
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

// GetScheduledMergeRequestByPullID gets a scheduled pull request merge by pull request id
func GetScheduledMergeRequestByPullID(pullID int64) (exists bool, scheduledPRM *ScheduledPullRequestMerge, err error) {
	scheduledPRM = &ScheduledPullRequestMerge{}
	exists, err = x.Where("pull_id", pullID).Get(scheduledPRM)
	return
}
