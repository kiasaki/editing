# Todo

- Progressive redisplay
- Make KeyPress.String() handle Ctrl-...

# The big list

- [ ] Auto indent
- [ ] Custom bindings
- [ ] Line numbers
- [ ] Search and replace
  - [ ] Search
  - [ ] Replace
- [ ] Tests
- [ ] Error handling
  - [ ] Fatal
  - [ ] Script/Runtime
  - [ ] User
- [ ] Unicode support
- [ ] Command execution
- [ ] Movement keys
- [ ] Visual mode
  - [ ] Normal
  - [ ] Line
- [ ] Help
  - [ ] General
  - [ ] Per command
  - [ ] Tutorial
- [ ] Options/Configuration
  - [ ] Save/Load
  - [ ] Tabs to space
  - [ ] Tab size
  - [ ] Color scheme
- [ ] Undo/Redo
- [ ] Clipboard
  - [ ] Copy
  - [ ] Paste
  - [ ] Cut
  - [ ] Registers
- [ ] Macros
- [ ] Syntax highlighting
- [ ] Color schemes

# Vim's Perl Interface for inspiration:

```
VIM::Msg({msg}, {group}?)
VIM::SetOption({arg})             Sets a vim option.
VIM::Buffers([{bn}...])           With no arguments, returns a list of all the buffers.
VIM::Windows([{wn}...])           With no arguments, returns a list of all the windows.
VIM::DoCommand({cmd})             Executes Ex command {cmd}.
VIM::Eval({expr})                 Evaluates {expr} and returns (success, val).
Window->SetHeight({height})
Window->Cursor({row}?, {col}?)
Window->Buffer()
Buffer->Name()                    Returns the filename for the Buffer.
Buffer->Number()                  Returns the number of the Buffer.
Buffer->Count()                   Returns the number of lines in the Buffer.
Buffer->Get({lnum}, {lnum}?, ...)
Buffer->Delete({lnum}, {lnum}?)
Buffer->Append({lnum}, {line}, {line}?, ...)
Buffer->Set({lnum}, {line}, {line}?, ...)
$main::curwin
$main::curbuf
```
