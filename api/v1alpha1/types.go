// +kubebuilder:object:generate=true
// +groupName=scalebit.moodykhalif23.github.com
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{Group: "scalebit.moodykhalif23.github.com", Version: "v1alpha1"}

func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Microservice{},
		&MicroserviceList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Microservice is the Schema for the microservices API
type Microservice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MicroserviceSpec   `json:"spec,omitempty"`
	Status MicroserviceStatus `json:"status,omitempty"`
}

// DeepCopyObject implements the runtime.Object interface
func (m *Microservice) DeepCopyObject() runtime.Object {
	if c := m.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy creates a deep copy of Microservice
func (m *Microservice) DeepCopy() *Microservice {
	if m == nil {
		return nil
	}
	out := new(Microservice)
	m.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies the receiver into out. in must be non-nil.
func (m *Microservice) DeepCopyInto(out *Microservice) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	m.Spec.DeepCopyInto(&out.Spec)
	m.Status.DeepCopyInto(&out.Status)
}

// MicroserviceSpec defines the desired state of Microservice
type MicroserviceSpec struct {
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Port int32 `json:"port"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`
}

// DeepCopyInto copies the receiver into out.
func (s *MicroserviceSpec) DeepCopyInto(out *MicroserviceSpec) {
	*out = *s
}

// DeepCopy creates a deep copy of MicroserviceSpec
func (s *MicroserviceSpec) DeepCopy() *MicroserviceSpec {
	if s == nil {
		return nil
	}
	out := new(MicroserviceSpec)
	s.DeepCopyInto(out)
	return out
}

// MicroserviceStatus defines the observed state of Microservice
type MicroserviceStatus struct {
	// The generation observed by the deployment controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Total number of ready pods targeted by this deployment.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

// DeepCopyInto copies the receiver into out.
func (s *MicroserviceStatus) DeepCopyInto(out *MicroserviceStatus) {
	*out = *s
}

// DeepCopy creates a deep copy of MicroserviceStatus
func (s *MicroserviceStatus) DeepCopy() *MicroserviceStatus {
	if s == nil {
		return nil
	}
	out := new(MicroserviceStatus)
	s.DeepCopyInto(out)
	return out
}

// +kubebuilder:object:root=true

// MicroserviceList contains a list of Microservice
type MicroserviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Microservice `json:"items"`
}

// DeepCopyObject implements the runtime.Object interface
func (m *MicroserviceList) DeepCopyObject() runtime.Object {
	if c := m.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopy creates a deep copy of MicroserviceList
func (m *MicroserviceList) DeepCopy() *MicroserviceList {
	if m == nil {
		return nil
	}
	out := new(MicroserviceList)
	m.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all the fields of this object into out
func (m *MicroserviceList) DeepCopyInto(out *MicroserviceList) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ListMeta.DeepCopyInto(&out.ListMeta)
	if m.Items != nil {
		in, out := &m.Items, &out.Items
		*out = make([]Microservice, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
