package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

const (
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

type ResourceIdentifier struct {
	Gvr        schema.GroupVersionResource `yaml:"gvr"`
	MinAge     float64                     `yaml:"minAge"`
	ApiVersion string                      `yaml:"apiVersion"`
	Kind       string                      `yaml:"kind"`
	Metadata   map[string]interface{}      `yaml:"metadata"`
	Spec       map[string]interface{}      `yaml:"spec"`
	Status     map[string]interface{}      `yaml:"status"`
}

func (ri *ResourceIdentifier) FlagForDeletion(client dynamic.Interface) error {
	resources := ListResources(client, *ri)
	if len(resources) == 0 {
		return errors.New(fmt.Sprintf("did not find any resources that match the criteria:\n%v", ri))
	}

	for _, resource := range resources {
		ns := resource.Object["metadata"].(map[string]interface{})["namespace"].(string)
		name := resource.Object["metadata"].(map[string]interface{})["name"].(string)
		patch := []byte(`{"metadata":{"annotations":{"kln.com/delete":"true"}}}`)
		_, err := client.Resource(ri.Gvr).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func ListResources(client dynamic.Interface, ri ResourceIdentifier) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	responseFromServer, err := client.Resource(ri.Gvr).List(context.TODO(), v1.ListOptions{})
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
		if mapIntersectionCheck(metadataFilter, objectMeta) {
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
		objectMeta := item.Object["status"].(map[string]interface{})
		if mapIntersectionCheck(statusFilter, objectMeta) {
			responseList = append(responseList, item)
		}
	}
	return responseList
}

func mapIntersectionCheck(mapA, mapB map[string]interface{}) bool {
	for k, vA := range mapA {
		if vB, ok := mapB[k]; ok && typeof(vA) == typeof(vB) {
			if typeof(vA) == "map[string]interface {}" {
				return mapIntersectionCheck(vA.(map[string]interface{}), vB.(map[string]interface{}))
			}
			if vA != vB {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
