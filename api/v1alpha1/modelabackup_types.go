package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import v1 "k8s.io/api/core/v1"

// ModelaBackupSpec defines the desired state of ModelaBackup
type ModelaBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The schedule follows the same format used in Kubernetes CronJobs,
	// see https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format
	// Cron Schedule to backup
	Schedule *string `json:"schedule"`
	// The modela cluster to backup
	ModelaRef v1.LocalObjectReference `json:"modelaRef"`
	//+kubebuilder:validation:Optional
	Suspended *bool `json:"suspended,omitempty"`
}

// ModelaBackupStatus defines the observed state of ModelaBackup
type ModelaBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The last time a backup completed
	LastBackup *metav1.Time `json:"startedAt,omitempty"`
	// Last error that occur in the backup.
	LastError string `json:"error,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ModelaBackup is the Schema for the modelabackups API
type ModelaBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelaBackupSpec   `json:"spec,omitempty"`
	Status ModelaBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModelaBackupList contains a list of ModelaBackup
type ModelaBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelaBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModelaBackup{}, &ModelaBackupList{})
}
