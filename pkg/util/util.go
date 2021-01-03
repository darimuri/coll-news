package util

func EmptyIfNilString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
