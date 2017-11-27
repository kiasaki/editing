package main

func init_term() {
	add_mode("term")
	bind("term", k("ESC"), term_exit_mode)
	bind("insert", k("$any"), term_input)
}

func term_exit_mode(vt *view_tree, b *buffer, kl *key_list) {
	b.remove_mode("term")
}

func term_input(vt *view_tree, b *buffer, kl *key_list) {
}
