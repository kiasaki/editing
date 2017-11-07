#!/usr/bin/env python3
import os
import sys
import curses

class Editor:
    def __init__(self):
        self.last_key = ""

    def open_file(self, path):
        pass

    def show(self):
        self.screen.clear()
        self.screen.addstr(0, 0, "Hello world" + self.last_key)
        self.screen.refresh()

    def read_input(self):
        key = self.screen.getkey()
        self.last_key = key
        if key == "\x03": # Ctrl-C
            self.quit()
            sys.exit(0)

    def run(self):
        self.running = True
        self.screen = curses.initscr()
        self.screen.keypad(True)
        curses.curs_set(False)
        curses.noecho()
        curses.cbreak()
        curses.raw()

        try:
            while self.running:
                self.show()
                self.read_input()
        except Exception as e:
            self.quit()
            raise e

    def quit(self):
        curses.noraw()
        curses.nocbreak()
        curses.echo()
        curses.curs_set(True)
        self.screen.keypad(False)
        curses.endwin()
        self.running = False

def run_rc_file(editor):
    config_file_path = os.path.expanduser('~/.ryrc')
    if not os.path.exists(config_file_path):
        return
    try:
        namespace = {}
        with open(config_file_path, 'r') as f:
            code = compile(f.read(), config_file_path, 'exec')
            exec(code, namespace, namespace)
        if 'configure' in namespace:
            namespace['configure'](editor)
    except Exception as e:
        traceback.print_exc()
        input('Press any key to continue...')

def main():
    editor = Editor()
    run_rc_file(editor)

    for arg in sys.argv[1:]:
        editor.open_file(arg)

    editor.run()

if __name__ == "__main__":
    main()
