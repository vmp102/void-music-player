package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	gray    = lipgloss.Color("#AAAAAA")
	special = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	black   = lipgloss.Color("#000000")

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("#333333")).
			Padding(1, 2)

	midStyle = lipgloss.NewStyle().
			Padding(1, 2)
)

func renderSidebar(folders []Folder, cursor int, searching bool, query string, h int) string {
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
	if searching {
		bar := lipgloss.NewStyle().
			Foreground(special).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(lipgloss.Color("#333333")).
			Render("/ " + query + "█")
		s.WriteString("\n" + bar)
	}
	return paneStyle.Copy().Width(30).Height(h).Render(s.String())
}

func renderQueue(songs []string, queueIdx int, h int) string {
	var s strings.Builder
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(special).Render("QUEUE") + "\n\n")
	count := 0
	const maxDisplay = 15
	if queueIdx >= 0 && queueIdx < len(songs)-1 {
		for i := queueIdx + 1; i < len(songs); i++ {
			if count >= maxDisplay {
				s.WriteString(lipgloss.NewStyle().Foreground(gray).Render(". . .") + "\n")
				break
			}
			trackName := getTrackName(songs[i])
			s.WriteString(lipgloss.NewStyle().Foreground(gray).Render("- "+trackName) + "\n")
			count++
		}
	}
	if count == 0 {
		s.WriteString(lipgloss.NewStyle().Foreground(gray).Italic(true).Render("Empty"))
	}
	return paneStyle.Copy().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		PaddingRight(4).
		Width(30).
		Height(h).
		Render(s.String())
}

func formatDuration(d time.Duration) string {
	totalSeconds := int(d.Seconds())
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func renderPlayer(title string, artist string, currentPath string, playing bool, vol int, shuf bool, repeat bool, loop bool, h int, w int) string {
	state := "⏸ PAUSED"
	if playing {
		state = "▶ PLAYING"
	}
	folderName := "Unknown"
	if currentPath != "" {
		folderName = filepath.Base(filepath.Dir(currentPath))
	}

	curr, total := getTimeInfo()
	mainWidth := w - 60
	if mainWidth < 40 { mainWidth = 40 }
	barWidth := mainWidth - 16
	if barWidth < 10 { barWidth = 10 }

	percent := 0.0
	if total > 0 { percent = float64(curr) / float64(total) }
	filled := int(float64(barWidth) * percent)
	bar := lipgloss.NewStyle().Foreground(special).Render(strings.Repeat("━", filled)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(strings.Repeat("━", barWidth-filled))

	statusBox := lipgloss.NewStyle().Background(special).Foreground(black).Bold(true).Padding(0, 1).Render(state)
	folderBox := lipgloss.NewStyle().Background(special).Foreground(black).Bold(true).Padding(0, 1).MarginLeft(1).Render("󰉋 " + folderName)

	// Volume Icon Logic
	volIcon := " "
	if vol == 0 {
		volIcon = " "
	} else if vol <= 60 {
		volIcon = " "
	}

	// Volume Bar: Using  + Space for breathing room
	numOn := vol / 2
	numOff := 50 - numOn
	
	// We use the square  followed by a space " " to ensure they don't touch
	volOnStr := strings.Repeat(" ", numOn)
	volOffStr := strings.Repeat(" ", numOff)

	volBar := lipgloss.NewStyle().Foreground(special).Render(volOnStr) +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")).Render(volOffStr)

	// Mode icons logic (Shuffle, Repeat, Loop)
	shufIcon := lipgloss.NewStyle().Foreground(gray).Render("")
	if shuf { shufIcon = lipgloss.NewStyle().Foreground(special).Render("") }
	
	repIcon := lipgloss.NewStyle().Foreground(gray).Render("")
	if repeat { repIcon = lipgloss.NewStyle().Foreground(special).Render("") }
	
	loopIcon := lipgloss.NewStyle().Foreground(gray).Render("")
	if loop { loopIcon = lipgloss.NewStyle().Foreground(special).Render("") }

	// Create the modes block with spacing between icons
	modes := lipgloss.NewStyle().MarginLeft(3).Render(
		fmt.Sprintf("%s   %s   %s", shufIcon, repIcon, loopIcon),
	)

	artistStyle := lipgloss.NewStyle().Foreground(gray).Italic(true)
	
	// Combine the volume icon, the 50 squares, percentage, and the mode icons
	volumeLine := lipgloss.JoinHorizontal(lipgloss.Center, 
		fmt.Sprintf("%s %s %d%%", volIcon, volBar, vol),
		modes,
	)

	return midStyle.Copy().Width(mainWidth).Height(h).Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center, statusBox, folderBox),
		"\n",
		lipgloss.NewStyle().Bold(true).Foreground(special).Render(title),
		artistStyle.Render("by "+artist),
		"\n",
		fmt.Sprintf("%s %s %s", formatDuration(curr), bar, formatDuration(total)),
		"\n",
		volumeLine,
	))
}
