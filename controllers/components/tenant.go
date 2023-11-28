package components

import (
	"context"
	"fmt"
	"github.com/metaprov/modela-operator/pkg/kube"
	"github.com/metaprov/modela-operator/pkg/vault"
	"golang.org/x/crypto/bcrypt"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kustomize/kyaml/kio"

	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
)

type Tenant struct {
	Name         string
	ManifestPath string
}

func NewTenant(name string) *Tenant {
	return &Tenant{
		Name:         name,
		ManifestPath: "tenant",
	}
}

func (t Tenant) GetInstallPhase() managementv1.ModelaPhase {
	return managementv1.ModelaPhaseInstallingTenant
}

func (t Tenant) IsEnabled(_ managementv1.Modela) bool {
	return true
}

func (t Tenant) Installed(ctx context.Context) (bool, error) {
	return kube.IsNamespaceCreated(t.Name)
}

func (t Tenant) Install(ctx context.Context, modela *managementv1.Modela, tenant *managementv1.TenantSpec) error {
	logger := log.FromContext(ctx)

	if err := kube.CreateNamespace(t.Name, modela.Name); err != nil && !k8serr.IsAlreadyExists(err) {
		logger.Error(err, "failed to create namespace")
		return err
	}

	var adminPassword string
	if tenant.AdminPassword != nil {
		adminPassword = *tenant.AdminPassword
	} else {
		adminPassword = "default"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.MinCost)
	if err != nil {
		return err
	}

	if err := vault.ApplySecret(modela, fmt.Sprintf("tenant/%s/accounts/admin", t.Name), map[string]interface{}{
		"password": string(hash),
	}); err != nil {
		return err
	}

	yaml, n, err := kube.LoadResources(t.ManifestPath, []kio.Filter{
		kube.LabelFilter{Labels: map[string]string{"management.modela.ai/operator": modela.Name}},
		kube.NamespaceFilter{Namespace: t.Name},
		kube.TenantFilter{TenantName: t.Name},
		kube.ConnectionFilter{
			PgvectorEnabled: modela.Spec.Database.InstallPgvector,
			MongoEnabled:    modela.Spec.Database.InstallMongoDB,
		},
	}, false)
	if err != nil {
		return err
	}

	if n > 0 {
		logger.Info("Applying tenant resources", "tenant", t.Name, "length", len(yaml))
		if err := kube.ApplyYaml(string(yaml)); err != nil {
			return err
		}
	}

	if values, err := kube.GetSecretValuesAsString("modela-system", "modela-storage-minio"); err == nil {
		accessKey, _ := values["root-user"]
		secretKey, _ := values["root-password"]

		logger.Info("Applying minio secret")
		if err := vault.ApplySecret(modela, fmt.Sprintf("tenant/%s/connections/minio-connection", t.Name), map[string]interface{}{
			"accessKey": accessKey,
			"secretKey": secretKey,
			"host":      "modela-storage-minio.modela-system.svc.cluster.local:9000",
		}); err != nil {
			return err
		}
	}

	if values, err := kube.GetSecretValuesAsString("modela-system", "modela-postgresql"); err == nil {
		password, _ := values["postgres-password"]

		logger.Info("Applying postgres secret")

		for _, conn := range []string{"postgres-connection", "postgres-vector-connection"} {
			if err := vault.ApplySecret(modela, fmt.Sprintf("tenant/%s/connections/%s", conn), map[string]interface{}{
				"username": "postgres",
				"password": password,
				"host":     "modela-postgresql.modela-system.svc.cluster.local",
				"port":     "5432",
			}); err != nil {
				return err
			}
		}
	}

	if values, err := kube.GetSecretValuesAsString("modela-system", "modela-mongodb"); err == nil {
		password, _ := values["mongodb-root-password"]

		logger.Info("Applying mongo secret")
		if err := vault.ApplySecret(modela, fmt.Sprintf("tenant/%s/connections/mongodb-connection", t.Name), map[string]interface{}{
			"username": "root",
			"password": password,
			"host":     "modela-mongodb.modela-system.svc.cluster.local",
			"port":     "27017",
		}); err != nil {
			return err
		}
	}

	return nil
}

func (t Tenant) Installing(ctx context.Context) (bool, error) {
	installed, err := t.Installed(ctx)
	if !installed {
		return installed, err
	}
	return false, nil
}

func (t Tenant) Ready(ctx context.Context) (bool, error) {
	if _, missing, err := kube.LoadResources(t.ManifestPath, []kio.Filter{kube.NamespaceFilter{Namespace: t.Name}, kube.TenantFilter{TenantName: t.Name}}, false); missing > 0 {
		return false, managementv1.ComponentMissingResourcesError
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (d Tenant) Uninstall(ctx context.Context, modela *managementv1.Modela) error {
	if created, err := kube.IsNamespaceCreatedByOperator(d.Name, modela.Name); !created {
		return managementv1.ComponentNotInstalledByModelaError
	} else if err != nil {
		return err
	}

	if err := kube.DeleteNamespace(d.Name); err != nil {
		return err
	}

	return nil
}
