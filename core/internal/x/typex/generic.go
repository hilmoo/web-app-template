package typex

// MapSlice converts every element of in with fn.
func MapSlice[S, D any](in []S, fn func(S) D) []D {
	out := make([]D, 0, len(in))
	for _, v := range in {
		out = append(out, fn(v))
	}
	return out
}
