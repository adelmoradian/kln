package kln

import (
	"context"
	"errors"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func ListResources(client dynamic.Interface, ri ResourceIdentifier) ([]unstructured.Unstructured, error) {
	var responseList []unstructured.Unstructured

	responseFromServer, err := client.Resource(ri.GVR).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	responseList, err = filterByAge(responseFromServer, ri.MinAge)
	if err != nil {
		return nil, err
	}

	if len(responseList) != 0 {
		responseList = filterByField(responseList, map[string]interface{}{"metadata": ri.Metadata, "spec": ri.Spec, "status": ri.Status})
	}
	return responseList, nil
}

func filterByAge(responseFromServer *unstructured.UnstructuredList, minAge float64) ([]unstructured.Unstructured, error) {
	var responseList []unstructured.Unstructured

	if minAge == 0 {
		return responseFromServer.Items, nil
	}

	if minAge < 0 {
		return nil, errors.New("minAge cannot be negative")
	}

	for _, item := range responseFromServer.Items {
		age := time.Since(item.GetCreationTimestamp().Time)
		if age.Hours() > minAge {
			responseList = append(responseList, item)
		}
	}
	return responseList, nil
}

func filterByField(responseFromServer []unstructured.Unstructured, filters map[string]interface{}) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured

	for field, v := range filters {
		filter := v.(map[string]interface{})
		if filter == nil {
			continue
		}

		for _, item := range responseFromServer {
			objectField := item.Object[field].(map[string]interface{})
			if mapIntersection(filter, objectField) && !unstructuredArrayInclude(responseList, item) {
				responseList = append(responseList, item)
			}
		}

		responseFromServer = responseList
	}
	return responseFromServer
}
