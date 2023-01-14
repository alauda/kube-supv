package output

import (
	"encoding/json"
	"errors"
	"io"
	"os"
)

var (
	ErrNeedStdOut = errors.New("need stdout")
)

func WriteJSON(stdout io.Writer, obj interface{}) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stdout == nil {
		return ErrNeedStdOut
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(obj); err != nil {
		return err
	}
	return nil
}
