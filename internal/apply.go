package internal

// ApplyFns
// Sequentially applies ff to x
func ApplyFns[T any](x T, ff ...func(x T) T) T {
	for _, f := range ff {
		x = f(x)
	}

	return x
}
