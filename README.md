# ry

_A basic modal text editor, written in Go w/ a Lisp runtime_

### installing

To build and install `ry` on your computer simply run
`go get github.com/kiasaki/ry`.

### developing

```bash
make run # builds and runs ry
```

### features

`ry` is a text editor aiming to provide an editing environment similar to `vim`
in terms of key bindings and modal editing while being as easily extended
as `emacs`.

**Currently implemented keybindings:**

- Normal Mode
  - <kbd>C-q</kbd> Quits editor
  - <kbd>i</kbd> Enters insert-mode
  - <kbd>a</kbd> Enters insert-mode then moves right 1 character
  - <kbd>A</kbd> Enters insert-mode then moves to the end of the line
  - <kbd>h</kbd> Moves cursor left
  - <kbd>h</kbd> Moves cursor right
  - <kbd>j</kbd> Moves cursor down
  - <kbd>k</kbd> Moves cursor up
  - <kbd>0</kbd> Moves cursor to the beginning of the line
  - <kbd>$</kbd> Moves cursor to the beginning of the line
- Insert mode
  - <kbd>any visible chars</kbd> Inserts character at cursor's position
  - <kbd>backspace</kbd> Deletes character to the left
  - <kbd>esc</kbd> Enters normal mode

### screenshot

![](https://raw.githubusercontent.com/kiasaki/ry/master/screenshot.png)

### todo (minimal usability)

- Handle line wrapping
- Loading files in args
- Window splits
- Window movement
- Visual mode
- Yank/Paste (1 register)
- Undo
- Search
- Most movement keys
- Handle a few commands
  - Quit
  - Read
  - Edit
  - Save
  - Force quit
  - Run shell command
  - Replace

### todo

- Plugin language
- Many settings
- Error reporting in fringe
- Syntax highlighting
- Defining new commands
- Binding/Rebinding
- Creating modes
- Shell mode
- Full unicode support
- Help
- Tutorial
- Multiple clipboard registers
- Macros

### license

See `LICENSE` file.

