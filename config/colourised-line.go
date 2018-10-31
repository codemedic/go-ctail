package config

import (
	"container/list"
	"sort"
	rbtree "github.com/emirpasic/gods/trees/redblacktree"
)

type colourisedFragment struct {
	text   string
	colour *Colour
}

type colourSegment struct {
	colour     *Colour
	start, end int
}

type ColourisedLine struct {
	WholeLine *Colour
	Partials  []*colourSegment
}

func (cl *ColourisedLine) FormatString(str string) string {
	partials := cl.Partials

	// if there are no partial matches, colourise the whole line
	if len(cl.Partials) == 0 {
		return cl.WholeLine.ColouriseString(str)
	}

	// sort the partial matches so that they are in the order of start-index and for any repeated start-index, by end-index
	sort.SliceStable(partials, func(i, j int) bool {
		if partials[i].start == partials[j].start {
			return partials[i].end < partials[j].end
		}
		return partials[i].start < partials[j].start
	})

	// |<-----<s1.....<s2...s2>......s1>---<s3.....s3>--->|
	// |<-----<s1..s1><s2...s2><s1...s1>---<s3.....s3>--->|

	finalised := list.New()
	closing := list.New()

	// add the whole-line colour

	closing.PushBack(
		finalised.PushBack(&colourSegment{colour: cl.WholeLine, start: 0, end: len(str) - 1}).Value)

	for i, p := range partials {
		finalisedBack, _ := finalised.Back().Value.(*colourSegment)
		closingBack, _ := closing.Back().Value.(*colourSegment)
		finalPartial := i == (len(partials) - 1)

		// if there is a gap after the previous one, we need to fill it with the outer colour
		if p.start > finalisedBack.end {
			closed := closing.Remove(closing.Back()).(*colourSegment)
			if p.start > (finalisedBack.end + 1) {
				finalised.PushBack(&colourSegment{colour: closed.colour, start: 0, end: len(str) - 1}).Value)
			}
		}

		if p.start < finalisedBack.end {
			finalisedBack.end = p.start - 1
			finalised.PushBack(p)
		}
	}
}

/*

1. Add the whole line as a fragment and sort fragments in the order of start-index
2. foreach fragment
3.   Push fragment.colour into colour-stack
4.


0-100 red
3-10 green
12-20 blue

*/
