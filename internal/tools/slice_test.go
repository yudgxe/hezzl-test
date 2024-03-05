package tools

import (
	"testing"
)

func TestMergeSlices(t *testing.T) {
	for _, test := range []struct {
		name     string
		base     []int
		add      []int
		index    int
		expected []int
	}{
		{
			name:  "add_in_back",
			base:  []int{1, 2, 3},
			add:   []int{4, 5},
			index: 1000,
		},
		{
			name:  "add_in_start",
			base:  []int{1, 2, 3},
			add:   []int{4, 5},
			index: 0,
		},

		{
			name:  "add_in_midle",
			base:  []int{1, 5},
			add:   []int{2, 3, 4},
			index: 1,
		},
		{
			name:  "negative_index",
			base:  []int{1, 4, 5},
			add:   []int{2, 3},
			index: -1,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := MergeSlices(test.base, test.add, test.index)
			t.Fatal(response)

		})
	}

}
