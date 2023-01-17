package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/liggitt/tabwriter"
)

type TableWriter struct {
	writer *tabwriter.Writer
}

func (w *TableWriter) Write(col ...string) error {
	_, err := fmt.Fprintln(w.writer, strings.Join(col, "\t"))
	if err != nil {
		return err
	}
	return nil
}

func (w *TableWriter) Flush() error {
	return w.writer.Flush()
}

func NewTable(w io.Writer, headers ...string) (*TableWriter, error) {
	table := &TableWriter{
		writer: tabwriter.NewWriter(w, 6, 4, 3, ' ', tabwriter.RememberWidths),
	}
	if err := table.Write(headers...); err != nil {
		return nil, err
	}
	return table, nil
}
