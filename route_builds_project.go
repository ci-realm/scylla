package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	macaron "gopkg.in/macaron.v1"
)

func getBuildsProject(ctx *macaron.Context) {
	projectName := ctx.Params("project")
	projectPath := filepath.Join("ci", projectName)
	projectIDs, _ := filepath.Glob(filepath.Join(projectPath, "*"))
	projectBuilds := make([]projectBuild, len(projectIDs))
	for i, projectID := range projectIDs {
		parts := strings.Split(projectID, "/")
		projectBuilds[i] = projectBuild{
			Project: parts[1],
			ID:      parts[2],
			ModTime: modTime(projectID),
		}
	}

	sort.SliceStable(projectBuilds, func(i int, j int) bool {
		return projectBuilds[i].ModTime.After(projectBuilds[j].ModTime)
	})

	ctx.Data["Name"] = projectPretty(projectName)
	ctx.Data["Builds"] = projectBuilds
	ctx.HTML(200, "builds_project")
}

type projectBuild struct {
	Project, ID string
	ModTime     time.Time
}

func modTime(path string) (t time.Time) {
	if file, err := os.Open(path); err == nil {
		if fileInfo, err := file.Stat(); err == nil {
			t = fileInfo.ModTime()
		}
	}

	return t
}
