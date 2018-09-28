package main

import (
	"github.com/go-macaron/binding"
	macaron "gopkg.in/macaron.v1"
)

func setupRouting(m *macaron.Macaron) {
	m.Get("/_system/alive", getAlive)
	m.Get("/", getIndex)
	m.Get("/builds", getBuilds)
	m.Get("/builds/:user/:repo", getBuildsProject)
	m.Get("/builds/:user/:repo/:id", getBuildsProjectId)
	m.Get("/projects", getProjects)
	m.Post("/hooks/github", binding.Bind(GithubHook{}), postHooksGithub)
}
