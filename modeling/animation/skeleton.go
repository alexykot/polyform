package animation

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/vector/vector3"
)

type jointDistItem struct {
	dist  float64
	joint int
	point vector3.Float64
}

type jointItemPriorityQueue []jointDistItem

func (pq jointItemPriorityQueue) Len() int { return len(pq) }

func (pq jointItemPriorityQueue) Less(i, j int) bool {
	return pq[i].dist < pq[j].dist
}

func (pq jointItemPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *jointItemPriorityQueue) Push(x any) {
	item := x.(jointDistItem)
	*pq = append(*pq, item)
}

func (pq *jointItemPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

type skeletonJoint struct {
	path           string
	worldPosition  vector3.Float64
	up, forward    vector3.Float64
	relativeMatrix mat.Matrix4x4
	children       []int
}

type Skeleton struct {
	joints   []skeletonJoint
	jointLUT map[string]int // mapping of path of joints to index in array
}

func (s Skeleton) JointCount() int {
	return len(s.joints)
}

func (s Skeleton) Lookup(name string) int {
	if index, ok := s.jointLUT[name]; ok {
		return index
	}
	panic(fmt.Errorf("skeleton did not contain a joint with the path: %s", name))
}

func (s Skeleton) Children(index int) []int {
	return s.joints[index].children
}

func (s Skeleton) ClosestJoints(point vector3.Float64) {
	const maxPointsToConsider = 3

	queue := make(jointItemPriorityQueue, 0)

	for i, n := range s.joints {
		dist := n.worldPosition.Distance(point)
		if queue.Len() < maxPointsToConsider {
			heap.Push(&queue, jointDistItem{
				dist:  dist,
				joint: i,
				point: n.worldPosition,
			})
		}
	}
}

func (s Skeleton) RelativeMatrix(index int) mat.Matrix4x4 {
	// j := s.joints[index]
	// return mat.MatFromDirs(j.up, j.forward, j.relativePosition)
	return s.joints[index].relativeMatrix
}

func (s Skeleton) InverseBindMatrix(index int) mat.Matrix4x4 {
	j := s.joints[index]
	return mat.MatFromDirs(j.up, j.forward, j.worldPosition).Inverse()
}

func flattenJoints(index int, curPath string, root Joint, parentMat mat.Matrix4x4) []skeletonJoint {
	if root.name == "" {
		panic("joint name can not be empty")
	}

	if strings.Contains(root.name, "/") {
		panic(fmt.Errorf("joint name '%s' can not contain the character '/'", root.name))
	}

	combinedName := root.name
	if curPath != "" {
		combinedName = fmt.Sprintf("%s/%s", curPath, root.name)
	}

	flattened := make([]skeletonJoint, 1)
	flattened[0] = skeletonJoint{
		path: combinedName,
		// relativePosition: root.worldPosition.Sub(parentPos),
		relativeMatrix: parentMat.
			Inverse().
			Multiply(root.Matrix()),
		worldPosition: root.worldPosition,
		up:            root.up,
		forward:       root.forward,
		children:      make([]int, 0),
	}

	offset := index + 1
	for _, child := range root.children {
		flattened[0].children = append(flattened[0].children, offset)
		nodes := flattenJoints(offset, combinedName, child, root.Matrix())
		offset += len(nodes)
		flattened = append(flattened, nodes...)
	}

	return flattened
}

func NewSkeleton(root Joint) Skeleton {
	nodes := flattenJoints(0, "", root, mat.Identity())

	lut := make(map[string]int)
	for i, n := range nodes {
		if _, ok := lut[n.path]; ok {
			panic(fmt.Errorf("skeleton requires unique names for joints that share the same parent, found duplicate %s", n.path))
		}
		lut[n.path] = i
	}

	return Skeleton{
		joints:   nodes,
		jointLUT: lut,
	}
}