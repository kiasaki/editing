package main

func init_visual() {
	add_mode("visual")
	bind("visual", k("ESC"), exit_visual_mode)
	bind("visual", k("y"), visual_mode_yank)
	bind("visual", k("d"), visual_mode_delete)
	bind("visual", k("p"), visual_mode_paste)

	add_mode("visual-line")
	bind("visual-line", k("ESC"), exit_visual_mode)
	bind("visual-line", k("y"), visual_mode_yank)
	bind("visual-line", k("d"), visual_mode_delete)
	bind("visual-line", k("p"), visual_mode_paste)

	hook_buffer("moved", visual_rehighlight)
}

// Run highlight when in visual mode and cursor mode as normall we
// only recompute highlights when the buffer changes
func visual_rehighlight(b *buffer) {
	if b.is_in_mode("visual") || b.is_in_mode("visual-line") {
		highlight_buffer(b)
	}
}

func visual_highlight(b *buffer, l, c int) bool {
	in_visual_line := b.is_in_mode("visual-line")
	if !b.is_in_mode("visual") && !in_visual_line {
		return false
	}

	l1, l2 := order_locations(b.cursor, get_mark('∫').loc)

	if in_visual_line {
		// compare using line numbers
		return l1.line <= l && l <= l2.line
	} else {
		// compare using line numbers + char position
		loc := new_location(l, c)
		return loc.after(l1) && loc.before(l2)
	}
}

func exit_visual_mode(vt *view_tree, b *buffer, kl *key_list) {
	b.remove_mode("visual")
	b.remove_mode("visual-line")
	highlight_buffer(b)
}

func enter_visual_mode(vt *view_tree, b *buffer, kl *key_list) {
	b.add_mode("visual")
	mark_create('∫', b)
}

func enter_visual_block_mode(vt *view_tree, b *buffer, kl *key_list) {
	b.add_mode("visual-line")
	mark_create('∫', b)
	highlight_buffer(b)
}

func visual_mode_selection(b *buffer) ([]rune, *location, *location) {
	in_visual_line := b.is_in_mode("visual-line")
	l1, l2 := order_locations(b.cursor, get_mark('∫').loc)
	data := []rune{}
	if in_visual_line {
		l1.char = 0
		l2.char = len(b.get_line(l2.line))
	}
	for l := l1.line; l <= l2.line; l++ {
		start_char := 0
		if l == l1.line {
			start_char = l1.char
		}
		line_data := b.get_line(l)
		end_char := len(line_data)
		if l == l2.line {
			end_char = l2.char
		}
		data = append(data, line_data[start_char:end_char]...)
		if l != l2.line || end_char == len(line_data) {
			data = append(data, '\n')
		}
	}
	return data, l1, l2
}

func visual_mode_yank(vt *view_tree, b *buffer, kl *key_list) {
	text, l1, _ := visual_mode_selection(b)
	clipboard_set(default_clipboard, text)

	b.move_to(l1.char, l1.line)
	exit_visual_mode(vt, b, kl)
}
func visual_mode_delete(vt *view_tree, b *buffer, kl *key_list) {
	text, l1, _ := visual_mode_selection(b)
	b.move_to(l1.char, l1.line)
	b.remove(len(text))

	exit_visual_mode(vt, b, kl)
}
func visual_mode_paste(vt *view_tree, b *buffer, kl *key_list) {
	text, l1, _ := visual_mode_selection(b)
	clipboard_text := clipboard_get(default_clipboard)
	b.move_to(l1.char, l1.line)
	b.remove(len(text))
	b.insert(clipboard_text)
	clipboard_set(default_clipboard, text)

	b.move_to(l1.char, l1.line)
	exit_visual_mode(vt, b, kl)
}
