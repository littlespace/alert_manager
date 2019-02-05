package listener

// parser is an interface that satisfies the parse method for parsing incoming URL data
// from various sources
type Parser interface {
	Name() string
	Parse(data []byte) (*WebHookAlertData, error)
}

var parsers []Parser

func AddParser(parser Parser) {
	parsers = append(parsers, parser)
}
