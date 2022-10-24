package kln

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func FlagForDeletion(client dynamic.Interface, ri ResourceIdentifier, undoSwitch bool) error {
	resources := ListResources(client, ri)
	if len(resources) == 0 {
		return errors.New(fmt.Sprintf("did not find any resources that match the criteria:\n%v", ri))
	}

	for _, resource := range resources {
		var patch []byte
		ns := resource.GetNamespace()
		name := resource.GetName()
		if undoSwitch {
			patch = []byte(`{"metadata":{"labels":{"kln.com/delete":"false"}}}`)
		} else {
			patch = []byte(`{"metadata":{"labels":{"kln.com/delete":"true"}}}`)
		}
		_, err := client.Resource(ri.GVR).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
