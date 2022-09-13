package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

func PrintGuestStatus(out io.Writer, id int, text string) {
	fmt.Fprintf(out, "Guest with id (%d) has been %s\n", id, text)
}

func PrintItemCreated(out io.Writer, id, text string) {
	fmt.Fprintf(out, "%s (%s) has been created\n", text, id)
}

func PrintItemUpdated(out io.Writer, id, text string) {
	fmt.Fprintf(out, "%s (%s) has been updated\n", text, id)
}

func PrintItemDeleted(out io.Writer, id, text string) {
	fmt.Fprintf(out, "%s (%s) has been deleted\n", text, id)
}

func PrintItemSet(out io.Writer, id, text string) {
	fmt.Fprintf(out, "%s (%s) has been configured\n", text, id)
}

func PrintRawJson(out io.Writer, input interface{}) {
	list, err := json.Marshal(input)
	LogFatalError(err)
	fmt.Fprintln(out, string(list))
}

func PrintFormattedJson(out io.Writer, input interface{}) {
	list, err := json.MarshalIndent(input, "", "  ")
	LogFatalError(err)
	fmt.Fprintln(out, string(list))
}
