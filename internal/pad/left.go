package pad

func Left(input string, length int, padChar rune) string {
	if len(input) >= length {
		return input
	}
	padding := make([]rune, length-len(input))
	for i := range padding {
		padding[i] = padChar
	}
	return string(padding) + input
}
