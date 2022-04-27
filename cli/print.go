package cli

import (
	"fmt"
	"encoding/json"
)

func PrintRawJson(input interface{}){
	list, err := json.Marshal(input)
	LogFatalError(err)
	fmt.Println(string(list))
}
