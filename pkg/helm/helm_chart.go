/*
 * Copyright (c) 2021.
 *
 * Metaprov.com
 */

package helm

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/metaprov/modela-operator/pkg/kube"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"os"
	"regexp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/kustomize/kyaml/kio"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/pkg/errors"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmloader "helm.sh/helm/v3/pkg/chart/loader"
	helmcli "helm.sh/helm/v3/pkg/cli"
	helmrelease "helm.sh/helm/v3/pkg/release"
)

var settings = helmcli.New()

type LabelPostRenderer struct {
	Labels map[string]string
}

func (lb LabelPostRenderer) Run(renderedManifests *bytes.Buffer) (modifiedManifests *bytes.Buffer, err error) {
	var output bytes.Buffer
	var writer = bufio.NewWriter(&output)
	rw := &kio.ByteReadWriter{
		Reader:                bytes.NewReader(renderedManifests.Bytes()),
		Writer:                writer,
		OmitReaderAnnotations: true,
		KeepReaderAnnotations: true,
	}
	p := kio.Pipeline{
		Inputs:  []kio.Reader{rw},
		Filters: []kio.Filter{kube.LabelFilter{lb.Labels}},
		Outputs: []kio.Writer{rw},
	}

	if err := p.Execute(); err != nil {
		return nil, err
	}

	_ = writer.Flush()
	return bytes.NewBuffer(output.Bytes()), nil
}

type HelmChart struct {
	RepoName        string
	Name            string // chart name
	Namespace       string // chart namespace
	ReleaseName     string // release name
	ChartVersion    string // chart version
	RepoUrl         string // repo url
	DryRun          bool
	CreateNamespace bool
	crt             *helmchart.Chart
	Values          map[string]interface{}
}

func NewHelmChart(repoName, repoUrl, name, namespace, releaseName, versionName string, dryRun bool) *HelmChart {
	return &HelmChart{
		RepoName:        repoName,
		Name:            name,
		Namespace:       namespace,
		ReleaseName:     releaseName,
		ChartVersion:    versionName,
		RepoUrl:         repoUrl,
		DryRun:          dryRun,
		CreateNamespace: false,
		Values:          make(map[string]interface{}),
	}
}

func (chart *HelmChart) GetConfig() (*helmaction.Configuration, error) {
	var kubeConfig *genericclioptions.ConfigFlags
	config := ctrl.GetConfigOrDie()
	kubeConfig = genericclioptions.NewConfigFlags(false)
	kubeConfig.APIServer = &config.Host
	kubeConfig.BearerToken = &config.BearerToken
	kubeConfig.CAFile = &config.CAFile
	kubeConfig.Namespace = &chart.Namespace

	actionConfig := new(helmaction.Configuration)
	if err := actionConfig.Init(kubeConfig, chart.Namespace, os.Getenv("HELM_DRIVER"), klog.Infof); err != nil {
		klog.Error(err, "Unable to initialize Helm")
		return nil, err
	}
	return actionConfig, nil
}

// Load the chart, and assign it to the crt field
func (chart *HelmChart) Load(ctx context.Context) error {
	logger := log.FromContext(ctx)
	config, err := chart.GetConfig()
	if err != nil {
		logger.Error(err, "Failed to get config")
		return err
	}

	client := helmaction.NewInstall(config)
	client.ChartPathOptions.RepoURL = chart.RepoUrl
	client.Namespace = chart.Namespace
	client.ReleaseName = chart.ReleaseName
	name := chart.Name //chart.RepoName + "/" + chart.Name

	chartFullPath, err := client.ChartPathOptions.LocateChart(name, settings)
	if err != nil {
		logger.Error(err, "Failed to locate Helm Chart", "name", chart.Name)
		return fmt.Errorf("Failed to locate Helm Chart '%s' due to %s", chart.Name, err)
	}
	result, err := helmloader.Load(chartFullPath)
	if err != nil {
		logger.Error(err, "Failed to load Helm Chart")
		return errors.Wrapf(err, "Failed to load resources from %s", chartFullPath)
	}
	chart.crt = result
	return nil
}

func (chart *HelmChart) Version() string {
	chartPackageSplit := chart.parsePackageName()
	chartVersion := chartPackageSplit[1]
	if chartPackageSplit[2] != "" {
		chartVersion = fmt.Sprintf("%s-%s", chartVersion, chartPackageSplit[2])
	}
	return chartVersion
}

func (chart *HelmChart) parsePackageName() []string {
	packageNameRegexp := regexp.MustCompile(`([a-z\-]+)-([0-9\.]*[0-9]+)(-([0-9]+))?`)
	packageSubstringSubmatch := packageNameRegexp.FindStringSubmatch(chart.Name)
	parsedOutput := []string{"", "", ""}
	if len(packageSubstringSubmatch) > 2 {
		parsedOutput[0] = packageSubstringSubmatch[1]
		parsedOutput[1] = packageSubstringSubmatch[2]
	}
	if len(packageSubstringSubmatch) > 4 {
		parsedOutput[2] = packageSubstringSubmatch[4]
	}

	return parsedOutput
}

func (chart *HelmChart) CanInstall(ctx context.Context) (bool, error) {
	err := chart.Load(ctx)
	if err != nil {
		return false, err
	}
	switch chart.crt.Metadata.Type {
	case "", "application":
		return true, err
	}
	return false, err
}

func (chart *HelmChart) Get(ctx context.Context) (*helmrelease.Release, error) {
	logger := log.FromContext(ctx)

	config, err := chart.GetConfig()
	if err != nil {
		logger.Error(err, "failed to get config")
		return nil, err
	}
	// Check if the Release Exists
	aList := helmaction.NewList(config) // NewGet provides bad error message if release doesn't exist
	aList.All = true
	charts, err := aList.Run()
	if err != nil {
		logger.Error(err, "failed to get config")
		return nil, errors.Wrap(err, "failed to run get")
	}
	for _, release := range charts {
		if release.Name == chart.ReleaseName && release.Namespace == chart.Namespace {
			return release, nil
		}
	}
	return nil, errors.Errorf("unable to find release '%s' in namespace '%s'", chart.ReleaseName, chart.Namespace)
}

// check if the chart is already installed
func (chart *HelmChart) IsInstalled(ctx context.Context) (bool, error) {
	logger := log.FromContext(ctx)
	err := chart.Load(ctx)
	if err != nil {
		logger.Error(err, "failed to load chart")
		return false, errors.Wrapf(err, "failed to load chart")
	}
	existingRelease, _ := chart.Get(ctx)
	/*if err != nil {
		logger.Error(err, "failed to get chart")
		return false, err
	}*/
	if existingRelease != nil {
		return true, nil
	}
	return false, nil
}

func (chart *HelmChart) GetStatus(ctx context.Context) (helmrelease.Status, error) {
	err := chart.Load(ctx)
	if err != nil {
		return helmrelease.StatusUnknown, errors.Wrapf(err, "failed to load chart")
	}
	existingRelease, err := chart.Get(ctx)
	if err != nil {
		return helmrelease.StatusUnknown, errors.Wrapf(err, "chart does not exist")
	}
	return existingRelease.Info.Status, nil

}

func (chart *HelmChart) Install(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("Installing Helm Chart", "release", chart.ReleaseName, "namespace", chart.Namespace, "name", chart.Name)
	err := chart.Load(ctx)
	if err != nil {
		logger.Error(err, "Failed to load chart")
		return errors.Wrapf(err, "Failed to load chart")
	}
	// Check if resource already exists
	existingRelease, _ := chart.Get(ctx)
	if existingRelease != nil {
		logger.Error(err, fmt.Sprintf("Release \"%s\" already exists in namespace \"%s\"", existingRelease.Name, existingRelease.Namespace))
		return errors.Wrapf(err, "Release '%s' already exists in namespace '%s'", existingRelease.Name, existingRelease.Namespace)
	}

	can, err := chart.CanInstall(ctx)
	if err != nil {
		logger.Error(err, "Failed to check if Helm Chart is installed", "namespace", existingRelease.Namespace)
		return errors.Wrapf(err, "Failed to check if Helm Chart is installed (namespace=%s)", existingRelease.Namespace)
	}
	if !can {
		return errors.Wrapf(err, "release at '%s' is not installable", chart.Name)
	}

	config, err := chart.GetConfig()
	if err != nil {
		logger.Error(err, "failed to get config")
		return errors.Wrap(err, "failed to get config")
	}

	inst := helmaction.NewInstall(config)
	if inst.Version == "" && inst.Devel {
		inst.Version = ">0.0.0-0"
	}
	inst.ReleaseName = chart.ReleaseName
	inst.Namespace = chart.Namespace
	inst.DryRun = chart.DryRun
	inst.CreateNamespace = chart.CreateNamespace
	inst.Version = chart.ChartVersion
	inst.PostRenderer = LabelPostRenderer{map[string]string{"app.kubernetes.io/created-by": "modela-operator"}}
	inst.Replace = true

	_, err = inst.Run(chart.crt, chart.Values)
	if err != nil {
		logger.Error(err, "failed to install")
		return fmt.Errorf("failed to run install due to %s", err)
	}
	return nil

}

func (chart *HelmChart) Upgrade(ctx context.Context) error {
	logger := log.FromContext(ctx)

	logger.Info("Enter upgrade")

	err := chart.Load(ctx)
	if err != nil {
		logger.Error(err, "failed to load chart")
		return errors.Wrapf(err, "failed to load chart")
	}
	// Check if resource already exists
	existingRelease, err := chart.Get(ctx)
	if existingRelease != nil {
		return errors.Wrapf(err, "release '%s' already exists in namespace '%s'", existingRelease.Name, existingRelease.Namespace)
	}

	can, err := chart.CanInstall(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to check if chart is installed '%s'", existingRelease.Namespace)
	}
	if !can {
		return errors.Wrapf(err, "release at '%s' is not installable", chart.Name)
	}

	config, err := chart.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get config")
	}

	isInstalled, err := chart.IsInstalled(ctx)
	if err != nil {
		return fmt.Errorf("failed to get installed state %s", err)
	}

	if !isInstalled {
		inst := helmaction.NewInstall(config)
		if inst.Version == "" && inst.Devel {
			inst.Version = ">0.0.0-0"
		}
		inst.ReleaseName = chart.ReleaseName
		inst.Namespace = chart.Namespace
		inst.DryRun = chart.DryRun
		inst.CreateNamespace = chart.CreateNamespace
		inst.Version = chart.ChartVersion

		_, err = inst.Run(chart.crt, chart.Values)
		if err != nil {
			return fmt.Errorf("failed to run install due to %s", err)
		}
		return nil
	} else {
		inst := helmaction.NewUpgrade(config)
		if inst.Version == "" && inst.Devel {
			inst.Version = ">0.0.0-0"
		}
		inst.DryRun = chart.DryRun
		inst.Version = chart.ChartVersion

		_, err = inst.Run(chart.ReleaseName, chart.crt, chart.Values)
		if err != nil {
			return fmt.Errorf("failed to run install due to %s", err)
		}
		return nil
	}

}

func (chart *HelmChart) Uninstall(ctx context.Context) error {
	err := chart.Load(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to load chart")
	}
	// Check if resource already exists
	existingRelease, _ := chart.Get(ctx)
	if existingRelease == nil {
		return nil
	}

	config, err := chart.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to get config")
	}

	inst := helmaction.NewUninstall(config)
	inst.DryRun = chart.DryRun

	_, err = inst.Run(chart.ReleaseName)
	if err != nil {
		return fmt.Errorf("failed to run uninstall due to %s", err)
	}
	return nil
}

func InstallChart(ctx context.Context, repoName, repoUrl, name string, ns string, releaseName string, versionName string, values map[string]interface{}) error {
	chart := NewHelmChart(repoName, repoUrl, name, ns, releaseName, versionName, false)
	chart.ChartVersion = versionName
	chart.ReleaseName = releaseName
	chart.Namespace = ns
	chart.Values = values

	canInstall, err := chart.CanInstall(ctx)
	if err != nil {
		return errors.Errorf("Failed to check if chart is installed ,err: %s", err)
	}
	if canInstall {
		err = chart.Install(ctx)
		if err != nil {
			return errors.Wrapf(err, "Error installing chart %s", name)
		}
	}
	return nil
}

func UninstallChart(ctx context.Context, repoName, repoUrl, name string, ns string, releaseName string, versionName string, values map[string]interface{}) error {
	chart := NewHelmChart(repoName, repoUrl, name, ns, releaseName, versionName, false)
	chart.ChartVersion = versionName
	chart.ReleaseName = releaseName
	chart.Namespace = ns
	chart.Values = values

	if installed, err := chart.IsInstalled(ctx); err != nil {
		return err
	} else if !installed {
		return nil
	}

	if err := chart.Uninstall(ctx); err != nil {
		return errors.Wrapf(err, "Error uninstalling chart %s", name)
	}

	return nil
}

func IsChartInstalled(ctx context.Context, repoName, repoUrl string, url string, ns string, releaseName string, versionName string) (bool, error) {
	// TODO(liam): Refactor these methods into the helm chart struct
	if err := AddRepo(repoName, repoUrl, false); err != nil {
		return false, err
	}
	fmt.Println("Added repo", repoName, repoUrl)

	chart := NewHelmChart(repoName, repoUrl, url, ns, releaseName, versionName, false)
	chartStatus, _ := chart.GetStatus(ctx)
	if chartStatus == helmrelease.StatusUnknown {
		return false, nil
	}
	if chartStatus != helmrelease.StatusDeployed {
		return false, errors.New("chart " + releaseName + " is not in deployed state")
	}
	return true, nil
}
