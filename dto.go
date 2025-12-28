package main

type BlockDto struct {
	Id      string `json:"id"`
	Content string `json:"content"`
	Indent  int    `json:"indent"`
}

type DocumentDto struct {
	Id     string     `json:"id"`
	Title  string     `json:"title"`
	Blocks []BlockDto `json:"blocks"`
}
