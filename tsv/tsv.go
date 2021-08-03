package tsv

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/jtolds/qod"
	"github.com/kr/pretty"
)

func Rows(r io.Reader) (header []string, vals <-chan []string) {
	ch := make(chan []string)
	go func() {
		defer close(ch)
		br := bufio.NewReader(r)
		for {
			l, err := br.ReadString('\n')
			if err == io.EOF {
				if l != "" {
					ch <- strings.Split(strings.TrimRight(l, "\n"), "\t")
				}
				break
			}
			qod.ANE(err)
			ch <- strings.Split(strings.TrimRight(l, "\n"), "\t")
		}
	}()
	return <-ch, ch
}

func Lookup(header []string, name string) int {
	for i, h := range header {
		if h == name {
			return i
		}
	}
	panic("not found")
}

func WriteRow(w io.Writer, vals []string) {
	for i, val := range vals {
		if i > 0 {
			qod.AI(w.Write([]byte("\t")))
		}
		qod.AI(w.Write([]byte(strings.ReplaceAll(val, "\t", "        "))))
	}
	qod.AI(w.Write([]byte("\n")))
}

type Row struct {
	header []string
	vals   []string
	index  map[string]int
}

func (r Row) V(name string) string {
	if i, ok := r.index[name]; ok {
		return r.vals[i]
	}
	return ""
}

func (r Row) W(w io.Writer) {
	WriteRow(w, r.vals)
}

func (r Row) S(k, v string) Row {
	valscp := append([]string(nil), r.vals...)

	idx, ok := r.index[k]
	if !ok {
		indexcp := make(map[string]int, len(r.index))
		for ik, iv := range r.index {
			indexcp[ik] = iv
		}
		indexcp[k] = len(r.vals)
		return Row{
			header: append(append([]string(nil), r.header...), k),
			vals:   append(valscp, v),
			index:  indexcp,
		}
	}

	valscp[idx] = v
	return Row{
		header: r.header,
		vals:   valscp,
		index:  r.index,
	}
}

func (r Row) AsMap() map[string]string {
	rv := make(map[string]string, len(r.header))
	for i, key := range r.header {
		rv[key] = r.vals[i]
	}
	return rv
}

func FancyRows(r io.Reader) (header []string, rows <-chan Row) {
	header, vals := Rows(r)
	index := make(map[string]int, len(header))
	for i, name := range header {
		index[name] = i
	}
	ch := make(chan Row)
	go func() {
		defer close(ch)
		for row := range vals {
			if len(header) != len(row) {
				panic(pretty.Sprint("row length mismatch on:", os.Args[0], r, header, row))
			}
			ch <- Row{header: header, vals: row, index: index}
		}
	}()
	return header, ch
}
