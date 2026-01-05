package gee
import "strings"

type node struct {
	pattern string // 待匹配路由 例如 /p/:lang
	part string // 路由中的一部分 例如 :lang
	children []*node // 子节点 例如[doc, tutorial, intro]
	isWild bool // 是否精准匹配 part含有 :/* 时为true
}

// 第一个匹配成功的节点 用于插入
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

// 所有匹配成功的节点，用于查找
func (n *node) matchChildren(part string) []*node {
	nodes := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			nodes = append(nodes, child)
		}
	}
	return nodes
}

// Trie树 节点的插入
/*
  递归查找每一层的节点，如果没有匹配到当前part的节点，则新建一个
  /p/:lang/doc只有在第三层节点，即doc节点，pattern才会设置为/p/:lang/doc
  p 和 :lang节点的pattern属性均为空
*/
func (n *node) insert(pattern string, parts []string, height int) {
	if len(parts) == height {
		n.pattern = pattern
		return
	}

	part := parts[height]
	child := n.matchChild(part)
	if child == nil {
		child = &node{
			part: part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height + 1)
}

// Trie树 节点的查询
/*
 递归查询每一层的节点，退出规则为 匹配到了* / 匹配失败 / 匹配到了第len(parts)层节点
*/
func (n *node) search(parts []string, height int) *node {
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		if n.pattern == "" {
			return nil
		}
		return n
	}

	part := parts[height]
	children := n.matchChildren(part)

	for _, child := range children {
		result := child.search(parts, height + 1)
		if result != nil {
			return result
		}
	}

	return nil
}