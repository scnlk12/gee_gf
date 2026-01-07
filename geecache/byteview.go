package geecache

// A Byteview holds an immutable view of bytes
type Byteview struct {
	b []byte // byte类型能够支持任意的数据类型的存储 字符串/图片等
}

// Len returns the view's length
// 实现lru.Cache时 要求被缓存对象必须实现Value接口 即len() int 方法
func (v Byteview) Len() int {
	return len(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// ByteSlice returns a copy of the data as a byte slice
// b只读，使用ByteSlice方法返回一个拷贝 防止缓存值被外部程序修改
func (v Byteview) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String returns the data as a string, making a copy if necessary
func (v Byteview) String() string {
	return string(v.b)
}