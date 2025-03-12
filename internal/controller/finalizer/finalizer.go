package finalizer

import (
	"context"
	v1alpha1 "github.com/oshribelay/namespace-label/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const namespaceLabelFinalizer = "finalizer.namespacelabel.dana.io"

func EnsureFinalizer(ctx context.Context, c client.Client, namespaceLabel *v1alpha1.NamespaceLabel) error {
	if !controllerutil.ContainsFinalizer(namespaceLabel, namespaceLabelFinalizer) {
		controllerutil.AddFinalizer(namespaceLabel, namespaceLabelFinalizer)
		if err := c.Update(ctx, namespaceLabel); err != nil {
			return err
		}
	}
	return nil
}

func RemoveFinalizer(ctx context.Context, c client.Client, namespaceLabel *v1alpha1.NamespaceLabel) error {
	if controllerutil.ContainsFinalizer(namespaceLabel, namespaceLabelFinalizer) {
		controllerutil.RemoveFinalizer(namespaceLabel, namespaceLabelFinalizer)
		if err := c.Update(ctx, namespaceLabel); err != nil {
			return err
		}
	}
	return nil
}
