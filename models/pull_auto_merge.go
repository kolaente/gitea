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

}

// MergeScheduledPullRequest merges a previously scheduled pull request when all checks succeeded
// TODO: How do I get that pull request?
func MergeScheduledPullRequest() (err error) {

}
