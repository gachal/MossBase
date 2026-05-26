package diff

import "github.com/sergi/go-diff/diffmatchpatch"

type DiffLine struct {
	Type  string // "added", "removed", "unchanged"
	Text  string
}

func Compute(oldText, newText string) []DiffLine {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(oldText, newText, true)
	var result []DiffLine
	for _, d := range diffs {
		var t string
		switch d.Type {
		case diffmatchpatch.DiffInsert:
			t = "added"
		case diffmatchpatch.DiffDelete:
			t = "removed"
		default:
			t = "unchanged"
		}
		result = append(result, DiffLine{Type: t, Text: d.Text})
	}
	return result
}
