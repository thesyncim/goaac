package fdkaac

func checkGroupedSfbOffsets(
	sfbOffset []int,
	sfbCnt int,
	sfbPerGroup int,
	maxSfbPerGroup int,
	requireActiveNonEmpty bool,
	invalidMsg string,
	shortMsg string,
	spectrumLens ...int,
) int {
	if len(sfbOffset) < sfbCnt+1 {
		panic(shortMsg)
	}
	maxOffset := 0
	for sfbGrp := 0; sfbGrp < sfbCnt; sfbGrp += sfbPerGroup {
		prev := sfbOffset[sfbGrp]
		if prev < 0 {
			panic(invalidMsg)
		}
		for sfb := 0; sfb < maxSfbPerGroup; sfb++ {
			next := sfbOffset[sfbGrp+sfb+1]
			if next < prev {
				panic(invalidMsg)
			}
			if requireActiveNonEmpty && next == prev {
				panic(invalidMsg)
			}
			prev = next
		}
		if prev > maxOffset {
			maxOffset = prev
		}
	}
	for _, spectrumLen := range spectrumLens {
		if maxOffset > spectrumLen {
			panic(shortMsg)
		}
	}
	return maxOffset
}
