package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Volume      int      `json:"volume"`
	Shuffle     bool     `json:"shuffle"`
	Queue       []string `json:"queue"`
	CurrentPath string   `json:"current_path"`
	QueueIdx    int      `json:"queue_idx"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "vmp")
	os.MkdirAll(path, 0755)
	return filepath.Join(path, "config.json")
}

func saveConfig(m model) {
	c := Config{
		Volume:      m.volume,
		Shuffle:     m.shuffle,
		Queue:       m.masterQueue,
		CurrentPath: m.currentPath,
		QueueIdx:    m.queueIdx,
	}
	data, _ := json.Marshal(c)
	os.WriteFile(getConfigPath(), data, 0644)
}

func loadConfig() Config {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return Config{Volume: 40, Shuffle: true}
	}
	var c Config
	json.Unmarshal(data, &c)
	return c
}
