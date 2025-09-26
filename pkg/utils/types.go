package utils

func PtrValueOrElse[T any](p *T, elseValue T) T {
	if p == nil {
		return elseValue
	}
	return *p
}

func PtrStrAlias[T, P ~string](t *T) *P {
	if t == nil {
		return nil
	}
	return Ptr(P(*t))
}

func StringPtr(s string) *string {
	return &s
}

func StringPtrValue[T ~string](p *T) T {
	if p == nil {
		return ""
	}
	return *p
}

func Int32PtrValue(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}

func Int64PtrValue(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

func Int64PtrToInt32Ptr(p *int64) *int32 {
	if p == nil {
		return nil
	}
	res := int32(*p)
	return &res
}

func StringPtrArray(s []string) []*string {
	r := make([]*string, len(s))
	for i := range s {
		r[i] = &s[i]
	}
	return r
}

func StringArray(s []*string) []string {
	r := make([]string, len(s))
	for i := range s {
		r[i] = *s[i]
	}
	return r
}

func Int64Ptr(i int64) *int64 {
	return &i
}

func Int64PtrArray(a []int64) []*int64 {
	r := make([]*int64, len(a))
	for i := range a {
		r[i] = &a[i]
	}
	return r
}

func Uint64Ptr(i uint64) *uint64 {
	return &i
}

func Int32Ptr(i int32) *int32 {
	return &i
}

func IntPtr(i int) *int {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}

func Float64Ptr(b float64) *float64 {
	return &b
}

func Float32Ptr(b float32) *float32 {
	return &b
}

func Ptr[T any](b T) *T {
	return &b
}

func BoolPtrValue(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func IsStringPtrEqual(x, y *string) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return *x == *y
}

func IsInt32PtrEqual(x, y *int32) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return *x == *y
}
