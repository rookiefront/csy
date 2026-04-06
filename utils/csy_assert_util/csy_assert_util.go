package csy_assert_util

import (
	"reflect"
)

// True 断言条件为真，返回是否通过
func True(condition bool) bool {
	return condition
}

// False 断言条件为假，返回是否通过
func False(condition bool) bool {
	return !condition
}

// Equal 断言两个值相等，返回是否通过
func Equal(expected, actual interface{}) bool {
	return reflect.DeepEqual(expected, actual)
}

// NotEqual 断言两个值不相等，返回是否通过
func NotEqual(expected, actual interface{}) bool {
	return !reflect.DeepEqual(expected, actual)
}

// Nil 断言值为 nil，返回是否通过
func Nil(obj interface{}) bool {
	if obj == nil {
		return true
	}
	val := reflect.ValueOf(obj)
	// 检查指针、接口、切片、map、chan、函数是否为 nil
	switch val.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return val.IsNil()
	default:
		panic("unhandled default case")
	}
	return false
}

// NotNil 断言值不为 nil，返回是否通过
func NotNil(obj interface{}) bool {
	return !Nil(obj)
}

// Zero 断言值为零值，返回是否通过
func Zero(obj interface{}) bool {
	if obj == nil {
		return true
	}
	return reflect.DeepEqual(obj, reflect.Zero(reflect.TypeOf(obj)).Interface())
}

// NotZero 断言值不为零值，返回是否通过
func NotZero(obj interface{}) bool {
	return !Zero(obj)
}

// Empty 断言切片、map、字符串为空，返回是否通过
func Empty(obj interface{}) bool {
	if obj == nil {
		return true
	}
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	case reflect.String:
		return val.Len() == 0
	default:
		return false
	}
}

// NotEmpty 断言切片、map、字符串不为空，返回是否通过
func NotEmpty(obj interface{}) bool {
	return !Empty(obj)
}

// Len 断言长度相等，返回是否通过
func Len(obj interface{}, length int) bool {
	val := reflect.ValueOf(obj)
	switch val.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array, reflect.String:
		return val.Len() == length
	}
	return false
}

// Error 断言错误不为 nil，返回是否通过
func Error(err error) bool {
	return err != nil
}

// NoError 断言错误为 nil，返回是否通过
func NoError(err error) bool {
	return err == nil
}

// Type 断言类型匹配，返回是否通过
func Type(expected, actual interface{}) bool {
	return reflect.TypeOf(expected) == reflect.TypeOf(actual)
}

// Implements 断言实现了接口，返回是否通过
func Implements(interfaceType interface{}, obj interface{}) bool {
	interfacePtr := reflect.TypeOf(interfaceType)
	if interfacePtr.Kind() != reflect.Ptr {
		return false
	}
	interfaceType = interfacePtr.Elem()
	return reflect.TypeOf(obj).Implements(interfaceType.(reflect.Type))
}

// Panics 断言函数会 panic，返回是否通过
/*
if assert.Panics(fnNoPanic) {
		fmt.Println("函数 panic 了")
	} else {
		fmt.Println("✓ 函数没有 panic")
	}

*/
func Panics(fn func()) (panicked bool) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
		}
	}()
	fn()
	return false
}

//
//func IsErrorReflect(v interface{}) bool {
//	// 获取 error 接口的反射类型
//	errorType := reflect.TypeOf((*error)(nil)).Elem()
//	// 获取传入值的反射类型
//	valueType := reflect.TypeOf(v)
//	// 检查是否实现了 error 接口
//	return valueType.Implements(errorType)
//}
