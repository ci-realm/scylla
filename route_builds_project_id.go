package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jackc/pgx"
	macaron "gopkg.in/macaron.v1"
)

func getBuildsProjectId(ctx *macaron.Context) {
	projectName := ctx.Params("user") + "/" + ctx.Params("repo")
	buildID := ctx.ParamsInt("id")

	withConn(ctx, func(conn *pgx.Conn) error {
		build, err := findBuildByProjectAndID(conn, projectName, buildID)
		if err != nil {
			return err
		}

		ctx.Data["Build"] = build
		ctx.HTML(200, "builds_project_id")
		return nil
	})

	return

	// projectName := ctx.Params("project")
	// buildID := ctx.Params("id")
	// idPath := filepath.Join("ci", projectName, buildID)
	// hookJSON := filepath.Join(idPath, "hook.json")
	// job := newGithubJobFromJSONFile(hookJSON)
	// if job == nil {
	// 	ctx.HTML(404, "not_found")
	// 	return
	// }

	// infos, err := buildIDInfos(idPath)
	// if err == nil {
	// 	ctx.Data["Status"] = "success"
	// 	ctx.Data["Results"] = infos
	// } else {
	// 	ctx.Error(500, err.Error())
	// 	return
	// }

	// if len(infos) == 0 {
	// 	sout, serr, err := job.nixLog()
	// 	if err == nil { // apparently the build failed
	// 		ctx.Data["NixLogStdout"] = sout
	// 		ctx.Data["NixLogStderr"] = serr
	// 		ctx.Data["Status"] = "failed"
	// 	} else { // we might not be done building it?
	// 		ctx.Data["Status"] = "pending"
	// 	}
	// }

	// ctx.Data["ProjectLink"] = "/builds/" + projectName
	// ctx.Data["Job"] = job
	// ctx.HTML(200, "builds_project_id")
}

type projectIDResult struct {
	NixPath              string
	NixStdout, NixStderr string
	NixError             error
}

func nixLog(nixPath string) (string, string, error) {
	stdoutBuf, stderrBuf, err := runCmd(exec.Command("nix", "log", nixPath))
	if err != nil {
		return "", "", err
	}
	return stdoutBuf.String(), stderrBuf.String(), nil
}

func buildIDInfos(idPath string) (map[string]projectIDResult, error) {
	resultLinks, _ := filepath.Glob(filepath.Join(idPath, "result*"))
	infos := map[string]projectIDResult{}
	for _, resultLink := range resultLinks {
		resolved, err := os.Readlink(resultLink)
		if err != nil {
			return nil, fmt.Errorf("Failed resolving result link %s: %s\n", resultLink, err)
		}
		no, ne, err := nixLog(resolved)
		infos[resolved] = projectIDResult{
			NixPath:   resolved,
			NixStdout: no,
			NixStderr: ne,
			NixError:  err,
		}
	}
	return infos, nil
}
