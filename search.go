package main

import (
	"regexp"
	"strings"
)

var (
	last_search         = ""
	last_search_index   = 0
	last_search_results = []*location{}
)

func init_search() {
	bind("normal", k("/"), search_start)
	bind("normal", k("p"), search_prev)
	bind("normal", k("n"), search_next)
}

func search_start(vt *view_tree, b *buffer, kl *key_list) {
	prompt("/", noop_complete, func(args []string) {
		search := strings.Join(args, " ")
		if len(search) > 0 {
			last_search = search
			re := regexp.MustCompile(regexp.QuoteMeta(search))

			last_search_results = []*location{}
			for i, line := range b.data {
				idxs := re.FindAllStringIndex(string(line), -1)
				for idx := range idxs {
					last_search_results = append(last_search_results, new_location(i, idx))
				}
			}

			// TODO start index at first match after cursor
			last_search_index = len(last_search_results) - 1
			search_next(vt, b, kl)
		}
	})
}

func search_prev(vt *view_tree, b *buffer, kl *key_list) {
	if len(last_search_results) == 0 {
		message("No search result.")
		return
	}
	if last_search_index == 0 {
		last_search_index = len(last_search_results) - 1
	} else {
		last_search_index--
	}
	b.cursor = last_search_results[last_search_index].clone()
}

func search_next(vt *view_tree, b *buffer, kl *key_list) {
	if len(last_search_results) == 0 {
		message("No search result.")
		return
	}
	if last_search_index+1 == len(last_search_results) {
		last_search_index = 0
	} else {
		last_search_index++
	}
	b.cursor = last_search_results[last_search_index].clone()
}
