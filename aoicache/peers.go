package aoicache

//PeerPicker 的 PickPeer() 方法用于根据传入的 key 选择相应节点 PeerGetter。
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

//PeerGetter Get() 方法用于从对应 group 查找缓存值。
//对应于上述流程中的 HTTP 客户端
type PeerGetter interface {
	Get(in *Request, out *Response) error
}
