package router

import "strconv"

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value
		}
	}
	return ""
}

// EqualsAny is true if the give parameter name equals any of the given values.
// An empty string is both a valid name and value.
func (ps Params) EqualsAny(name string, values ...string) bool {
	value := ps.ByName(name)
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}

// AsID returns the value of the requested param as an int64
func (ps Params) AsID(name string) int64 {
	id, _ := strconv.ParseInt(ps.ByName(name), 10, 64)
	return id
}
