package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"

	"golang.org/x/tools/go/analysis/passes/nilfunc"
)

// A Getter loads data for a key
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function
type GetterFunc func(key string) ([]byte, error)

// A Group is a cache namespace and associated data loaded spread over
// 一个Group可以认为是一个缓存的命名空间
type Group struct {
	name      string // 每个Group拥有一个唯一的名称name
	getter    Getter // 缓存未命中时获取源数据的回调
	mainCache cache  // 一开始实现的并发缓存
	peers     PeerPicker
	// use singleflight.Group to make sure that
	// each key is only fetched once
	loader *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 将 实现了PeerPicker接口的HTTPPool注入到Group中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 使用PickPeer()方法选择节点，若非本机节点，则调用getFromPeer()从远程获取
// 若是本机节点或失败 则回退到getLocally()
func (g *Group) load(key string) (value Byteview, err error) {
	// each key is only fetched once (either locally or remotely)
	// regardless of the number of concurrent callers.
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(Byteview), nil
	}
	return
}

// 使用实现了PeerGetter接口的httpGetter从访问远程节点，获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (Byteview, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return Byteview{}, err
	}
	return Byteview{b: bytes}, nil
}

// Get implements Getter interface function
// 回调函数 当缓存不存在时，调用这个函数，得到源数据
// 函数类型实现某一个接口，称之为接口型函数，方便使用者在调用时既能够传入函数作为参数，
// 也能够传入实现了该接口的结构体作为参数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// NewGroup create a new instance of Group
// 实例化Group 并将group存储在全局变量groups中
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		// 延迟初始化
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named previously created with NewGroup,
// or nil if there's no such groups
// 获取特定名称的Group
func GetGroup(name string) *Group {
	// 使用只读锁 因为不涉及任何冲突变量的写操作
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

/* -----------------------------------------
                            是
接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                |  否                         是
                |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                            |  否
                            |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

   -----------------------------------------*/
// Get value for a key from cache
func (g *Group) Get(key string) (Byteview, error) {
	if key == "" {
		return Byteview{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// 调用用户回调函数g.getter.Get()获取源数据 并将源数据添加到缓存mainCache中
func (g *Group) getLocally(key string) (Byteview, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return Byteview{}, err
	}

	value := Byteview{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value Byteview) {
	g.mainCache.add(key, value)
}
