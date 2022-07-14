package kube

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
	"time"

	catalog "github.com/metaprov/modelaapi/pkg/apis/catalog/v1alpha1"
	infra "github.com/metaprov/modelaapi/pkg/apis/infra/v1alpha1"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	nwv1 "k8s.io/api/networking/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	ClientScheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(infra.AddKnownTypes(ClientScheme))
	utilruntime.Must(catalog.AddKnownTypes(ClientScheme))

	utilruntime.Must(appsv1.AddToScheme(ClientScheme))
	utilruntime.Must(corev1.AddToScheme(ClientScheme))
	utilruntime.Must(rbacv1.AddToScheme(ClientScheme))
	utilruntime.Must(nwv1.AddToScheme(ClientScheme))

}

// check if a pod is running, return not nil error if not.
func IsPodRunning(ns string, prefix string) (bool, error) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	pods, err := clientSet.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, v := range pods.Items {
		if strings.Contains(v.Name, prefix) {
			return v.Status.Phase == "Running", nil
		}
	}
	return false, nil
}

func IsDeploymentCreatedByModela(ns string, name string) (bool, error) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	deployment, err := clientSet.AppsV1().Deployments(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if val, ok := deployment.GetLabels()["app.kubernetes.io/created-by"]; ok {
		if val == "modela-operator" {
			return true, nil
		}
	}
	return false, nil
}

func IsStatefulSetCreatedByModela(ns string, name string) (bool, error) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	statefulSet, err := clientSet.AppsV1().StatefulSets(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if val, ok := statefulSet.GetLabels()["app.kubernetes.io/created-by"]; ok {
		if val == "modela-operator" {
			return true, nil
		}
	}
	return false, nil
}

func GetCRDVersion(name string) string {
	clientSet := apiextensions.NewForConfigOrDie(ctrl.GetConfigOrDie())
	crd, err := clientSet.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	for _, version := range crd.Spec.Versions {
		if version.Storage {
			return version.Name
		}
	}
	return ""
}

func CreateNamespace(name string, operatorName string) error {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	_, err := clientSet.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if k8serr.IsNotFound(err) {
		_, err = clientSet.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: map[string]string{"management.modela.ai/operator": operatorName},
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return errors.Errorf("Failed to create namespace %s, err: %s", name, err)
		}
	} else {
		return err
	}
	return nil
}

func IsNamespaceCreated(name string) (bool, error) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	_, err := clientSet.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	return !k8serr.IsNotFound(err), nil
}

func IsNamespaceCreatedByOperator(name string, operator string) (bool, error) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	namespace, err := clientSet.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	if val, ok := namespace.GetLabels()["management.modela.ai/operator"]; ok {
		if val == operator {
			return true, nil
		}
	}
	return false, nil
}

func DeleteNamespace(name string) error {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	err := clientSet.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !k8serr.IsNotFound(err) {
		return errors.Wrapf(err, "Failed to delete namespace %s", name)
	}
	return nil
}

func CreateOrUpdateSecret(ns string, name string, values map[string]string) error {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	secret, err := clientSet.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
	if k8serr.IsNotFound(err) {
		s := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
			StringData: values,
		}
		_, err = clientSet.CoreV1().Secrets(ns).Create(context.Background(), s, metav1.CreateOptions{})
		if err != nil {
			return errors.Errorf("Failed to create namespace %s, err: %s", name, err)
		}
	} else if err != nil {
		return errors.Errorf("Error getting namespace %s, err: %s", name, err)
	} else {
		for k, v := range values {
			secret.Data[k] = []byte(v)
		}
		_, err = clientSet.CoreV1().Secrets(ns).Update(context.Background(), secret, metav1.UpdateOptions{})
		if err != nil {
			return errors.Errorf("Failed to create namespace %s, err: %s", name, err)
		}
	}
	return nil
}

func DeleteSecret(ns string, name string) error {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	err := clientSet.CoreV1().Secrets(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	if k8serr.IsNotFound(err) {
		return nil
	}
	return err
}

func GetSecret(ns string, name string) (*v1.Secret, error) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	return clientSet.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func GetSecretValuesAsString(ns string, name string) (map[string]string, error) {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	s, err := clientSet.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range s.Data {
		result[k] = string(v)
	}
	return result, nil
}

func CreateOrUpdateLicense(ns string, name string, license *infra.License) error {
	k8sClient, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: ClientScheme,
	})
	if err != nil {
		return err
	}

	if err = k8sClient.Get(context.Background(), client.ObjectKey{ns, name}, &infra.License{}); k8serr.IsNotFound(err) {
		err = k8sClient.Create(context.Background(), license)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	if err := k8sClient.Update(context.Background(), license); err != nil {
		return err
	}
	return nil
}

func CreateOrUpdateConnection(ns string, name string, conn *infra.Connection) error {
	/*
		k8sClient, err := client.New(config.GetConfigOrDie(), client.Options{
			Scheme: runtimescheme,
		})
		if err != nil {
			return err
		}
		k8sdb := k8s.NewK8sDb(k8sClient, runtimescheme, 3)
		connRepo := k8sdb.ConnectionDb()
		current, err := connRepo.GetConnection(context.Background(), ns, name)
		if k8serr.IsNotFound(err) {
			err = connRepo.CreateConnection(context.Background(), current)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		return connRepo.UpdateConnection(context.Background(), conn)

	*/
	return nil
}

func GetConnection(ns string, name string) (*infra.Connection, error) {
	/*
		k8sClient, err := client.New(config.GetConfigOrDie(), client.Options{
			Scheme: runtimescheme,
		})
		if err != nil {
			return nil, err
		}

		k8sdb := k8s.NewK8sDb(k8sClient, runtimescheme, 3)
		connRepo := k8sdb.ConnectionDb()
		return connRepo.GetConnection(context.Background(), ns, name)

	*/
	return nil, nil
}

func WaitForPod(ns string, name string) error {
	clientSet := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie())
	checks := 0
	for checks = 0; checks < 20; checks++ {
		pod, err := clientSet.CoreV1().Pods(ns).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			break
		}
		podstatusPhase := string(pod.Status.Phase)
		if podstatusPhase == "Running" {
			break
		}
		time.Sleep(10 * time.Second)
	}
	if checks == 20 {
		return errors.New("failed to start all the pods")
	}
	return nil
}

func IsPodRunningWithWait(ns string, name string) (bool, error) {
	counter := 0
	var err error
	for counter = 0; counter < 20; counter++ {
		running, err := IsPodRunning(ns, name)
		if running {
			break
		}
		if err != nil {
			return false, err
		} else {
			time.Sleep(10 * time.Second)
		}
	}
	if counter == 20 {
		return false, err
	}
	return true, nil

}

type RESTClientGetter struct {
	RestConfig *rest.Config
}

func (p RESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return nil
}

func (p RESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return p.RestConfig, nil
}

func (p RESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	home := homedir.HomeDir()
	var httpCacheDir = filepath.Join(home, ".kube", "http-cache")
	discoveryCacheDir := filepath.Join(home, ".kube", "cache", "discovery")
	return disk.NewCachedDiscoveryClientForConfig(p.RestConfig, discoveryCacheDir, httpCacheDir, 10*time.Minute)
}

func (p RESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, _ := p.ToDiscoveryClient()
	if discoveryClient != nil {
		mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
		expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
		return expander, nil
	}

	return nil, fmt.Errorf("no restmapper")
}
