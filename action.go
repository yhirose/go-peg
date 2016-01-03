package peg

import "strconv"

func ActionToStr(sv *SemanticValues, dt Any) (Any, error) { return sv.S, nil }
func ActionToInt(sv *SemanticValues, dt Any) (Any, error) { return strconv.Atoi(sv.S) }
