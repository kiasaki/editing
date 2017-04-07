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
				if high_len := search_highlight(l, c); high_len > 0 {
					for i := 0; i < high_len; i++ {
						if style_map[l][c+i] == 0 {
							style_map[l][c+i] = sse
						}
					}
				}
				if style_map[l][c] != 0 {
					continue
				}
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
