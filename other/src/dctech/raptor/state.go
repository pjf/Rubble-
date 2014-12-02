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

package raptor

import "fmt"
import "io"
import "os"
import "regexp"
import "strings"
import "errors"

// State handles all global script data and settings.
// The majority of the fields are exported for the use of commands only.
type State struct {
	Commands   *CommandStore
	NameSpaces *NameSpaceStore
	Types      *TypeStore
	NoRecover  bool         // Do not recover errors, this makes it easier to debug the parser.
	Debug      []DbgHandler // This stores the debug handlers, normally nil unless a debugger is installed.
	Output     io.Writer    // Normally set to os.Stdout, this can be changed to redirect to a log file or the like.
}

// NewState creates (and initializes) a new state.
func NewState() *State {
	rtn := new(State)
	rtn.Commands = NewCommandStore()
	rtn.NameSpaces = NewNameSpaceStore()
	rtn.Types = NewTypeStore()
	rtn.Debug = nil
	rtn.Output = os.Stdout
	return rtn
}

// Variables

// NewNamespacedVar creates a new script variable in a namespace.
// For use in global initialization.
func (this *State) NewNamespacedVar(name string, value *Value) {
	space, itemname := this.ParseName(name)
	if space == nil {
		panic(fmt.Sprintf("Variable: \"%v\" not in a namespace", name))
	}
	if space.Vars.Exist(itemname) {
		panic(fmt.Sprintf("Variable: \"%v\" already declared", name))
	}
	space.Vars.Store(itemname, value)
}

// Namespaces

// NewNameSpace creates a new namespace.
func (this *State) NewNameSpace(name string) {
	space, itemname := this.ParseName(name)
	
	if space != nil {
		if space.NameSpaces.Exist(itemname) {
			panic(fmt.Sprintf("Namespace: \"%v\" already declared", name))
		}
		space.NameSpaces.Store(itemname, NewNameSpace())
		return
	}
	
	if this.NameSpaces.Exist(itemname) {
		panic(fmt.Sprintf("Namespace: \"%v\" already declared", name))
	}
	this.NameSpaces.Store(itemname, NewNameSpace())
}

// DeleteNameSpace deletes a namespace.
func (this *State) DeleteNameSpace(name string) {
	space, itemname := this.ParseName(name)

	if space != nil {
		space.NameSpaces.Delete(itemname)
		return
	}
	this.NameSpaces.Delete(itemname)
}

// Commands

// NewNativeCommand adds a new native command.
func (this *State) NewNativeCommand(name string, handler NativeCommand) {
	space, itemname := this.ParseName(name)
	
	rtn := new(Command)
	rtn.Native = true
	rtn.Handler = handler
	
	if space != nil {
		space.Commands.Store(itemname, rtn)
		return
	}
	this.Commands.Store(itemname, rtn)
}

// NewUserCommand adds a new user command (what else would it do?).
func (this *State) NewUserCommand(name string, code *Value, params []*Value) {
	rtn := new(Command)

	rtn.Code = code.CompiledScript()

	if params == nil {
		rtn.VarParams = true
	} else {
		for _, val := range params {
			rtn.Params = append(rtn.Params, val.String())
		}
	}

	space, itemname := this.ParseName(name)
	if space != nil {
		space.Commands.Store(itemname, rtn)
		return
	}
	this.Commands.Store(itemname, rtn)
}

// GetCommand fetches a command by it's name.
func (this *State) GetCommand(name string) *Command {
	space, itemname := this.ParseName(name)
	if space != nil {
		if space.Commands.Exist(itemname) {
			return space.Commands.Fetch(itemname)
		}
		panic("Undeclared command: " + name)
	}

	if this.Commands.Exist(itemname)  {
		return this.Commands.Fetch(itemname)
	}
	panic("Undeclared command: " + name)
}

// DeleteCommand removes a command.
func (this *State) DeleteCommand(name string) {
	space, itemname := this.ParseName(name)
	
	if space != nil {
		space.Commands.Delete(itemname)
		return
	}
	this.Commands.Delete(itemname)
}

// Indexables

// RegisterType registers a new Indexable or the like with a specific name.
// Registering a type allows it to be created with an object literal.
func (this *State) RegisterType(name string, handler ObjectFactory) {
	space, itemname := this.ParseName(name)
	
	if space != nil {
		space.Types.Store(itemname, handler)
		return
	}
	this.Types.Store(itemname, handler)
}

// GetType retrieves a named types ObjectFactory.
func (this *State) GetType(name string) ObjectFactory {
	space, itemname := this.ParseName(name)
	if space != nil {
		if space.Types.Exist(itemname) {
			return space.Types.Fetch(itemname)
		}
		panic("Undeclared type: " + name)
	}

	if this.Types.Exist(itemname)  {
		return this.Types.Fetch(itemname)
	}
	panic("Undeclared type: " + name)
}

// Name Parsing

// (namespace:namespace:)(name)
var nameSplitRegex = regexp.MustCompile("^((?:[^:]+:)*)([^:.]+)$")

// ParseName parses a name and returns a namespace or nil and the base name.
// Eg. "test:object" will parse to a pointer to the namespace "test", and the string "object".
func (this *State) ParseName(name string) (*NameSpace, string) {
	if !strings.Contains(name, ":") {
		return nil, name
	}

	namesplit := nameSplitRegex.FindStringSubmatch(name)
	namespacelist := strings.TrimRight(namesplit[1], ":")
	basename := namesplit[2]
	namespace := this.ParseNameSpaceName(namespacelist)
	return namespace, basename
}

// ParseNameSpaceName is just like ParseName but for just the name of a namespace.
func (this *State) ParseNameSpaceName(name string) *NameSpace {
	if !strings.Contains(name, ":") {
		if !this.NameSpaces.Exist(name) {
			panic("Undeclared Namespace: " + name)
		}
		return this.NameSpaces.Fetch(name)
	}

	names := strings.Split(name, ":")
	
	if !this.NameSpaces.Exist(names[0]) {
		panic("Undeclared Namespace: " + name)
	}
	return this.fetchNameSpaceName(names[1:], this.NameSpaces.Fetch(names[0]))
}

func (this *State) fetchNameSpaceName(names []string, namespace *NameSpace) *NameSpace {
	if len(names) == 1 {
		if !namespace.NameSpaces.Exist(names[0]) {
			panic(fmt.Sprintf("Undeclared Namespace: \"%v\"", names[0]))
		}
		return namespace.NameSpaces.Fetch(names[0])
	}
	return this.fetchNameSpaceName(names[1:], namespace.NameSpaces.Fetch(names[0]))
}

// Debugger

func (this *State) RegisterDbgCallback(typ int, handler DbgHandler) {
	if this.Debug == nil {
		this.Debug = make([]DbgHandler, dbgrMaxType)
	}
	if typ >= dbgrMaxType || typ < 0 {
		panic("Callback type out of range.")
	}
	this.Debug[typ] = handler
}

func (this *State) DbgCallback(typ int) {
	if this.Debug == nil {
		return
	}
	if typ >= dbgrMaxType || typ < 0 {
		panic("Callback type out of range.")
	}
    
	if this.Debug[typ] == nil {
		return
	}
	this.Debug[typ](this)
}

// Output

func (this *State) Printf(format string, msg ...interface{}) {
	fmt.Fprintf(this.Output, format, msg...)
}

func (this *State) Println(msg ...interface{}) {
	fmt.Fprintln(this.Output, msg...)
}

func (this *State) Print(msg ...interface{}) {
	fmt.Fprint(this.Output, msg...)
}

// Script Execution

// RunCommand runs the specified Raptor command.
// You must NEVER, NEVER, NEVER nest calls to RunCommand unless you make the nested calls with a new Script.
func (this *State) RunCommand(script *Script, command string, params ...*Value) (ret *Value, err error) {
	// Most of this function is identical to Run
	// It's just easier to copy than refactor
	
	script.Host = this
	
	err = nil

	defer func() {
		ret = script.RetVal
		
		if !this.NoRecover {
			if x := recover(); x != nil {
				switch i := x.(type) {
				case error:
					err = i
				case string:
					err = fmt.Errorf("%v Near %v", i, script.Code.Last().Position())
				default:
					err = errors.New(fmt.Sprint(i))
				}
			}
		}
		
		// Normal Cleanup
		script.Code.Clear()
		script.Envs.Clear()
		script.Host = nil
		//script.RetVal = nil
		script.This = nil
		script.Exit = false
		script.Return = false
		script.Break = false
		script.BreakLoop = false
	}()

	this.GetCommand(command).Call(script, params)
	return // Required :(
}

// Run executes a Raptor script.
// You must NEVER, NEVER, NEVER nest calls to Run unless you make the nested calls with a new Script.
func (this *State) Run(script *Script) (ret *Value, err error) {
	// Most of this function is identical to RunCommand
	// It's just easier to copy than refactor
	
	script.Host = this
	
	err = nil

	defer func() {
		ret = script.RetVal
		
		if !this.NoRecover {
			if x := recover(); x != nil {
				switch i := x.(type) {
				case error:
					err = i
				case string:
					err = fmt.Errorf("%v Near %v", i, script.Code.Last().Position())
				default:
					err = errors.New(fmt.Sprint(i))
				}
			}
		}
		
		// Normal Cleanup
		script.Code.Clear()
		script.Envs.Clear()
		script.Host = nil
		//script.RetVal = nil
		script.This = nil
		script.Exit = false
		script.Return = false
		script.Break = false
		script.BreakLoop = false
	}()

	script.Exec()
	return // Required :(
}
