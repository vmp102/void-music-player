package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"github.com/charmbracelet/lipgloss"
)

var (
	gray    = lipgloss.Color("#AAAAAA")
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	black   = lipgloss.Color("#000000")

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(1, 2).
			Width(35).
			Height(20)

	midStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Width(50).
			Height(20)
)

func renderSidebar(folders []Folder, cursor int) string {
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(special).Render("FOLDERS") + "\n\n")
	
	for i, f := range folders {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(gray)
		if cursor == i {
			prefix = "> "
			style = style.Foreground(special).Bold(true)
		}
		s.WriteString(style.Render(prefix+f.Name) + "\n")
	}
	return paneStyle.Render(s.String())
}

func renderQueue(songs []string, currentPath string) string {
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(special).Render("QUEUE") + "\n\n")
	
	count, found := 0, false
	for _, song := range songs {
		if song == currentPath {
			found = true
			continue
		}
		if found && count < 15 {
			s.WriteString(lipgloss.NewStyle().Foreground(gray).Render("- "+filepath.Base(song)) + "\n")
			count++
		}
	}
	
	if count == 0 && !found {
		s.WriteString(lipgloss.NewStyle().Foreground(gray).Italic(true).Render("Empty"))
	}

	return paneStyle.Copy().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Render(s.String())
}

func renderPlayer(song string, playing bool, vol int, shuf bool) string {
	state := "⏸ PAUSED"; if playing { state = "▶ PLAYING" }
	shufLabel := "OFF"; if shuf { shufLabel = "ON" }
	curr, total := getTimeInfo()
	
	percent := 0.0
	if total > 0 { percent = float64(curr) / float64(total) }
	barWidth := 30
	filled := int(float64(barWidth) * percent)
	
	bar := lipgloss.NewStyle().Foreground(special).Render(strings.Repeat("━", filled)) + 
	       lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("━", barWidth-filled))

	keyStyle := lipgloss.NewStyle().Foreground(gray)
	help := lipgloss.JoinVertical(lipgloss.Left,
		"\n",
		keyStyle.Render("[j/k]   Navigation    [/]   Search"),
		keyStyle.Render("[Enter] Play Folder   [e/r] Seek 5s"),
		keyStyle.Render("[p]     Pause"),
		keyStyle.Render("[b/n]   Prev/Next"),
		keyStyle.Render("[+/-]   Vol Up/Down"),
		keyStyle.Render("[s]     Shuffle"),
		keyStyle.Render("[q]     Quit"),
	)

	return midStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Background(special).Foreground(black).Bold(true).Render(" "+state+" "),
		"\n",
		lipgloss.NewStyle().Bold(true).Foreground(special).Render(song),
		"\n",
		fmt.Sprintf("%s %s %s", curr.String(), bar, total.String()),
		"\n",
		fmt.Sprintf("Vol: %d%% | Shuffle: %s", vol, shufLabel),
		help,
	))
}
