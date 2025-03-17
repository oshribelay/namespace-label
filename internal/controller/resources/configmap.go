package resources

import (
	"context"
	"fmt"
	"github.com/oshribelay/namespace-label/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateConfigMap(namespaceLabel *v1alpha1.NamespaceLabel, c client.Client, ctx context.Context) (corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-protected-labels-configmap", namespaceLabel.Name),
			Namespace: namespaceLabel.Namespace,

			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(namespaceLabel, schema.GroupVersionKind{
					Group:   "namespacelabel.dana.io",
					Version: "v1alpha1",
					Kind:    "NamespaceLabel",
				}),
			},
		},
		Data: map[string]string{
			"k8s.io":        "",
			"kubernetes.io": "",
			"openshift.io":  "",
		},
	}

	if err := c.Create(ctx, configMap); err != nil {
		return corev1.ConfigMap{}, err
	}

	return *configMap, nil
}
