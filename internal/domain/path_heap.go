package domain

import "container/heap"

type pathHeap []Path

func (paths pathHeap) Len() int { return len(paths) }

func (paths pathHeap) Less(i, j int) bool {
	return pathBetter(paths[j], paths[i])
}

func (paths pathHeap) Swap(i, j int) { paths[i], paths[j] = paths[j], paths[i] }

func (paths *pathHeap) Push(value any) { *paths = append(*paths, value.(Path)) }

func (paths *pathHeap) Pop() any {
	old := *paths
	last := old[len(old)-1]
	*paths = old[:len(old)-1]
	return last
}

func (paths *pathHeap) add(path Path, limit int) {
	if paths.Len() < limit {
		heap.Push(paths, path)
		return
	}
	if pathBetter(path, (*paths)[0]) {
		(*paths)[0] = path
		heap.Fix(paths, 0)
	}
}
