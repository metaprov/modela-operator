package controllers

import (
	"context"
	"fmt"
	infra "github.com/metaprov/modelaapi/pkg/apis/infra/v1alpha1"
	"github.com/pkg/errors"
	helmrelease "helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	"time"
)

func InstallChart(repoName, repoUrl, name string, ns string, releaseName string, versionName string) error {
	chart := NewHelmChart(repoName, repoUrl, name, ns, releaseName, versionName, false)
	chart.ChartVersion = versionName
	chart.ReleaseName = releaseName
	chart.Namespace = ns
	canInstall, err := chart.CanInstall()
	if err != nil {
		return errors.Errorf("Failed to check if chart is installed ,err: %s", err)
	}
	if canInstall {
		err = chart.Install()
		if err != nil {
			return errors.Errorf("Error installing chart %s, err: %s", name, err)
		}
	}
	return nil
}

func InstallChartWithValues(repoName, repoUrl, name string, ns string, releaseName string, versionName string, values map[string]interface{}) error {
	chart := NewHelmChart(repoName, repoUrl, name, ns, releaseName, versionName, false)
	chart.ChartVersion = versionName
	chart.ReleaseName = releaseName
	chart.Namespace = ns
	chart.Values = values
	canInstall, err := chart.CanInstall()
	if err != nil {
		return errors.Errorf("Failed to check if chart is installed ,err: %s", err)
	}
	if canInstall {
		err = chart.Install()
		if err != nil {
			return errors.Errorf("Error installing chart %s, err: %s", name, err)
		}
	}
	return nil
}

func UninstallChart(repoName string, repoUrl string, url string, ns string, releaseName string, versionName string) error {

	chart := NewHelmChart(repoName, repoUrl, url, ns, releaseName, versionName, false)
	installed, err := chart.IsInstalled()
	if err != nil {
		if !installed {
			return nil
		}
	}
	err = chart.Uninstall()
	if err != nil {
		return errors.Errorf("Error uninstalling chart %s, err: %s", url, err)
	}
	return nil
}

func IsChartInstalled(repoName, repoUrl string, url string, ns string, releaseName string, versionName string) (bool, error) {
	chart := NewHelmChart(repoName, repoUrl, url, ns, releaseName, versionName, false)
	chartStatus, _ := chart.GetStatus()
	if chartStatus == helmrelease.StatusUnknown {
		return false, nil
	}
	if chartStatus != helmrelease.StatusDeployed {
		return false, errors.New("chart " + releaseName + " is not in deployed state")
	}
	return true, nil
}

// check if a pod is running, return not nil error if not.
func IsPodRunning(ns string, prefix string) (bool, error) {
	conf, err := RestClient()
	if err != nil {
		return false, errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
	pods, err := clientSet.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, v := range pods.Items {
		if strings.Contains(v.Name, prefix) {
			if v.Status.Phase != "Running" {
				return false, nil
			} else {
				return true, nil
			}
		}
	}
	return false, nil

}

func CreateNamespace(name string) error {
	conf, err := RestClient()
	if err != nil {
		return errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
	_, err = clientSet.CoreV1().Namespaces().Get(context.Background(), name, metav1.GetOptions{})
	if k8serr.IsNotFound(err) {
		_, err = clientSet.CoreV1().Namespaces().Create(context.Background(), &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return errors.Errorf("Failed to create namespace %s, err: %s", name, err)
		}
	} else if err != nil {
		return errors.Errorf("Error getting namespace %s, err: %s", name, err)
	}
	return nil
}

func DeleteNamespace(name string) error {
	conf, err := RestClient()
	if err != nil {
		return errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
	err = clientSet.CoreV1().Namespaces().Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !k8serr.IsNotFound(err) {
		return errors.Errorf("Failed to delete namespace %s, err: %s", name, err)
	}
	return nil
}

func CreateOrUpdateSecret(ns string, name string, values map[string]string) error {
	conf, err := RestClient()
	if err != nil {
		return errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
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
	conf, err := RestClient()
	if err != nil {
		return errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
	err = clientSet.CoreV1().Secrets(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	if k8serr.IsNotFound(err) {
		return nil
	}
	return err
}

func GetSecret(ns string, name string) (*v1.Secret, error) {
	conf, err := RestClient()
	if err != nil {
		return nil, errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
	return clientSet.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func GetSecretValuesAsString(ns string, name string) (map[string]string, error) {
	conf, err := RestClient()
	if err != nil {
		return nil, errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
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
	conf, err := RestClient()
	if err != nil {
		return errors.Errorf("Error fetching rest client: %s", err)
	}
	// Get v1 interface to our cluster. Do or die trying
	clientSet := kubernetes.NewForConfigOrDie(conf)
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

func InstallCrd(url string) error {
	runner := NewRealRunner("kubectl", ".", false)
	_, _, err := runner.Run([]string{"apply", "-f", url})
	if err != nil {
		return err
	}
	return nil
}

func AddRepo(name string, url string, dryrun bool) error {
	repo := NewHelmRepo(name, url, dryrun, false)
	_, _, err := repo.Add()
	if err != nil {
		return err
	}
	fmt.Println("added repo " + name)
	_, _, err = repo.Update()
	if err != nil {
		return err
	}
	fmt.Println("repo updated " + name)
	fmt.Println("downloading index for repo  " + name)
	//err = repo.DownloadIndex()
	//if err != nil {
	//	return err
	//}
	fmt.Println("index updated for repo" + name)

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

func RestClient() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here

	configOverrides := &clientcmd.ConfigOverrides{}
	// if you want to change override values or bind them to flags, there are methods to help you

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return config, nil

}