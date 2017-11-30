package main

func init_term() {
	addMode("term")
	bind("term", k("ESC"), term_exit_mode)
	bind("insert", k("$any"), term_input)
}

func term_exit_mode(vt *ViewTree, b *Buffer, kl *KeyList) {
	b.RemoveMode("term")
}

func term_input(vt *ViewTree, b *Buffer, kl *KeyList) {
}
