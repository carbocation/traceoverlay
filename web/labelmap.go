package main

import "sort"

type Label struct {
	Label string
	ID    uint   `json:"id"`
	Color string `json:"color"`
}

type LabelMap map[string]Label

// Valid ensures that the LabelMap is valid by testing that it is bijective,
// starts with 0, and has no gaps. If not, it's invalid.
func (l LabelMap) Valid() bool {
	inverse := make(map[uint]string)
	for k, v := range l {
		inverse[v.ID] = k
	}

	// Bijective?
	if !(len(l) == len(inverse)) {
		return false
	}

	// Starts with 0 and has consecutive integers?
	for i := 0; i < len(inverse); i++ {
		if _, exists := inverse[uint(i)]; !exists {
			return false
		}
	}

	return true
}

func (l LabelMap) Sorted() []Label {
	out := make([]Label, 0, len(l))

	for k, v := range l {
		v.Label = k
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[j].ID > out[i].ID
	})

	return out
}
