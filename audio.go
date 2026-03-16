package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/flac"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type Folder struct {
	Name  string
	Path  string
	Songs []string
}

var (
	ctrl          *beep.Ctrl
	volumeControl *effects.Volume
	streamer      beep.StreamSeekCloser
	format        beep.Format
)

func initAudio() {
	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Millisecond*200)) 
}

func setVolume(steps int) {
	if volumeControl == nil { return }
	speaker.Lock()
	
	if steps <= 0 {
		volumeControl.Volume = -100.0 
	} else {
		volumeControl.Volume = float64(steps-100) / 20.0
	}
	speaker.Unlock()
}

func playFile(path string, onDone func()) error {
	f, err := os.Open(path)
	if err != nil { return err }

	ext := strings.ToLower(filepath.Ext(path))
	var dErr error
	if ext == ".mp3" {
		streamer, format, dErr = mp3.Decode(f)
	} else {
		streamer, format, dErr = flac.Decode(f)
	}
	if dErr != nil { return dErr }

	resampled := beep.Resample(4, format.SampleRate, 44100, streamer)
	volumeControl = &effects.Volume{Streamer: resampled, Base: 2, Volume: 0}
	ctrl = &beep.Ctrl{Streamer: volumeControl, Paused: false}
	
	speaker.Clear()
	speaker.Play(beep.Seq(ctrl, beep.Callback(onDone)))
	return nil
}

func seekAudio(seconds int) {
	if streamer == nil { return }
	
	speaker.Lock()
	newPos := streamer.Position() + format.SampleRate.N(time.Duration(seconds)*time.Second)
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= streamer.Len() {
		newPos = streamer.Len() - 1
	}
	streamer.Seek(newPos)
	speaker.Unlock()
}

func getTimeInfo() (time.Duration, time.Duration) {
	if streamer == nil { return 0, 0 }
	cp := format.SampleRate.D(streamer.Position())
	tt := format.SampleRate.D(streamer.Len())
	return cp.Round(time.Second), tt.Round(time.Second)
}

func scanMusic() ([]Folder, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	musicPath := filepath.Join(home, "Music")
	
	if _, err := os.Stat(musicPath); os.IsNotExist(err) {
		altPath := filepath.Join(home, "Documents", "My Music")
		if _, err := os.Stat(altPath); err == nil {
			musicPath = altPath
		}
	}

	folderMap := make(map[string]*Folder)
	exts := map[string]bool{".mp3": true, ".flac": true}

	err = filepath.Walk(musicPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		ext := strings.ToLower(filepath.Ext(path))
		if !info.IsDir() && exts[ext] {
			dir := filepath.Dir(path)
			if _, exists := folderMap[dir]; !exists {
				folderMap[dir] = &Folder{
					Name: filepath.Base(dir), 
					Path: dir,
				}
			}
			folderMap[dir].Songs = append(folderMap[dir].Songs, path)
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}

	var folders []Folder
	for _, f := range folderMap {
		folders = append(folders, *f)
	}
	return folders, nil
}
