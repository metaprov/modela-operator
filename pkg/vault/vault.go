package vault

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/kubernetes"
	managementv1 "github.com/metaprov/modela-operator/api/v1alpha1"
	"github.com/metaprov/modela-operator/pkg/kube"
	"github.com/pkg/errors"
)

func GetUnauthenticatedClientInCluster() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = "http://modela-vault.modela-system.svc.cluster.local:8200"
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	_, err = client.Sys().Health()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetUnauthenticatedClient(modela *managementv1.Modela) (*api.Client, error) {
	var address string
	if modela.Spec.Vault.VaultAddress == nil || *modela.Spec.Vault.VaultAddress == "" {
		address = "http://modela-vault.modela-system.svc.cluster.local:8200"
	} else {
		address = *modela.Spec.Vault.VaultAddress
	}

	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	_, err = client.Sys().Health()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func GetAuthenticatedClient(modela *managementv1.Modela) (*api.Client, error) {
	var address string
	if modela.Spec.Vault.VaultAddress == nil || *modela.Spec.Vault.VaultAddress == "" {
		address = "http://modela-vault.modela-system.svc.cluster.local:8200"
	} else {
		address = *modela.Spec.Vault.VaultAddress
	}

	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	auth, err := kubernetes.NewKubernetesAuth("modela")
	if err != nil {
		return nil, err
	}

	// Skip Kubernetes authentication if we already have a root token on our cluster
	if exists, err := kube.IsNamespaceCreated("modela-system"); exists && err == nil {
		if secret, err := kube.GetSecret("modela-system", "vault-root-token"); err == nil {
			if token, ok := secret.Data["token"]; ok {
				client.SetToken(string(token))
				return client, nil
			}
		}
	}

	if _, err := client.Auth().Login(context.Background(), auth); err != nil {
		return nil, err
	}

	return client, nil
}

func ApplySecret(modela *managementv1.Modela, key string, value map[string]interface{}) error {
	client, err := GetAuthenticatedClient(modela)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to apply vault secret %s", key))
	}

	kv := client.KVv2(modela.Spec.Vault.MountPath)
	if _, err = kv.Put(context.Background(), key, value); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to apply vault secret %s", key))
	}

	return nil
}
