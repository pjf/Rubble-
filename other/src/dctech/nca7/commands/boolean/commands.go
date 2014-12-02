/*
Copyright 2012-2013 by Milo Christiansen

This software is provided 'as-is', without any express or implied warranty. In
no event will the authors be held liable for any damages arising from the use of
this software.

Permission is granted to anyone to use this software for any purpose, including
commercial applications, and to alter it and redistribute it freely, subject to
the following restrictions:

1. The origin of this software must not be misrepresented; you must not claim
that you wrote the original software. If you use this software in a product, an
acknowledgment in the product documentation would be appreciated but is not
required.

2. Altered source versions must be plainly marked as such, and must not be
misrepresented as being the original software.

3. This notice may not be removed or altered from any source distribution.
*/

// NCA v7 Boolean Commands.
package boolean

import "dctech/nca7"

// Adds the boolean commands to the state.
// The boolean commands are:
//	int:and
//	int:or
//	int:not
func Setup(state *nca7.State) {
	state.NewNameSpace("bool")
	state.NewNativeCommand("bool:and", CommandBool_And)
	state.NewNativeCommand("bool:or", CommandBool_Or)
	state.NewNativeCommand("bool:not", CommandBool_Not)
}

// Ands two values.
// 	bool:and a b
// Returns a && b
func CommandBool_And(state *nca7.State, params []*nca7.Value) {
	if len(params) != 2 {
		panic("Wrong number of params to bool:and.")
	}

	state.RetVal = nca7.NewValueBool(params[0].Bool() && params[1].Bool())
	return
}

// Ors two values.
// 	bool:or a b
// Returns a || b
func CommandBool_Or(state *nca7.State, params []*nca7.Value) {
	if len(params) != 2 {
		panic("Wrong number of params to bool:or.")
	}

	state.RetVal = nca7.NewValueBool(params[0].Bool() || params[1].Bool())
	return
}

// Inverts a value.
// 	bool:not a
// Returns !a
func CommandBool_Not(state *nca7.State, params []*nca7.Value) {
	if len(params) != 1 {
		panic("Wrong number of params to bool:not.")
	}

	state.RetVal = nca7.NewValueBool(!params[0].Bool())
	return
}
