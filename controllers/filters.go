package controllers

import (
	"sigs.k8s.io/kustomize/kyaml/yaml"
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
