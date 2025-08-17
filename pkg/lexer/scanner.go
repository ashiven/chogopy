package lexer

type Scanner struct {
	stream     string
	peekBuffer string
	offset     int
}

func NewScanner(stream string) Scanner {
	return Scanner{
		stream:     stream,
		peekBuffer: "",
		offset:     0,
	}
}

func (s *Scanner) Peek() string {
	if s.peekBuffer == "" {
		s.peekBuffer = s.Consume()
		// 1) we never want to shift the offset when only peeking
		s.offset -= 1
	}

	return s.peekBuffer
}

func (s *Scanner) Consume() string {
	if s.peekBuffer != "" {
		nextChar := s.peekBuffer
		s.peekBuffer = ""
		// 2) we only want to shift the offset once the peeked symbol is actually consumed
		s.offset += 1
		return nextChar
	}

	if len(s.stream) == 0 {
		s.offset += 1
		return ""
	}

	nextChar := string(s.stream[0])
	s.stream = s.stream[1:]
	s.offset += 1

	return nextChar
}
