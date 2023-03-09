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
	"encoding/json"
	"errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// The current phase of a Modela installation
type ModelaPhase string

const (
	ModelaPhaseInstallingVault         = "InstallingVault"
	ModelaPhaseInstallingCertManager   = "InstallingCertManager"
	ModelaPhaseInstallingObjectStorage = "InstallingObjectStorage"
	ModelaPhaseInstallingOnlineStore   = "InstallingOnlineStore"
	ModelaPhaseInstallingNginx         = "InstallingNginx"
	ModelaPhaseInstallingPrometheus    = "InstallingPrometheus"
	ModelaPhaseInstallingGrafana       = "InstallingGrafana"
	ModelaPhaseInstallingLoki          = "InstallingLoki"
	ModelaPhaseInstallingDatabase      = "InstallingSystemDatabase"
	ModelaPhaseInstallingModela        = "InstallingModela"
	ModelaPhaseInstallingTenant        = "InstallingTenant"
	ModelaPhaseReady                   = "Ready"
	ModelaPhaseUninstalling            = "UninstallingComponent"
	ModelaPhaseFailed                  = "Failed"
)

var (
	ComponentNotInstalledByModelaError = errors.New("component not installed by Modela Operator")
	ComponentMissingResourcesError     = errors.New("component missing resources")
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

// Unstructured values for rendering Helm Charts
// +k8s:deepcopy-gen=false
type ChartValues struct {
	// Object is a JSON compatible map with string, float, int, bool, []interface{}, or
	// map[string]interface{} children.
	Object map[string]interface{} `json:"-"`
}

// MarshalJSON ensures that the unstructured object produces proper
// JSON when passed to Go's standard JSON library.
func (u *ChartValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Object)
}

// UnmarshalJSON ensures that the unstructured object properly decodes
// JSON when passed to Go's standard JSON library.
func (u *ChartValues) UnmarshalJSON(data []byte) error {
	m := make(map[string]interface{})
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	u.Object = m

	return nil
}

// Declaring this here prevents it from being generated.
func (u *ChartValues) DeepCopyInto(out *ChartValues) {
	out.Object = runtime.DeepCopyJSON(u.Object)
}

// ModelaAccessSpec defines the configuration for Modela to be exposed externally through Ingress resources.
// The Kubernetes Ingress Class annotation (kubernetes.io/ingress.class) must be defined in the parent Modela resource
// for Ingress resources to be created.
type ModelaAccessSpec struct {
	// IngressEnabled indicates if Ingress resources will be created to expose the Modela API gateway, proxy, and frontend.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`
	// +kubebuilder:default:=false
	// InstallNginx indicates if the NGINX Ingress Controller will be installed
	InstallNginx bool `json:"installNginx,omitempty"`
	// NginxValues is the set of Helm values that is used to render the Nginx Ingress Chart.
	// Values are determined from https://artifacthub.io/packages/helm/ingress-nginx/ingress-nginx
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	NginxValues ChartValues `json:"nginxValues,omitempty"`
	// Hostname specifies the host domain which will be used as the hostname for rules in Ingress resources managed
	// by the Modela operator. By default, the hostname will default to a localhost alias.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=""
	Hostname *string `json:"hostname,omitempty"`
}

type ApiGatewaySpec struct {
	// Define the number of API Gateway replicas
	// +kubebuilder:default:=0
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	// Resources specifies resource requests and limits for the data plane deployment.
	// Default values: 100m CPU request, 200m CPU limit, 128Mi memory request, 256Mi memory limit.
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type ControlPlaneSpec struct {
	// The number of Control Plane replicas
	// +kubebuilder:default:=0
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Resources specifies resource requests and limits for the control plane deployment.
	// Default values: 256m CPU request, 512m CPU limit, 256Mi memory request, 512Mi memory limit.
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type DataPlaneSpec struct {
	// The number of Data Plane replicas
	// +kubebuilder:default:=0
	// +kubebuilder:validation:Optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Resources specifies resource requests and limits for the data plane deployment.
	// Default values: 100m CPU request, 200m CPU limit, 256Mi memory request, 512Mi memory limit.
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

type CertManagerSpec struct {
	// Indicates if cert-manager should be installed.
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	Install bool `json:"install"`

	// ChartValues is the set of Helm values that is used to render the Cert Manager Chart.
	// Values are determined from https://artifacthub.io/packages/helm/cert-manager/cert-manager.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	Values ChartValues `json:"values,omitempty"`
}

type VaultSpec struct {
	// Indicates if Vault should be installed. Enabling installation will initialize Vault on the modela-system
	// namespace and configure it with the appropriate secret engine and policies. This option is not recommended
	// for production environments as the root token and vault keys will be stored inside Kubernetes secrets.
	// When installed this way, the Modela Operator will automatically unseal the Vault when necessary.
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	Install bool `json:"install"`

	// MountPath specifies the path where secrets consumed by Modela will be stored.
	// +kubebuilder:default:="modela/secrets"
	MountPath string `json:"mountPath,omitempty"`

	// VaultAddress specifies the address for an external Vault server. If specified, the Vault server
	// must be configured with a KVv2 secret engine mounted at MountPath. It must also be configured to
	// authorize the modela-operator-controller-manager ServiceAccount with read/write permissions
	// +kubebuilder:validation:Optional
	VaultAddress *string `json:"vaultAddress,omitempty"`

	// ChartValues is the set of Helm values that are used to render the Vault Chart.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	Values ChartValues `json:"values,omitempty"`
}

type ObjectStorageSpec struct {
	// Indicates if Minio should be installed.
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	Install bool `json:"install"`

	// ChartValues is the set of Helm values that is used to render the Minio Chart.
	Values ChartValues `json:"values,omitempty"`
}

type SystemDatabaseSpec struct {
	// ChartValues is the set of Helm values that is used to render the Postgres Chart.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	Values ChartValues `json:"values,omitempty"`
}

type OnlineStoreSpec struct {
	// Indicates if Redis should be installed as part of the built-in online store.
	// +kubebuilder:default:=true
	// +kubebuilder:validation:Optional
	Install bool `json:"install"`

	// ChartValues is the set of Helm values that is used to render the Redis Chart.
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	Values ChartValues `json:"values,omitempty"`
}

type ObservabilitySpec struct {
	// Prometheus indicates if the Prometheus Helm Chart will be installed
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Prometheus bool `json:"installPrometheus,omitempty"`
	// ChartValues is the set of Helm values that is used to render the Prometheus Chart.
	// Values are determined from https://artifacthub.io/packages/helm/prometheus-community/prometheus
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	PrometheusValues ChartValues `json:"prometheusValues,omitempty"`

	// Loki indicates if the Loki Helm Chart will be installed
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Loki bool `json:"installLoki,omitempty"`
	// ChartValues is the set of Helm values that is used to render the Loki Chart.
	// Values are determined from https://artifacthub.io/packages/helm/grafana/loki
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	LokiValues ChartValues `json:"lokiValues,omitempty"`

	// Grafana indicates if the Grafana Helm Chart will be installed
	// +kubebuilder:default:=false
	//+kubebuilder:validation:Optional
	Grafana bool `json:"installGrafana,omitempty"`
	// ChartValues is the set of Helm values that is used to render the Grafana Chart.
	// Values are determined from https://artifacthub.io/packages/helm/grafana/grafana
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Optional
	GrafanaValues ChartValues `json:"grafanaValues,omitempty"`
}

type ModelaLicenseSpec struct {
	//+kubebuilder:validation:Optional
	LicenseKey *string `json:"licenseKey,omitempty"`

	// If LinkLicense is enabled, the Modela Operator will open a linking session through modela.ai which a
	// system administrator can use to log in to their modela.ai account and link their license. The URL which
	// must be opened by the administrator will be stored in the status of the Modela resource.
	LinkLicense *bool `json:"linkLicense,omitempty"`
}

type TenantSpec struct {
	// The name of the Tenant. This will determine the name of the namespace containing the Tenant's resources.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// The password for the default admin account (with the username "admin"). If empty, then the
	// Modela Operator will generate a random, 16-character long password that will be logged to
	// the output of the operator's container.
	// +kubebuilder:validation:Optional
	AdminPassword *string `json:"adminPassword,omitempty"`
}

// ModelaSpec defines the desired state of Modela
type ModelaSpec struct {
	// Distribution denotes the desired version of Modela. This version will determine the
	// Docker image tags for all Modela images provided by Metaprov
	// +kubebuilder:default:="develop"
	// +kubebuilder:validation:Required
	Distribution string `json:"distribution"`

	// Observability specifies the configuration to install monitoring tools (Prometheus, Loki, Grafana)
	Observability ObservabilitySpec `json:"observability,omitempty"`

	// Ingress specifies the configuration to install Ingress resources that will expose Modela externally
	Ingress ModelaAccessSpec `json:"ingress,omitempty"`

	// License specifies the license information
	// that will be applied to the installation of Modela
	License ModelaLicenseSpec `json:"license,omitempty"`

	// Tenants contains the collection of tenants that will be installed
	//+kubebuilder:validation:Optional
	Tenants []*TenantSpec `json:"tenants,omitempty"`

	//+kubebuilder:validation:Optional
	CertManager CertManagerSpec `json:"certManager,omitempty"`

	//+kubebuilder:validation:Optional
	ObjectStore ObjectStorageSpec `json:"objectStore,omitempty"`

	//+kubebuilder:validation:Optional
	SystemDatabase SystemDatabaseSpec `json:"systemDatabase,omitempty"`

	//+kubebuilder:validation:Optional
	OnlineStore OnlineStoreSpec `json:"onlineStore,omitempty"`

	//+kubebuilder:validation:Optional
	ControlPlane ControlPlaneSpec `json:"controlPlane,omitempty"`

	//+kubebuilder:validation:Optional
	DataPlane DataPlaneSpec `json:"dataPlane,omitempty"`

	//+kubebuilder:validation:Optional
	ApiGateway ApiGatewaySpec `json:"apiGateway,omitempty"`

	//+kubebuilder:validation:Optional
	Vault VaultSpec `json:"vault,omitempty"`
}

// ModelaStatus defines the observed state of Modela
type ModelaStatus struct {
	// InstalledVersion denotes the live image tags of all Modela images
	InstalledVersion string `json:"installedVersion,omitempty"`

	// InstalledTenants contains the names of all Tenants installed by the Modela resource
	Tenants []string `json:"installedTenants,omitempty"`

	// LicenseToken contains the reference to the license token generated by the license linking process, which
	// can be used to fetch the active license of a modela.ai
	LicenseToken *v1.ObjectReference `json:"licenseTokenRef,omitempty"`

	// LinkLicenseUrl contains the URL which the system administrator must open in order to link their
	// https://modela.ai account to the Modela Operator. Once linked, the Modela Operator will automatically
	// fetch the license of their account in the case that will it will expire.
	LinkLicenseUrl *string `json:"linkLicenseUrl,omitempty"`

	Phase ModelaPhase `json:"phase,omitempty"`

	// The Modela resource controller will update FailureMessage with an error message in the case of a failure
	FailureMessage *string `json:"failureMessage,omitempty"`

	// The last time the Modela resource was updated
	//+kubebuilder:validation:Optional
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +kubebuilder:validation:Optional
	Conditions []ModelaCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,8,rep,name=conditions"`
}

// Modela defines the configuration of the Modela operator
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=modelas,singular=modela,shortName="md",categories={data,modela,all}
type Modela struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelaSpec   `json:"spec,omitempty"`
	Status ModelaStatus `json:"status,omitempty"`
}

// ModelaList contains a list of Modela
// +kubebuilder:object:root=true
type ModelaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Modela `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Modela{}, &ModelaList{})
}
