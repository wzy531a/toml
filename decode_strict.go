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

	if !structAsTypeOk {
		structAsValue = reflect.ValueOf(thestruct)
		structAsValueType = structAsValue.Type()
		structAsValueKind = structAsValueType.Kind()
		structAsValueKind = structAsValueType.Kind()
	}
	fmt.Printf("dType: %s\n", dType)
	fmt.Printf("dKind: %s\n", dKind)
	fmt.Printf("structAsTypeOk: %t\n", structAsTypeOk)
	fmt.Printf("structAsValue: %s\n", structAsValue)
	fmt.Printf("structAsValueType: %s\n", structAsValueType)
	fmt.Printf("structAsValueKind: %s\n", structAsValueKind)
	fmt.Printf("structAsValueKind: %s\n", structAsValueKind)

	// TODO:
	// Special case. Go's `time.Time` is a struct, which we don't want
	// to confuse with a user struct.
	if reflect.ValueOf(thestruct).Type().AssignableTo(rvalue(time.Time{}).Type()) {
		// TODO: deal with time.Time types
		if dType.AssignableTo(rvalue(time.Time{}).Type()) {
			fmt.Printf("Time type detected on incoming and assignable\n")
			return nil
		}

		return fmt.Errorf("Invalid type came in for a time.Time gotyp")
	}

	if structAsTypeOk {
		fmt.Printf("struct cast to reflect.Type [%s]\n", structAsType)
		return checkTypeStructAsType(data, structAsType)
	} else {
		return checkTypeStructAsValue(data, structAsValue)
	}
	return nil
}

func checkTypeStructAsValue(data interface{}, structAsValue reflect.Value) (err error) {
	fmt.Printf("-------\ncheckTypeStructAsValue data: [%s]\n\tstruct: [%s]\n", data, structAsValue)

	// switch on the type for structAsValue and call the right check
	structKind := structAsValue.Kind()
	switch structKind {
	case reflect.Map:
		dataAsMap := data.(map[string]interface{})
		// TODO: make sure all keys from dataMap are valid
		for _, k := range structAsValue.MapKeys() {
			keyString := k.Interface().(string)
			fmt.Printf("Map Key: %s\n", keyString)
			fmt.Printf("Data @ key: [%s]\n", dataAsMap[keyString])
			structAtKey := structAsValue.MapIndex(k).Interface()
			fmt.Printf("Struct @ key: [%s]\n", structAtKey)
			fmt.Printf("Struct @ key Type: [%s]\n", structAsValue.MapIndex(k).Type())
			fmt.Printf("Struct @ key Kind: [%s]\n", structAsValue.MapIndex(k).Type().Kind())
			err = CheckType(dataAsMap[keyString], structAtKey)
			if err != nil {
				return err
			}
		}
		// everything is ok
		return nil
	case reflect.Slice:
		// TODO:
		return fmt.Errorf("*** Not done yet! Slice")
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
		// TODO:
		return fmt.Errorf("*** Not done yet! Array")
	case reflect.Struct:
		typeOfStruct := structAsValue.Type()

		dataMap := data.(map[string]interface{})
		// TODO: need to iterate over each key in the data to make
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
			}
			fmt.Println("--End Field info")
		}
		return nil
	default:
		return fmt.Errorf("Unrecognized struct kind: [%s]", structKind)
	}

	return nil
}

func checkTypeStructAsType(data interface{}, structAsType reflect.Type) (er error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Error casting input data to map[string]interface{}")
	}
	fmt.Printf("dataMap: [%s]\n", dataMap)
	return nil
}

func OldCheckType(data interface{}, thestruct interface{}) (err error) {

	dType := reflect.TypeOf(data)
	dKind := dType.Kind()

	structAsType, ok := thestruct.(reflect.Type)
	if !ok {
		// named struct types won't implement Type
	}

	fmt.Printf("=============Checking data   : %s\n", dKind)
	fmt.Printf("=============Checking struct : %s\n", thestruct)
	fmt.Printf("=============Checking struct type: %s\n", structAsType)
	fmt.Printf("=============Checking struct kind: %s\n", reflect.TypeOf(thestruct).Kind())

	// Special case. Go's `time.Time` is a struct, which we don't want
	// to confuse with a user struct.
	if reflect.ValueOf(thestruct).Type().AssignableTo(rvalue(time.Time{}).Type()) {
		// TODO: deal with time.Time types
		fmt.Printf("TODO: handle time.Time types\n")
		return nil
	}

	if dKind >= reflect.Int && dKind <= reflect.Uint64 {
		return fmt.Errorf("int Not implemented")
	}
	switch dKind {
	case reflect.Map:
		// When the incoming data is a map, we're either mapping to a
		// data to an unnamed map type, or a struct.

		// Deal with the case where we've got a map first.
		structType := reflect.TypeOf(thestruct)
		if structType.Kind() == reflect.Map {
			fmt.Printf("Elem Kind: [%s]\n", structType.Elem().Kind())
			// toml was mapped into an interface type
			// Ok, this is an interface on the *struct* - just
			// leave it alone since we want that to be allowed
			// TODO: iterate over the k/v pairs in the data
			dataMap := data.(map[string]interface{})
			for _, v := range dataMap {
				sType := structType.Elem()

				// 2 broad cases.  It's map to interface{} or a
				// map to something else.
				if sType.Kind() == reflect.Interface {
					fmt.Printf("NumMethods: [%d]\n",
						sType.NumMethod())
					if sType.NumMethod() == 0 {
						fmt.Printf("Sweet.  The struct is an empty interface. Terminate!\n")
						return nil
					} else {
						return fmt.Errorf("We don't write data to non-empty interfaces around here\n")
					}
				}

				// For all non-interface{} elem maps, we just
				// recurse down with another call to CheckType
				fmt.Printf("Passing [%s] with [%s] to CheckType\n", v, sType)
				if err = CheckType(v, sType); err != nil {
					return err
				}
			}
			return nil
		}

		// the other case - map data to a struct type
		dataMap := data.(map[string]interface{})
		for k, v := range dataMap {
			fmt.Printf("CheckTypeMap: k=[%s]\n", k)
			fmt.Printf("CheckTypeMap: v=[%s]\n", v)

			// Build dictionaries up to do field lookups
			var fieldNames = make([]string, 0)
			var origFieldNames = make(map[string]string)

			var structAsType reflect.Type
			var ok bool
			structAsType, ok = thestruct.(reflect.Type)
			fmt.Printf("structAsType is : [%s]\n", structAsType)
			if ok {
				// thestruct is an actual struct
				fmt.Printf("CheckTypeMap: typeOf(thestruct)=[%s]\n", reflect.TypeOf(thestruct))
				fmt.Printf("CheckTypeMap: cast to a type = %s\n", structAsType)
				fmt.Printf("structAsType NumMethod: %d\n", structAsType.NumMethod())
				fmt.Printf("structAsType.Kind(): %s\n", structAsType.Kind())

				fmt.Printf("dataMap: %s\n", dataMap)
				if structAsType.Kind() != reflect.Interface {
					return checkStructType(dataMap, structAsType)
				}
			}
			s := reflect.ValueOf(thestruct)
			typeOfT := s.Type()

			fmt.Printf("s NumMethod: %d\n", s.NumMethod())
			fmt.Printf("s NumField: %d\n", s.NumField())

			for i := 0; i < s.NumField(); i++ {
				lFieldname := strings.ToLower(typeOfT.Field(i).Name)
				fieldNames = append(fieldNames, lFieldname)
				origFieldNames[lFieldname] = typeOfT.Field(i).Name
			}
			// Find all keys from map in the datastructure
			// Maps can map down to either golang
			// map[string]interface{} types or to actual golang
			// structs

			for k, _ := range dataMap {
				lKeyName := strings.ToLower(k)
				if !Contains(fieldNames, lKeyName) {
					return fmt.Errorf("Can't find field [%s] in struct\n", k)
				} else {
					fmt.Printf("Found [%s] in struct as [%s]\n", k, origFieldNames[lKeyName])
					f, ok := typeOfT.FieldByName(origFieldNames[lKeyName])
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
		return nil
	case reflect.Slice:
		dataSlice := data.([]interface{})
		// Get the underlying type of the slice in the struct
		structSliceElem := reflect.ValueOf(thestruct).MethodByName("Elem").Call(nil)[0].Interface()

		fmt.Printf("structType: [%s]\n", thestruct)
		fmt.Printf("DynElem(): %s\n", structSliceElem)

		for k, v := range dataSlice {
			fmt.Printf("CheckTypeSlice: key=[%s] data=%r\n", k, dataSlice)
			fmt.Printf("CheckTypeSlice: checking subkey=[%s]\n", v)

			// Build dictionaries up to do field lookups

			fmt.Printf("Items in slice : %d\n", len(dataSlice))
			// Check each of the items in our dataslice against the
			// underlying type of the slice type we are mapping onto
			if err = CheckType(v, structSliceElem); err != nil {
				return err
			}
		}
		return nil
	case reflect.String:
		structType := thestruct.(reflect.Type)
		if structType.Kind() == reflect.String {
			fmt.Printf("golang's reflect API will kill me.  string type matched.\n")
			return nil
		}
		dataStr := data.(string)
		return fmt.Errorf("Error mapping [%s] to type [string]\n", dataStr)
	case reflect.Bool:
		// TODO: do the same thing as reflect.String
		return fmt.Errorf("Not implemented")
	case reflect.Interface:
		// we only support empty interfaces.
		if dType.NumMethod() > 0 {
			e("Unsupported type '%s'.", dKind)
		}
		return nil
	case reflect.Float32, reflect.Float64:
		// TODO: do the same thing as reflect.String
		return fmt.Errorf("Not implemented")
	case reflect.Array:
		// TODO: do the same thing as reflect.Slice
		return fmt.Errorf("Not implemented")
	case reflect.Struct:
		// TODO: pretty sure this is impossible, the incoming data
		// isn't going to be a struct except for maybe datetime which
		// should be handled before the switch/case statement
		return fmt.Errorf("Not implemented")
	default:
		return fmt.Errorf("Unrecognized Type in the parsed data. data: [%s]  type:[%s]",
			data, dType)
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
