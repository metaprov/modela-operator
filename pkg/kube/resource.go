package kube

import (
	"context"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"strings"
)

var kustomizer = krusty.MakeKustomizer(krusty.MakeDefaultOptions())

func LoadResources(folder string, filters []kio.Filter, loadAll bool) ([]byte, int, error) {
	path, _ := filepath.Abs("./manifests")
	resMap, err := kustomizer.Run(filesys.MakeFsOnDisk(), filepath.Join(path, folder))
	if err != nil {
		return nil, 0, err
	}
	for _, filter := range filters {
		if err = resMap.ApplyFilter(filter); err != nil {
			return nil, 0, err
		}
	}

	var missing = 0
	k8sclient, err := client.New(config.GetConfigOrDie(), client.Options{})
	if loadAll {
		goto skipFilter
	}
	for _, res := range resMap.Resources() {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   res.GetGvk().Group,
			Version: res.GetGvk().Version,
			Kind:    res.GetGvk().Kind,
		})
		err := k8sclient.Get(context.Background(), client.ObjectKey{Namespace: res.GetNamespace(), Name: res.GetName()}, obj)
		if err != nil {
			missing++
		} else {
			_ = resMap.Remove(res.OrgId())
		}
	}

skipFilter:
	if yaml, err := resMap.AsYaml(); err == nil {
		return yaml, missing, nil
	} else {
		return nil, missing, err
	}
}

func ApplyYaml(yaml string) error {
	f := util.NewFactory(RESTClientGetter{RestConfig: ctrl.GetConfigOrDie()})
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
		ForceConflicts:    true,
	}

	err = applyOptions.Run()
	if err != nil {
		return err
	}

	return nil

}

func ApplyUrlKustomize(url string) error {
	f := util.NewFactory(RESTClientGetter{RestConfig: ctrl.GetConfigOrDie()})
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
