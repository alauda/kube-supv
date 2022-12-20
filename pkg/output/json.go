package output

import (
	"encoding/json"
	"io"
	"os"

	"github.com/alauda/kube-supv/pkg/errors"
)

func WriteJSON(stdout io.Writer, obj interface{}) error {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stdout == nil {
		return errors.ErrNeedStdOut
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(obj); err != nil {
		return err
	}
	return nil
}
