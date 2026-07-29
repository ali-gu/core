package ptr

func To[T any](v T) *T {
	return &v
}

func From[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func ToPtrOrNil[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}
