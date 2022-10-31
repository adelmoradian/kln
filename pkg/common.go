package kln

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var WarningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

const (
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

type ResourceIdentifier struct {
	GVR         schema.GroupVersionResource `yaml:"gvr"`
	MinAge      float64                     `yaml:"minAge"`
	ApiVersion  string                      `yaml:"apiVersion"`
	Metadata    map[string]interface{}      `yaml:"metadata"`
	Spec        map[string]interface{}      `yaml:"spec"`
	Status      map[string]interface{}      `yaml:"status"`
	Name        string                      `yaml:"name"`
	Description string                      `yaml:"description"`
}

func mapIntersection(mapA, mapB map[string]interface{}) bool {
	for k, vA := range mapA {
		if vB, ok := mapB[k]; ok && typeof(vA) == typeof(vB) {
			if typeof(vA) == "map[string]interface {}" {
				return mapIntersection(vA.(map[string]interface{}), vB.(map[string]interface{}))
			}
			if typeof(vA) == "[]interface {}" {
				return arrayIntersection(vA.([]interface{}), vB.([]interface{}))
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

func arrayIntersection(sliceSmall, sliceBig []interface{}) bool {
	if len(sliceSmall) > len(sliceBig) {
		return false
	}
	for _, vSmall := range sliceSmall {
		if !contains(sliceBig, vSmall) {
			return false
		}
	}
	return true
}

func contains(sliceBig []interface{}, vSmall interface{}) bool {
	for _, vBig := range sliceBig {
		if typeof(vBig) != typeof(vSmall) {
			return false
		}
		if typeof(vBig) == "map[string]interface {}" {
			return mapIntersection(vSmall.(map[string]interface{}), vBig.(map[string]interface{}))
		}
		if typeof(vBig) == "[]interface {}" {
			return arrayIntersection(vSmall.([]interface{}), vBig.([]interface{}))
		}
		if vSmall != vBig {
			return false
		}
	}
	return true
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func GetDynamicClient(kubeconfig string) dynamic.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return client
}

func ReadFile(file string) []byte {
	filename, _ := filepath.Abs(file)
	config, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return config
}

func arrayInclude(array []unstructured.Unstructured, element unstructured.Unstructured) bool {
	for _, e := range array {
		if equality.Semantic.DeepEqual(e.Object, element.Object) {
			return true
		}
	}
	return false
}
