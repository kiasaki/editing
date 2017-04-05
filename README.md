# ry

_A simple modal text editor, written in Go, scriptable in Lua_

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
as `emacs`. It's is built with the day-to-day usage of the original author in
mind but hopefully flexible enough for anybody with Vim experience to adopt and
mold to their image.

**Currently implemented keybindings:**

- Normal Mode
  - <kbd>C-q</kbd> Quits editor
  - <kbd>:</kbd> Enters command mode
  - <kbd>C-c</kbd> Cancels keys entered
  - <kbd>C-g</kbd> Cancels keys entered
  - <kbd>ESC ESC</kbd> Cancels keys entered
  - <kbd>h</kbd> Moves cursor left
  - <kbd>h</kbd> Moves cursor right
  - <kbd>j</kbd> Moves cursor down
  - <kbd>k</kbd> Moves cursor up
  - <kbd>0</kbd> Moves cursor to the beginning of the line
  - <kbd>$</kbd> Moves cursor to the beginning of the line
  - <kbd>g g</kbd> Moves to the beginning of the buffer
  - <kbd>G</kbd> Moves to the end of the buffer
  - <kbd>C-u</kbd> Move 15 lines up
  - <kbd>C-d</kbd> Moves 15 lines down
  - <kbd>z z</kbd> Centers current line in view
  - <kbd>w</kbd> Moves forward to next beginning of a word
  - <kbd>b</kbd> Moves backwards to next beginning of a word
  - <kbd>i</kbd> Enters insert-mode
  - <kbd>I</kbd> Enters insert-mode at beginning of line
  - <kbd>a</kbd> Enters insert-mode then moves right 1 character
  - <kbd>A</kbd> Enters insert-mode then moves to the end of the line
  - <kbd>o</kbd> Enters insert-mode and creates a new line under the current one
  - <kbd>O</kbd> Enters insert-mode and creates a new line on top of the current one
  - <kbd>u</kbd> Undo last change
  - <kbd>C-r</kbd> Redo last change
  - <kbd>x</kbd> Delete char under cursor
  - <kbd>d d</kbd> Deletes line under cursor
  - <kbd>y y</kbd> Copies line under cursor
  - <kbd>p</kbd> Pastes from clipboard
  - <kbd>m $alpha</kbd> Set mark at cursor
  - <kbd>' $alpha</kbd> Jump to mark
  - <kbd>C-w s</kbd> Splits buffer horizontally
  - <kbd>C-w v</kbd> Splits buffer vertically
  - <kbd>C-w h</kbd> Move to the window to the left
  - <kbd>C-w j</kbd> Move to the window to the bottom
  - <kbd>C-w k</kbd> Move to the window to the top
  - <kbd>C-w l</kbd> Move to the window to the right
- Insert mode
  - <kbd>$any</kbd> Inserts character at cursor's position
  - <kbd>BAK</kbd> Deletes character to the left
  - <kbd>RET</kbd> Inserts a new line at cursor position
  - <kbd>ESC</kbd> Enters normal mode
- Prompt mode
  - <kbd>$any</kbd> Inserts character
  - <kbd>BAK</kbd> Deletes character
  - <kbd>C-c</kbd> Enters normal mode
  - <kbd>ESC</kbd> Enters normal mode
  - <kbd>RET</kbd> Execute command and go back to normal mode
  - <kbd>C-u</kbd> Clear entered command

**Currently implemented command**

- `edit <filename>` (aliased as `e`) Edit a file in a new buffer
- `write <filename?>` (aliased as `w`) Write buffer to disk, optionally setting it's path
- `quit` (aliased as `q`) Close current buffer (making sure it's saved before)
- `quit!` (aliased as `q!`) Close current buffer (ignoring unsaved changes)
- `writequit` (aliased as `wq`) Writes buffer to disk then closes it

### screenshot

![](https://raw.githubusercontent.com/kiasaki/ry/master/screenshot.png)

### license

See `LICENSE` file.

