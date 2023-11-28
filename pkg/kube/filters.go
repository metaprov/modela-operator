package kube

import (
	"encoding/base64"
	"github.com/Masterminds/goutils"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"strings"
)

type LabelFilter struct {
	Labels map[string]string
}

func (l LabelFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	filters := make([]yaml.Filter, 0)
	for k, v := range l.Labels {
		filters = append(filters, yaml.SetLabel(k, v))
	}
	for _, node := range nodes {
		if _, err := node.Pipe(filters...); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

type ContainerVersionFilter struct {
	Version string
}

func (cv ContainerVersionFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		containers, err := node.Pipe(yaml.Lookup("spec", "template", "spec", "containers"))
		if err != nil || containers == nil {
			continue
		}

		// Set ModelaSystem release
		_ = node.PipeE(yaml.Lookup("spec", "release"), yaml.Set(yaml.NewStringRNode(cv.Version)))

		// Visit each container and apply the container version
		_ = containers.VisitElements(func(node *yaml.RNode) error {
			imageNode, _ := node.Pipe(yaml.Lookup("image"))
			image, _ := imageNode.String()
			image = strings.Replace(image, "\n", "", -1)
			// Skip data-dock image; this container is to be deprecated in a future release
			if strings.Contains(image, "ghcr.io/metaprov/modela-data-dock") {
				return nil
			}

			image = strings.Split(image, ":")[0] + ":" + cv.Version
			_ = node.PipeE(
				yaml.Lookup("image"),
				yaml.Set(yaml.NewStringRNode(image)))

			return nil
		})
	}

	return nodes, nil
}

type NamespaceFilter struct {
	Namespace string
}

func (nf NamespaceFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		_ = node.SetNamespace(nf.Namespace)
		_ = node.PipeE(yaml.Lookup("spec", "tenantRef", "name"), yaml.Set(yaml.NewStringRNode(nf.Namespace)))
		_ = node.PipeE(yaml.Lookup("spec", "secretRef", "namespace"), yaml.Set(yaml.NewStringRNode(nf.Namespace)))
	}

	return nodes, nil
}

type TenantFilter struct {
	TenantName string
}

func (t TenantFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		switch node.GetKind() {
		case "DataProduct":
			if err := node.PipeE(
				yaml.Lookup("spec", "defaultLabName"),
				yaml.Set(yaml.NewStringRNode(t.TenantName+"-lab"))); err != nil {
				return nil, err
			}

			if err := node.PipeE(
				yaml.Lookup("spec", "defaultServingSiteName"),
				yaml.Set(yaml.NewStringRNode(t.TenantName+"-serving-site"))); err != nil {
				return nil, err
			}
		case "Lab":
			if err := node.SetName(t.TenantName + "-lab"); err != nil {
				return nil, err
			}
		case "ServingSite":
			if err := node.SetName(t.TenantName + "-serving-site"); err != nil {
				return nil, err
			}
		case "Tenant":
			if err := node.SetName(t.TenantName); err != nil {
				return nil, err
			}
			if err := node.SetNamespace("modela-system"); err != nil {
				return nil, err
			}

			if err := node.PipeE(
				yaml.Lookup("spec", "defaultLabName"),
				yaml.Set(yaml.NewStringRNode(t.TenantName+"-lab"))); err != nil {
				return nil, err
			}

			if err := node.PipeE(
				yaml.Lookup("spec", "defaultServingSiteName"),
				yaml.Set(yaml.NewStringRNode(t.TenantName+"-serving-site"))); err != nil {
				return nil, err
			}

		}
	}
	return nodes, nil
}

type ManagedImageFilter struct {
	Version string
}

func (mi ManagedImageFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		if node.GetKind() == "ManagedImage" {
			_ = node.PipeE(yaml.Lookup("spec", "tag"), yaml.Set(yaml.NewStringRNode(mi.Version)))
		}
	}
	return nodes, nil
}

type JwtSecretFilter struct{}

func (j JwtSecretFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		if node.GetName() == "modela-auth-token" {
			str, _ := goutils.RandomAlphaNumeric(32)
			b64 := base64.StdEncoding.EncodeToString([]byte(str))
			_ = node.PipeE(yaml.Lookup("data", "jwt-secret"), yaml.Set(yaml.NewStringRNode(b64)))
		}
	}
	return nodes, nil
}

type RedisSecretFilter struct {
	Password string
}

func (r RedisSecretFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		if node.GetName() == "redis-secret" {
			_ = node.PipeE(
				yaml.Lookup("data", "redis-password"),
				yaml.Set(yaml.NewStringRNode(base64.StdEncoding.EncodeToString([]byte(r.Password)))),
			)
		}
	}

	return nodes, nil
}

type ModelaConfigFilter struct {
	VaultAddress   string
	VaultMountPath string
}

func (m ModelaConfigFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		if node.GetName() == "modela-config" {
			_ = node.PipeE(
				yaml.Lookup("spec", "vaultAddress"),
				yaml.Set(yaml.NewStringRNode(m.VaultAddress)),
			)
			_ = node.PipeE(
				yaml.Lookup("spec", "vaultMountPath"),
				yaml.Set(yaml.NewStringRNode(m.VaultMountPath)),
			)
		}
	}

	return nodes, nil
}

type OwnerReferenceFilter struct {
	Owner          string
	OwnerNamespace string
	UID            string
}

func (o OwnerReferenceFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range nodes {
		if o.OwnerNamespace != node.GetNamespace() {
			continue
		}
		ownerReference := yaml.NewMapRNode(&map[string]string{
			"apiVersion":         "management.modela.ai/v1alpha1",
			"kind":               "Modela",
			"name":               o.Owner,
			"uid":                o.UID,
			"blockOwnerDeletion": "true",
			"controller":         "true",
		})
		_ = node.PipeE(yaml.LookupCreate(yaml.SequenceNode, "metadata", "ownerReferences"),
			yaml.Append(ownerReference.YNode()))
	}
	return nodes, nil
}

type SkipCertManagerFilter struct{}

func (o SkipCertManagerFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	var outNodes []*yaml.RNode
	_, err := GetCRDVersion("issuers.cert-manager.io")
	var certManagerMissing = k8serr.IsNotFound(err)
	for _, node := range nodes {
		if certManagerMissing && node.GetApiVersion() == "cert-manager.io/v1" {
			continue
		}
		outNodes = append(outNodes, node)
	}

	return outNodes, nil
}

type ConnectionFilter struct {
	PgvectorEnabled bool
	MongoEnabled    bool
}

func (cf ConnectionFilter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	var outNodes []*yaml.RNode
	for _, node := range nodes {
		if node.GetName() == "postgres-vector-connection" && !cf.PgvectorEnabled {
			continue
		}
		if node.GetName() == "mongodb-connection" && !cf.MongoEnabled {
			continue
		}
		outNodes = append(outNodes, node)
	}

	return outNodes, nil
}
