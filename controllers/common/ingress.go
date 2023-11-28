package common

import (
	"errors"
	"fmt"
	managementv1alpha1 "github.com/metaprov/modela-operator/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const IngressClassAnnotationKey = "kubernetes.io/ingress.class"

func BuildFrontendIngress(hostname string, modela managementv1alpha1.Modela) (*networkingv1.Ingress, error) {
	if _, ok := modela.Annotations[IngressClassAnnotationKey]; !ok {
		return nil, errors.New("modela missing ingress class annotation (kubernetes.io/ingress.class)")
	}

	annotation := map[string]string{}
	for key, value := range modela.Annotations {
		annotation[key] = value
	}

	prefixPathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "modela-frontend-ingress",
			Namespace: "modela-system",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by":  "modela-operator",
				"management.modela.ai/operator": modela.Name,
			},
			Annotations: annotation,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: fmt.Sprintf("modela-app.%s", hostname),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &prefixPathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "modela-frontend",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Host: fmt.Sprintf("modela-api.%s", hostname),
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &prefixPathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "modela-api-gateway",
											Port: networkingv1.ServiceBackendPort{
												Number: 8081,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress, nil
}
