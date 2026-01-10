package pointers

// CoalesceInt64 returns the dereferenced custom value if non-nil, otherwise defaultVal
func CoalesceInt64(custom *int64, defaultVal int64) int64 {
	if custom != nil {
		return *custom
	}
	return defaultVal
}

// CoalesceFloat64 returns the dereferenced custom value if non-nil,
// otherwise the dereferenced defaultVal if non-nil, otherwise 0
func CoalesceFloat64(custom *float64, defaultVal *float64) float64 {
	if custom != nil {
		return *custom
	}
	if defaultVal != nil {
		return *defaultVal
	}
	return 0
}

// DerefFloat64 returns the dereferenced value or 0 if nil
func DerefFloat64(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}
