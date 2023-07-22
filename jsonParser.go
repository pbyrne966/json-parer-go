package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// JSONValue represents a JSON value
type JSONValue struct {
	Type  string
	Value interface{}
}

// JSONParser represents the custom JSON parser
type JSONParser struct {
	input        *bytes.Buffer
	currentToken string
}

// NewJSONParser creates a new JSONParser instance
func NewJSONParser(input []byte) *JSONParser {
	return &JSONParser{input: bytes.NewBuffer(input), currentToken: ""}
}

// readNextToken reads the next JSON token from the input
func (p *JSONParser) readNextToken() bool {
	p.skipWhitespaces()

	if p.input.Len() == 0 {
		return false
	}

	// Read the next character
	currentChar := p.input.Next(1)

	// Check the type of the token
	switch currentChar[0] {
	case '{', '}', '[', ']', ':', ',':
		p.currentToken = string(currentChar)
	case 'n': // Check for null
		if p.input.Len() >= 4 && string(p.input.Next(4)) == "null"  {
			p.currentToken = "null"
		} else {
			return false
		}
	case 't': // Check for true
		if p.input.Len() >= 4 && string(p.input.Next(4)) == "true" {
			p.currentToken = "true"
		} else {
			return false
		}
	case 'f': // Check for false
		if p.input.Len() >= 5 && string(p.input.Next(5)) == "false" {
			p.currentToken = "false"
		} else {
			return false
		}
	case '"': // Check for string
		start := p.input.Len()
		for p.input.Len() > 0 {
			currentChar := p.input.Next(1)
			if currentChar[0] == '"' {
				p.currentToken = p.input.String()[start:p.input.Len()-1]
				break
			}
		}
	default: // Check for number
		start := p.input.Len()
		for p.input.Len() > 0 {
			currentChar := p.input.Next(1)
			if bytes.IndexByte([]byte("0123456789+-.eE"), currentChar[0]) < 0 {
				p.input.UnreadByte() // Unread the non-numeric character
				p.currentToken = p.input.String()[start:p.input.Len()]
				break
			}
		}
	}

	return true
}

// skipWhitespaces skips whitespaces in the input buffer
func (p *JSONParser) skipWhitespaces() {
	for p.input.Len() > 0 {
		currentChar := p.input.Next(1)
		if currentChar[0] != ' ' && currentChar[0] != '\n' && currentChar[0] != '\r' && currentChar[0] != '\t' {
			p.input.UnreadByte() // Unread the non-whitespace character
			break
		}
	}
}

// parseValue parses a JSON value
func (p *JSONParser) parseValue() JSONValue {
	p.readNextToken()

	switch p.currentToken {
	case "null", "true", "false":
		return JSONValue{Type: p.currentToken, Value: nil}
	case "{":
		return p.parseObject()
	case "[":
		return p.parseArray()
	default:
		// Check if the token is a number
		if _, err := json.Number(p.currentToken).Float64(); err == nil {
			// Convert the number to float64
			num, _ := json.Number(p.currentToken).Float64()
			return JSONValue{Type: "number", Value: num}
		}

		// Otherwise, it must be a string
		return JSONValue{Type: "string", Value: p.currentToken}
	}
}

// parseObject parses a JSON object
func (p *JSONParser) parseObject() JSONValue {
	object := make(map[string]JSONValue)

	p.readNextToken()
	for p.currentToken != "}" {
		key := p.currentToken

		// Read the ':' separator
		p.readNextToken()
		if p.currentToken != ":" {
			panic("Expected ':'")
		}

		// Parse the value and add it to the object
		p.readNextToken()
		value := p.parseValue()
		object[key] = value

		// Read the next token (',' or '}')
		p.readNextToken()
	}

	return JSONValue{Type: "object", Value: object}
}

// parseArray parses a JSON array
func (p *JSONParser) parseArray() JSONValue {
	array := make([]JSONValue, 0)

	p.readNextToken()
	for p.currentToken != "}" {
		// Parse the value and add it to the array
		value := p.parseValue()
		array = append(array, value)

		// Read the next token (',' or ']')
		p.readNextToken()
	}

	return JSONValue{Type: "array", Value: array}
}

func main() {
	// Example JSON data
	jsonData := []byte(`{
		"name": "John Doe",
		"age": 30,
		"email": "john.doe@example.com",
		"active": true,
		"address": {
			"city": "New York",
			"zip": "10001"
		},
		"tags": ["golang", "json", "parser", [1, 2, 3]]
	}`)

	// Create a new JSONParser instance
	parser := NewJSONParser(jsonData)

	// Parse the JSON data
	result := parser.parseValue()

	// Print the parsed JSON data
	fmt.Printf("%+v\n", result)
}
