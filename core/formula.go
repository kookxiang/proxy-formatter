package core

import (
	"bytes"
	"fmt"
	"io"
)

type Formula struct {
	Name    string
	actions []Action
}

func (formula *Formula) Parse(data []byte) error {
	scanner := &FormulaScanner{
		data: data,
	}

	for scanner.HasNext() {
		if action, err := scanner.ScanAction(); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("formula parse failed: %w at line %d", err, scanner.Line)
		} else {
			formula.actions = append(formula.actions, action)
		}
	}
	return nil
}

func (formula *Formula) Execute(ctx *ExecuteContext) error {
	for _, action := range formula.actions {
		if err := action.Execute(ctx); err != nil {
			return err
		}
	}
	return nil
}

type FormulaScanner struct {
	Line    int
	data    []byte
	current int
}

func (s *FormulaScanner) HasNext() bool {
	return s.current < len(s.data)
}

func (s *FormulaScanner) Next() byte {
	s.current++
	if s.Current() == '\n' {
		s.Line++
	}
	return s.Current()
}

func (s *FormulaScanner) Current() byte {
	return s.data[s.current-1]
}

func (s *FormulaScanner) PreviewNext() byte {
	return s.data[s.current]
}

func (s *FormulaScanner) ScanAction() (Action, error) {
	buf := &bytes.Buffer{}
	isLeadingSpaces := true
	params := make([]string, 0)
	for s.HasNext() {
		char := s.Next()
		if isLeadingSpaces {
			if char == ' ' {
				continue
			} else {
				isLeadingSpaces = false
			}
		}
		if char == '\n' && buf.Len() == 0 {
			// ignore empty lines
			continue
		} else if char == '\n' {
			params = append(params, buf.String())
			return CreateAction(params)
		} else if char == '#' && buf.Len() == 0 {
			s.ScanLineComment()
			continue
		} else if char == ' ' {
			params = append(params, buf.String())
			for s.Current() != '\n' {
				params = append(params, s.ScanParams())
			}
			return CreateAction(params)
		} else {
			buf.WriteByte(char)
		}
	}
	if buf.Len() == 0 {
		return nil, io.EOF
	}
	params = append(params, buf.String())
	return CreateAction(params)
}

func (s *FormulaScanner) ScanParams() string {
	buf := &bytes.Buffer{}
	inQuote := s.PreviewNext() == '"'
	escape := false
	if inQuote {
		s.Next()
	}
	for s.HasNext() {
		char := s.Next()
		if char == '\n' {
			return buf.String()
		} else if char == '\\' {
			escape = true
		} else if escape {
			escape = false
			buf.WriteByte(char)
		} else if char == '"' && inQuote {
			s.Next()
			return buf.String()
		} else if char == ' ' && !inQuote {
			return buf.String()
		} else {
			buf.WriteByte(char)
		}
	}
	return buf.String()
}

func (s *FormulaScanner) ScanLineComment() {
	for s.HasNext() {
		char := s.Next()
		if char == '\n' {
			return
		}
	}
}
