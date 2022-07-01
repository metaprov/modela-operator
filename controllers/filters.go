package controllers

import (
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

		stakeholders, err := node.Pipe(yaml.Lookup("spec", "permissions", "stakeholders"))
		if err != nil || stakeholders == nil {
			continue
		}

		_ = stakeholders.VisitElements(func(node *yaml.RNode) error {
			roles, _ := node.Pipe(yaml.Lookup("roles"))
			_ = roles.VisitElements(func(node *yaml.RNode) error {
				err = node.PipeE(
					yaml.Lookup("namespace"),
					yaml.Set(yaml.NewStringRNode(nf.Namespace)))

				return nil
			})

			return nil
		})
	}

	return nodes, nil
}
