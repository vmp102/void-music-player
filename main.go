package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dhowden/tag"
)

type nextSongMsg struct{}
type tickMsg time.Time

type model struct {
	terminalProgram *tea.Program

	folders       []Folder
	cursor        int
	playing       bool
	volume        int
	shuffle       bool
	currentSong   string
	currentPath   string
	currentTitle  string
	currentArtist string
	displayQueue  []string
	queueIdx      int
	searching     bool
	searchQuery   string
	allFolders    []Folder

	termWidth  int
	termHeight int
}

var p *tea.Program

func initialModel() model {
	f, _ := scanMusic()
	sort.Slice(f, func(i, j int) bool { return f[i].Name < f[j].Name })

	conf := loadConfig()

	m := model{
		folders:       f,
		allFolders:    f,
		cursor:        conf.SidebarCursor,
		volume:        conf.Volume,
		shuffle:       conf.Shuffle,
		displayQueue:  conf.Queue,
		currentPath:   conf.CurrentPath,
		currentTitle:  conf.CurrentTitle,
		currentArtist: conf.CurrentArtist,
		queueIdx:      conf.QueueIdx,
		playing:       false,
	}

	if m.currentPath != "" {
		playFile(m.currentPath, func() {
			if p != nil {
				p.Send(nextSongMsg{})
			}
		})
		pauseAudio()
		setVolume(m.volume)
		seekAudio(conf.Offset)
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Every(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

	case tickMsg:
		// Auto-save every 10 seconds
		t := time.Time(msg)
		if t.Second()%10 == 0 {
			saveConfig(m)
		}
		return m, tick()

	case nextSongMsg:
		if m.queueIdx < len(m.displayQueue)-1 {
			m.queueIdx++
			m.playCurrent()
		} else {
			m.playing = false
		}
		return m, nil

	case tea.KeyMsg:
		s := msg.String()

		if m.searching {
			if s == "enter" || s == "esc" {
				m.searching = false
				return m, nil
			}
			if s == "backspace" {
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
			} else if len(s) == 1 {
				m.searchQuery += s
			}

			m.folders = nil
			for _, f := range m.allFolders {
				if strings.Contains(strings.ToLower(f.Name), strings.ToLower(m.searchQuery)) {
					m.folders = append(m.folders, f)
				}
			}
			m.cursor = 0
			return m, nil
		}

		switch s {
		case KeyQuit, "ctrl+c":
			saveConfig(m)
			return m, tea.Quit

		case KeySearch:
			m.searching = true
			return m, nil

		case KeyClearSearch:
			m.searchQuery = ""
			m.folders = m.allFolders
			m.cursor = 0
			return m, nil

		case KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}
		case KeyDown:
			if m.cursor < len(m.folders)-1 {
				m.cursor++
			}

		case KeyPause, KeyPauseAlt, " ":
			m.playing = !m.playing
			if m.playing {
				resumeAudio()
			} else {
				pauseAudio()
			}

		case KeyNext:
			if m.queueIdx < len(m.displayQueue)-1 {
				m.queueIdx++
				m.playCurrent()
			}

		case KeyPrev:
			if m.queueIdx > 0 {
				m.queueIdx--
				m.playCurrent()
			}

		case KeySeekBack:
			seekAudio(-5)
		case KeySeekForward:
			seekAudio(5)

		case KeyShuffle:
			m.shuffle = !m.shuffle

		case KeyClearQueue:
			m.displayQueue = nil
			m.queueIdx = -1
			if m.currentPath != "" {
				m.displayQueue = []string{m.currentPath}
				m.queueIdx = 0
			}

		case KeyPlayFolder:
			if len(m.folders) > 0 {
				sel := m.folders[m.cursor]
				songs := make([]string, len(sel.Songs))
				copy(songs, sel.Songs)
				sort.Slice(songs, func(i, j int) bool {
					return filepath.Base(songs[i]) < filepath.Base(songs[j])
				})

				m.displayQueue = songs

				if m.shuffle {
					rand.Seed(time.Now().UnixNano())
					rand.Shuffle(len(m.displayQueue), func(i, j int) {
						m.displayQueue[i], m.displayQueue[j] = m.displayQueue[j], m.displayQueue[i]
					})
				}
				m.queueIdx = 0
				m.playCurrent()
			}

		case KeyVolUp:
			if m.volume < 130 {
				m.volume += 5
				if m.volume > 130 {
					m.volume = 130
				}
				setVolume(m.volume)
			}
		case KeyVolDown:
			if m.volume > 0 {
				m.volume -= 5
				if m.volume < 0 {
					m.volume = 0
				}
				setVolume(m.volume)
			}
		}
	}
	return m, nil
}

func (m *model) playCurrent() {
	if m.queueIdx < 0 || m.queueIdx >= len(m.displayQueue) {
		return
	}
	path := m.displayQueue[m.queueIdx]
	m.currentPath = path
	m.playing = true

	m.currentTitle = getTrackName(path)
	m.currentArtist = "Unknown Artist"

	f, err := os.Open(path)
	if err == nil {
		tags, err := tag.ReadFrom(f)
		if err == nil {
			if tags.Title() != "" {
				m.currentTitle = tags.Title()
			}
			if tags.Artist() != "" {
				m.currentArtist = tags.Artist()
			}
		}
		f.Close()
	}

	playFile(path, func() {
		if m.terminalProgram != nil {
			m.terminalProgram.Send(nextSongMsg{})
		}
	})
	setVolume(m.volume)
}

func (m model) View() string {
	sidebar := renderSidebar(m.folders, m.cursor, m.searching, m.searchQuery, m.termHeight-2)
	player := renderPlayer(m.currentTitle, m.currentArtist, m.currentPath, m.playing, m.volume, m.shuffle, m.termHeight-2, m.termWidth)
	queue := renderQueue(m.displayQueue, m.queueIdx, m.termHeight-2)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, player, queue)
}

func main() {
	initAudio()
	m := initialModel()
	p = tea.NewProgram(m, tea.WithAltScreen())
	m.terminalProgram = p

	initMPRIS(&m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func getTrackName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}
