package utils

import "sort"

type SortableSlice []int64

func (a SortableSlice) Len() int {
	return len([]int64(a))
}

func (a SortableSlice) Less(i, j int) bool {
	return a[i] < a[j]
}

func (a SortableSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func SortInt64List(a []int64) {
	sort.Sort(SortableSlice(a))
}
