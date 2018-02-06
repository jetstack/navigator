package util

import "reflect"

func Contains(list interface{}, elem interface{}) bool {
	v := reflect.ValueOf(list)
	for i := 0; i < v.Len(); i++ {
		if reflect.DeepEqual(v.Index(i).Interface(), elem) {
			return true
		}
	}
	return false
}

func ContainsAll(list interface{}, elems interface{}) bool {
	listV := reflect.ValueOf(list)
	listLen := listV.Len()
	elemsV := reflect.ValueOf(elems)
	elemsLen := elemsV.Len()
	if listLen < elemsLen {
		return false
	}
Outer:
	for i := 0; i < elemsLen; i++ {
		elem := elemsV.Index(i).Interface()
		for i := 0; i < listLen; i++ {
			if reflect.DeepEqual(listV.Index(i).Interface(), elem) {
				continue Outer
			}
		}
		return false
	}
	return true
}
