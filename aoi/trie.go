package aoi

import "strings"

type node struct {
	pattern  string  //待匹配的路由
	part     string  //路由中的部分
	children []*node //子节点
	isWild   bool    //是否使用通配符匹配
}

//matchChild 查询第一个匹配到的节点
func (n *node) matchChild(part string) *node {
	for _, child := range n.children {
		if child.part == part || child.isWild {
			return child
		}
	}
	return nil
}

//matchChildren 查询所有匹配的节点
func (n *node) matchChildren(part string) []*node {
	cs := make([]*node, 0)
	for _, child := range n.children {
		if child.part == part || child.isWild {
			cs = append(cs, child)
		}
	}
	return cs
}

//insert 将对应节点插入至前缀树中去
func (n *node) insert(pattern string, parts []string, height int) {
	//首先判断是否已经到达末尾了
	if len(parts) == height {
		n.pattern = pattern
		return
	}
	//如果没有到末尾
	part := parts[height]
	child := n.matchChild(part) //找到第一个匹配part的节点
	if child == nil {           //说明没有找到合适的节点
		child = &node{
			part:   part,
			isWild: part[0] == ':' || part[0] == '*',
		}
		n.children = append(n.children, child)
	}
	child.insert(pattern, parts, height+1)
}

//search 查找匹配的节点
func (n *node) search(parts []string, height int) *node {
	//说明已经到达了最底层，或者遇到了*通配符
	if len(parts) == height || strings.HasPrefix(n.part, "*") {
		//此时匹配到的为非叶子节点
		if n.pattern == "" {
			return nil
		}
		return n
	}
	part := parts[height]
	childs := n.matchChildren(part) //匹配下一层的所有节点
	for _, child := range childs {
		//所有的孩子节点开始寻找
		result := child.search(parts, height+1)
		if result != nil {
			return result
		}
	}
	return nil
}
