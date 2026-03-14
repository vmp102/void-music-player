package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep/speaker"
)

type nextSongMsg struct{}
type tickMsg time.Time

type model struct {
	folders      []Folder
	cursor       int
	playing      bool
	volume       int
	shuffle      bool
	currentSong  string
	currentPath  string
	masterQueue  []string 
	displayQueue []string 
	queueIdx     int
}

var p *tea.Program

func initialModel() model {
	f, _ := scanMusic()
	sort.Slice(f, func(i, j int) bool { return f[i].Name < f[j].Name })
	
	conf := loadConfig()
	
	return model{
		folders:      f,
		volume:       conf.Volume,
		shuffle:      conf.Shuffle,
		masterQueue:  conf.Queue,
		currentPath:  conf.CurrentPath,
		queueIdx:     conf.QueueIdx,
		displayQueue: conf.Queue, 
	}
}

func (m model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd { 
	return tea.Every(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) }) 
}

func (m *model) playCurrent() {
	if len(m.displayQueue) > 0 {
		m.currentPath = m.displayQueue[m.queueIdx]
		m.currentSong = filepath.Base(m.currentPath)
		
		playFile(m.currentPath, func() {
			if p != nil {
				p.Send(nextSongMsg{})
			}
		})
		
		m.playing = true
		setVolume(m.volume)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nextSongMsg:
		if len(m.displayQueue) > 0 {
			m.queueIdx = (m.queueIdx + 1) % len(m.displayQueue)
			m.playCurrent()
		}
		return m, nil

	case tickMsg:
		return m, tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			saveConfig(m)
			return m, tea.Quit
		case "j": 
			if m.cursor < len(m.folders)-1 { m.cursor++ }
		case "k": 
			if m.cursor > 0 { m.cursor-- }
		case "n":
			if len(m.displayQueue) > 0 {
				m.queueIdx = (m.queueIdx + 1) % len(m.displayQueue)
				m.playCurrent()
			}
		case "b":
			if len(m.displayQueue) > 0 {
				m.queueIdx = (m.queueIdx - 1 + len(m.displayQueue)) % len(m.displayQueue)
				m.playCurrent()
			}
		case "s":
			m.shuffle = !m.shuffle
			if len(m.masterQueue) > 0 {
				if m.shuffle {
					shuf := make([]string, len(m.masterQueue))
					copy(shuf, m.masterQueue)
					rand.Seed(time.Now().UnixNano())
					rand.Shuffle(len(shuf), func(i, j int) { shuf[i], shuf[j] = shuf[i], shuf[j] })
					m.displayQueue = shuf
				} else {
					m.displayQueue = make([]string, len(m.masterQueue))
					copy(m.displayQueue, m.masterQueue)
				}
				for i, path := range m.displayQueue {
					if path == m.currentPath { m.queueIdx = i; break }
				}
			}
		case "enter":
			sel := m.folders[m.cursor]
			songs := make([]string, len(sel.Songs))
			copy(songs, sel.Songs)
			sort.Slice(songs, func(i, j int) bool { return filepath.Base(songs[i]) < filepath.Base(songs[j]) })
			m.masterQueue = songs
			if m.shuffle {
				m.displayQueue = make([]string, len(songs))
				copy(m.displayQueue, songs)
				rand.Seed(time.Now().UnixNano())
				rand.Shuffle(len(m.displayQueue), func(i, j int) { m.displayQueue[i], m.displayQueue[j] = m.displayQueue[j], m.displayQueue[i] })
			} else {
				m.displayQueue = songs
			}
			m.queueIdx = 0
			m.playCurrent()
		case "+", "=":
			if m.volume < 100 { m.volume += 5; setVolume(m.volume) }
		case "-", "_":
			if m.volume > 0 { m.volume -= 5; setVolume(m.volume) }
		case "p":
			if ctrl != nil {
				speaker.Lock()
				ctrl.Paused = !ctrl.Paused
				m.playing = !ctrl.Paused
				speaker.Unlock()
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		renderSidebar(m.folders, m.cursor),
		renderPlayer(m.currentSong, m.playing, m.volume, m.shuffle),
		renderQueue(m.displayQueue, m.currentPath),
	)
}

func main() {
	initAudio()
	p = tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
