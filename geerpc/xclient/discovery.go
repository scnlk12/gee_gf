package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// SelectMode 表示不同的负载均衡策略
type SelectMode int

const (
	RandomSelect     SelectMode = iota // select randomly
	RoundRobinSelect                   // select using Robbin algorithm
)

type Discovery interface {
	// 从注册中心更新服务列表
	Refresh() error // refresh from remote registry
	// 手动更新服务列表
	Update(servers []string) error
	// 根据负载均衡策略，选择一个服务实例
	Get(mode SelectMode) (string, error)
	// 返回所有的服务实例
	GetAll() ([]string, error)
}

// 实现一个不需要注册中心，服务列表由手工维护的服务发现的结构体
// MultiServersDiscovery is a discovery for multi servers without a registery center
// user provides the server addresses explicitly instead
type MultiServersDiscovery struct {
	r       *rand.Rand   // generate random number
	mu      sync.RWMutex // protect following
	servers []string
	index   int // record the selected position for robin algorithm
}

// NewMultiServerDiscovery creates a MultiServerDiscovery instance
func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		// 产生随机数的实例，初始化时使用时间戳设定随机数种子，避免每次产生相同的随机数序列
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	// index记录Round Robin算法已经轮询到的位置，为了避免每次从0开始，初始化时随机设定一个值
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

// 实现Discovery接口
var _ Discovery = (*MultiServersDiscovery)(nil)

// Refresh doesn't make sense for MultiServerDiscovery, so ignore it.
func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

// Update the servers of discovery dynamically if needed
func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

// Get a server according to mode
func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n] // servers could be updated, so mode n to ensure safety
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

// returns all servers in discovery
func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	// return a copy of d.servers
	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
