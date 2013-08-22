package matcher


type RISMatcher struct{

}

func Btoi(inp bool) int {
	if inp == true{
		return 1
	}
	return 0
}

func (ris *RISMatcher) Match(collection []string, needle string, tolerance int) string{
	match := ""
	bestDistance := 0
	low := 0
	high := len(collection) - 1
	bestIndex := 0

	for ;low <= high; {
		mid := (low + high) / 2
		find := collection[mid]
		distance := ris.longestCommonPrefixLength(needle,find)
		if distance >= tolerance && distance > bestDistance {
			bestIndex = mid
			match = find
			bestDistance = distance
		}
		cmp := Btoi(find > needle) - Btoi(needle > find)
		if cmp < 0{
			low = mid + 1
		} else if cmp > 0{
			high = mid - 1
		} else {
			break
		}

	}
	if bestDistance < tolerance{
		return ""
	}
	if bestIndex == 0 {
		return match
	}
	return ris.firstOfTheBests(collection,needle,bestIndex,bestDistance)

}

func (ris *RISMatcher) firstOfTheBests(collection []string, needle string, bestIndex int, bestDistance int) string {
	for;(bestIndex > 0) && (ris.longestCommonPrefixLength(collection[bestIndex - 1],needle) == bestDistance);{
		bestIndex--
	}
	return collection[bestIndex]
}

func (ris *RISMatcher) longestCommonPrefixLength(s,t string) int{
	length := len(s)
	if length > len(t){
		length = len(t)
	}
	var i = 0
	for ;i < length; {
		if(s[i] != t[i]){
			break
		}
		i++
	}
	return i
}