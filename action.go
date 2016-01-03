package peg

import "strconv"

func ActionToStr(v *Values, d Any) (Any, error) { return v.S, nil }
func ActionToInt(v *Values, d Any) (Any, error) { return strconv.Atoi(v.S) }
