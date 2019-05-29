// Copyright 2019 The Gitea Authors. All rights reserved.
// Copyright 2018 Jonas Franz. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package base

// Downloader downloads the site repo informations
type Downloader interface {
	GetRepoInfo() (*Repository, error)
	GetMilestones() ([]*Milestone, error)
	GetReleases() ([]*Release, error)
	GetLabels() ([]*Label, error)
	GetIssues(start, limit int) ([]*Issue, error)
	GetComments(issueNumber int64) ([]*Comment, error)
	GetPullRequests(start, limit int) ([]*PullRequest, error)
}

// DownloaderFactory defines an interface to match a downloader implementation and create a downloader
type DownloaderFactory interface {
	Match(opts MigrateOptions) (bool, error)
	New(opts MigrateOptions) (Downloader, error)
}
