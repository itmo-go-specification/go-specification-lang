package specir

type Span struct {
	File  string
	Start int
	End   int
}

type Position struct {
	File   string // filename, if any
	Offset int    // offset, starting at 0
	Line   int    // line number, starting at 1
	Column int    // column number, starting at 1 (byte count)
}
