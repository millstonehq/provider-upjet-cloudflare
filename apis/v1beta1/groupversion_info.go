// Package v1beta1 contains API Schema definitions for the Cloudflare provider v1beta1 API group.
// +kubebuilder:object:generate=true
// +groupName=cloudflare.millstone.tech
// +versionName=v1beta1
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// Group is the API group for Cloudflare resources.
	Group = "cloudflare.millstone.tech"

	// SchemeGroupVersion is group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
