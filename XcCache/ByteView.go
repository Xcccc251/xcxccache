package XcCache

// 只读
type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回一份[]byte的副本（深拷贝）
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
