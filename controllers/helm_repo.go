/*
 * Copyright (c) 2021.
 *
 * Metaprov.com
 */

package controllers

import (
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

var (
	getters = getter.Providers{
		getter.Provider{
			Schemes: []string{"http", "https"},
			New:     getter.NewHTTPGetter,
		},
	}
)

type HelmRepo struct {
	Name      string
	Url       string
	Namespace string
	DryRun    bool
	Debug     bool
}

func NewHelmRepo(name string, url string, dryRun bool, debug bool) *HelmRepo {
	return &HelmRepo{
		Name:   name,
		Url:    url,
		DryRun: dryRun,
		Debug:  debug,
	}
}

func (r *HelmRepo) Add() (string, string, error) {
	runner := NewRealRunner("helm", ".", r.Debug)
	return runner.Run([]string{"repo", "add", r.Name, r.Url})
}

func (r *HelmRepo) Update() (string, string, error) {
	runner := NewRealRunner("helm", ".", r.Debug)
	return runner.Run([]string{"repo", "update"})
}

func (r *HelmRepo) DownloadIndex() error {
	entry := &repo.Entry{URL: r.Url}
	chartRepo, err := repo.NewChartRepository(entry, getters)
	if err != nil {
		return err
	}

	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return err
	}
	chartRepo.IndexFile.SortEntries()
	return nil

}
