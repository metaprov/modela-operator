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
	ModelaPhasePending                 = "Pending"
	ModelaPhaseInstallingCertManager   = "InstallingCertManager"
	ModelaPhaseInstallingPormetous     = "InstallingPormetous"
	ModelaPhaseInstallingDatabase      = "InstallingSystemDatabase"
	ModelaPhaseInstallingLoki          = "InstallingLoki"
	ModelaPhaseInstallingModela        = "InstallingModela"
	ModelaPhaseInstallingDefaultTenant = "InstallingDefaultTenant"
	ModelaPhaseRunning                 = "Running"
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

// Define how the modela cluster is exposed.
type ModelaAccessSpec struct {
	// +kubebuilder:default:=8080
	Port *int `json:"port,omitempty"`

	NodePort *int32 `json:"nodeport,omitempty"`
}

type ApiGatewaySpec struct {
	// Define the number of api gateway replicas
	// +kubebuilder:default:=1
	//+kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Template to be used to generate the Persistent Volume Claim for the api gateway
	//+kubebuilder:validation:Optional
	PersistentVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"pvcTemplate,omitempty"`
}

type ControlPlaneSpec struct {
	// Define the control plane replicas
	// +kubebuilder:default:=1
	//+kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`
}

type ObjectStorageSpec struct {
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	Enabled *bool `json:"enabled"`

	// Minio Connection Reference
	//+kubebuilder:validation:Optional
	ConnectionRef v1.ObjectReference `json:"connectionRef,omitempty"`
}

type SystemDatabaseSpec struct {
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	Enabled *bool `json:"enabled"`

	// Minio Connection Reference
	//+kubebuilder:validation:Optional
	ConnectionRef v1.ObjectReference `json:"connectionRef,omitempty"`
}

type CertManagerSpec struct {
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	Enabled *bool `json:"enabled"`

	// Desired cert manager version
	//+kubebuilder:validation:Optional
	Version v1.ObjectReference `json:"version,omitempty"`
}

type BackupSpec struct {
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
	//+kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	Enabled *bool `json:"enabled"`
	// If true install the Prometheus helm chart (if not installed)
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Prometheus *bool `json:"prometheus"`
	// If true install the loki helm chart (if not installed)
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Loki *bool `json:"loki"`
}

type DataPlaneSpec struct {
	// +kubebuilder:default:=1
	//+kubebuilder:validation:Optional
	Replicas *int `json:"replicas,omitempty"`

	// StorageClass to use for data plane data
	//+kubebuilder:validation:Optional
	StorageClass *string `json:"storageClass,omitempty"`

	// Template to be used to generate the Persistent Volume Claim for the api gateway
	//+kubebuilder:validation:Optional
	PersistentVolumeClaimTemplate *v1.PersistentVolumeClaimSpec `json:"pvcTemplate,omitempty"`
}

type ControlPlaneStatus struct {
	// The status of the control plane
	DeploymentStatus string `json:"deploymentStatus,omitempty"`

	ServiceStatus string `json:"serviceStatus,omitempty"`
}

type DataPlaneStatus struct {
	// The status of the control plane
	DeploymentStatus string `json:"deploymentStatus,omitempty"`

	ServiceStatus string `json:"serviceStatus,omitempty"`
}

type ApiGatewayStatus struct {
	// The status of the control plane
	DeploymentStatus string `json:"deploymentStatus,omitempty"`

	ServiceStatus string `json:"serviceStatus,omitempty"`
}

// Define the license details
type ModelaLicenseSpec struct {
}

// ModelaSpec defines the desired state of Modela
type ModelaSpec struct {

	// The current version of modela cluster
	//+kubebuilder:validation:Optional
	Version *string `json:"version,omitempty"`

	// If true, install the modela cluster is not installed
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	Installed *bool `json:"installed,omitempty"`

	// If true, configure monitoring.
	//+kubebuilder:validation:Optional
	Observability ObservabilitySpec `json:"observability,omitempty"`

	// Define how to access modela cluster
	Access ModelaAccessSpec `json:"access,omitempty"`

	// Desired state of modela licensing
	License ModelaLicenseSpec `json:"license,omitempty"`

	// If true install the default tenant.
	// +kubebuilder:default:=true
	//+kubebuilder:validation:Optional
	DefaultTenant *bool `json:"defaultTenant,omitempty"`

	// Desired state of object storage
	//+kubebuilder:validation:Optional
	ObjectStore ObjectStorageSpec `json:"objectStore,omitempty"`

	// If true, install cert manager if not exist
	//+kubebuilder:validation:Optional
	UseCertManager CertManagerSpec `json:"certManager,omitempty"`

	// Desired state of the system database
	SystemDatabase SystemDatabaseSpec `json:"systemDatabase,omitempty"`

	// Setting of the control plane
	ControlPlane ControlPlaneSpec `json:"controlPlane,omitempty"`

	// Desired state of the data plane
	DataPlane DataPlaneSpec `json:"dataPlane,omitempty"`

	// Desired state of the api gateway
	ApiGateway ApiGatewaySpec `json:"apiGateway,omitempty"`

	Pod *v1.PodTemplate `json:"podTemplate,omitempty"`
}

// ModelaStatus defines the observed state of Modela
type ModelaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Actual Version is the actual modela version
	ActualModelaVersion string `json:"actualModelaVersion,omitempty"`

	// Status of the control plane
	ControlPlane ControlPlaneStatus `json:"control,omitempty"`

	// Status of data plane
	DataPlane DataPlaneStatus `json:"data,omitempty"`

	// Status of the api gateway
	Gateway ApiGatewayStatus `json:"gateway,omitempty"`

	Phase ModelaPhase `json:"phase,omitempty"`

	// ObservedGeneration is the last generation that was acted on
	//+kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Last time the modela installation was upgraded
	//+kubebuilder:validation:Optional
	LastUpgraded *metav1.Time `json:"lastUpgraded,omitempty"`

	// In the case of failure, the DataSource resource controller will set this field with a failure message
	//+kubebuilder:validation:Optional
	FailureMessage *string `json:"failureMessage,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +kubebuilder:validation:Optional
	Conditions []ModelaCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,8,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:path=modelas,singular=modela,shortName="md",categories={data,modela,all}
// Modela is the Schema for the modelas API
type Modela struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelaSpec   `json:"spec,omitempty"`
	Status ModelaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModelaList contains a list of Modela
type ModelaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Modela `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Modela{}, &ModelaList{})
}
