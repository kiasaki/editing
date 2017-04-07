package main

var (
	hooks_buffer map[string][]func(*buffer)
)

func hook_buffer(name string, f func(*buffer)) {
	if _, ok := hooks_buffer[name]; !ok {
		hooks_buffer[name] = []func(*buffer){}
	}
	hooks_buffer[name] = append(hooks_buffer[name], f)
}

func hook_trigger_buffer(name string, b *buffer) {
	if hooks, ok := hooks_buffer[name]; ok {
		for _, f := range hooks {
			f(b)
		}
	}
}

func init_hooks() {
	hooks_buffer = map[string][]func(*buffer){}

	hook_buffer("moved", func(b *buffer) {
		if current_view_tree.leaf.buf == b {
			current_view_tree.leaf.adjust_scroll(b.last_render_width, b.last_render_height)
		}
	})
}
