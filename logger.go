package vesper

type LogPrinter interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type NoOpPrinter struct{}

func (n NoOpPrinter) Print(...interface{}) {}

func (n NoOpPrinter) Printf(string, ...interface{}) {}

func (n NoOpPrinter) Println(...interface{}) {}
