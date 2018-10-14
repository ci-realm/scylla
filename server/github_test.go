package server

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	os.Setenv("GITHUB_USER", "nobody")
	os.Setenv("GITHUB_TOKEN", "invalid")
	os.Setenv("GITHUB_URL", "https://custom.github.io")
	os.Setenv("BUILDERS", "none x86_64-linux")
	os.Setenv("PRIVATE_SSH_KEY", "empty")
	os.Setenv("DATABASE_URL", "nothing")
	parseConfig()
}

func TestGithubJob(t *testing.T) {
	job := &githubJob{Hook: &GithubHook{}}
	job.Hook.Repository.FullName = "user/repo"
	job.Hook.PullRequest.Head.Sha = "sha"
	job.Host = "http://example.com"

	Convey("pname creation", t, func() {
		So(job.pname(), ShouldEqual, "user_repo-sha")
	})

	Convey("sourceDir", t, func() {
		So(job.sourceDir(), ShouldEqual, "ci/user_repo/sha/source")
	})

	Convey("resultLink", t, func() {
		So(job.resultLink(), ShouldEqual, "ci/user_repo/sha/result")
	})

	Convey("progressURL", t, func() {
		So(job.targetURL(), ShouldEqual, "http://example.com/builds/user/repo/sha")
	})
}

func TestLockID(t *testing.T) {
	job := &githubJob{Hook: &GithubHook{}}
	Convey("job lockID can be the highest Int64", t, func() {
		job.Hook.PullRequest.Head.Sha = "ffffffffffffffffffffffffffffffffffffffff"
		So(job.lockID(), ShouldEqual, 1934001156059249939)
	})

	Convey("job lockID should always work", t, func() {
		job.Hook.PullRequest.Head.Sha = "0000000000000000000000000000000000000000"
		So(job.lockID(), ShouldEqual, -800969582777417106)
	})
}

func TestGithubAuth(t *testing.T) {
	Convey("Create correct configuration", t, func() {
		So(githubAuthKey("https://source.xing.com/", "mytoken"), ShouldEqual,
			"url.https://mytoken:x-oauth-basic@source.xing.com/.insteadOf")
	})
}
