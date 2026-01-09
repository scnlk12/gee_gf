package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
// http服务端
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self       string
	basePath   string
	mu         sync.Mutex // guards peers and httpGetters
	peers      *consistenthash.Map // 一致性哈希算法的Map 根据具体key选择节点
	// 映射远程节点与对应的httpGetter 每一个远程节点对应一个httpGetter
	// 因为httpGetter与远程节点的地址baseURL有关
	httpGetter map[string]*httpGetter // keyed by e.g. "http://10.0.0.2:8008"
}

// http客户端
type httpGetter struct {
	baseURL string // 表示将要访问的远程节点的地址
}

// NewHTTPPool initializes an HTTP pool of peers
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,            // 记录自己的地址 包括主机名/IP和端口
		basePath: defaultBasePath, // 节点间通讯地址的前缀
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

// Set updates the pool's list of peers.
// 实例化了一致性哈希算法，并且添加了传入的节点 为每一个节点创建了一个HTTP客户端HttpGetter
func (p *HTTPPool) Set(peers ...string)  {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetter = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetter[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
// 包装了一致性哈希算法的Get方法 根据具体的key选择节点，返回节点对应的HTTP客户端
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetter[peer], true
	}
	return nil, false
}

// 编译时检查*httpGetter类型是否实现了PeerGetter接口
var _ PeerGetter = (*httpGetter)(nil)
var _ PeerPicker = (*HTTPPool)(nil)
