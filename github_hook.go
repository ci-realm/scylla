package main

// GithubHook contains all the data needed to trigger a build from a Github Hook
type GithubHook struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		StatusesURL string `json:"statuses_url"`
		Head        struct {
			Sha  string `json:"sha"`
			Repo struct {
				CloneURL string `json:"clone_url"`
			} `json:"repo"`
		} `json:"head"`
	} `json:"pull_request"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}
