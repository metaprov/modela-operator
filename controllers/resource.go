package controllers

import (
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/kubectl/pkg/cmd/apply"
	k8sdelete "k8s.io/kubectl/pkg/cmd/delete"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
	"log"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"strings"
)

func LoadResources(folder string, filters []kio.Filter) ([]byte, error) {
	kustomizer := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	path, _ := filepath.Abs("../manifests")
	resMap, err := kustomizer.Run(filesys.MakeFsOnDisk(), filepath.Join(path, folder))
	if err != nil {
		return nil, err
	}
	for _, filter := range filters {
		if err = resMap.ApplyFilter(filter); err != nil {
			return nil, err
		}
	}
	if yaml, err := resMap.AsYaml(); err == nil {
		return yaml, nil
	} else {
		return nil, err
	}
}

func ApplyYaml(yaml string) error {
	restClient, err := RestClient()
	if err != nil {
		return err
	}
	f := util.NewFactory(RESTClientGetter{RestConfig: restClient})
	mapper, err := f.ToRESTMapper()
	if err != nil {
		return err
	}
	dynamicClient, err := f.DynamicClient()
	if err != nil {
		return err
	}
	openAPISchema, _ := f.OpenAPISchema()

	tmpfile, _ := ioutil.TempFile("", "*kubectl_manifest.yaml")
	_, _ = tmpfile.Write([]byte(yaml))
	_ = tmpfile.Close()

	applyOptions := &apply.ApplyOptions{
		IOStreams: genericclioptions.IOStreams{
			In:     strings.NewReader(yaml),
			Out:    log.Writer(),
			ErrOut: log.Writer(),
		},
		Builder:       f.NewBuilder(),
		DynamicClient: dynamicClient,
		Mapper:        mapper,
		OpenAPISchema: openAPISchema,
		Recorder:      genericclioptions.NoopRecorder{},
		ToPrinter: func(string) (printers.ResourcePrinter, error) {
			return &printers.NamePrinter{Operation: "serverside-applied"}, nil

		},
		DeleteOptions: &k8sdelete.DeleteOptions{
			FilenameOptions: resource.FilenameOptions{
				Filenames: []string{tmpfile.Name()},
			},
		},
		EnforceNamespace: false,

		VisitedUids:       sets.NewString(),
		VisitedNamespaces: sets.NewString(),
		PrintFlags:        genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		Overwrite:         true,
		ServerSideApply:   true,
		FieldManager:      "kubectl",
	}

	err = applyOptions.Run()
	if err != nil {
		return err
	}

	return nil

}

func ApplyUrlKustomize(url string) error {
	restClient, err := RestClient()
	if err != nil {
		return err
	}
	f := util.NewFactory(RESTClientGetter{RestConfig: restClient})
	mapper, err := f.ToRESTMapper()
	if err != nil {
		return err
	}
	dynamicClient, err := f.DynamicClient()
	if err != nil {
		return err
	}
	openAPISchema, _ := f.OpenAPISchema()
	applyOptions := &apply.ApplyOptions{
		IOStreams: genericclioptions.IOStreams{
			Out:    log.Writer(),
			ErrOut: log.Writer(),
		},
		Builder:       f.NewBuilder(),
		DynamicClient: dynamicClient,
		Mapper:        mapper,
		OpenAPISchema: openAPISchema,
		Recorder:      genericclioptions.NoopRecorder{},
		ToPrinter: func(string) (printers.ResourcePrinter, error) {
			return &printers.NamePrinter{Operation: "serverside-applied"}, nil
		},
		DeleteOptions: &k8sdelete.DeleteOptions{
			FilenameOptions: resource.FilenameOptions{
				Kustomize: url,
			},
		},
		EnforceNamespace: false,

		VisitedUids:       sets.NewString(),
		VisitedNamespaces: sets.NewString(),
		PrintFlags:        genericclioptions.NewPrintFlags("created").WithTypeSetter(scheme.Scheme),
		Overwrite:         true,
		ServerSideApply:   true,
		FieldManager:      "kubectl",
	}

	err = applyOptions.Run()
	if err != nil {
		return err
	}

	return nil

}
