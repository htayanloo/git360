package tui

import (
	"fmt"
	"strings"
)

type BranchBrowser struct {
	Items  []string
	IsTags bool
	Cursor int
	Offset int
}

func NewBranchBrowser() *BranchBrowser {
	return &BranchBrowser{
		Items:  []string{},
		Cursor: 0,
		Offset: 0,
	}
}

func (bb *BranchBrowser) SetItems(items []string, isTags bool) {
	bb.Items = items
	bb.IsTags = isTags
	bb.Cursor = 0
	bb.Offset = 0
}

func (bb *BranchBrowser) ScrollUp() {
	if bb.Cursor > 0 {
		bb.Cursor--
	}
}

func (bb *BranchBrowser) ScrollDown() {
	if bb.Cursor < len(bb.Items)-1 {
		bb.Cursor++
	}
}

func (bb *BranchBrowser) SelectedItem() string {
	if len(bb.Items) == 0 || bb.Cursor < 0 || bb.Cursor >= len(bb.Items) {
		return ""
	}
	return bb.Items[bb.Cursor]
}

func (bb *BranchBrowser) View(styles Styles, width int, height int) string {
	if len(bb.Items) == 0 {
		if bb.IsTags {
			return "No tags found."
		}
		return "No branches found."
	}

	// Adjust offset for scrolling
	if bb.Cursor < bb.Offset {
		bb.Offset = bb.Cursor
	} else if bb.Cursor >= bb.Offset+height-2 {
		bb.Offset = bb.Cursor - height + 3
	}
	if bb.Offset < 0 {
		bb.Offset = 0
	}

	var sb strings.Builder
	
	// Draw list
	end := bb.Offset + height
	if end > len(bb.Items) {
		end = len(bb.Items)
	}

	for i := bb.Offset; i < end; i++ {
		item := bb.Items[i]
		isSelected := i == bb.Cursor

		icon := "  "
		if bb.IsTags {
			icon = "🏷️ "
		} else {
			if strings.HasPrefix(item, "origin/") || strings.Contains(item, "/") {
				icon = "📡 "
			} else {
				icon = "🌿 "
			}
		}

		itemName := item
		maxItemLen := width - 10
		if maxItemLen > 5 && len(itemName) > maxItemLen {
			itemName = itemName[:maxItemLen-3] + "..."
		}

		lineContent := fmt.Sprintf("  %s %s", icon, itemName)

		if isSelected {
			sb.WriteString(styles.SelectedLineStyle.Width(width).Render(lineContent) + "\n")
		} else {
			sb.WriteString(styles.NormalLineStyle.Render(lineContent) + "\n")
		}
	}

	return sb.String()
}
