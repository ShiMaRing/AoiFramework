package aoicache

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
)

//httpGetter  http客户端
type httpGetter struct {
	baseURL string //用来拼接请求
}

// Get 客户端实现获取kv,服务器还需要实现响应的逻辑，进行分布式请求
func (h *httpGetter) Get(in *Request, out *Response) error {
	//首先拼接url,使用url转义提高安全性，第一个不需要加因为会拼接上basePath
	url := fmt.Sprintf("%v%v/%v",
		h.baseURL, url.QueryEscape(in.Group),
		url.QueryEscape(in.Key),
	)
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)

	err = proto.Unmarshal(bytes, out)

	if err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}
