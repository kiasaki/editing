package main

import (
	"strings"

	"github.com/gdamore/tcell"
)

var (
	style_maps                  = map[*buffer][][]tcell.Style{}
	highlighting_reserved_words = []string{
		"func", "function", "fn", "lambda",
		"var", "let", "const", "def",
		"type", "struct", "interface", "class",
		"if", "else", "for", "of", "in", "while", "break", "continue", "goto", "end",
		"select", "switch", "case",
		"import", "export", "package", "from",
		"go", "async", "await",
		"raise", "throw", "try", "catch", "except", "finally",
	}
	highlighting_special_words = []string{
		"self", "this", "true", "false", "True", "False", "nil", "null", "None",
	}
)

func highlighting_styles(b *buffer) [][]tcell.Style {
	return style_maps[b]
}

func init_highlighting() {
	hook_buffer("modified", highlight_buffer)
}

func highlight_buffer(b *buffer) {
	s := style("default")
	ss := style("special")
	sse := style("search")
	sts := style("text.string")
	stn := style("text.number")
	stc := style("text.comment")
	str := style("text.reserved")
	stsp := style("text.special")

	style_map := make([][]tcell.Style, len(b.data))
	in_string := rune(0)
	for l := range b.data {
		in_line_comment := false
		word := ""
		style_map[l] = make([]tcell.Style, len(b.data[l])+1)
		for c, char := range b.data[l] {
			prev_char := rune(0)
			if c > 0 {
				prev_char = b.data[l][c-1]
			}

			// for numbers
			passed_alpha := false
			if is_alpha(prev_char) {
				passed_alpha = true
			}
			// for special words
			if is_word(char) {
				word += string(char)
			} else {
				word = ""
			}

			if high_len := search_highlight(l, c); high_len > 0 {
				for i := 0; i < high_len; i++ {
					// TODO check bounds (limit to current buff)
					if style_map[l][c+i] == 0 {
						style_map[l][c+i] = sse
					}
				}
			}
			if in_line_comment {
				style_map[l][c] = stc
				continue
			}
			if in_string > 0 && c-1 > 0 && b.data[l][c-1] == '\\' && (c-2 < 0 || b.data[l][c-2] != '\\') {
				style_map[l][c] = sts
				continue
			}
			if char == '/' && prev_char == '/' {
				in_line_comment = true
				style_map[l][c] = stc
				style_map[l][c-1] = stc
			}
			if char == '\'' || char == '"' {
				if in_string == char {
					in_string = rune(0)
				} else if in_string == rune(0) {
					in_string = char
				}
				style_map[l][c] = sts
			}
			if style_map[l][c] != 0 {
				continue
			}
			if in_string > 0 {
				style_map[l][c] = sts
			} else if list_contains_string(highlighting_special_words, word) && c+1 < len(b.data[l]) && !is_word(b.data[l][c+1]) {
				for i := len(word) - 1; i >= 0; i-- {
					style_map[l][c-i] = stsp
				}
			} else if list_contains_string(highlighting_reserved_words, word) && c+1 < len(b.data[l]) && !is_word(b.data[l][c+1]) {
				for i := len(word) - 1; i >= 0; i-- {
					style_map[l][c-i] = str
				}
			} else if !passed_alpha && is_num(char) {
				style_map[l][c] = stn
			} else if strings.ContainsRune(special_chars, b.data[l][c]) {
				style_map[l][c] = ss
			} else {
				style_map[l][c] = s
			}
		}
	}

	style_maps[b] = style_map
}
