package model

import "testing"

func TestDoneCount(t *testing.T) {
	c := Change{Tasks: []Task{
		{Num: 1, Done: true},
		{Num: 2, Done: false},
		{Num: 3, Done: true},
	}}
	if got := c.DoneCount(); got != 2 {
		t.Errorf("DoneCount() = %d, want 2", got)
	}
}
