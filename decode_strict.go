package toml

// Same as PrimitiveDecode but adds a strict verification
func PrimitiveDecodeStrict(primValue Primitive, v interface{}, ignore_fields []string) (err error) {
	err = verify(primValue, rvalue(v), []string{})
	if err != nil {
		return
	}

	err = unify(primValue, rvalue(v))
	return
}

// The same as Decode, except that parsed data that cannot be mapped
// will throw an error
func DecodeStrict(data string, v interface{}) (m MetaData, err error) {
	m, err = Decode(data, v)
	if err != nil {
		return
	}

	err = verify(m.mapping, rvalue(v), []string{})
	return
}
