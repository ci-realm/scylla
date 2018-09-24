package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	macaron "gopkg.in/macaron.v1"
)

func getBuilds(ctx *macaron.Context) {
	metas := []meta{}

	filepath.Walk("ci", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Println(err)
			return err
		}

		if filepath.Base(path) == "result" {
			resolved, err := filepath.EvalSymlinks(path + "/meta.json")
			if err != nil {
				logger.Println(err)
			}
			content, err := ioutil.ReadFile(resolved)
			if err != nil {
				logger.Println(err)
			}
			m := meta{Path: path}
			json.NewDecoder(bytes.NewBuffer(content)).Decode(&m)
			metas = append(metas, m)
		}
		return nil
	})

	ctx.Data["metas"] = metas
	ctx.HTML(200, "builds")
}

type meta struct {
	Path string
	Rev  string `json:"rev"`
	URL  string `json:"url"`
}
