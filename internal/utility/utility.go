package utility

import "fmt"

func MapIntersection(mapA, mapB map[string]interface{}) bool {
	for k, vA := range mapA {
		if vB, ok := mapB[k]; ok && Typeof(vA) == Typeof(vB) {
			if Typeof(vA) == "map[string]interface {}" {
				return MapIntersection(vA.(map[string]interface{}), vB.(map[string]interface{}))
			}
			if Typeof(vA) == "[]interface {}" {
				return ArrayIntersection(vA.([]interface{}), vB.([]interface{}))
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

func ArrayIntersection(sliceSmall, sliceBig []interface{}) bool {
	if len(sliceSmall) > len(sliceBig) {
		return false
	}
	for _, vSmall := range sliceSmall {
		if !Contains(sliceBig, vSmall) {
			return false
		}
	}
	return true
}

func Contains(sliceBig []interface{}, vSmall interface{}) bool {
	for _, vBig := range sliceBig {
		if Typeof(vBig) != Typeof(vSmall) {
			return false
		}
		if Typeof(vBig) == "map[string]interface {}" {
			return MapIntersection(vSmall.(map[string]interface{}), vBig.(map[string]interface{}))
		}
		if Typeof(vBig) == "[]interface {}" {
			return ArrayIntersection(vSmall.([]interface{}), vBig.([]interface{}))
		}
		if vSmall != vBig {
			return false
		}
	}
	return true
}

func Typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
