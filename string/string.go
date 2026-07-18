package string

/*
文字列sとtのハミング距離を計算します
*/
func ComputeHammingDistance(s, t string) int {
	dist := 0
	for i := 0; i < len(s); i++ {
		if s[i] != t[i] {
			dist++
		}
	}
	return dist
}
