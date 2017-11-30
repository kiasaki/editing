package main

func init_visual() {
	addMode("visual")
	bind("visual", k("ESC"), exit_visual_mode)
	bind("visual", k("y"), visual_mode_yank)
	bind("visual", k("d"), visual_mode_delete)
	bind("visual", k("p"), visual_mode_paste)
	bind("visual", k("c"), visual_mode_change)

	addMode("visual-line")
	bind("visual-line", k("ESC"), exit_visual_mode)
	bind("visual-line", k("y"), visual_mode_yank)
	bind("visual-line", k("d"), visual_mode_delete)
	bind("visual-line", k("p"), visual_mode_paste)
	bind("visual-line", k("c"), visual_mode_change)

	hook_buffer("moved", visual_rehighlight)
}

// Run highlight when in visual mode and cursor mode as normal we
// only recompute highlights when the buffer changes
func visual_rehighlight(b *Buffer) {
	if b.IsInMode("visual") || b.IsInMode("visual-line") {
		highlight_buffer(b)
	}
}

func visual_highlight(b *Buffer, l, c int) bool {
	in_visual_line := b.IsInMode("visual-line")
	if !b.IsInMode("visual") && !in_visual_line {
		return false
	}

	l1, l2 := orderLocations(b.Cursor, getMark('∫').Loc)

	if in_visual_line {
		// compare using line numbers
		return l1.Line <= l && l <= l2.Line
	} else {
		// compare using line numbers + char position
		loc := NewLocation(l, c)
		return loc.After(l1) && loc.Before(l2) || loc.Equal(l1) || loc.Equal(l2)
	}
}

func exit_visual_mode(vt *ViewTree, b *Buffer, kl *KeyList) {
	b.RemoveMode("visual")
	b.RemoveMode("visual-line")
	highlight_buffer(b)
}

func enter_visual_mode(vt *ViewTree, b *Buffer, kl *KeyList) {
	b.AddMode("visual")
	markCreate('∫', b)
}

func enter_visual_block_mode(vt *ViewTree, b *Buffer, kl *KeyList) {
	b.AddMode("visual-line")
	markCreate('∫', b)
	highlight_buffer(b)
}

func visual_mode_selection(b *Buffer) ([]rune, *Location, *Location) {
	in_visual_line := b.IsInMode("visual-line")
	l1, l2 := orderLocations(b.Cursor, getMark('∫').Loc)
	data := []rune{}
	if in_visual_line {
		l1.Char = 0
		l2.Char = len(b.GetLine(l2.Line))
	}
	for l := l1.Line; l <= l2.Line; l++ {
		start_char := 0
		if l == l1.Line {
			start_char = l1.Char
		}
		line_data := b.GetLine(l)
		end_char := len(line_data)
		if l == l2.Line {
			end_char = min(l2.Char+1, len(line_data))
		}
		data = append(data, line_data[start_char:end_char]...)
		if l != l2.Line || end_char == len(line_data) {
			data = append(data, '\n')
		}
	}
	return data, l1, l2
}

func visual_mode_yank(vt *ViewTree, b *Buffer, kl *KeyList) {
	text, l1, _ := visual_mode_selection(b)
	clipboardSet(defaultClipboard, text)

	b.MoveTo(l1.Char, l1.Line)
	exit_visual_mode(vt, b, kl)
}
func visual_mode_delete(vt *ViewTree, b *Buffer, kl *KeyList) {
	text, l1, _ := visual_mode_selection(b)
	b.MoveTo(l1.Char, l1.Line)
	b.Remove(len(text))

	exit_visual_mode(vt, b, kl)
}
func visual_mode_paste(vt *ViewTree, b *Buffer, kl *KeyList) {
	text, l1, _ := visual_mode_selection(b)
	clipboard_text := clipboardGet(defaultClipboard)
	b.MoveTo(l1.Char, l1.Line)
	b.Remove(len(text))
	b.Insert(clipboard_text)
	clipboardSet(defaultClipboard, text)

	b.MoveTo(l1.Char, l1.Line)
	exit_visual_mode(vt, b, kl)
}
func visual_mode_change(vt *ViewTree, b *Buffer, kl *KeyList) {
	text, l1, _ := visual_mode_selection(b)
	b.MoveTo(l1.Char, l1.Line)
	b.Remove(len(text))

	exit_visual_mode(vt, b, kl)
	enterInsertMode(vt, b, kl)
}
