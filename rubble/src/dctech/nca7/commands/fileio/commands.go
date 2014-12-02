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

// NCA v7 File IO Commands.
package fileio

import "dctech/nca7"
import "io/ioutil"

// Adds the file io commands to the state.
// The file io commands are:
//	fileio:read
//	fileio:write
func Setup(state *nca7.State) {
	state.NewNameSpace("fileio")
	state.NewNativeCommand("fileio:read", CommandFileIO_Read)
	state.NewNativeCommand("fileio:write", CommandFileIO_Write)
}

// Read from file.
// 	fileio:read path
// Returns file contents or an error message. May set the Error flag.
func CommandFileIO_Read(state *nca7.State, params []*nca7.Value) {
	if len(params) != 1 {
		panic("Wrong number of params to fileio:read.")
	}

	file, err := ioutil.ReadFile(params[0].String())
	if err != nil {
		state.Error = true
		state.RetVal = nca7.NewValueString("error:" + err.Error())
		return
	}
	state.RetVal = nca7.NewValueString(string(file))
}

// Write to file.
// 	fileio:write path contents
// Returns unchanged or an error message. May set the Error flag.
func CommandFileIO_Write(state *nca7.State, params []*nca7.Value) {
	if len(params) != 2 {
		panic("Wrong number of params to fileio:write.")
	}

	// I have no idea what "0600" means but I saw it in an example.
	// well I do know that it is a file permission.
	err := ioutil.WriteFile(params[0].String(), []byte(params[1].String()), 0600)
	if err != nil {
		state.Error = true
		state.RetVal = nca7.NewValueString("error:" + err.Error())
		return
	}
}
