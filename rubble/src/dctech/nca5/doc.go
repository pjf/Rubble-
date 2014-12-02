// NCA Version 5.
// 
// No Clever Acronym :p.
// 
// NCA is a simple command launguage similar in concept to TCL.
// 
// What's new from v4:
//	Arrays
//	Maps are now treated like Values
// 
// Error message position info may be wrong. The line is almost always correct
// but the column will mostly be a little after the problem.
// If lexing a double-quote string (that is being used as code) that has \n escape sequences
// the position will not match the source.
// In any case the error reporting is light-years better than that of version 3 and before.
// 
package nca5

// TODO:
// Write Documentation
// 

import "fmt"

// CommandTest is a simple command handler for testing (wow... I would have never guessed).
// 
//	// Register via: 
//	state.NewNativeCommand("test", CommandTest)
func CommandTest(state *State, params []*Value) {
	fmt.Println("Test command handler called.")
	fmt.Print("params[")
	for _, val := range params {
		fmt.Print(" \"" + val.String() + "\" ")
	}
	fmt.Print("]\n")
	return
}