package cli

import (
	"fmt"
	"io"
	"encoding/json"
)

func PrintItemCreated(out io.Writer, id, text string){
	fmt.Fprintf(out, "%s (%s) has been created\n", text, id)
}

func PrintItemUpdated(out io.Writer, id, text string){
	fmt.Fprintf(out, "%s (%s) has been updated\n", text, id)
}

func PrintItemDeleted(out io.Writer, id, text string){
	fmt.Fprintf(out, "%s (%s) has been deleted\n", text, id)
}

func PrintItemSet(out io.Writer, id, text string){
	fmt.Fprintf(out, "%s (%s) has been configured\n", text, id)
}

func PrintRawJson(out io.Writer, input interface{}){
	list, err := json.Marshal(input)
	LogFatalError(err)
	fmt.Fprintln(out,string(list))
}
