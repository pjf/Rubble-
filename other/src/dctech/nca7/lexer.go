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

package nca7

//import "fmt"
import "strconv"

// Lexer states
const (
	stReset = iota
	stEatComment
	stReadDQStr
	stReadCBStr
	stReadStr
)

// Token values
const (
	TknINVALID = iota
	TknString
	TknCmdBegin // (
	TknCmdEnd // )
	TknDerefBegin // [
	TknDerefEnd // ]
	TknObjLitBegin // <
	TknObjLitEnd // >
	TknObjLitSplit // =
)

// Lexer is a NCA lexer reading from a string.
type Lexer struct {
	look    *Token
	current *Token
	
	source string
	
	line int
	column int
	
	index int
	state int
	
	lexeme []byte
	
	token int
	tokenline int
	tokencolumn int
	
	strdepth int
	objdepth int
}

// Returns a new Lexer.
func NewLexer(input string, startline, startcolumn int) *Lexer {
	this := new(Lexer)
	
	this.source = input
	
	if startline < 0 {
		startline = -1
		startcolumn = -1
	}
	this.line = startline
	this.column = startcolumn
	
	this.index = 0
	this.state = stReset
	
	this.lexeme = make([]byte, 0, 20)
	
	this.token = TknINVALID
	this.tokenline = startline
	this.tokencolumn = startcolumn
	
	this.strdepth = 0
	this.objdepth = 0
	
	this.Advance()
	this.Advance()
	
	return this
}

func (this *Lexer) Advance() {
	
	if this.index > len(this.source) {
		this.current = this.look
		this.look = &Token{"INVALID", TknINVALID, this.tokenline, this.tokencolumn}
		return
	}
	
	for ; this.index < len(this.source); this.index++ {
		dat := this.source
		i := this.index
		lookok := len(this.source) - this.index
		
		if dat[i] == '\n' && this.line >= 0 {
			this.line++
			this.column = 0
		} else {
			this.column++
		}
		
		// Lexing Begin
		//======================================
		
		// Escape
		if dat[i] == '\\' && this.state == stReadDQStr {
			if lookok < 1 {
				panic("Unexpected end of stream.")
			}
			
			switch dat[i+1] {
			case 'n':
				this.lexeme = append(this.lexeme, '\n')
			case 'r':
				this.lexeme = append(this.lexeme, '\r')
			case 't':
				this.lexeme = append(this.lexeme, '\t')
			case '\'':
				this.lexeme = append(this.lexeme, '\'')
			case '"':
				this.lexeme = append(this.lexeme, '"')
			case '\\':
				this.lexeme = append(this.lexeme, '\\')
			case 'x':
				if lookok < 3 {
					panic("Unexpected end of stream.")
				}
				rep, err := strconv.ParseInt(string([]byte { dat[i+2], dat[i+3] } ), 16, 8)
				if err != nil {
					panic("Invalid escape sequence: \\x" + string(dat[i+2]) + string(dat[i+3]) + ".")
				}
				this.lexeme = append(this.lexeme, byte(rep))
				this.index += 2
				
			default:
				panic("Invalid escape sequence: \\" + string(dat[i+1]) + ".")
			}
			this.index++
			continue
		}
		
		// Double Quote Strings
		if this.state != stReadCBStr && this.state != stEatComment {
			if dat[i] == '"' {
				if this.state == stReadDQStr {
					this.state = stReset
					continue
				}
				this.current = this.look
				this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
				this.state = stReadDQStr
				this.token = TknString
				this.tokenline = this.line
				this.tokencolumn = this.column
				this.lexeme = this.lexeme[0:0]
				this.index++
				return
			}
			if this.state == stReadDQStr {
				this.lexeme = append(this.lexeme, dat[i])
				continue
			}
		}
		
		// Curly bracket Strings
		if this.state != stReadDQStr && this.state != stEatComment {
			if dat[i] == '{' {
				this.strdepth++
				if this.state == stReadCBStr {
					this.lexeme = append(this.lexeme, dat[i])
					continue
				}
				this.current = this.look
				this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
				this.state = stReadCBStr
				this.token = TknString
				this.tokenline = this.line
				this.tokencolumn = this.column
				this.lexeme = this.lexeme[0:0]
				this.index++
				return
			}
			if dat[i] == '}' {
				this.strdepth--
				if this.strdepth == 0 {
					this.state = stReset
					continue
				}
				this.lexeme = append(this.lexeme, dat[i])
				continue
			}
			if this.state == stReadCBStr {
				this.lexeme = append(this.lexeme, dat[i])
				continue
			}
		}
		
		// Comments
		if dat[i] == '\'' {
			if this.state == stEatComment {
				this.state = stReset
				continue
			}
			this.state = stEatComment
			continue
		}
		if this.state == stEatComment {
			continue
		}

		// Delimiters
		if dat[i] == '\n' || dat[i] == '\r' || dat[i] == ' ' || dat[i] == '\t' || dat[i] == ',' {
			this.state = stReset
			continue
		}
		
		// Parentheses
		if dat[i] == '(' {
			this.current = this.look
			this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
			this.state = stReset
			this.token = TknCmdBegin
			this.tokenline = this.line
			this.tokencolumn = this.column
			this.lexeme = this.lexeme[0:1]
			this.lexeme[0] = '('
			this.index++
			return
		}
		if dat[i] == ')' {
			this.current = this.look
			this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
			this.state = stReset
			this.token = TknCmdEnd
			this.tokenline = this.line
			this.tokencolumn = this.column
			this.lexeme = this.lexeme[0:1]
			this.lexeme[0] = ')'
			this.index++
			return
		}

		// Square Brackets
		if dat[i] == '[' {
			this.current = this.look
			this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
			this.state = stReset
			this.token = TknDerefBegin
			this.tokenline = this.line
			this.tokencolumn = this.column
			this.lexeme = this.lexeme[0:1]
			this.lexeme[0] = '['
			this.index++
			return
		}
		if dat[i] == ']' {
			this.current = this.look
			this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
			this.state = stReset
			this.token = TknDerefEnd
			this.tokenline = this.line
			this.tokencolumn = this.column
			this.lexeme = this.lexeme[0:1]
			this.lexeme[0] = ']'
			this.index++
			return
		}
		
		// Angle Brackets
		// No need to check the state, all the delimited tokens are already handled at this point
		if dat[i] == '<' { 
			this.objdepth++
			this.current = this.look
			this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
			this.state = stReset
			this.token = TknObjLitBegin
			this.tokenline = this.line
			this.tokencolumn = this.column
			this.lexeme = this.lexeme[0:1]
			this.lexeme[0] = '<'
			this.index++
			return
		}
		if dat[i] == '>' {
			this.objdepth--
			this.current = this.look
			this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
			this.state = stReset
			this.token = TknObjLitEnd
			this.tokenline = this.line
			this.tokencolumn = this.column
			this.lexeme = this.lexeme[0:1]
			this.lexeme[0] = '>'
			this.index++
			return
		}
		
		if dat[i] == '=' && this.objdepth > 0 {
			this.current = this.look
			this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
			this.state = stReset
			this.token = TknObjLitSplit
			this.tokenline = this.line
			this.tokencolumn = this.column
			this.lexeme = this.lexeme[0:1]
			this.lexeme[0] = '='
			this.index++
			return
		}

		// Raw Strings
		if this.state == stReadStr {
			this.lexeme = append(this.lexeme, dat[i])
			continue
		}
		
		this.current = this.look
		this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
		this.state = stReadStr
		this.token = TknString
		this.tokenline = this.line
		this.tokencolumn = this.column
		this.lexeme = this.lexeme[0:1]
		this.lexeme[0] = dat[i]
		this.index++
		return
	}
	
	if this.index == len(this.source) {
		this.current = this.look
		this.look = &Token{string(this.lexeme), this.token, this.tokenline, this.tokencolumn}
		this.index++
		return
	}
}

// CurrentTkn returns the current token.
func (this *Lexer) CurrentTkn() *Token {
	return this.current
}

// LookAhead returns the lookahead token.
func (this *Lexer) LookAhead() *Token {
	return this.look
}

// Line returns the line of the last token fetched.
func (this *Lexer) Line() int {
	return this.tokenline
}

// Column returns the column of the last token fetched.
func (this *Lexer) Column() int {
	return this.tokencolumn
}

// Gets the next token, and panics with an error if it's not of type tokenType.
// May cause a panic if the lexer encounters an error
// Used as a type checked Advance
func (this *Lexer) GetToken(tokenTypes ...int) {
	this.Advance()
	
	for _, val := range tokenTypes {
		if this.current.Type == val {
			return
		}
	}

	ExitOnTokenExpected(this.current, tokenTypes...)
}

// Checks to see if the look ahead is one of tokenTypes and if so returns true
func (this *Lexer) CheckLookAhead(tokenTypes ...int) bool {
	for _, val := range tokenTypes {
		if this.look.Type == val {
			return true
		}
	}
	return false
}
