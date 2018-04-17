package ptr

func Int32(i int32) *int32 {
	return &i
}

func Int64(i int64) *int64 {
	return &i
}

func Bool(b bool) *bool {
	return &b
}

func String(s string) *string {
	return &s
}

func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
