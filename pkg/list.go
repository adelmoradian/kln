package kln

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func ListResources(client dynamic.Interface, ri ResourceIdentifier) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	// InfoLog.Printf("--- GVR: %s, Name: %s, Description: %s\n", ri.GVR, ri.Name, ri.Description)
	responseFromServer, err := client.Resource(ri.GVR).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return responseList
	}
	responseList, err = filterByAge(responseFromServer, ri.MinAge)
	if len(responseList) != 0 {
		responseList = filter(responseList, map[string]interface{}{"metadata": ri.Metadata, "spec": ri.Spec, "status": ri.Status})
	}
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

func filter(responseFromServer []unstructured.Unstructured, filters map[string]interface{}) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	for k, v := range filters {
		filter := v.(map[string]interface{})
		if filter == nil {
			continue
		}
		for _, item := range responseFromServer {
			objectField := item.Object[k].(map[string]interface{})
			if mapIntersection(filter, objectField) && !arrayInclude(responseList, item) {
				responseList = append(responseList, item)
			}
		}
		responseFromServer = responseList
	}
	return responseFromServer
}
