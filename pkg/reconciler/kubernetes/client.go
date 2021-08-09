package kubernetes

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Resource struct {
	Kind      string
	Name      string
	Namespace string
}

func (r *Resource) String() string {
	return fmt.Sprintf("KubernetesResource [Kind:%s,Namespace:%s,Name:%s]", r.Kind, r.Namespace, r.Name)
}

type ResourceInterceptor interface {
	Intercept(resource *unstructured.Unstructured) error
}

type Client interface {
	Deploy(ctx context.Context, manifest string, interceptors ...ResourceInterceptor) ([]*Resource, error)
	Delete(ctx context.Context, manifest string) ([]*Resource, error)
	Clientset() (kubernetes.Interface, error)
	Config() *rest.Config
}
