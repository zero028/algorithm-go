package leetcode

import "testing"

func Test(t *testing.T) {
	t.Log(findTilt(construct([]int{4, 2, 9, 3, 5, 0, 7})))
	t.Log(findTilt(construct([]int{1, 2, 3})))
	t.Log(findTilt(construct([]int{21, 7, 14, 1, 1, 2, 2, 3, 3})))
}
