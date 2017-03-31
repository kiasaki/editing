package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/go-errors/errors"
	runewidth "github.com/mattn/go-runewidth"
)

// fatal(pp.Sprintln(value))

const (
	special_chars = "[]{}()/\\"
)

var (
	keys_entered                     = new_key_list("")
	last_key                         = new_key_list("")
	term_events                      = make(chan tcell.Event, 20)
	clipboard                        = []rune("")
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

	init_screen()
	init_term_events()
	init_buffers()
	init_views()
top:
	for {
		select {
		case ev := <-term_events:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlQ {
					screen.Fini()
					break top
				} else {
					keys_entered.add_key(new_key_from_event(ev))

					mode := modes[editor_mode]
					for kl, f := range mode.bindings {
						if matched := keys_entered.has_suffix(kl); matched != nil {
							keys_entered = k("")
							f(current_view_tree, current_view_tree.leaf.buf, matched)
							last_key = matched
						}
					}
				}
			case *tcell.EventResize:
				editor_width, editor_height = screen.Size()
			}
		}
		render()
	}

}

// {{{ mode
type command_fn func(*view_tree, *buffer, *key_list)

type mode struct {
	name     string
	bindings map[*key_list]command_fn
}

var modes = map[string]*mode{
	"normal": &mode{name: "normal", bindings: map[*key_list]command_fn{
		k("h"):       move_left,
		k("j"):       move_down,
		k("k"):       move_up,
		k("l"):       move_right,
		k("0"):       move_line_beg,
		k("$"):       move_line_end,
		k("g g"):     move_top,
		k("G"):       move_bottom,
		k("C-u"):     move_jump_up,
		k("C-d"):     move_jump_down,
		k("z z"):     move_center_line,
		k("w"):       move_word_forward,
		k("b"):       move_word_backward,
		k("C-c"):     cancel_keys_entered,
		k("C-g"):     cancel_keys_entered,
		k("ESC ESC"): cancel_keys_entered,
		k("i"):       enter_insert_mode,
		k("a"):       enter_insert_mode_append,
		k("A"):       enter_insert_mode_eol,
		k("o"):       enter_insert_mode_nl,
		k("O"):       enter_insert_mode_nl_up,
		k("x"):       remove_char,
		k("d d"):     remove_line,
		k("u"):       command_undo,
		k("C-r"):     command_redo,
	}},
	"insert": &mode{name: "insert", bindings: map[*key_list]command_fn{
		k("ESC"):  enter_normal_mode,
		k("RET"):  insert_enter,
		k("BAK"):  insert_backspace,
		k("$any"): insert,
	}},
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

func enter_normal_mode(vt *view_tree, b *buffer, kl *key_list) {
	move_left(vt, b, kl)
	editor_mode = "normal"
}
func enter_insert_mode(vt *view_tree, b *buffer, kl *key_list) {
	editor_mode = "insert"
}
func enter_insert_mode_append(vt *view_tree, b *buffer, kl *key_list) {
	move_right(vt, b, kl)
	editor_mode = "insert"
}
func enter_insert_mode_eol(vt *view_tree, b *buffer, kl *key_list) {
	move_line_end(vt, b, kl)
	editor_mode = "insert"
}
func enter_insert_mode_nl(vt *view_tree, b *buffer, kl *key_list) {
	move_line_end(vt, b, kl)
	b.insert([]rune("\n"))
	b.move(0, 1) // ensure a valid position
	editor_mode = "insert"
}
func enter_insert_mode_nl_up(vt *view_tree, b *buffer, kl *key_list) {
	move_line_beg(vt, b, kl)
	b.insert([]rune("\n"))
	b.move(0, 0) // ensure a valid position
	editor_mode = "insert"
}

func insert_enter(vt *view_tree, b *buffer, kl *key_list) {
	b.insert([]rune("\n"))
	move_down(vt, b, kl)
	move_line_beg(vt, b, kl)
}
func insert_backspace(vt *view_tree, b *buffer, kl *key_list) {
	if b.cursor.char == 0 {
		if b.cursor.line != 0 {
			move_up(vt, b, kl)
			move_line_end(vt, b, kl)
			b.remove(1)
		}
	} else {
		move_left(vt, b, kl)
		b.remove(1)
	}
}
func insert(vt *view_tree, b *buffer, kl *key_list) {
	k := kl.keys[len(kl.keys)-1]
	if k.key == tcell.KeyRune && k.mod == 0 {
		b.insert([]rune{k.chr})
		move_right(vt, b, kl)
	}
}

func remove_char(vt *view_tree, b *buffer, kl *key_list) {
	b.remove(1)
}
func remove_line(vt *view_tree, b *buffer, kl *key_list) {
	move_line_beg(vt, b, kl)
	b.remove(len(b.data[b.cursor.line]) + 1)
}

func command_undo(vt *view_tree, b *buffer, kl *key_list) {
	b.undo()
}
func command_redo(vt *view_tree, b *buffer, kl *key_list) {
	b.redo()
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

func (loc *location) clone() *location {
	return &location{line: loc.line, char: loc.char}
}

type char_range struct {
	beg int
	ent int
}

func new_char_range(b, e int) *char_range {
	return &char_range{b, e}
}

type buffer struct {
	data          [][]rune
	history       []*action
	history_index int
	name          string
	path          string
	modified      bool
	cursor        *location
}

func new_buffer(name string, path string) *buffer {
	return &buffer{
		data:          [][]rune{{}},
		history:       []*action{},
		history_index: -1,
		name:          name,
		path:          path,
		modified:      false,
		cursor:        new_location(0, 0),
	}
}

func (b *buffer) char_at(l, c int) rune {
	line := b.data[l]
	if c < len(line) {
		return line[c]
	} else {
		return '\n'
	}
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
				b.cursor.line++
				b.cursor.char = 0
				break
			}
		}

		for is_word(c) && c != '\n' {
			b.cursor.char++
			c = b.char_under_cursor()
		}

		if c == '\n' {
			continue
		}
		break
	}

	c := b.char_under_cursor()
	for !is_word(c) && c != '\n' {
		b.cursor.char++
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
				b.cursor.line--
				b.cursor.char = len(b.data[b.cursor.line])
				continue
			}
		}

		for !is_word(c) && b.cursor.char != 0 {
			b.cursor.char--
			c = b.char_under_cursor()
		}

		if b.cursor.char == 0 {
			continue
		}
		break
	}

	c := b.char_under_cursor()
	for is_word(c) && b.cursor.char != 0 {
		b.cursor.char--
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

func (b *buffer) remove(n int) []rune {
	a := new_action(action_type_remove, b.cursor.clone(), make([]rune, n))
	b.history_index++
	b.history = try_merge_history(b.history[:b.history_index], a)
	a.apply(b)
	return a.data
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

func try_merge_history(al []*action, a *action) []*action {
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
}

func (a *action) revert(b *buffer) {
	a.do(b, -a.typ)
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
func open_buffer(name, path string) {
	buf := new_buffer(name, path)
	if path != "" {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			message_error("Error reading file '" + path + "'")
			return
		}
		buf.data = [][]rune{}
		for _, line := range strings.Split(string(contents), "\n") {
			buf.data = append(buf.data, []rune(line))
		}
	}
	buffers = append(buffers, buf)
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
	marks      map[rune]*location
}

func new_view(buf *buffer) *view {
	return &view{

		buf:            buf,
		line_offset:    0,
		center_pending: false,
		highlights:     []*view_highlight{},
		marks:          map[rune]*location{},
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
			Foreground(tcell.ColorRed).
			Background(tcell.ColorDefault)
	}
	if name == "statusbar" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorAqua)
	}
	if name == "statusbar.highlight" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorYellow)
	}
	if name == "linenumber" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorYellow).
			Background(tcell.ColorDefault)
	}
	if name == "special" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorYellow).
			Background(tcell.ColorDefault)
	}
	if name == "cursor" {
		return tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorWhite)
	}
	return tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorDefault)
}

// }}}

// {{{ render
func render() {
	width, height := editor_width, editor_height
	s := style("default")

	screen.Clear()

	render_view_tree(root_view_tree, 0, 0, width, height-1)

	// Message bar
	if editor_message != "" {
		write(s, 0, height-1, editor_message)
	} else {
		write(s, 0, height-1, keys_entered.String())
	}
	last_key_text := last_key.String()
	write(s, width-len(last_key_text)-1, height-1, last_key_text)

	screen.Show()
}

func render_view_tree(vt *view_tree, x, y, w, h int) {
	if vt.leaf != nil {
		render_view(vt.leaf, x, y, w, h)
		return
	}
	panic("unreachable")
}

func render_view(v *view, x, y, w, h int) {
	s := style("default")
	sc := style("cursor")
	ss := style("special")
	sln := style("linenumber")
	ssb := style("statusbar")
	ssbh := style("statusbar.highlight")
	b := v.buf

	v.adjust_scroll(w, h)

	gutterw := len(strconv.Itoa(len(b.data))) + 1
	sy := y
	line := v.line_offset
	for line < len(b.data) && sy < y+h-1 {
		write(sln, x, sy, padl(strconv.Itoa(line+1), gutterw-1, ' '))

		sx := x + gutterw
		for c, char := range b.data[line] {
			if v == current_view_tree.leaf && line == b.cursor.line && c == b.cursor.char {
				sx += write(sc, sx, sy, string(char))
			} else if strings.ContainsRune(special_chars, char) {
				sx += write(ss, sx, sy, string(char))
			} else {
				sx += write(s, sx, sy, string(char))
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

	mode_status := " " + editor_mode + " "
	write(ssbh, x, y+h-1, mode_status)
	cur_status := fmt.Sprintf("(%d,%d) %d ", b.cursor.char+1, b.cursor.line+1, len(b.data))
	write(ssb, x+w-len(cur_status), y+h-1, cur_status)
	write(ssb, x+len(mode_status), y+h-1, padr(" "+b.name, w-len(cur_status)-len(mode_status), ' '))
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
		open_buffer(arg, arg)
	}
	if len(buffers) == 0 {
		open_buffer("*scratch*", "")
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
		k = tcell.Key(' ')
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
	case tcell.Key(' '):
		name = "SPC"
	case tcell.KeyEscape:
		name = "ESC"
	case tcell.KeyTab:
		name = "TAB"
	}

	return strings.Join(append(mods, name), "-")
}

func (k1 *key) matches(k2 *key) bool {
	if k1.key == key_type_catchall || k2.key == key_type_catchall {
		return true
	}
	// TODO implement num & alpha match
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
			tabWidth := 4

			// Print first tab char
			s.SetContent(x+i, y, '>', nil, style.Foreground(tcell.ColorAqua))
			i++

			// Add space till we reach tab column or tabWidth
			for j := 0; j < tabWidth-1 || i%tabWidth == 0; j++ {
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

func is_space(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n'
}

/// }}}
