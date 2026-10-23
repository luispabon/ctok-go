package ctok

func isUpper(r rune) bool { return properties(r).flags&4 != 0 }
