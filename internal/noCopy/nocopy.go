// Made this it's own package to make it easier to track all the structs that can't be copied.
// This way the IDE can find it without doing a text search.
package noCopy

type True struct{}

func (*True) Lock()   {}
func (*True) Unlock() {}
