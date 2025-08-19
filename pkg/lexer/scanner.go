package lexer

import "strings"

type Scanner struct {
	stream       string
	streamLookup string
	peekBuffer   string
	offset       int
}

func NewScanner(stream string) Scanner {
	// This is an unmodified copy of the input stream which can
	// later be used to search for token locations etc. (for example in Lexer.GetLocation)
	streamLookup := strings.Clone(stream)

	return Scanner{
		stream:       stream,
		streamLookup: streamLookup,
		peekBuffer:   "",
		offset:       0,
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
