/*
Copyright 2014 by Milo Christiansen

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

// Rex AXIS VFS Commands
package axisrex

import "dctech/rex"
import "dctech/axis"

// Adds the AXIS VFS commands to the state.
// The AXIS VFS commands are:
//	axis:read
//	axis:write
//	axis:exists
//	axis:isdir
//	axis:del
//	axis:walkdirs
//	axis:walkfiles
func Setup(state *rex.State) (err error) {
	defer func() {
		if !state.NoRecover {
			if x := recover(); x != nil {
				if y, ok := x.(rex.ScriptError); ok {
					err = y
					return
				}
				panic(x)
			}
		}
	}()
	
	mod := state.RegisterModule("axis")
	mod.RegisterCommand("read", Command_Read)
	mod.RegisterCommand("write", Command_Write)
	
	mod.RegisterCommand("exists", Command_Exists)
	mod.RegisterCommand("isdir", Command_IsDir)
	
	mod.RegisterCommand("del", Command_Del)
	
	mod.RegisterCommand("walkdirs", Command_WalkDirs)
	mod.RegisterCommand("walkfiles", Command_WalkFiles)
	
	return nil
}

// Read from a AXIS file.
// 	axis:read ds path
// Returns file contents or an error message. May set the Error flag.
func Command_Read(script *rex.Script, params []*rex.Value) {
	if len(params) != 2 {
		rex.ErrorParamCount("axis:read", "2")
	}
	
	ds, ok := params[0].Data.(axis.DataSource)
	if !ok {
		rex.ErrorGeneralCmd("axis:read", "Param 0 is not an axis.DataSource")
	}

	file, err := ds.ReadAll(params[1].String())
	if err != nil {
		script.Error = true
		script.RetVal = rex.NewValueString("error:" + err.Error())
		return
	}
	script.RetVal = rex.NewValueString(string(file))
}

// Write to a AXIS file.
// 	axis:write ds path content
// If the directories in the path do not exist axis:write tries to create them.
// Returns unchanged or an error message. May set the Error flag.
func Command_Write(script *rex.Script, params []*rex.Value) {
	if len(params) != 3 {
		rex.ErrorParamCount("axis:write", "3")
	}
	
	ds, ok := params[0].Data.(axis.DataSource)
	if !ok {
		rex.ErrorGeneralCmd("axis:write", "Param 0 is not an axis.DataSource")
	}

	err := ds.Create(params[1].String())
	if err != nil {
		script.Error = true
		script.RetVal = rex.NewValueString("error:" + err.Error())
		return
	}
	err = ds.WriteAll(params[1].String(), []byte(params[2].String()))
	if err != nil {
		script.Error = true
		script.RetVal = rex.NewValueString("error:" + err.Error())
		return
	}
}

// Does a AXIS file exist?
// 	axis:exists ds path
// Returns true or false.
func Command_Exists(script *rex.Script, params []*rex.Value) {
	if len(params) != 2 {
		rex.ErrorParamCount("axis:exists", "2")
	}
	
	ds, ok := params[0].Data.(axis.DataSource)
	if !ok {
		rex.ErrorGeneralCmd("axis:exists", "Param 0 is not an axis.DataSource")
	}

	script.RetVal = rex.NewValueBool(ds.Exists(params[1].String()))
}

// Is a AXIS file a directory?
// 	axis:isdir ds path
// Returns true or false.
func Command_IsDir(script *rex.Script, params []*rex.Value) {
	if len(params) != 2 {
		rex.ErrorParamCount("axis:isdir", "2")
	}
	
	ds, ok := params[0].Data.(axis.DataSource)
	if !ok {
		rex.ErrorGeneralCmd("axis:isdir", "Param 0 is not an axis.DataSource")
	}

	script.RetVal = rex.NewValueBool(ds.IsDir(params[1].String()))
}

// Delete a AXIS file or directory.
// 	axis:del ds path
// Returns true or false.
func Command_Del(script *rex.Script, params []*rex.Value) {
	if len(params) != 2 {
		rex.ErrorParamCount("axis:del", "2")
	}
	
	ds, ok := params[0].Data.(axis.DataSource)
	if !ok {
		rex.ErrorGeneralCmd("axis:del", "Param 0 is not an axis.DataSource")
	}

	err := ds.Delete(params[1].String())
	if err != nil {
		script.RetVal = rex.NewValueString(err.Error())
		script.Error = true
	}
}

// Iterate over all the directories in a AXIS directory.
// 	axis:walkdirs ds path code
// Runs code for every directory found, params:
//	path
// code MUST be a block created via a block declaration!
// Returns unchanged.
func Command_WalkDirs(script *rex.Script, params []*rex.Value) {
	if len(params) != 3 {
		rex.ErrorParamCount("axis:walkdirs", "3")
	}
	
	ds, ok := params[0].Data.(axis.DataSource)
	if !ok {
		rex.ErrorGeneralCmd("axis:walkdirs", "Param 0 is not an axis.DataSource")
	}

	if params[2].Type != rex.TypCode {
		rex.ErrorGeneralCmd("axis:walkdirs", "Attempt to run non-executable Value.")
	}
	block := params[2].Data.(*rex.Code)
	
	files := ds.ListDir(params[1].String())
	
	script.Locals.Add(block)
	for _, file := range files {
		script.Locals.Set(0, rex.NewValueString(file))
		script.Exec(block)
		script.Return = false
	}
	script.Locals.Remove()
	script.RetVal = rex.NewValue()
}

// Iterate over all the files in a AXIS directory.
// 	axis:walkfiles ds path code
// Runs code for every directory found, params:
//	path
// code MUST be a block created via a block declaration!
// Returns unchanged.
func Command_WalkFiles(script *rex.Script, params []*rex.Value) {
	if len(params) != 3 {
		rex.ErrorParamCount("axis:walkfiles", "3")
	}
	
	ds, ok := params[0].Data.(axis.DataSource)
	if !ok {
		rex.ErrorGeneralCmd("axis:walkfiles", "Param 0 is not an axis.DataSource")
	}

	if params[2].Type != rex.TypCode {
		rex.ErrorGeneralCmd("axis:walkfiles", "Attempt to run non-executable Value.")
	}
	block := params[2].Data.(*rex.Code)
	
	files := ds.ListFile(params[1].String())
	
	script.Locals.Add(block)
	for _, file := range files {
		script.Locals.Set(0, rex.NewValueString(file))
		script.Exec(block)
		script.Return = false
	}
	script.Locals.Remove()
	script.RetVal = rex.NewValue()
}
