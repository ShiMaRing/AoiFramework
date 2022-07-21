package aoicache

//存储数据

// Data 保存底层数据
type Data struct {
	bytes []byte
}

//Len 返回数据长度
func (d Data) Len() int {
	return len(d.bytes)
}

//ByteSlice 返回数据克隆备份，避免修改底层数据
func (d Data) ByteSlice() []byte {
	data := make([]byte, len(d.bytes))
	copy(data, d.bytes)
	return data
}

//String 以字符串形式返回数据
func (d Data) String() string {
	return string(d.bytes)
}

func clone(b []byte) []byte {
	data := make([]byte, len(b))
	copy(data, b)
	return data
}
