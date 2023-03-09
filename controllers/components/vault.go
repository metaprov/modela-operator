package components

import (
	"context"
	"fmt"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/helm"
	"github.com/metaprov/modela-operator/pkg/kube"
	"github.com/pkg/errors"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	"github.com/hashicorp/vault/api"
)

const PolicyTemplate = `
path "%s/*" {
  capabilities = ["create", "read", "update", "patch", "delete", "list"]
}
`

// Modela system represent the model core system
type Vault struct {
	Namespace     string
	Name          string
	ReleaseName   string
	PodNamePrefix string

	VaultAddress string
	MountPath    string
}

func NewVault(address, mountPath string) *Vault {
	if address == "" {
		address = "modela-vault.modela-system.svc.cluster.local:8200"
	}

	if mountPath == "" {
		mountPath = "modela/secrets"
	}

	return &Vault{
		Namespace:     "modela-system",
		Name:          "vault",
		ReleaseName:   "modela-vault",
		PodNamePrefix: "vault-0",
		VaultAddress:  address,
		MountPath:     mountPath,
	}
}

func (v Vault) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingVault
}

func (v Vault) IsEnabled(modela managementv1.Modela) bool {
	return modela.Spec.Vault.Install
}

func (v Vault) Installed(ctx context.Context) (bool, error) {
	if belonging, err := kube.IsStatefulSetCreatedByModela(v.Namespace, "vault"); err == nil && !belonging {
		return true, managementv1.ComponentNotInstalledByModelaError
	}

	if installed, err := helm.IsChartInstalled(ctx, v.Name, v.Namespace, v.ReleaseName); !installed {
		return false, err
	}

	return true, nil
}

func (v Vault) Install(ctx context.Context, modela *managementv1.Modela) error {
	logger := log.FromContext(ctx)

	if err := kube.CreateNamespace(v.Namespace, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	logger.Info("Applying Vault Helm Chart")
	values := modela.Spec.Vault.Values.Object
	if values == nil {
		values = make(map[string]interface{})
		if _, ok := values["injector"]; !ok { // Disable the agent injector by default
			injectorValues := make(map[string]interface{})
			injectorValues["enabled"] = false
			values["injector"] = injectorValues
		}
	}

	return helm.InstallChart(ctx, v.Name, v.Namespace, v.ReleaseName, values)
}

func (v Vault) ConfigureVault(ctx context.Context) error {
	config := api.DefaultConfig()
	config.Address = v.VaultAddress
	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	sys := client.Sys()
	initialized, err := sys.InitStatus()
	if err != nil {
		return err
	}

	if !initialized {
		initResponse, err := sys.Init(&api.InitRequest{
			SecretShares:    1,
			SecretThreshold: 1,
		})

		if err != nil {
			return errors.Wrap(err, "Failed to initialize Vault server")
		}

		if err := kube.CreateOrUpdateSecret("modela-system", "vault-keys", map[string]string{
			"key": initResponse.Keys[0],
		}); err != nil {
			return errors.Wrap(err, "Failed to create Vault keys secret")
		}

		if err := kube.CreateOrUpdateSecret("modela-system", "vault-root-token", map[string]string{
			"token": initResponse.RootToken,
		}); err != nil {
			return errors.Wrap(err, "Failed to create Vault keys secret")
		}

		// Unseal the vault
		if _, err := sys.Unseal(initResponse.Keys[0]); err != nil {
			return errors.Wrap(err, "Failed to unseal Vault")
		}

		client.SetToken(initResponse.RootToken)

		// Mount the KVv2 secret engine
		if err := sys.Mount(v.MountPath, &api.MountInput{
			Type:    "kv",
			Options: map[string]string{"version": "2"},
		}); err != nil {
			return errors.Wrap(err, "Failed to mount secret engine")
		}

		// Create the policy for the mount
		var policy = fmt.Sprintf(PolicyTemplate, v.MountPath)
		if err := sys.PutPolicy("modela-policy", policy); err != nil {
			return errors.Wrap(err, "Failed to create policy")
		}

		// Configure Kubernetes authentication
		if err := sys.EnableAuthWithOptions("kubernetes", &api.EnableAuthOptions{Type: "kubernetes"}); err != nil {
			return errors.Wrap(err, "Failed to enable Kubernetes authentication")
		}

		c := client.Logical()
		if _, err := c.Write("/auth/kubernetes/config", map[string]interface{}{
			"kubernetes_host": "https://kubernetes.default.svc",
		}); err != nil {
			return errors.Wrap(err, "Failed to configure Kubernetes authentication")
		}

		// Configure the Kubernetes auth method role
		if _, err := c.Write("/auth/kubernetes/role/modela", map[string]interface{}{
			"name": "modela",
			"bound_service_account_names": []string{"lab-job-sa", "modela-apigateway", "modela-dataplane",
				"modela-datadock", "modela-control-plane", "servingsite-job-sa", "modela-operator-controller-manager"},
			"bound_service_account_namespaces": []string{"*"},
			"policies":                         []string{"modela-policy"},
		}); err != nil {
			return errors.Wrap(err, "Failed to configure Kubernetes authentication roles")
		}
	}

	return nil
}

// Check if we are still installing the database
func (v Vault) Installing(ctx context.Context) (bool, error) {
	installed, err := v.Installed(ctx)
	if !installed {
		return installed, err
	}
	running, err := kube.IsPodRunning(v.Namespace, v.PodNamePrefix)
	if err != nil {
		return false, err
	}
	return !running, nil
}

func (v Vault) Ready(ctx context.Context) (bool, error) {
	installing, err := v.Installing(ctx)
	if err != nil && err != managementv1.ComponentNotInstalledByModelaError {
		return false, err
	}

	return !installing, nil
}

func (v Vault) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	return helm.UninstallChart(ctx, v.Name, v.Namespace, v.ReleaseName, map[string]interface{}{})
}

func performAutoUnseal() {
	// Check if we are running inside the cluster. If not, abort as we have no way to communicate with Vault
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); errors.Is(err, os.ErrNotExist) {
		return
	}

	// Check if we have the Vault keys
	if exists, err := kube.IsNamespaceCreated("modela-system"); !exists || err != nil {
		return
	}

	secret, err := kube.GetSecret("modela-system", "vault-keys")
	if err != nil {
		return
	}

	key, ok := secret.Data["key"]
	if !ok {
		return
	}

	config := api.DefaultConfig()
	config.Address = "http://modela-vault.modela-system.svc.cluster.local:8200"
	client, err := api.NewClient(config)
	if err != nil {
		return
	}

	sys := client.Sys()
	sealed, err := sys.SealStatus()
	if err != nil {
		klog.ErrorS(err, "Failed to check if Vault is sealed")
		return
	}

	if sealed.Sealed {
		klog.Info("Attempting to unseal Vault server")
		_, _ = sys.Unseal(string(key))
	}
}

func StartAutoUnseal(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			performAutoUnseal()
		}
	}
}
