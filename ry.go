package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/go-errors/errors"
	runewidth "github.com/mattn/go-runewidth"
	zclip "github.com/zyedidia/clipboard"
)

// fatal(pp.Sprintln(value))

const (
	special_chars = "[]{}()/\\"
)

var (
	keys_entered                     = new_key_list("")
	last_key                         = new_key_list("")
	term_events                      = make(chan tcell.Event, 500)
	default_clipboard                = '_'
	clipboards                       = map[rune][]rune{'_': []rune{}}
	editor_mode                      = "normal"
	editor_message                   = ""
	editor_message_type              = "info"
	editor_width                     = 0
	editor_height                    = 0
	buffers                          = []*buffer{}
	screen              tcell.Screen = nil
	root_view_tree      *view_tree   = nil
	current_view_tree   *view_tree   = nil
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-v" {
		fmt.Println("ry v0.0.0")
		os.Exit(0)
	}

	defer handle_panics()

	init_modes()
	init_commands()

	init_config()
	init_hooks()
	init_highlighting()
	init_search()
	init_visual()

	init_screen()
	init_term_events()
	init_buffers()
	init_views()

	render()

top:
	for {
		select {
		case ev := <-term_events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlQ {
					screen.Fini()
					break top
					/*
						} else if ev.Key() == tcell.KeyEscape {
							kl := k("ESC")
							enter_normal_mode(current_view_tree, current_view_tree.leaf.buf, kl)
							last_key = kl
							keys_entered = k("")
					*/
				} else {
					keys_entered.add_key(new_key_from_event(ev))

					buf := current_view_tree.leaf.buf
					for _, mode_name := range buf.modes {
						if matched := mode_handle(must_find_mode(mode_name), keys_entered); matched != nil {
							keys_entered = k("")
							last_key = matched
							continue top
						}
					}
					if matched := mode_handle(must_find_mode(editor_mode), keys_entered); matched != nil {
						keys_entered = k("")
						last_key = matched
						continue top
					}
				}
			case *tcell.EventResize:
				editor_width, editor_height = screen.Size()
			}
		default:
			render()
		}
	}

}

// {{{ mode

type command_fn func(*view_tree, *buffer, *key_list)

type mode_binding struct {
	k *key_list
	f command_fn
}

type mode struct {
	name     string
	bindings []*mode_binding
}

var modes = map[string]*mode{}

func mode_handle(m *mode, kl *key_list) *key_list {
	var match *key_list = nil
	var match_binding *mode_binding = nil
	for _, binding := range m.bindings {
		if matched := kl.has_suffix(binding.k); matched != nil {
			if match == nil || len(matched.keys) > len(match.keys) {
				match_binding = binding
				match = matched
			}
		}
	}
	if match != nil {
		match_binding.f(current_view_tree, current_view_tree.leaf.buf, match)
		return match
	}
	return nil
}

func find_mode(name string) *mode {
	if m, ok := modes[name]; ok {
		return m
	} else {
		return nil
	}
}

func must_find_mode(name string) *mode {
	mode := find_mode(name)
	if mode == nil {
		panic(fmt.Sprintf("no mode named '%s'", name))
	}
	return mode
}

// Adds a new empty mode to the mode list, if not already present
func add_mode(name string) {
	if _, ok := modes[name]; !ok {
		modes[name] = &mode{name: name, bindings: []*mode_binding{}}
	}
}

func bind(mode_name string, k *key_list, f command_fn) {
	mode := must_find_mode(mode_name)

	// If this key is bound, update bound function
	for _, binding := range mode.bindings {
		if k.String() == binding.k.String() {
			binding.f = f
		}
	}
	// Else, it's a new binding, add it
	mode.bindings = append(mode.bindings, &mode_binding{k: k, f: f})
}

func init_modes() {
	add_mode("normal")
	bind("normal", k("m $alpha"), command_mark)
	bind("normal", k("' $alpha"), command_move_to_mark)
	bind("normal", k(":"), prompt_command)
	bind("normal", k("p"), command_paste)
	bind("normal", k("h"), move_left)
	bind("normal", k("j"), move_down)
	bind("normal", k("k"), move_up)
	bind("normal", k("l"), move_right)
	bind("normal", k("0"), move_line_beg)
	bind("normal", k("$"), move_line_end)
	bind("normal", k("g g"), move_top)
	bind("normal", k("G"), move_bottom)
	bind("normal", k("C-u"), move_jump_up)
	bind("normal", k("C-d"), move_jump_down)
	bind("normal", k("z z"), move_center_line)
	bind("normal", k("w"), move_word_forward)
	bind("normal", k("b"), move_word_backward)
	bind("normal", k("C-c"), cancel_keys_entered)
	bind("normal", k("C-g"), cancel_keys_entered)
	bind("normal", k("ESC ESC"), cancel_keys_entered)
	bind("normal", k("i"), enter_insert_mode)
	bind("normal", k("a"), enter_insert_mode_append)
	bind("normal", k("A"), enter_insert_mode_eol)
	bind("normal", k("o"), enter_insert_mode_nl)
	bind("normal", k("O"), enter_insert_mode_nl_up)
	bind("normal", k("x"), remove_char)
	bind("normal", k("d d"), remove_line)
	bind("normal", k("u"), command_undo)
	bind("normal", k("C-r"), command_redo)
	bind("normal", k("y y"), command_copy_line)
	bind("normal", k("p"), command_paste)
	bind("normal", k("v"), enter_visual_mode)
	bind("normal", k("V"), enter_visual_block_mode)

	add_mode("insert")
	bind("insert", k("ESC"), enter_normal_mode)
	bind("insert", k("RET"), insert_enter)
	bind("insert", k("BAK"), insert_backspace)
	bind("insert", k("$any"), insert)

	add_mode("prompt")
	bind("prompt", k("C-c"), prompt_cancel)
	bind("prompt", k("C-g"), prompt_cancel)
	bind("prompt", k("ESC"), prompt_cancel)
	bind("prompt", k("RET"), prompt_finish)
	bind("prompt", k("BAK"), prompt_backspace)
	bind("prompt", k("$any"), prompt_insert)

	add_mode("buffers")
	bind("buffers", k("q"), func(vt *view_tree, b *buffer, kl *key_list) {
		close_current_buffer(true)
	})
	bind("buffers", k("RET"), func(vt *view_tree, b *buffer, kl *key_list) {
		close_current_buffer(true)
		show_buffer(string(b.data[b.cursor.line]))
	})

	add_mode("directory")
	bind("directory", k("q"), func(vt *view_tree, b *buffer, kl *key_list) {
		close_current_buffer(true)
	})
	bind("directory", k("RET"), func(vt *view_tree, b *buffer, kl *key_list) {
		close_current_buffer(true)
		file_path := filepath.Join(b.path, string(b.data[b.cursor.line]))
		run_command([]string{"edit", file_path})
	})

	// TODO Remove once I have user configurable bindings
	bind("normal", k("SPC b"), func(vt *view_tree, b *buffer, kl *key_list) {
		run_command([]string{"buffers"})
	})
	bind("normal", k("SPC f"), func(vt *view_tree, b *buffer, kl *key_list) {
		if b.path == "" {
			run_command([]string{"edit", "."})
		} else {
			run_command([]string{"edit", filepath.Dir(b.path)})
		}
	})
	bind("normal", k("SPC n"), func(vt *view_tree, b *buffer, kl *key_list) {
		run_command([]string{"clearsearch"})
	})
}

func move_left(vt *view_tree, b *buffer, kl *key_list) {
	b.move(-1, 0)
}
func move_right(vt *view_tree, b *buffer, kl *key_list) {
	b.move(1, 0)
}
func move_up(vt *view_tree, b *buffer, kl *key_list) {
	b.move(0, -1)
}
func move_down(vt *view_tree, b *buffer, kl *key_list) {
	b.move(0, 1)
}
func move_line_beg(vt *view_tree, b *buffer, kl *key_list) {
	b.move_to(0, b.cursor.line)
}
func move_line_end(vt *view_tree, b *buffer, kl *key_list) {
	b.move_to(len(b.data[b.cursor.line]), b.cursor.line)
}
func move_top(vt *view_tree, b *buffer, kl *key_list) {
	b.move_to(0, 0)
}
func move_bottom(vt *view_tree, b *buffer, kl *key_list) {
	b.move_to(0, len(b.data)-1)
}
func move_jump_up(vt *view_tree, b *buffer, kl *key_list) {
	b.move(0, -15)
}
func move_jump_down(vt *view_tree, b *buffer, kl *key_list) {
	b.move(0, 15)
}
func move_center_line(vt *view_tree, b *buffer, kl *key_list) {
	vt.leaf.center_pending = true
}
func move_word_backward(vt *view_tree, b *buffer, kl *key_list) {
	b.move_word_backward()
}
func move_word_forward(vt *view_tree, b *buffer, kl *key_list) {
	b.move_word_forward()
}

func cancel_keys_entered(vt *view_tree, b *buffer, kl *key_list) {
	keys_entered = k("")
}

// Enter in a new mode
func enter_mode(mode string) {
	editor_mode = mode
	// TODO maybe not the best place to clear this
	message("")
}

func enter_normal_mode(vt *view_tree, b *buffer, kl *key_list) {
	move_left(vt, b, kl)
	enter_mode("normal")
}
func enter_insert_mode(vt *view_tree, b *buffer, kl *key_list) {
	enter_mode("insert")
}
func enter_insert_mode_append(vt *view_tree, b *buffer, kl *key_list) {
	move_right(vt, b, kl)
	enter_mode("insert")
}
func enter_insert_mode_eol(vt *view_tree, b *buffer, kl *key_list) {
	move_line_end(vt, b, kl)
	enter_mode("insert")
}
func enter_insert_mode_nl(vt *view_tree, b *buffer, kl *key_list) {
	move_line_end(vt, b, kl)
	b.insert([]rune("\n"))
	b.move(0, 1) // ensure a valid position
	enter_mode("insert")
}
func enter_insert_mode_nl_up(vt *view_tree, b *buffer, kl *key_list) {
	move_line_beg(vt, b, kl)
	b.insert([]rune("\n"))
	b.move(0, 0) // ensure a valid position
	enter_mode("insert")
}

func insert_enter(vt *view_tree, b *buffer, kl *key_list) {
	i := 0
	for ; i < len(b.data[b.cursor.line]) && is_space(b.data[b.cursor.line][i]); i++ {
	}
	b.insert([]rune("\n" + strings.Repeat(" ", i)))
	b.move_to(i, b.cursor.line+1)
}
func insert_backspace(vt *view_tree, b *buffer, kl *key_list) {
	if b.cursor.char == 0 {
		if b.cursor.line != 0 {
			move_up(vt, b, kl)
			move_line_end(vt, b, kl)
			b.remove(1)
		}
	} else {
		if b.char_at_left() != ' ' {
			b.move(-1, 0)
			b.remove(1)
			return
		}
		// handle spaces
		delete_n := 1
		if config_get_bool("tab_to_spaces", b) {
			delete_n = int(config_get_number("tab_width", b))
		}
		for i := 0; i < delete_n && b.char_at_left() == ' '; i++ {
			b.move(-1, 0)
			b.remove(1)
		}
	}
}
func insert(vt *view_tree, b *buffer, kl *key_list) {
	k := kl.keys[len(kl.keys)-1]
	if k.key == tcell.KeyTab {
		if config_get_bool("tab_to_spaces", b) {
			tab_width := int(config_get_number("tab_width", b))
			message(strconv.Itoa(tab_width))
			b.insert([]rune(strings.Repeat(" ", tab_width)))
			b.move(tab_width, 0)
		} else {
			b.insert([]rune{'\t'})
			b.move(1, 0)
		}
	} else if k.key == tcell.KeyRune && k.mod == 0 {
		b.insert([]rune{k.chr})
		move_right(vt, b, kl)
	} else {
		message("Can't insert '" + kl.String() + "'")
	}
}

func remove_char(vt *view_tree, b *buffer, kl *key_list) {
	removed := b.remove(1)
	clipboard_set(default_clipboard, removed)
}
func remove_line(vt *view_tree, b *buffer, kl *key_list) {
	move_line_beg(vt, b, kl)
	removed := b.remove(len(b.data[b.cursor.line]) + 1)
	clipboard_set(default_clipboard, removed)
}

func command_undo(vt *view_tree, b *buffer, kl *key_list) {
	b.undo()
}
func command_redo(vt *view_tree, b *buffer, kl *key_list) {
	b.redo()
}
func command_copy_line(vt *view_tree, b *buffer, kl *key_list) {
	value := make([]rune, len(b.data[b.cursor.line])+1)
	copy(value, b.data[b.cursor.line])
	value[len(value)-1] = '\n'
	clipboard_set(default_clipboard, value)
}
func command_paste(vt *view_tree, b *buffer, kl *key_list) {
	value := clipboard_get(default_clipboard)
	if len(value) == 0 {
		message("Nothing to paste!")
		return
	}
	b.insert(value)
}

func command_mark(vt *view_tree, b *buffer, kl *key_list) {
	mark_letter := kl.keys[len(kl.keys)-1].chr
	mark_create(mark_letter, b)
}
func command_move_to_mark(vt *view_tree, b *buffer, kl *key_list) {
	mark_letter := kl.keys[len(kl.keys)-1].chr
	mark_jump(mark_letter)
}

// }}}

// {{{ prompt
var (
	editor_is_prompt_active                           = false
	editor_prompt                                     = ""
	editor_prompt_value                               = ""
	editor_prompt_callback_fn   func([]string)        = nil
	editor_prompt_completion_fn func(string) []string = nil
)

func prompt(prompt string, comp_fn func(string) []string, cb_fn func([]string)) {
	editor_prompt = prompt
	editor_prompt_value = ""
	editor_prompt_callback_fn = cb_fn
	editor_prompt_completion_fn = comp_fn
	enter_mode("prompt")
}

func noop_complete(prefix string) []string {
	return []string{}
}

func prompt_update_completion() {
	// TODO
}

func prompt_cancel(vt *view_tree, b *buffer, kl *key_list) {
	enter_mode("normal")
}

func prompt_finish(vt *view_tree, b *buffer, kl *key_list) {
	enter_mode("normal")
	// TODO better args parsing
	editor_prompt_callback_fn(strings.Split(editor_prompt_value, " "))
}

func prompt_backspace(vt *view_tree, b *buffer, kl *key_list) {
	if len(editor_prompt_value) > 0 {
		editor_prompt_value = editor_prompt_value[:len(editor_prompt_value)-1]
		prompt_update_completion()
	}
}

func prompt_insert(vt *view_tree, b *buffer, kl *key_list) {
	k := kl.keys[len(kl.keys)-1]
	if k.key == tcell.KeyRune && k.mod == 0 {
		editor_prompt_value += string(k.chr)
		prompt_update_completion()
	}
}

func prompt_command(vt *view_tree, b *buffer, kl *key_list) {
	prompt(":", func(prefix string) []string {
		// TODO provide command suggestions
		return []string{}
	}, run_command)
}

// }}}

// {{{ clipboard
func clipboard_get(register rune) []rune {
	if register == default_clipboard {
		if value, err := zclip.ReadAll("clipboard"); err == nil {
			return []rune(value)
		}
	}
	if value, ok := clipboards[register]; ok {
		return value
	}
	return []rune{}
}

func clipboard_set(register rune, value []rune) {
	if register == default_clipboard {
		if err := zclip.WriteAll(string(value), "clipboard"); err != nil {
			message_error("Error clipboard_get: " + err.Error())
			clipboards[register] = value
		}
	} else {
		clipboards[register] = value
	}
}

// }}}

// {{{ mark
type mark struct {
	loc         *location
	buffer_name string
}

var marks = map[rune]*mark{}

func mark_create(mark_letter rune, b *buffer) *mark {
	m := &mark{loc: b.cursor.clone(), buffer_name: b.name}
	marks[mark_letter] = m
	return m
}

func get_mark(mark_letter rune) *mark {
	return marks[mark_letter]
}

func mark_jump(mark_letter rune) {
	if m, ok := marks[mark_letter]; ok {
		if b := show_buffer(m.buffer_name); b != nil {
			b.move_to(m.loc.char, m.loc.line)
		} else {
			message_error("Can't find buffer named '" + m.buffer_name + "'")
		}
	}
}

// }}}

// {{{ buffer
type location struct {
	line int
	char int
}

func new_location(l, c int) *location {
	return &location{line: l, char: c}
}

func order_locations(l1, l2 *location) (*location, *location) {
	if l1.line < l2.line {
		return l1, l2
	} else if l1.line > l2.line {
		return l2, l1
	} else {
		if l1.char > l2.char {
			return l2, l1
		} else {
			return l1, l2
		}
	}
}

func (l1 *location) before(l2 *location) bool {
	ol1, _ := order_locations(l1, l2)
	return l1 == ol1
}

func (l1 *location) after(l2 *location) bool {
	_, ol2 := order_locations(l1, l2)
	return l1 == ol2
}

func (loc *location) clone() *location {
	return &location{line: loc.line, char: loc.char}
}

type char_range struct {
	beg int
	end int
}

func new_char_range(b, e int) *char_range {
	return &char_range{b, e}
}

type buffer struct {
	data               [][]rune
	history            []*action
	history_index      int
	name               string
	path               string
	modified           bool
	cursor             *location
	modes              []string
	last_render_width  int
	last_render_height int
}

func new_buffer(name string, path string) *buffer {
	b := &buffer{
		data:          [][]rune{{}},
		history:       []*action{},
		history_index: -1,
		modified:      false,
		cursor:        new_location(0, 0),
		modes:         []string{},
	}

	if path == "" {
		// TODO ensure uniqueness
		b.name = name
	} else {
		b.set_path(path)
		// TODO ensure uniqueness
		b.name = name
	}

	return b
}

func (b *buffer) is_in_mode(name string) bool {
	for _, n := range b.modes {
		if n == name {
			return true
		}
	}
	return false
}

func (b *buffer) char_at(l, c int) rune {
	line := b.data[l]
	if c < 0 {
		return rune(0)
	} else if c < len(line) {
		return line[c]
	} else {
		return '\n'
	}
}

func (b *buffer) get_line(l int) []rune {
	return b.data[l]
}

func (b *buffer) char_at_left() rune {
	return b.char_at(b.cursor.line, b.cursor.char-1)
}

func (b *buffer) char_under_cursor() rune {
	return b.char_at(b.cursor.line, b.cursor.char)
}

func (b *buffer) first_line() bool {
	return b.cursor.line == 0
}

func (b *buffer) last_line() bool {
	return b.cursor.line == len(b.data)-1
}

func (b *buffer) move_to(c, l int) {
	b.cursor.line = max(min(l, len(b.data)-1), 0)
	b.cursor.char = max(min(c, len(b.data[b.cursor.line])), 0)
	hook_trigger_buffer("moved", b)
}

func (b *buffer) move(c, l int) {
	b.move_to(b.cursor.char+c, b.cursor.line+l)
}

func (b *buffer) move_word_forward() bool {
	for {
		c := b.char_under_cursor()
		if c == '\n' {
			if b.last_line() {
				return false
			} else {
				b.move_to(0, b.cursor.line+1)
				break
			}
		}

		for is_word(c) && c != '\n' {
			b.move(1, 0)
			c = b.char_under_cursor()
		}

		if c == '\n' {
			continue
		}
		break
	}

	c := b.char_under_cursor()
	for !is_word(c) && c != '\n' {
		b.move(1, 0)
		c = b.char_under_cursor()
	}

	return true
}

func (b *buffer) move_word_backward() bool {
	for {
		c := b.char_under_cursor()
		if b.cursor.char == 0 {
			if b.first_line() {
				return false
			} else {
				b.move_to(len(b.data[b.cursor.line-1]), b.cursor.line-1)
				continue
			}
		}

		for !is_word(c) && b.cursor.char != 0 {
			b.move(-1, 0)
			c = b.char_under_cursor()
		}

		if b.cursor.char == 0 {
			continue
		}
		break
	}

	c := b.char_under_cursor()
	for is_word(c) && b.cursor.char != 0 {
		b.move(-1, 0)
		c = b.char_under_cursor()
	}

	return true
}

func (b *buffer) insert(data []rune) {
	a := new_action(action_type_insert, b.cursor.clone(), data)
	b.history_index++
	b.history = try_merge_history(b.history[:b.history_index], a)
	a.apply(b)
}

func (b *buffer) remove_at(loc *location, n int) []rune {
	a := new_action(action_type_remove, loc.clone(), make([]rune, n))
	b.history_index++
	b.history = try_merge_history(b.history[:b.history_index], a)
	a.apply(b)
	return a.data
}

func (b *buffer) remove(n int) []rune {
	return b.remove_at(b.cursor, n)
}

func (b *buffer) undo() {
	if b.history_index >= 0 && b.history_index != -1 {
		b.history[b.history_index].revert(b)
		b.history_index--
	} else {
		message("Noting to undo!")
	}
}

func (b *buffer) redo() {
	if b.history_index+1 < len(b.history) && len(b.history) > b.history_index+1 {
		b.history_index++
		b.history[b.history_index].apply(b)
	} else {
		message("Noting to redo!")
	}
}

func (b *buffer) set_path(path string) {
	var err error
	b.path, err = filepath.Abs(path)
	if err != nil {
		b.path = filepath.Clean(path)
	}
	b.name = ""
	name := filepath.Base(b.path)

	i := 1
check_name:
	for _, b2 := range buffers {
		if b2.name == name {
			b.name = name + " " + strconv.Itoa(i)
			i++
			goto check_name
		}
	}
	if b.name == "" {
		b.name = name
	}
}

func (b *buffer) add_mode(name string) {
	if b.is_in_mode(name) {
		return
	}
	b.modes = append(b.modes, name)
}

func (b *buffer) remove_mode(name string) {
	for i, n := range b.modes {
		if n == name {
			b.modes = append(b.modes[:i], b.modes[i+1:]...)
			return
		}
	}
}

func (b *buffer) contents() string {
	ret := ""
	for _, line := range b.data {
		ret += string(line) + "\n"
	}
	return ret
}

func (b *buffer) nice_path() string {
	return strings.Replace(b.path, os.Getenv("HOME"), "~", -1)
}

func (b *buffer) save() {
	if b.path == "" {
		message_error("Can't save a buffer without a path.")
		return
	}
	err := ioutil.WriteFile(b.path, []byte(b.contents()), 0666)
	if err != nil {
		message_error("Error saving buffer: " + err.Error())
	} else {
		b.modified = false
		message("Buffer written to '" + b.nice_path() + "'")
	}
}

func try_merge_history(al []*action, a *action) []*action {
	// TODO save end location on actions so that we can merge them here
	return append(al, a)
}

// }}}

// {{{ action
type action_type int

const (
	action_type_insert action_type = 1
	action_type_remove             = -1
)

type action struct {
	typ  action_type
	loc  *location
	data []rune
}

func new_action(typ action_type, loc *location, data []rune) *action {
	return &action{typ: typ, loc: loc, data: data}
}

func (a *action) apply(b *buffer) {
	a.do(b, a.typ)
	b.modified = true
	hook_trigger_buffer("modified", b)
}

func (a *action) revert(b *buffer) {
	a.do(b, -a.typ)
	b.modified = true
	hook_trigger_buffer("modified", b)
}

func (a *action) do(b *buffer, typ action_type) {
	if typ == action_type_insert {
		a.insert(b)
	} else {
		a.remove(b)
	}
}

func (a *action) insert(b *buffer) {
	c, l := a.loc.char, a.loc.line
	for i := len(a.data) - 1; i >= 0; i-- {
		ch := a.data[i]
		if ch == '\n' {
			rest := append([]rune(nil), b.data[l][c:]...)
			b.data[l] = b.data[l][:c]
			b.data = append(b.data[:l+1],
				append([][]rune{rest}, b.data[l+1:]...)...)
		} else {
			b.data[l] = append(b.data[l][:c],
				append([]rune{ch}, b.data[l][c:]...)...)
		}
	}
}

func (a *action) remove(b *buffer) {
	n := len(a.data)
	c, l := a.loc.char, a.loc.line
	removed := []rune{}
	for i := 0; i < n; i++ {
		removed = append(removed, b.char_at(l, c))
		if b.char_at(l, c) == '\n' {
			if len(b.data)-1 == l {
				a.data = removed
				return
			}
			b.data[l] = append(b.data[l], b.data[l+1]...)
			b.data = append(b.data[:l+1], b.data[l+2:]...)
		} else {
			b.data[l] = append(b.data[l][:c], b.data[l][c+1:]...)
		}
	}
	a.data = removed
}

// }}}

// {{{ commands
func open_buffer_from_file(path string) *buffer {
	if file_info, err := os.Stat(path); os.IsNotExist(err) || err != nil {
		return open_buffer_named(filepath.Base(path))
	} else {
		if file_info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				message_error("Error opening directory: " + err.Error())
				return nil
			}
			files, err := file.Readdir(0)
			if err != nil {
				message_error("Error opening directory: " + err.Error())
				return nil
			}
			buf := new_buffer(filepath.Base(path), path)
			buf.data = [][]rune{}
			file_names := []string{}
			for _, file_info := range files {
				if file_info.IsDir() {
					file_names = append(file_names, " "+file_info.Name())
				} else {
					file_names = append(file_names, file_info.Name())
				}
			}
			file_names = append(file_names, " ..")
			sort.Strings(file_names)
			for _, file_name := range file_names {
				if file_name[0] == ' ' { // is dir
					buf.data = append(buf.data, []rune(file_name[1:]+"/"))
				} else {
					buf.data = append(buf.data, []rune(file_name))
				}
			}
			buf.add_mode("directory")
			buffers = append(buffers, buf)
			hook_trigger_buffer("modified", buf)
			return buf
		}
	}

	buf := new_buffer(filepath.Base(path), path)
	if buf.path != "" {
		contents, err := ioutil.ReadFile(buf.path)
		if err != nil {
			message_error("Error reading file '" + buf.nice_path() + "'")
			return nil
		}
		buf.data = [][]rune{}
		for _, line := range strings.Split(string(contents), "\n") {
			buf.data = append(buf.data, []rune(line))
		}
		if len(buf.data) > 1 {
			buf.data = buf.data[:len(buf.data)-1]
		}
	}
	buffers = append(buffers, buf)
	hook_trigger_buffer("modified", buf)
	return buf
}

func open_buffer_named(name string) *buffer {
	buf := new_buffer(name, "")
	buffers = append(buffers, buf)
	hook_trigger_buffer("modified", buf)
	return buf
}

// TODO check if is shown first, then if not create split not replace
func show_buffer(buffer_name string) *buffer {
	for _, b := range buffers {
		if b.name == buffer_name {
			if current_view_tree.leaf.buf == b {
				return b // already shown
			}
			current_view_tree = new_view_tree_leaf(nil, new_view(b))
			root_view_tree = current_view_tree
			return b
		}
	}
	return nil
}

func close_current_buffer(force bool) {
	b := current_view_tree.leaf.buf
	if b.modified && !force {
		message_error("Save buffer before closing it.")
		return
	}
	for i, b2 := range buffers {
		if b == b2 {
			buffers = append(buffers[:i], buffers[i+1:]...)
			break
		}
	}
	if len(buffers) == 0 {
		// TODO Use method here (don't handcode screen.Fini())
		screen.Fini()
		os.Exit(0)
	} else {
		// TODO call method to open left over buffer (don't set root too)
		current_view_tree = new_view_tree_leaf(nil, new_view(buffers[0]))
		root_view_tree = current_view_tree
	}
}

func find_buffer(name string) *buffer {
	for _, b := range buffers {
		if b.name == name {
			return b
		}
	}
	return nil
}

var commands = map[string]func([]string){}
var command_aliases = map[string]string{}

func run_command(args []string) {
	if len(args) == 0 {
		message_error("No command given!")
		return
	}
	command_name := args[0]
	if full_command_name, ok := command_aliases[command_name]; ok {
		command_name = full_command_name
	}
	if c, ok := commands[command_name]; ok {
		c(args)
	} else {
		message_error("No command named '" + command_name + "'")
	}
}

func add_command(name string, fn func([]string)) {
	commands[name] = fn
}
func add_alias(alias, name string) {
	command_aliases[alias] = name
}

func init_commands() {
	add_command("quit", func(args []string) {
		close_current_buffer(false)
	})
	add_alias("q", "quit")
	add_command("quit!", func(args []string) {
		close_current_buffer(true)
	})
	add_alias("q!", "quit!")
	add_command("write", func(args []string) {
		b := current_view_tree.leaf.buf
		if len(args) > 1 {
			b.set_path(args[1])
		}
		b.save()
	})
	add_alias("w", "write")
	add_command("edit", func(args []string) {
		if len(args) < 2 {
			message_error("Can't open buffer without a name or file path.")
		} else {
			path, err := filepath.Abs(args[1])
			if err != nil {
				path = args[1]
			}
			if b := open_buffer_from_file(path); b != nil {
				show_buffer(b.name)
			}
		}

	})
	add_alias("e", "edit")
	add_alias("o", "edit")
	add_command("writequit", func(args []string) {
		run_command([]string{"write"})
		run_command([]string{"quit"})
	})
	add_alias("wq", "writequit")
	add_command("buffers", func(args []string) {
		var b *buffer
		if b = find_buffer("*buffers*"); b == nil {
			b = open_buffer_named("*buffers*")
			b.add_mode("buffers")
		}
		b.data = [][]rune{}
		for _, buf := range buffers {
			if buf.name != "*buffers*" {
				b.data = append(b.data, []rune(buf.name))
			}
		}
		hook_trigger_buffer("modified", b)
		show_buffer(b.name)
	})
	add_alias("b", "buffers")
}

// }}}

// {{{ view
type view_highlight struct {
	beg   *location
	end   *location
	style tcell.Style
}

type view struct {
	buf            *buffer
	line_offset    int
	center_pending bool

	highlights []*view_highlight
}

func new_view(buf *buffer) *view {
	return &view{
		buf:            buf,
		line_offset:    0,
		center_pending: false,
		highlights:     []*view_highlight{},
	}
}

func (v *view) adjust_scroll(w, h int) {
	l := v.buf.cursor.line
	if v.center_pending {
		v.line_offset = max(l-int(math.Floor(float64(h-1)/2)), 1)
		v.center_pending = false
		return
	}
	// too low
	// (h-2) as height includes status bar and moving to 0 based
	if l > h-2+v.line_offset {
		v.line_offset = max(l-h+2, 0)
	}
	// too high
	if l < v.line_offset {
		v.line_offset = l
	}
}

// }}}

// {{{ view_tree
type view_tree struct {
	parent *view_tree
	left   *view_tree
	right  *view_tree
	top    *view_tree
	bottom *view_tree
	leaf   *view
	size   int
}

func new_view_tree_leaf(parent *view_tree, v *view) *view_tree {
	return &view_tree{parent: parent, leaf: v, size: 50}
}

// }}}

// {{{ message
func message(m string) {
	editor_message = m
	editor_message_type = "info"
}

func message_error(m string) {
	editor_message = m
	editor_message_type = "error"
}

// }}}

// {{{ styles
func style(name string) tcell.Style {
	// TODO make table based and configurable
	if name == "message.error" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorMaroon)
	}
	if name == "statusbar" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.Color(6))
	}
	if name == "statusbar.highlight" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.Color(5))
	}
	if name == "linenumber" {
		return tcell.StyleDefault.
			Foreground(tcell.Color(6))
	}
	if name == "search" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.ColorOlive)
	}
	if name == "visual" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorWhite).
			Background(tcell.Color(0))
	}
	if name == "special" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorPurple)
	}
	if name == "text.string" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorOlive)
	}
	if name == "text.number" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorNavy)
	}
	if name == "text.comment" {
		return tcell.StyleDefault.
			Foreground(10)
	}
	if name == "text.reserved" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorPurple)
	}
	if name == "text.special" {
		return tcell.StyleDefault.
			Foreground(tcell.Color(6))
	}
	if name == "cursor" {
		return tcell.StyleDefault.Reverse(true)
	}
	return tcell.StyleDefault.Foreground(tcell.Color(0))
}

// }}}

// {{{ render
func render() {
	width, height := editor_width, editor_height

	screen.Clear()

	render_view_tree(root_view_tree, 0, 0, width, height-1)

	render_message_bar(width, height)

	screen.Show()
}

func render_message_bar(width, height int) {
	s := style("default")

	if editor_mode == "prompt" {
		p := editor_prompt + editor_prompt_value
		write(s, 0, height-1, p)
		write(s.Reverse(true), len(p), height-1, " ")
		return
	}

	smb := s
	if editor_message_type == "error" {
		smb = style("message.error")
	}
	if editor_message != "" {
		write(smb, 0, height-1, editor_message)
	} else {
		write(smb, 0, height-1, keys_entered.String())
	}
	last_key_text := last_key.String()
	write(s, width-len(last_key_text)-1, height-1, last_key_text)
}

func render_view_tree(vt *view_tree, x, y, w, h int) {
	if vt.leaf != nil {
		render_view(vt.leaf, x, y, w, h)
		return
	}
	panic("unreachable")
}

func render_view(v *view, x, y, w, h int) {
	sc := style("cursor")
	sln := style("linenumber")
	ssb := style("statusbar")
	ssbh := style("statusbar.highlight")
	b := v.buf

	b.last_render_width = w
	b.last_render_height = h

	style_map := highlighting_styles(b)

	gutterw := len(strconv.Itoa(len(b.data))) + 1
	sy := y
	line := v.line_offset
	for line < len(b.data) && sy < y+h-1 {
		write(sln, x, sy, padl(strconv.Itoa(line+1), gutterw-1, ' '))

		sx := x + gutterw
		for c, char := range b.data[line] {
			if v == current_view_tree.leaf && line == b.cursor.line && c == b.cursor.char {
				sx += write(sc, sx, sy, string(char))
			} else {
				sx += write(style_map[line][c], sx, sy, string(char))
			}
			if sx >= x+w {
				break
			}
		}
		if v == current_view_tree.leaf &&
			line == b.cursor.line &&
			b.cursor.char == len(b.data[b.cursor.line]) {
			write(sc, sx, sy, " ")
		}

		line++
		sy++
	}

	// Current mode
	mode_status := editor_mode
	for _, mode_name := range b.modes {
		mode_status += "+" + mode_name
	}
	mode_status = " " + mode_status + " "
	write(ssbh, x, y+h-1, mode_status)

	// Position
	status_right := fmt.Sprintf("(%d,%d) %d ", b.cursor.char+1, b.cursor.line+1, len(b.data))
	write(ssb, x+w-len(status_right), y+h-1, status_right)
	// File name
	status_left := " " + b.name
	if b.modified {
		status_left += " [+]"
	}
	write(ssb, x+len(mode_status), y+h-1, padr(status_left, w-len(status_right)-len(mode_status), ' '))
}

// }}}

// {{{ init
func init_screen() {
	var err error
	screen, err = tcell.NewScreen()
	fatal_error(err)
	err = screen.Init()
	fatal_error(err)

	encoding.Register()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	screen.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorDefault))
	screen.Clear()

	editor_width, editor_height = screen.Size()
}

func init_term_events() {
	go func() {
		for {
			if screen == nil {
				break
			}
			term_events <- screen.PollEvent()
		}
	}()
}

func init_buffers() {
	for _, arg := range os.Args[1:] {
		open_buffer_from_file(arg)
	}
	if len(buffers) == 0 {
		open_buffer_named("*scratch*")
	}
}

func init_views() {
	view := new_view(buffers[0])
	root_view_tree = &view_tree{leaf: view}
	current_view_tree = root_view_tree
}

// }}}

// {{{ key
type key struct {
	mod tcell.ModMask
	key tcell.Key
	chr rune
}

const (
	key_type_catchall tcell.Key = iota + 5000
	key_type_alpha
	key_type_num
	key_type_alpha_num
)

func new_key_from_event(ev *tcell.EventKey) *key {
	k, r, m := ev.Key(), ev.Rune(), ev.Modifiers()

	key_name := ev.Name()
	if strings.HasPrefix(key_name, "Ctrl+") {
		k = tcell.KeyRune
		r = unicode.ToLower([]rune(key_name[5:6])[0])
	}

	// Handle Ctrl-h
	if k == tcell.KeyBackspace {
		m |= tcell.ModCtrl
		k = tcell.KeyRune
		r = 'h'
	}

	if k != tcell.KeyRune {
		r = 0
	}

	return &key{mod: ev.Modifiers(), key: k, chr: r}
}

func new_key(rep string) *key {
	if rep == "$any" {
		return &key{key: key_type_catchall}
	} else if rep == "$num" {
		return &key{key: key_type_num}
	} else if rep == "$alpha" {
		return &key{key: key_type_alpha}
	} else if rep == "$alphanum" {
		return &key{key: key_type_alpha_num}
	}

	parts := strings.Split(rep, "-")

	// Modifiers
	mod_mask := tcell.ModNone
	for _, part := range parts[:len(parts)-1] {
		switch part {
		case "C":
			mod_mask |= tcell.ModCtrl
		case "S":
			mod_mask |= tcell.ModShift
		case "A":
			mod_mask |= tcell.ModAlt
		case "M":
			mod_mask |= tcell.ModMeta
		}
	}

	// Key
	var r rune = 0
	var k tcell.Key
	last_part := parts[len(parts)-1]
	switch last_part {
	case "DEL":
		k = tcell.KeyDelete
	case "BAK":
		k = tcell.KeyBackspace2
	case "RET":
		k = tcell.KeyEnter
	case "SPC":
		k = tcell.KeyRune
		r = ' '
	case "ESC":
		k = tcell.KeyEscape
	case "TAB":
		k = tcell.KeyTab
	default:
		k = tcell.KeyRune
		r = []rune(last_part)[0]
	}

	return &key{mod: mod_mask, key: k, chr: r}
}

func (k *key) String() string {
	mods := []string{}
	if k.mod&tcell.ModCtrl != 0 {
		mods = append(mods, "C")
	}
	if k.mod&tcell.ModShift != 0 {
		mods = append(mods, "S")
	}
	if k.mod&tcell.ModAlt != 0 {
		mods = append(mods, "A")
	}
	if k.mod&tcell.ModMeta != 0 {
		mods = append(mods, "M")
	}

	name := string(k.chr)
	switch k.key {
	case tcell.KeyDelete:
		name = "DEL"
	case tcell.KeyBackspace2:
		name = "BAK"
	case tcell.KeyEnter:
		name = "RET"
	case tcell.KeyEscape:
		name = "ESC"
	case tcell.KeyTab:
		name = "TAB"
	}
	if k.key == tcell.KeyRune && k.chr == ' ' {
		name = "SPC"
	}

	return strings.Join(append(mods, name), "-")
}

func (k *key) is_rune() bool {
	return k.mod == 0 && k.key == tcell.KeyRune
}

// TODO implement alphanum match
func (k1 *key) matches(k2 *key) bool {
	if k1.key == key_type_catchall || k2.key == key_type_catchall {
		return true
	}
	if k1.key == key_type_alpha && k2.is_rune() && is_alpha(k2.chr) {
		return true
	}
	if k2.key == key_type_alpha && k1.is_rune() && is_alpha(k1.chr) {
		return true
	}
	if k1.key == key_type_num && k2.is_rune() && is_num(k2.chr) {
		return true
	}
	if k2.key == key_type_num && k1.is_rune() && is_num(k1.chr) {
		return true
	}
	return k1.mod == k2.mod && k1.key == k2.key && k1.chr == k2.chr
}

type key_list struct {
	keys []*key
}

func new_key_list(rep string) *key_list {
	kl := &key_list{[]*key{}}
	parts := strings.Split(rep, " ")
	for _, part := range parts {
		if part != "" {
			kl.keys = append(kl.keys, new_key(part))
		}
	}
	return kl
}

var k = new_key_list

func (kl *key_list) String() string {
	rep := []string{}
	for _, k := range kl.keys {
		rep = append(rep, k.String())
	}
	return strings.Join(rep, " ")
}

func (kl *key_list) add_key(k *key) {
	kl.keys = append(kl.keys, k)
}

func (kl1 *key_list) matches(kl2 *key_list) bool {
	if len(kl1.keys) != len(kl2.keys) {
		return false
	}
	for i := range kl1.keys {
		if !kl1.keys[i].matches(kl2.keys[i]) {
			return false
		}
	}
	return true
}

func (kl1 *key_list) has_suffix(kl2 *key_list) *key_list {
	for i := len(kl1.keys) - 1; i >= 0; i-- {
		tmp_kl := key_list{kl1.keys[i:]}
		if tmp_kl.matches(kl2) {
			return &tmp_kl
		}
	}
	return nil
}

// }}}

// {{{ utils
func fatal_error(err error) {
	if err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	if screen != nil {
		screen.Fini()
		screen = nil
	}
	fmt.Printf("%v\n", message)
	os.Exit(1)
}

func handle_panics() {
	if err := recover(); err != nil {
		fatal(fmt.Sprintf("ry fatal error:\n%v\n%s", err, errors.Wrap(err, 2).ErrorStack()))
	}
}

func write(style tcell.Style, x, y int, str string) int {
	s := screen
	i := 0
	var deferred []rune
	dwidth := 0
	for _, r := range str {
		// Handle tabs
		if r == '\t' {
			// TODO setting
			tab_width := int(config_get_number("tab_width", nil))

			// Print first tab char
			s.SetContent(x+i, y, '>', nil, style.Foreground(tcell.ColorAqua))
			i++

			// Add space till we reach tab column or tab_width
			for j := 0; j < tab_width-1 || i%tab_width == 0; j++ {
				s.SetContent(x+i, y, ' ', nil, style)
				i++
			}

			deferred = nil
			continue
		}

		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}

	if len(deferred) != 0 {
		s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}

	// i is the real width of what we just outputed
	return i
}

func list_contains_string(list []string, search string) bool {
	for _, item := range list {
		if item == search {
			return true
		}
	}
	return false
}

func padr(str string, length int, padding rune) string {
	for utf8.RuneCountInString(str) < length {
		str = str + string(padding)
	}
	return str
}

func padl(str string, length int, padding rune) string {
	for utf8.RuneCountInString(str) < length {
		str = string(padding) + str
	}
	return str
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func is_word(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || strings.ContainsRune("_", r)
}

func is_space(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n'
}

func is_alpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func is_num(r rune) bool {
	return r >= '0' && r <= '9'
}

/// }}}
