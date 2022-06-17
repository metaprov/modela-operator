package controllers

import (
	"fmt"
	"github.com/metaprov/modela-operator/internal/pkg/util"
)

// Modela system represent the model core system
type Prometheus struct {
	Namespace     string
	Version       string
	ReleaseName   string
	RepoUrl       string
	RepoName      string
	Name          string
	PodNamePrefix string
	Dryrun        bool
}

func NewPrometheus() *Prometheus {
	return &Prometheus{
		Namespace:     "prometheus-community",
		Version:       "2.8.4",
		ReleaseName:   "kube-prometheus-stack",
		RepoName:      "prometheus-community",
		Name:          "kube-prometheus-stack",
		PodNamePrefix: "prometheus",
		RepoUrl:       "https://prometheus-community.github.io/helm-charts",
		Dryrun:        false,
	}
}

// Check if the database installed
func (m Prometheus) Installed() (bool, error) {
	return util.IsChartInstalled(
		m.RepoName,
		m.RepoUrl,
		m.ReleaseName,
		m.Namespace,
		m.ReleaseName,
		m.Version,
	)
}

func (m Prometheus) Install() error {

	if err := util.AddRepo(m.RepoName, m.RepoUrl, m.Dryrun); err != nil {
		return err
	}
	fmt.Println("\u2713 added repo " + m.RepoName)
	// install namespace modela-system
	if err := util.CreateNamespace(m.Namespace); err != nil {
		return err
	}
	fmt.Println("\u2713 created namespace " + m.Namespace)

	return util.InstallChart(
		m.RepoName,
		m.RepoUrl,
		m.ReleaseName,
		m.Namespace,
		m.ReleaseName,
		m.Version,
	)
}

// Check if we are still installing the database
func (m Prometheus) Installing() (bool, error) {
	installed, err := m.Installed()
	if !installed {
		return installed, err
	}
	running, err := util.IsPodRunning(m.Namespace, m.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

// Check if the database is ready
func (m Prometheus) Ready() (bool, error) {
	installed, err := m.Installed()
	if !installed {
		return installed, err
	}
	running, err := util.IsPodRunning(m.Namespace, m.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return running, nil
}

func (m Prometheus) Uninstall() error {
	return util.UninstallChart(
		m.RepoName,
		m.RepoUrl,
		"",
		m.Namespace,
		m.ReleaseName,
		m.Version,
	)
}
