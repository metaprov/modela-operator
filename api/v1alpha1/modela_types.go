/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// The phases of modela cluster
type ModelaPhase string

const (
	ModelaPhaseInstallingCertManager   = "InstallingCertManager"
	ModelaPhaseInstallingObjectStorage = "InstallingObjectStorage"
	ModelaPhaseInstallingPrometheus    = "InstallingPrometheus"
	ModelaPhaseInstallingDatabase      = "InstallingSystemDatabase"
	ModelaPhaseInstallingLoki          = "InstallingLoki"
	ModelaPhaseInstallingModela        = "InstallingModela"
	ModelaPhaseInstallingTenant        = "InstallingTenant"
	ModelaPhaseReady                   = "Ready"
	ModelaPhaseUninstalling            = "UninstallingComponent"
	ModelaPhaseFailed                  = "Failed"
)

// ConditionStatus defines conditions of resources
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition;
// "ConditionFalse" means a resource is not in the condition; "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

// ClusterConditionType is of string type
type ModelaConditionType string

// ClusterCondition describes the state of a cluster object at a certain point
type ModelaCondition struct {
	// Type of the condition.
	Type ModelaConditionType `json:"type,omitempty"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// ModelaAccessSpec defines the configuration for Modela to be exposed externally
type ModelaAccessSpec struct {
	// IngressEnabled indicates if Ingress resources will be created to expose the Modela API gateway, proxy, and frontend.
	// +kubebuilder:default:=true
	IngressEnabled *bool `json:"ingressEnabled,omitempty"`
	// InstallNginx indicates if the Nginx Ingress Controller will be installed. If enabled, Ingress resources created
	// by the Modela operator will use Nginx as their Ingress class.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	InstallNginx *bool `json:"installNginx,omitempty"`
	// Hostname specifies the host domain which will be used as the hostname for rules in Ingress resources managed
	// by the Modela operator. By default, the hostname will default to a localhost alias.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="vcap.me"
	Hostname *string `json:"hostname,omitempty"`
	// IngressClass specifies the Ingress class which will be applied to Ingress resources managed by the Modela operator.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="nginx"
	IngressClass *string `json:"ingressClass,omitempty"`
}

type ApiGatewaySpec struct {
	// Define the number of api gateway replicas
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`
	// +kubebuilder:validation:Optional

}

type ControlPlaneSpec struct {
	// The number of control plane replicas
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`
}

type DataPlaneSpec struct {
	// The number of data plane replicas
	// +kubebuilder:default:=1
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`
}

type CertManagerSpec struct {
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	Install *bool `json:"install"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="v1.7.1"
	CertManagerChartVersion *string `json:"chartVersion,omitempty"`
}

type ObjectStorageSpec struct {
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	Install *bool `json:"install"`

	//+kubebuilder:validation:Optional
	MinioChartVersion *string `json:"chartVersion,omitempty"`
}

type SystemDatabaseSpec struct {
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	Install *bool `json:"installed"`
	// +kubebuilder:default:="10.9.2"
	//+kubebuilder:validation:Optional
	PostgresChartVersion *string `json:"chartVersion,omitempty"`
}

type BackupSpec struct {
	// If true enable backups
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Enabled *bool `json:"enabled"`
	//+kubebuilder:validation:Optional
	CronSchedule string `json:"schedule"`
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Suspended *bool `json:"suspended"`
}

type ObservabilitySpec struct {
	// Prometheus indicates if the Prometheus Helm Chart will be installed
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Prometheus *bool `json:"installPrometheus"`
	// +kubebuilder:default:="2.8.4"
	//+kubebuilder:validation:Optional
	PrometheusVersion *string `json:"chartVersion"`
	// Loki indicates if the Loki Helm Chart will be installed
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Loki *bool `json:"installLoki"`
	//+kubebuilder:validation:Optional
	LokiVersion *string `json:"lokiChartVersion"`
}

type ModelaLicenseSpec struct {
	//+kubebuilder:validation:Optional
	LicenseKey *string `json:"licenseKey,omitempty"`

	// If LinkLicense is enabled, the Modela Operator will open a linking session through modela.ai which a
	// system administrator can use to log in to their modela.ai account and link their license. The URL which
	// must be opened by the administrator will be logged in the output of the operator's container.
	LinkLicense *bool `json:"linkLicense,omitempty"`
}

type TenantSpec struct {
	Name *string `json:"name"`
}

type TenantStatus struct {
	Name      *string `json:"name"`
	Installed *bool   `json:"installed"`
}

type ModelaSystemSpec struct {
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	Installed *bool `json:"installed"`

	//+kubebuilder:validation:Optional
	ChartVersion *string `json:"chartVersion"`
}

// ModelaSpec defines the desired state of Modela
type ModelaSpec struct {
	// ReleaseVersion denotes the desired version of Modela. This version will determine the
	// Docker image tags for all Modela images provided by Metaprov
	// +kubebuilder:validation:Required
	ReleaseVersion string `json:"releaseVersion"`

	// Observability specifies the configuration to install monitoring tools (Prometheus and Loki)
	Observability ObservabilitySpec `json:"observability,omitempty"`

	// Access specifies the configuration to install Ingress resources that will expose Modela externally
	Access ModelaAccessSpec `json:"access,omitempty"`

	// License specifies the license (or modela.ai account) information
	// that will be applied to the installation of Modela
	License ModelaLicenseSpec `json:"license,omitempty"`

	// If true install the default tenant.
	//+kubebuilder:validation:Optional
	ModelaChart ModelaSystemSpec `json:"modelaChart,omitempty"`

	// Tenants contains the collection of tenants that will be installed
	//+kubebuilder:validation:Optional
	Tenants []TenantSpec `json:"defaultTenantChart,omitempty"`

	// Desired state of object storage
	//+kubebuilder:validation:Optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`

	// Desired state of object storage
	//+kubebuilder:validation:Optional
	ObjectStore ObjectStorageSpec `json:"objectStore,omitempty"`

	SystemDatabase SystemDatabaseSpec `json:"systemDatabase,omitempty"`

	ControlPlane ControlPlaneSpec `json:"controlPlane,omitempty"`

	DataPlane DataPlaneSpec `json:"dataPlane,omitempty"`

	ApiGateway ApiGatewaySpec `json:"apiGateway,omitempty"`

	Pod *v1.PodTemplate `json:"podTemplate,omitempty"`
}

// ModelaStatus defines the observed state of Modela
type ModelaStatus struct {
	InstalledVersion string `json:"installedVersion,omitempty"`

	// Status of the control plane
	ModelaSystemInstalled bool `json:"modelaSystemInstalled,omitempty"`

	// Status of data plane
	Tenants []TenantStatus `json:"tenants,omitempty"`

	LicenseToken *v1.ObjectReference `json:"licenseTokenRef,omitempty"`

	Phase ModelaPhase `json:"phase,omitempty"`

	// ObservedGeneration is the last generation that was acted on
	//+kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	FailureMessage *string `json:"failureMessage,omitempty"`

	// Last time the modela installation was upgraded
	//+kubebuilder:validation:Optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +kubebuilder:validation:Optional
	Conditions []ModelaCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,8,rep,name=conditions"`
}

// Modela defines the configuration of the Modela operator
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:path=modelas,singular=modela,shortName="md",categories={data,modela,all}
type Modela struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelaSpec   `json:"spec,omitempty"`
	Status ModelaStatus `json:"status,omitempty"`
}

// ModelaList contains a list of Modela
//+kubebuilder:object:root=true
type ModelaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Modela `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Modela{}, &ModelaList{})
}
