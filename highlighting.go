package main

import (
	"strings"

	"github.com/gdamore/tcell"
)

var (
	style_maps = map[*buffer][][]tcell.Style{}
)

func highlighting_styles(b *buffer) [][]tcell.Style {
	return style_maps[b]
}

func init_highlighting() {
	hook_buffer("changed", func(b *buffer) {
		s := style("default")
		ss := style("special")
		// sse := style("search")

		style_map := make([][]tcell.Style, len(b.data))
		for l := range b.data {
			style_map[l] = make([]tcell.Style, len(b.data[l])+1)
			for c := range b.data[l] {
				if strings.ContainsRune(special_chars, b.data[l][c]) {
					style_map[l][c] = ss
				} else {
					style_map[l][c] = s
				}
			}
		}

		style_maps[b] = style_map
	})
}
