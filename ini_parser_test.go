package ssm_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	sm "git.sr.ht/~mariusor/ssm"
)

// ip holds the INI parser state
type ip struct {
	// source is a byte scanner implementation that allows us to move in the
	// string one byte forwards and backwards.
	source io.ByteScanner
	// read is the counter of the read bytes. It's mostly used for error reporting.
	read int64
	// line is the counter for the number of parsed lines in the string.
	line int64
	// char is the counter for the number of parsed characters in the current line.
	char int64
}

// error signals the state machine to exit with error.
func (i *ip) error(err error) sm.Fn {
	return sm.ErrorEnd(fmt.Errorf("parse error at line %d pos %d: %w", i.line, i.char-1, err))
}

// endLine signals an end of line character, it moves the state machine to
// the next state, which is the parseLine state.
func (i *ip) endLine(_ context.Context) sm.Fn {
	i.char = 0
	i.line++
	return i.parseLine
}

// parseKeyValuePair parses a line as a key/value pair
// At first it reads the key name, when '=' is encountered it reads the value.
// When it reaches the end of the line, it moves to the endLine state.
func (i *ip) parseKeyValuePair(_ context.Context) sm.Fn {
	// We need to unwind the source tokenizer for the character we read to accumulate all the bytes in the kv.
	i.unreadCurrent()

	key := new(strings.Builder)
	val := new(strings.Builder)

	target := key
	for {
		buff, errState := i.readNextOrError()
		if errState != nil {
			return errState
		}

		switch {
		case buff == '=':
			target = val
		case buff == '\n' && target == key:
			return i.error(fmt.Errorf("invalid character in key name: %q", '\n'))
		case buff == ' ' && target == key:
			return i.error(fmt.Errorf("invalid character in key name: %q", ' '))
		case buff == '\n':
			fmt.Printf("Key: %s\nValue: %s", key.String(), val.String())
			return i.endLine
		case utf8.ValidRune(rune(buff)):
			target.WriteRune(rune(buff))
		default:
			break
		}
	}
	return i.error(fmt.Errorf("unkown error when parsing key value"))
}

// parseComment parses a line as a comment.
// When it reaches the end of the line, it moves to the endLine state
func (i *ip) parseComment(_ context.Context) sm.Fn {
	comment := new(strings.Builder)

	for {
		buff, errState := i.readNextOrError()
		if errState != nil {
			return errState
		}

		switch {
		case buff == '\n':
			fmt.Printf("start comment: %s // end comment", comment.String())
			return i.endLine
		case utf8.ValidRune(rune(buff)):
			comment.WriteRune(rune(buff))
		default:
			break
		}
	}
	return i.error(fmt.Errorf("unkown error when parsing comment"))
}

// parseGroup parses a line as a group name.
// When it reaches the symbol that marks the end of the name, it stops accumulating the name.
// When it reaches the end of the line it moves to the endLine state.
func (i *ip) parseGroup(_ context.Context) sm.Fn {
	group := new(strings.Builder)
	for {
		buff, errState := i.readNextOrError()
		if errState != nil {
			return errState
		}

		switch {
		case buff == ' ':
			return i.error(fmt.Errorf("invalid character in group name: %q", ' '))
		case buff == ']':
			c, _ := fmt.Printf("Group: %s\n", group.String())
			fmt.Printf("%s", strings.Repeat("-", c-1))
		case buff == '\n':
			return i.endLine
		case utf8.ValidRune(rune(buff)):
			group.WriteRune(rune(buff))
		default:
			break
		}
	}
	return i.error(fmt.Errorf("unkown error when parsing group name"))
}

func (i *ip) parseEnd(_ context.Context) sm.Fn {
	return sm.End
}

func (i *ip) readNextOrError() (byte, sm.Fn) {
	buff, err := i.source.ReadByte()
	if err != nil {
		if err != io.EOF {
			return buff, i.error(fmt.Errorf("parse error: %w", err))
		}
		return buff, i.parseEnd
	}
	i.read++
	i.char++

	return buff, sm.End
}

func (i *ip) unreadCurrent() {
	i.read--
	i.char--
	i.source.UnreadByte()
}

func (i *ip) parseLine(_ context.Context) sm.Fn {
	var next sm.Fn

	buff, errState := i.readNextOrError()
	if errState != nil {
		return errState
	}

	switch {
	case buff == ';':
		next = i.parseComment
		fmt.Println()
	case buff == '[':
		next = i.parseGroup
		fmt.Println()
	case buff == '\n':
		next = i.endLine
	case utf8.ValidRune(rune(buff)):
		next = i.parseKeyValuePair
		fmt.Println()
	default:
		break
	}

	return next
}

// parse starts the parsing by moving to the parseLine state
func (i *ip) parse(_ context.Context) sm.Fn {
	return i.parseLine
}

// ParseINI parses the content of the s source.
// It initializes the state based ini parser and runs its first state: parse
func ParseINI(ctx context.Context, s io.ByteScanner) error {
	i := ip{source: s}

	return sm.Run(ctx, i.parse)
}

const testINI = `
; comment
[group-name]
first-item=first value
second=0.555
`

func ExampleParseINI() {
	s := bytes.NewReader([]byte(testINI))
	err := ParseINI(context.Background(), s)

	// No parsing errors for the basic testINI example
	fmt.Printf("\nreturn error: %v", err)

	// Output:
	// start comment:  comment // end comment
	// Group: group-name
	// -----------------
	// Key: first-item
	// Value: first value
	// Key: second
	// Value: 0.555
	// return error: <nil>
}
