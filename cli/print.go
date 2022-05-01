package cli

import (
	"fmt"
	"io"
	"encoding/json"
)

func PrintRawJson(out io.Writer, input interface{}){
	list, err := json.Marshal(input)
	LogFatalError(err)
	fmt.Fprintln(out,string(list))
}
