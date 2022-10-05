package flag

import (
	"context"
	"errors"
	"fmt"
	"time"

	kutility "github.com/adelmoradian/kln/internal/utility"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func FlagForDeletion(client dynamic.Interface, ri kutility.ResourceIdentifier, undoSwitch bool) error {
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
		kutility.InfoLog.Printf("Labelling %s %s in %s ns - dryRun %v", name, resource.GetKind(), ns, undoSwitch)
		_, err := client.Resource(ri.GVR).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func ListResources(client dynamic.Interface, ri kutility.ResourceIdentifier) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	responseFromServer, err := client.Resource(ri.GVR).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return responseList
	}
	responseList, err = filterByAge(responseFromServer, ri.MinAge)
	responseList = filterByMetadata(responseList, ri.Metadata)
	responseList = filterByStatus(responseList, ri.Status)
	return responseList
}

func filterByAge(responseFromServer *unstructured.UnstructuredList, minAgeFilter float64) ([]unstructured.Unstructured, error) {
	var responseList []unstructured.Unstructured
	if minAgeFilter == 0 {
		for _, item := range responseFromServer.Items {
			responseList = append(responseList, item)
		}
		return responseList, nil
	}
	for _, item := range responseFromServer.Items {
		age := time.Since(item.GetCreationTimestamp().Time)
		if age.Hours() > minAgeFilter {
			responseList = append(responseList, item)
		}
	}
	return responseList, nil
}

func filterByMetadata(responseFromServer []unstructured.Unstructured, metadataFilter map[string]interface{}) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	if metadataFilter == nil {
		return responseFromServer
	}
	for _, item := range responseFromServer {
		objectMeta := item.Object["metadata"].(map[string]interface{})
		if kutility.MapIntersection(metadataFilter, objectMeta) {
			responseList = append(responseList, item)
		}
	}
	return responseList
}

func filterByStatus(responseFromServer []unstructured.Unstructured, statusFilter map[string]interface{}) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	if statusFilter == nil {
		return responseFromServer
	}
	for _, item := range responseFromServer {
		objectStatus := item.Object["status"].(map[string]interface{})
		if kutility.MapIntersection(statusFilter, objectStatus) {
			responseList = append(responseList, item)
		}
	}
	return responseList
}
