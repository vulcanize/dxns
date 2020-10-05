//
// Copyright 2020 Wireline, Inc.
//

package utils

import (
	"bytes"
	"encoding/binary"

	set "github.com/deckarep/golang-set"
)

func Int64ToBytes(num int64) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, num)
	return buf.Bytes()
}

func SetToSlice(set set.Set) []string {
	names := []string{}

	for name := range set.Iter() {
		if name, ok := name.(string); ok && name != "" {
			names = append(names, name)
		}
	}

	return names
}

func SliceToSet(names []string) set.Set {
	set := set.NewThreadUnsafeSet()

	for _, name := range names {
		if name != "" {
			set.Add(name)
		}
	}

	return set
}

func AppendUnique(list []string, element string) []string {
	set := SliceToSet(list)
	set.Add(element)

	return SetToSlice(set)
}
