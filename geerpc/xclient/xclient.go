package xclient

import (
	"context"
	"geerpc"
	"io"
	"reflect"
	"sync"
)

// 支持负载均衡的客户端
type XClient struct {
	d    Discovery
	mode SelectMode
	opt  *geerpc.Option
	mu   sync.Mutex // protect following
	// 为了尽量复用已经创建好的Socket连接，使用clients保存创建成功的Client实例
	clients map[string]*geerpc.Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *geerpc.Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*geerpc.Client)}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}

// 将复用Client的能力封装在dial方法中
func (xc *XClient) dial(rpcAddr string) (*geerpc.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcAddr]
	// 检查xc.clients是否有缓存的Client，如果有，检查是否是可用状态，如果是则返回缓存的Client，
	// 如果不可用，则从缓存中删除
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = geerpc.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

// Call invokes the named function, waits for it to complete.
// and returns its error status.
// xc will choose a proper server.
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

// Broadcast invokes the named function for every server registered in discovery
// Broadcast 将请求广播到所有的服务实例，如果任意一个实例发生错误，则返回其中一个错误；如果调用成功，则返回其中一个的结果
// 1. 为了提升性能，请求是并发的
// 2. 并发情况下需要使用互斥锁保证error和reply能被正确赋值
// 3. 借助context.WithCancel确保有错误发生时，快速失败
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex // protect e and replyDone
	var e error
	replyDone := reply == nil // if reply is nil, don't need to set value
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				// cancel() // if any call failed, cancel unfinished calls
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
