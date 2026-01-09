package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
// 定义函数类型Hash 采取依赖注入的方式 允许用于替换成自定义的Hash函数 也方便测试时替换 默认是crc.ChecksumIEEE算法
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	hash     Hash
	replicas int            // 虚拟节点倍数
	keys     []int          //Sorted 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射表 key: 虚拟节点哈希值 value: 真实节点的名称
}

// New creates a Map instance
// 允许自定义虚拟节点倍数 和 Hash函数
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash
// 添加真实节点/机器 允许传入0或多个真实节点的名称
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys) // 环上哈希值排序
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// 计算key的哈希值
	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	// 顺时针找到第一个匹配的虚拟节点下标idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 如果idx == len(m.keys) 说明应该选择m.keys[0]
	// 因为m.keys是环状结构，所以用取余的方式处理这种情况
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
