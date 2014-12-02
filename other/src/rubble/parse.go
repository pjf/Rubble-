/*
Copyright 2013 by Milo Christiansen

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

package main

import "fmt"
import "strings"
import "io/ioutil"
import "regexp"

func ReadConfig(path string) {
	fmt.Println("Reading Config File:", path)
	file, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error:", err)
		panic(err)
	}
	
	lines := strings.Split(string(file), "\n")
	for i := range lines {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			continue
		}
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		
		parts := strings.SplitN(lines[i], "=", 2)
		if len(parts) != 2 {
			panic("Malformed config line.")
		}
		
		parts[0] = strings.TrimSpace(parts[0])
		VariableData[parts[0]] = strings.TrimSpace(parts[1])
	}
}

// This is the stage parser
func StageParse(input string) string {
	if ParseStage == 0 {
		return PreParse(input)
	} else if ParseStage == 1 {
		return Parse(input)
	} else if ParseStage == 2 {
		return PostParse(input)
	}
	panic("Invalid ParseStage")
}

func PreParse(input string) string {
	out := ""
	lex := NewLexer(input)
	
	for {
		lex.Advance()
		if lex.Current.Type == tknString {
			out += lex.Current.Lexeme
		} else if lex.Current.Type == tknTagBegin {
			lex.GetToken(tknString)
			if lex.Current.Lexeme[0] != '!' {
				// Not a pre tag, copy over until we get a tag end
				out += "{" + lex.Current.Lexeme
				for lex.Advance() {
					if lex.Current.Type == tknTagEnd {
						out += lex.Current.Lexeme
						break
					}
					out += lex.Current.Lexeme
				}
				continue
			}
			name := lex.Current.Lexeme
			params := make([]string, 0, 5)
			for lex.CheckLookAhead(tknDelimiter) {
				lex.GetToken(tknDelimiter)
				lex.GetToken(tknString)
				params = append(params, lex.Current.Lexeme)
			}
			lex.GetToken(tknTagEnd)
			
			if _, ok := Templates[name]; !ok {
				panic("Invalid template: " + name)
			}
			out += Templates[name].Call(params)
		} else if lex.Current.Type == tknINVALID {
			break
		}
	}
	
	return out
}

func Parse(input string) string {
	out := ""
	lex := NewLexer(input)
	
	for {
		lex.Advance()
		if lex.Current.Type == tknString {
			out += lex.Current.Lexeme
		} else if lex.Current.Type == tknTagBegin {
			lex.GetToken(tknString)
			if lex.Current.Lexeme[0] == '#' {
				// Post tag, copy over until we get a tag end
				out += "{" + lex.Current.Lexeme
				for lex.Advance() {
					if lex.Current.Type == tknTagEnd {
						out += lex.Current.Lexeme
						break
					}
					out += lex.Current.Lexeme
				}
				continue
			}
			name := lex.Current.Lexeme
			params := make([]string, 0, 5)
			for lex.CheckLookAhead(tknDelimiter) {
				lex.GetToken(tknDelimiter)
				lex.GetToken(tknString)
				params = append(params, lex.Current.Lexeme)
			}
			lex.GetToken(tknTagEnd)
			
			if _, ok := Templates[name]; !ok {
				panic("Invalid template: " + name)
			}
			out += Templates[name].Call(params)
		} else if lex.Current.Type == tknINVALID {
			break
		}
	}
	
	return out
}

func PostParse(input string) string {
	out := ""
	lex := NewLexer(input)
	
	for {
		lex.Advance()
		if lex.Current.Type == tknString {
			out += lex.Current.Lexeme
		} else if lex.Current.Type == tknTagBegin {
			lex.GetToken(tknString)
			if lex.Current.Lexeme[0] != '#' {
				// Not a post tag, copy over until we get a tag end
				out += "{" + lex.Current.Lexeme
				for lex.Advance() {
					if lex.Current.Type == tknTagEnd {
						out += lex.Current.Lexeme
						break
					}
					out += lex.Current.Lexeme
				}
				continue
			}
			name := lex.Current.Lexeme
			params := make([]string, 0, 5)
			for lex.CheckLookAhead(tknDelimiter) {
				lex.GetToken(tknDelimiter)
				lex.GetToken(tknString)
				params = append(params, lex.Current.Lexeme)
			}
			lex.GetToken(tknTagEnd)
			
			if _, ok := Templates[name]; !ok {
				panic("Invalid template: " + name)
			}
			out += Templates[name].Call(params)
		} else if lex.Current.Type == tknINVALID {
			break
		}
	}
	
	return out
}

var varNameSimpleRegEx = regexp.MustCompile("\\$[a-zA-Z_][a-zA-Z0-9_]*")

// This is a modified version of os.Expand
func ExpandVars(input string) string {
	buf := make([]byte, 0, len(input))
	
	depth := 0
	i := 0
	for j := 0; j < len(input); j++ {
		if input[j] == '{' {
			depth++
		}
		if input[j] == '}' && depth > 0 {
			depth--
		}
		
		if input[j] == '$' && j+1 < len(input) && depth == 0 {
			buf = append(buf, input[i:j]...)
			name, w := getVarName(input[j+1:])
			if name == "" {
				buf = append(buf, '{')
			} else {
				buf = append(buf, VariableData[name]...)
			}
			j += w
			i = j + 1
		}
	}
	
	return string(buf) + input[i:]
}

func getVarName(input string) (string, int) {
	if input[0] == '{' {
		// Scan to closing brace
		for i := 1; i < len(input); i++ {
			if input[i] == '}' {
				return input[1:i], i + 1
			}
		}
		return "", 1 // Bad syntax
	}
	// Scan alphanumerics.
	var i int
	for i = 0; i < len(input) && isAlphaNum(input[i]); i++ {
	}
	return input[:i], i
}

func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}