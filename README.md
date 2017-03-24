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
  - <kbd>:</kbd> Enters command mode
  - <kbd>i</kbd> Enters insert-mode
  - <kbd>i</kbd> Enters insert-mode at beginning of line
  - <kbd>a</kbd> Enters insert-mode then moves right 1 character
  - <kbd>A</kbd> Enters insert-mode then moves to the end of the line
  - <kbd>o</kbd> Enters insert-mode and creates a new line under the current one
  - <kbd>O</kbd> Enters insert-mode and creates a new line on top of the current one
  - <kbd>h</kbd> Moves cursor left
  - <kbd>h</kbd> Moves cursor right
  - <kbd>j</kbd> Moves cursor down
  - <kbd>k</kbd> Moves cursor up
  - <kbd>0</kbd> Moves cursor to the beginning of the line
  - <kbd>$</kbd> Moves cursor to the beginning of the line
  - <kbd>x</kbd> Delete char under cursor
  - <kbd>dd</kbd> Deletes line under cursor
  - <kbd>gg</kbd> Moves to the beginning of the buffer
  - <kbd>G</kbd> Moves to the end of the buffer
  - <kbd>C-w s</kbd> Splits buffer horizontally
  - <kbd>C-w v</kbd> Splits buffer vertically
  - <kbd>C-w h</kbd> Move to the window to the left
  - <kbd>C-w j</kbd> Move to the window to the bottom
  - <kbd>C-w k</kbd> Move to the window to the top
  - <kbd>C-w l</kbd> Move to the window to the right
- Insert mode
  - <kbd>any visible chars</kbd> Inserts character at cursor's position
  - <kbd>backspace</kbd> Deletes character to the left
  - <kbd>esc</kbd> Enters normal mode
- Command mode
  - <kbd>any visible chars</kbd> Inserts character
  - <kbd>backspace</kbd> Deletes character
  - <kbd>esc</kbd> Enters normal mode
  - <kbd>enter</kbd> Execute command and go back to normal mode
  - <kbd>C-u</kbd> Clear entered command

**Currently implemented command**

- `write/w` Save the current buffer
- `quit/q` Close the current buffer (closes ry when no buffers are left)
- `wq` Save and close the current buffer

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

