/*
 * Copyright (c) 2021.
 *
 * Metaprov.com
 */

package helm

import (
	"fmt"
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
	Name   string
	Url    string
	DryRun bool
	Debug  bool
}

func NewHelmRepo(name string, url string, dryRun bool, debug bool) *HelmRepo {
	return &HelmRepo{
		Name:   name,
		Url:    url,
		DryRun: dryRun,
		Debug:  debug,
	}
}

func (r *HelmRepo) DownloadIndex() error {
	entry := &repo.Entry{Name: r.Name, URL: r.Url}
	chartRepo, err := repo.NewChartRepository(entry, getters)
	if err != nil {
		return err
	}

	fmt.Println("Beginning index file dl")
	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return err
	}
	fmt.Println("Finished download")
	var f repo.File
	f.Update(entry)

	chartRepo.IndexFile.SortEntries()
	return nil

}

func AddRepo(name string, url string, dryrun bool) error {
	return nil
}
