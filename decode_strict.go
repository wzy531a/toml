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

	err = verify(primValue, rvalue(v), ignore_fields)
	if err != nil {
		return
	}

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

	err = verify(m.mapping, rvalue(v), ignore_fields)
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

func checkStructType(data interface{}, s reflect.Type) (err error) {
	fmt.Printf("====== checkStructType \n")
	// the other case - map data to a struct type
	dataMap := data.(map[string]interface{})
	for k, v := range dataMap {
		fmt.Printf("CheckTypeMap: k=[%s]\n", k)
		fmt.Printf("CheckTypeMap: v=[%s]\n", v)

		// Build dictionaries up to do field lookups
		var fieldNames = make([]string, 0)
		var origFieldNames = make(map[string]string)

		fmt.Printf("s NumMethod: %d\n", s.NumMethod())
		fmt.Printf("s NumField: %d\n", s.NumField())

		for i := 0; i < s.NumField(); i++ {
			lFieldname := strings.ToLower(s.Field(i).Name)
			fieldNames = append(fieldNames, lFieldname)
			origFieldNames[lFieldname] = s.Field(i).Name
		}
		// Find all keys from map in the datastructure
		// Maps can map down to either golang
		// map[string]interface{} types or to actual golang
		// structs

		for k, _ := range dataMap {
			fmt.Printf("dataMap key: [%s]\n", k)
		}

		for _, v := range fieldNames {
			fmt.Printf("field name: [%s]\n", v)
		}

		for k, _ := range dataMap {
			lKeyName := strings.ToLower(k)
			if !Contains(fieldNames, lKeyName) {
				return fmt.Errorf("Can't find field [%s] in struct\n", k)
			} else {
				fmt.Printf("Found [%s] in struct as [%s]\n", k, origFieldNames[lKeyName])
				f, ok := s.FieldByName(origFieldNames[lKeyName])
				if !ok {
					return fmt.Errorf("Can't find original field [%s]\n", origFieldNames[lKeyName])
				}
				fmt.Printf("Nested type is : [%s]\n\t[%s]\n", f, f.Type)
				if err = CheckType(dataMap[k], f.Type); err != nil {
					return err
				}
			}
		}
	}
	fmt.Printf("====== exit checkStructType ok\n")
	return nil

}

func CheckType(data interface{}, thestruct interface{}) (err error) {
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

	if structAsTypeOk {
		// this should never happen
		fmt.Printf("struct cast to reflect.Type [%s]\n", structAsType)
		return checkTypeStructAsType(data, structAsType)
	} else {
		return checkTypeStructAsValue(data, structAsValue)
	}
}

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

func checkTypeStructAsType(data interface{}, structAsType reflect.Type) (err error) {
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
			if err = CheckType(v, elemType); err != nil {
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
		for i := 0; i < structAsType.NumField(); i++ {
			f := structAsType.Field(i)
			fieldName := structAsType.Field(i).Name
			fmt.Println("--Field info")
			fmt.Printf("Field name in struct: [%s]\n", fieldName)
			//fieldInterface := f.Interface()
			//fmt.Printf("f.Interface() : %s\n", fieldInterface)
			//fmt.Printf("Type of field interface: %s\n", reflect.ValueOf(fieldInterface).Type())
			// Now look up the data map
			mapdata, ok := insensitiveGet(dataMap, fieldName)
			if ok {
				fmt.Printf("Map data @ key [%s]  [%s]\n", fieldName, mapdata)
				return fmt.Errorf("Need to get the type of the field: %s", f)
				/*
					err = CheckType(mapdata, f.Interface())
				*/
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

/////////////
// verify performs a sort of type unification based on the structure of `rv`,
// which is the client representation.
//
// Any type mismatch produces an error. Finding a type that we don't know
// how to handle produces an unsupported type error.
//
// This code is patterned after the unify() function in toml/decode.go
func verify(data interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {
	// Special case. Look for a `Primitive` value.
	if rv.Type() == reflect.TypeOf((*Primitive)(nil)).Elem() {
		return verifyAnything(data, rv, ignore_fields)
	}

	// Special case. Go's `time.Time` is a struct, which we don't want
	// to confuse with a user struct.
	if rv.Type().AssignableTo(rvalue(time.Time{}).Type()) {
		return verifyDatetime(data, rv, ignore_fields)
	}

	k := rv.Kind()

	// laziness
	if k >= reflect.Int && k <= reflect.Uint64 {
		return verifyInt(data, rv, ignore_fields)
	}
	switch k {
	case reflect.Struct:
		return verifyStruct(data, rv, ignore_fields)
	case reflect.Map:
		return verifyMap(data, rv, ignore_fields)
	case reflect.Slice:
		result := verifySlice(data, rv, ignore_fields)
		return result
	case reflect.String:
		return verifyString(data, rv, ignore_fields)
	case reflect.Bool:
		return verifyBool(data, rv, ignore_fields)
	case reflect.Interface:
		// we only support empty interfaces.
		if rv.NumMethod() > 0 {
			e("Unsupported type '%s'.", rv.Kind())
		}
		result := verifyAnything(data, rv, ignore_fields)
		return result
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return verifyFloat64(data, rv, ignore_fields)
	}
	return e("Unsupported type '%s'.", rv.Kind())
}

func verifyStruct(mapping interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	tmap, ok := mapping.(map[string]interface{})
	if !ok {
		return mismatch(rv, "map", mapping)
	}

	rt := rv.Type()

	struct_keys := make(map[string]interface{})
	for i := 0; i < rt.NumField(); i++ {

		sft := rt.Field(i)
		kname := sft.Tag.Get("toml")
		if len(kname) == 0 {
			kname = sft.Name
		}

		struct_keys[kname] = nil
	}

	for k, _ := range tmap {
		if _, ok := insensitiveGet(struct_keys, k); !ok {
			if _, ok = insensitiveGet(ignore_fields, k); !ok {
				return e("Configuration contains key [%s] "+
					"which doesn't exist in struct", k)
			}
		}
	}

	for i := 0; i < rt.NumField(); i++ {
		// A little tricky. We want to use the special `toml` name in the
		// struct tag if it exists. In particular, we need to make sure that
		// this struct field is in the current map before trying to
		// verify it.
		sft := rt.Field(i)
		kname := sft.Tag.Get("toml")
		if len(kname) == 0 {
			kname = sft.Name
		}
		if datum, ok := insensitiveGet(tmap, kname); ok {
			sf := indirect(rv.Field(i))

			// Don't try to mess with unexported types and other such things.
			if sf.CanSet() {
				if err := verify(datum, sf, ignore_fields); err != nil {
					return err
				}
			} else if len(sft.Tag.Get("toml")) > 0 {
				// Bad user! No soup for you!
				return e("Field '%s.%s' is unexported, and therefore cannot "+
					"be loaded with reflection.", rt.String(), sft.Name)
			}
		}
	}
	return nil
}

func verifyMap(mapping interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	tmap, ok := mapping.(map[string]interface{})
	if !ok {
		return badtype("map", mapping)
	}

	if rv.IsNil() {
		rv.Set(reflect.MakeMap(rv.Type()))
	}
	// Just verify each of the keys
	for _, v := range tmap {
		rvval := indirect(reflect.New(rv.Type().Elem()))
		if err := verify(v, rvval, ignore_fields); err != nil {
			return err
		}
	}
	return nil
}

func verifySlice(data interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	slice, ok := data.([]interface{})
	if !ok {
		return badtype("slice", data)
	}

	if rv.IsNil() {
		rv.Set(reflect.MakeSlice(rv.Type(), len(slice), len(slice)))
	} else {
	}

	for i, _ := range slice {
		if err := verify(slice[i], rvalue(slice[i]), ignore_fields); err != nil {
			return err
		}
	}
	return nil
}

func verifyDatetime(data interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	if _, ok := data.(time.Time); ok {
		return nil
	}
	return badtype("time.Time", data)
}

func verifyString(data interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	if _, ok := data.(string); ok {
		return nil
	}
	return badtype("string", data)
}

func verifyFloat64(data interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	if _, ok := data.(float64); ok {
		switch rv.Kind() {
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			return nil
		default:
			panic("bug")
		}
		return nil
	}
	return badtype("float", data)
}

func verifyInt(data interface{}, rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	if _, ok := data.(int64); ok {
		switch rv.Kind() {
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			return nil

		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			return nil

		default:
			panic("bug")
		}
		return nil
	}
	return badtype("integer", data)
}

func verifyBool(data interface{},
	rv reflect.Value,
	ignore_fields map[string]interface{}) error {

	if _, ok := data.(bool); ok {
		return nil
	}
	return badtype("integer", data)
}

func verifyAnything(data interface{}, rv reflect.Value,
	ignore_fields map[string]interface{}) error {
	return nil
}
