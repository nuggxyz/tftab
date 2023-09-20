// Code generated by go generate. DO NOT EDIT.

package opt

import (
	"github.com/algolia/algoliasearch-client-go/v3/algolia/opt"
)

// ExtractReplicas returns the first found ReplicasOption from the
// given variadic arguments or nil otherwise.
func ExtractReplicas(opts ...interface{}) *opt.ReplicasOption {
	for _, o := range opts {
		if v, ok := o.(*opt.ReplicasOption); ok {
			return v
		}
	}
	return nil
}
