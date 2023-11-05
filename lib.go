package sqlmy

func P[V any](v V) *V {
	return &v
}
