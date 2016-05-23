/*
  Nukata Lisp Light 1.4 in Go 1.6 by SUZUKI Hisao (H27.5/11, H28.2/22)
  This is a Lisp interpreter written in Go.
  It differs from the previous version(*1) in that all numbers are
  64-bit floats and the whole interpreter consists of only one file.
  It is a "light" version.
  Intentionally it implements the same language as Nukata Lisp Light
  1.23 in TypeScript 1.7(*2) except that it has also two concurrent
  constructs, future and force.  See *3.
    *1: http://www.oki-osk.jp/esc/golang/lisp3.html (in Japanese)
    *2: http://www.oki-osk.jp/esc/typescript/lisp-en.html
    *3: http://www.oki-osk.jp/esc/golang/lisp4-en.html
*/
package lang
