package kln

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func DeleteResources(client dynamic.Interface, gvr schema.GroupVersionResource) error {
	items, _ := client.Resource(gvr).List(context.TODO(), v1.ListOptions{LabelSelector: "kln.com/delete=true"})
	if len(items.Items) == 0 {
		InfoLog.Printf("Nothing to delete in %s", gvr)
	}
	for _, item := range items.Items {
		name := item.GetName()
		ns := item.GetNamespace()
		kind := item.GetKind()
		InfoLog.Printf("Deleting %s %s in %s ns", name, kind, ns)
		err := client.Resource(gvr).Namespace(ns).Delete(context.TODO(), name, v1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
