package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	macaron "gopkg.in/macaron.v1"
)

func getProjects(ctx *macaron.Context) {
	projectNames, _ := filepath.Glob("ci/*")
	projectInfos := make([]projectInfo, len(projectNames))

	for i, name := range projectNames {
		projectName := projectPretty(name)
		buildCount := subdirCount(name)
		projectInfos[i] = projectInfo{
			Link:       filepath.Base(name),
			Name:       projectName,
			BuildCount: buildCount,
		}
	}

	ctx.Data["Projects"] = projectInfos
	ctx.HTML(200, "projects")
}

type projectInfo struct {
	Name, Link string
	BuildCount int
}

func projectPretty(name string) string {
	return strings.Replace(filepath.Base(name), "_", "/", 1)
}

func subdirCount(name string) int {
	info, err := os.Stat(name)
	if err != nil {
		logger.Printf("Failed to stat %s: %s\n", name, err)
		return 0
	}

	rvalue := reflect.ValueOf(info.Sys())
	nlink := rvalue.Elem().FieldByName("Nlink")
	if !nlink.IsValid() { // use the wasteful path
		count, _ := filepath.Glob(filepath.Join(name, "/*/"))
		return len(count)
	}
	return int(nlink.Uint() - 2)
}
