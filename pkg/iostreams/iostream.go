package iostreams

import (
	"io"
	"os"
)

type IOStream struct {
	Out io.Writer
}

func NewIOStream() *IOStream {
	return &IOStream{
		Out: os.Stdout,
	}
}
