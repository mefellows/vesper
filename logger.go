package vesper

type logPrinter interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

type noOpPrinter struct{}

func (n noOpPrinter) Print(...interface{}) {}

func (n noOpPrinter) Printf(string, ...interface{}) {}

func (n noOpPrinter) Println(...interface{}) {}
