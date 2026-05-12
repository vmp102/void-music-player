package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Volume        int      `json:"volume"`
	Shuffle       bool     `json:"shuffle"`
	Repeat        bool     `json:"repeat"`
	Loop          bool     `json:"loop"`
	Queue         []string `json:"queue"`
	CurrentPath   string   `json:"current_path"`
	QueueIdx      int      `json:"queue_idx"`
	Offset        int      `json:"offset"`
	CurrentTitle  string   `json:"current_title"`
	CurrentArtist string   `json:"current_artist"`
	SidebarCursor int      `json:"sidebar_cursor"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "vmp")
	os.MkdirAll(path, 0755)
	return filepath.Join(path, "config.json")
}

func saveConfig(m model) {
	curr, _ := getTimeInfo()
	c := Config{
		Volume:        m.volume,
		Shuffle:       m.shuffle,
		Repeat:        m.repeat,
		Loop:          m.loop,
		Queue:         m.displayQueue,
		CurrentPath:   m.currentPath,
		CurrentTitle:  m.currentTitle,
		CurrentArtist: m.currentArtist,
		QueueIdx:      m.queueIdx,
		Offset:        int(curr.Seconds()),
		SidebarCursor: m.cursor,
	}
	data, _ := json.Marshal(c)
	os.WriteFile(getConfigPath(), data, 0644)
}

func loadConfig() Config {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return Config{
			Volume:  100,
			Shuffle: false,
			Repeat:  false,
			Loop:    false,
			Queue:   []string{},
		}
	}
	var c Config
	json.Unmarshal(data, &c)
	return c
}
