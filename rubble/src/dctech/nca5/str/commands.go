// NCA v5 String Commands.
package str

import "dctech/nca5"
//import "fmt"
import "strconv"
import "strings"

// Adds the string commands to the state.
// The string commands are:
//	append
//	strlen
//	strchar
//	fmtstr
func Setup(state *nca5.State) {
	state.NewNativeCommand("append", CommandAppend)
	state.NewNativeCommand("trimspace", CommandTrimSpace)
	state.NewNativeCommand("strlen", CommandStrLen)
	state.NewNativeCommand("strchar", CommandStrChar)
	state.NewNativeCommand("fmtstr", CommandFmtStr)
}

// Appends two or more strings together.
// 	append a b [c...]
// Returns "0" or "-1"
func CommandAppend(state *nca5.State, params []*nca5.Value) {
	if len(params) <= 1 {
		panic("append needs at least two params.")
	}

	result := ""
	for _, val := range params {
		result += val.String()
	}

	state.RetVal = nca5.NewValue(result)
	return
}

// Trims leading and trailing whitespace from a string.
// 	trimspace str
// Returns str with leading and trailing whitespace removed.
func CommandTrimSpace(state *nca5.State, params []*nca5.Value) {
	if len(params) != 1 {
		panic("Wrong param count to trimspace.")
	}

	state.RetVal = nca5.NewValue(strings.TrimSpace(params[0].String()))
	return
}

// Gets the length of a string.
// 	strlen a
// Returns the length
func CommandStrLen(state *nca5.State, params []*nca5.Value) {
	if len(params) != 1 {
		panic("Wrong param count to strlen.")
	}

	state.RetVal = nca5.NewValueFromI64(int64(len(params[0].String())))
	return
}

// Gets char at pos.
// 	strchar a pos
// Opperand pos is converted to a 64 bit integer. Invalid strings are given the value "0"
// If the position is out of range returns unchanged and sets the error flag.
// Returns the character
func CommandStrChar(state *nca5.State, params []*nca5.Value) {
	if len(params) != 2 {
		panic("Wrong param count to strchar.")
	}

	pos := params[1].Int64()
	if pos >= int64(len(params[0].String())) {
		state.Error = true
		return
	}
	
	state.RetVal = nca5.NewValue(string(params[0].String()[pos]))
	return
}

// Formats a string.
// 	fmtstr format values...
// valid format "verbs" for fmtstr:
//	%% literal percent
//	%s the raw string
//	%d as a decimal number
//	%x as a lowercase hexadecimal number
//	%X as an upercase hexadecimal number
// Returns the formated string
func CommandFmtStr(state *nca5.State, params []*nca5.Value) {
	if len(params) < 2 {
		panic("Wrong param count to fmtstr.")
	}
	
	paramcount := len(params)-1
	escapecount := 0
	escape := false
	output := make([]rune, 0, 100)
	for _, val := range params[0].String() {
		if val == '%' && escape {
			output = append(output, val)
			escape = false
			continue
		}
		if val == '%' && !escape {
			escape = true
			continue
		}
		
		if escape && val == 's' {
			escapecount++
			if paramcount < escapecount {
				panic("More format codes than params to fmtstr.")
			}
			output = append(output, []rune(params[escapecount].String())...)
			escape = false
			continue
		}
		
		if escape && val == 'd' {
			escapecount++
			if paramcount < escapecount {
				panic("More format codes than params to fmtstr.")
			}
			temp := strconv.FormatInt(params[escapecount].Int64(), 10)
			output = append(output, []rune(temp)...)
			escape = false
			continue
		}
		
		if escape && val == 'x' {
			escapecount++
			if paramcount < escapecount {
				panic("More format codes than params to fmtstr.")
			}
			temp := strconv.FormatInt(params[escapecount].Int64(), 16)
			//output = append(output, '0', 'x')
			output = append(output, []rune(temp)...)
			escape = false
			continue
		}
		
		if escape && val == 'X' {
			escapecount++
			if paramcount < escapecount {
				panic("More format codes than params to fmtstr.")
			}
			temp := strconv.FormatInt(params[escapecount].Int64(), 16)
			//output = append(output, '0', 'X')
			output = append(output, []rune(strings.ToUpper(temp))...)
			escape = false
			continue
		}
		
		if escape {
			panic("Invalid format code: %" + string(val) + " to fmtstr.")
		}
		
		output = append(output, val)
	}
	
	state.RetVal = nca5.NewValue(string(output))
	return
}