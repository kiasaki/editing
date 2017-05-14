package main

func init_visual() {
	add_mode("visual")
	bind("visual", k("ESC"), exit_visual_mode)

	add_mode("visual-line")
	bind("visual-line", k("ESC"), exit_visual_mode)

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
	is_in_visual_line := b.is_in_mode("visual-line")
	if !b.is_in_mode("visual") && !is_in_visual_line {
		return false
	}

	l1, l2 := order_locations(b.cursor, get_mark('∫').loc)

	if is_in_visual_line {
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
}
