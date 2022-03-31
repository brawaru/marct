package launcher

import "github.com/imdario/mergo"

func MergeVersions(a, b Version) (res Version, err error) {
	res = a
	err = mergo.Merge(&res, b, mergo.WithOverride, mergo.WithAppendSlice)
	return
}
