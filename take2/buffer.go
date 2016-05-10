package main

import "github.com/kiasaki/go-rope"

type Buffer struct {
	r *rope.Rope
}

func NewBuffer() *Buffer {
	return &Buffer{r: rope.New("")}
}

func (b *Buffer) Insert(text string) {

}
