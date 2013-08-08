package toml

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

var typeOfStringSlice = reflect.TypeOf([]string(nil))
var typeOfIntSlice = reflect.TypeOf([]int(nil))

// Same as PrimitiveDecode but adds a strict verification
func PrimitiveDecodeStrict(primValue Primitive,
	v interface{},
	ignore_fields map[string]interface{}) (err error) {

	fmt.Println("============================")
	fmt.Printf("pre         v: [ %s ]\n", v)
	fmt.Printf("pre primValue: [ %s ]\n", primValue)

	err = PrimitiveDecode(primValue, v)
	if err != nil {
		return
	}

	fmt.Printf("post PrimitiveDecode: \n\tprimValue = { %s } \n\tv = { %s }\n", primValue, v)

	thestruct := reflect.ValueOf(v).Elem().Interface()
	err = CheckType(primValue, thestruct, ignore_fields)
	return
}

// The same as Decode, except that parsed data that cannot be mapped
// will throw an error
func DecodeStrict(data string,
	v interface{},
	ignore_fields map[string]interface{}) (m MetaData, err error) {

	m, err = Decode(data, v)
	if err != nil {
		return
	}

	thestruct := reflect.ValueOf(v).Elem().Interface()
	err = CheckType(m.mapping, thestruct, ignore_fields)
	return
}

func Contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

func CheckType(data interface{},
	thestruct interface{},
	ignore_fields map[string]interface{}) (err error) {

	var dType reflect.Type
	var dKind reflect.Kind
	var structAsType reflect.Type
	var structAsTypeOk bool
	var structAsValue reflect.Value
	var structAsValueType reflect.Type
	var structAsValueKind reflect.Kind

	fmt.Println("=============CheckType")
	dType = reflect.TypeOf(data)
	dKind = dType.Kind()

	structAsType, structAsTypeOk = thestruct.(reflect.Type)

	structAsValue = reflect.ValueOf(thestruct)
	structAsValueType = structAsValue.Type()
	structAsValueKind = structAsValueType.Kind()
	structAsValueKind = structAsValueType.Kind()

	// Special case. Go's `time.Time` is a struct, which we don't want
	// to confuse with a user struct.
	timeType := rvalue(time.Time{}).Type()

	if dType == timeType && thestruct == timeType {
		fmt.Printf("Time type detected on thestruct and dType\n")
		return nil
	}
	fmt.Printf("dType: %s\n", dType)
	fmt.Printf("dKind: %s\n", dKind)
	fmt.Printf("data: %s\n", data)
	fmt.Printf("thestruct : %s\n", thestruct)
	fmt.Printf("structAsType: %t\n", structAsType)
	fmt.Printf("structAsTypeOk: %t\n", structAsTypeOk)
	fmt.Printf("structAsValue: %s\n", structAsValue)
	fmt.Printf("structAsValueType: %s\n", structAsValueType)
	fmt.Printf("structAsValueKind: %s\n", structAsValueKind)

	_, structAsValueTypeOK := structAsValueType.(reflect.Type)

	if structAsTypeOk {
		// this should never happen
		fmt.Printf("struct cast to reflect.Type [%s]\n", structAsType)
		return checkTypeStructAsType(data,
			structAsType,
			ignore_fields)
	} else if structAsValueTypeOK {
		return checkTypeStructAsType(data,
			structAsValueType,
			ignore_fields)
	} else {
		return fmt.Errorf("this shouldn't happen")
	}
}

/*
func checkTypeStructAsValue(data interface{}, structAsValue reflect.Value) (err error) {
	fmt.Printf("-------\ncheckTypeStructAsValue data: [%s]\n\tstruct: [%s]\n", data, structAsValue)

	// switch on the type for structAsValue and call the right check
	structKind := structAsValue.Kind()

	switch structKind {
	case reflect.Map:
		dataAsMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Expected data to be a map: [%s]", data)
		}
		// make sure all keys from dataMap are valid
		for _, k := range structAsValue.MapKeys() {
			keyString := k.Interface().(string)
			fmt.Printf("Map Key: %s\n", keyString)
			fmt.Printf("Data @ key: [%s]\n", dataAsMap[keyString])
			structAtKey := structAsValue.MapIndex(k).Interface()
			fmt.Printf("Struct @ key: [%s]\n", structAtKey)
			fmt.Printf("Struct @ key Type: [%s]\n", structAsValue.MapIndex(k).Type())
			fmt.Printf("Struct @ key Kind: [%s]\n", structAsValue.MapIndex(k).Type().Kind())
			fmt.Printf("Struct @ key ValueOf: [%s]\n", reflect.ValueOf(structAtKey))
			fmt.Printf("Struct @ key ValueOf.Type(): [%s]\n", reflect.ValueOf(structAtKey).Type())
			structType := reflect.ValueOf(structAtKey).Type()
			err = CheckType(dataAsMap[keyString], structType)
			if err != nil {
				return err
			}
		}
		// everything is ok
		return nil
	case reflect.Slice:
		dataSlice := data.([]interface{})
		// Get the underlying type of the slice in the struct
		structSliceElem := structAsValue.Type().Elem()

		fmt.Printf("structType: [%s]\n", structAsValue.Type())
		fmt.Printf("DynElem(): %s\n", structSliceElem)

		fmt.Printf("Items in slice : %d\n", len(dataSlice))
		for k, v := range dataSlice {
			fmt.Printf("CheckTypeSlice: k=[%s]\n", k)
			fmt.Printf("CheckTypeSlice: v=[%s]\n", v)

			// Check each of the items in our dataslice against the
			// underlying type of the slice type we are mapping onto
			elemType := structSliceElem.(reflect.Type)
			fmt.Printf("v: [%s] structSliceElem: [%s]\n", v, structSliceElem)
			fmt.Printf("value of structAsValue: [%s]\n", reflect.ValueOf(structAsValue))
			fmt.Printf("type of structAsValue: [%s]\n", reflect.ValueOf(structAsValue).Type())
			fmt.Printf("interface value of structAsValue: [%s]\n", reflect.ValueOf(structAsValue).Interface())
			fmt.Printf("elemType: [%s]\n", elemType)
			if err = CheckType(v, elemType); err != nil {
				return err
			}

		}
		return nil

	case reflect.String:
		_, ok := data.(string)
		if ok {
			fmt.Println("strings are ok")
			return nil
		}
		return fmt.Errorf("Incoming type didn't match gotype string")
	case reflect.Bool:
		_, ok := data.(bool)
		if ok {
			fmt.Println("bool is ok")
			return nil
		}
		return fmt.Errorf("Incoming type didn't match gotype bool")
	case reflect.Interface:
		if structAsValue.NumMethod() == 0 {
			// Anything would be ok from the data side - just accept
			// it

			return nil
		} else {
			return fmt.Errorf("We don't write data to non-empty interfaces around here")
		}
	case reflect.Float32, reflect.Float64:
		var ok bool
		_, ok = data.(float32)
		if ok {
			fmt.Println("float32 is ok")
			return nil
		}
		_, ok = data.(float64)
		if ok {
			fmt.Println("float64 is ok")
			return nil
		}
		return fmt.Errorf("Incoming type didn't match gotype float64")
	case reflect.Array:
		return fmt.Errorf("*** This shouldn't happen")
	case reflect.Struct:
		typeOfStruct := structAsValue.Type()
		dataMap := data.(map[string]interface{})
		// need to iterate over each key in the data to make
		// sure it exists in typeOfStruct
		for i := 0; i < structAsValue.NumField(); i++ {
			f := structAsValue.Field(i)
			fieldName := typeOfStruct.Field(i).Name
			fmt.Println("--Field info")
			fmt.Printf("Field name in struct: [%s]\n", fieldName)
			fieldInterface := f.Interface()
			fmt.Printf("f.Interface() : %s\n", fieldInterface)
			fmt.Printf("Type of field interface: %s\n", reflect.ValueOf(fieldInterface).Type())
			// Now look up the data map
			mapdata, ok := insensitiveGet(dataMap, fieldName)
			if ok {
				fmt.Printf("Map data @ key [%s]  [%s]\n", fieldName, mapdata)
				err = CheckType(mapdata, f.Interface())
				if err != nil {
					return err
				}
			} else {
				fmt.Printf("Can't find data for : [%s]\n\tdata: [%s]", fieldName, dataMap)
			}
			fmt.Println("--End Field info")
		}
		return nil
	default:
		return fmt.Errorf("Unrecognized struct kind: [%s]", structKind)
	}

	return nil
}
*/

func checkTypeStructAsType(data interface{},
	structAsType reflect.Type,
	ignore_fields map[string]interface{}) (err error) {
	fmt.Printf("type: %s\n", reflect.ValueOf(data).Type())
	fmt.Printf("structAsType: %s\n", structAsType)
	dType := reflect.ValueOf(data).Type()
	dKind := dType.Kind()

	// Handle all the int types
	dIsInt := (dKind >= reflect.Int && dKind <= reflect.Uint64)
	sIsInt := (structAsType.Kind() >= reflect.Int && structAsType.Kind() <= reflect.Uint64)
	if dIsInt && sIsInt {
		fmt.Printf("woot. It's an int\n")
		return nil
	}

	structKind := structAsType.Kind()
	switch structKind {
	case reflect.Map:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Expected data to be a map: [%s]", data)
		}
		// Check the elem
		structMapElem := structAsType.Elem()

		fmt.Printf("structAsType: [%s]\n", structAsType)
		fmt.Printf("DynElem(): %s\n", structMapElem)

		fmt.Printf("Items in map: %d\n", len(dataMap))
		for k, v := range dataMap {
			fmt.Printf("CheckTypeMap: k=[%s]\n", k)
			fmt.Printf("CheckTypeMap: v=[%s]\n", v)

			// Check each of the items in our dataMap against the
			// underlying type of the slice type we are mapping onto
			elemType := structMapElem.(reflect.Type)
			fmt.Printf("map elemType: [%s]\n", elemType)
			if err = CheckType(v, elemType, ignore_fields); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		dataSlice := data.([]interface{})
		// Get the underlying type of the slice in the struct
		structSliceElem := structAsType.Elem()

		fmt.Printf("structAsType: [%s]\n", structAsType)
		fmt.Printf("DynElem(): %s\n", structSliceElem)

		fmt.Printf("Items in slice : %d\n", len(dataSlice))
		for k, v := range dataSlice {
			fmt.Printf("CheckTypeSlice: k=[%s]\n", k)
			fmt.Printf("CheckTypeSlice: v=[%s]\n", v)

			// Check each of the items in our dataslice against the
			// underlying type of the slice type we are mapping onto
			elemType := structSliceElem.(reflect.Type)
			fmt.Printf("elemType: [%s]\n", elemType)
			if err = CheckType(v, elemType, ignore_fields); err != nil {
				return err
			}

		}
		return nil
	case reflect.String:
		_, ok := data.(string)
		if ok {
			fmt.Println("strings are ok")
			return nil
		}
		return fmt.Errorf("Incoming type didn't match gotype string")
	case reflect.Bool:
		_, ok := data.(bool)
		if ok {
			fmt.Println("bool is ok")
			return nil
		}
		return fmt.Errorf("Incoming type didn't match gotype bool")
	case reflect.Interface:
		if structAsType.NumMethod() == 0 {
			// Anything would be ok from the data side - just accept
			// it
			fmt.Printf("Empty interface - accepting anything is ok\n")
			return nil
		} else {
			return fmt.Errorf("We don't write data to non-empty interfaces around here")
		}
	case reflect.Float32, reflect.Float64:
		var ok bool
		_, ok = data.(float32)
		if ok {
			fmt.Println("float32 is ok")
			return nil
		}
		_, ok = data.(float64)
		if ok {
			fmt.Println("float64 is ok")
			return nil
		}
		return fmt.Errorf("Incoming type didn't match gotype float64")
	case reflect.Array:
		return fmt.Errorf("*** This shouldn't happen")
	case reflect.Struct:
		dataMap := data.(map[string]interface{})
		// need to iterate over each key in the data to make
		// sure it exists in structAsType
		mapKeys := make([]string, 0)
		for k, _ := range dataMap {
			mapKeys = append(mapKeys, strings.ToLower(k))
		}
		structKeys := make([]string, 0)
		var fieldName string
		for i := 0; i < structAsType.NumField(); i++ {
			f := structAsType.Field(i)

			fieldName = f.Tag.Get("toml")
			if len(fieldName) == 0 {
				fieldName = f.Name
			}
			structKeys = append(structKeys, strings.ToLower(fieldName))
		}

		for _, k := range mapKeys {
			if !Contains(structKeys, k) {
				if _, ok := insensitiveGet(ignore_fields, k); !ok {
					return e("Configuration contains key [%s] "+
						"which doesn't exist in struct", k)
				}
			}
		}

		for i := 0; i < structAsType.NumField(); i++ {
			f := structAsType.Field(i)
			fieldName := f.Name
			fmt.Println("--Field info")
			fmt.Printf("Field name in struct: [%s]\n", fieldName)
			// Now look up the data map
			mapdata, ok := insensitiveGet(dataMap, fieldName)
			if ok {
				fmt.Printf("key [%s]  mapdata: [%s] f.Type[%s]\n", fieldName, mapdata, f.Type)
				err = CheckType(mapdata, f.Type, ignore_fields)
				if err != nil {
					return err
				}
			} else {
				fmt.Printf("Can't find data for : [%s]\n\tdata: [%s]", fieldName, dataMap)
			}
			fmt.Println("--End Field info")
		}
		return nil
	default:
		return fmt.Errorf("Unrecognized struct kind: [%s]", structKind)
	}

	return nil
}
