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

		// visit each container and apply the cpu and memory reservations
		_ = containers.VisitElements(func(node *yaml.RNode) error {
			imageNode, _ := node.Pipe(yaml.Lookup("image"))
			image, _ := imageNode.String()
			image = strings.Replace(image, "\n", "", -1)
			if image == "ghcr.io/metaprov/modela-data-dock" {
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
