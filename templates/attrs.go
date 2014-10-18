package templates

import ()

type Attrs map[string]interface{}

func (a Attrs) Merge(b map[string]interface{}) {
	for key, value := range b {
		a[key] = value
	}
}
