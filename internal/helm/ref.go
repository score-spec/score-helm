package helm

func Ref[k any](input k) *k {
	return &input
}
