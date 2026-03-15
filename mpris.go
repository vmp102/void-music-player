package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

func initMPRIS(m *model) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return
	}

	reply, err := conn.RequestName("org.mpris.MediaPlayer2.vmp", dbus.NameFlagReplaceExisting)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		return
	}

	mp := &mprisPlayer{m: m}

	propsSpec := map[string]map[string]*prop.Prop{
		"org.mpris.MediaPlayer2": {
			"CanQuit":      {Value: true, Writable: false, Emit: prop.EmitTrue},
			"Identity":     {Value: "Void Music Player", Writable: false, Emit: prop.EmitTrue},
			"DesktopEntry": {Value: "vmp", Writable: false, Emit: prop.EmitTrue},
		},
		"org.mpris.MediaPlayer2.Player": {
			"PlaybackStatus": {Value: "Paused", Writable: false, Emit: prop.EmitTrue},
			"Volume":         {Value: 1.0, Writable: true, Emit: prop.EmitTrue},
			"Metadata":       {Value: map[string]dbus.Variant{}, Writable: false, Emit: prop.EmitTrue},
			"CanGoNext":      {Value: true, Writable: false, Emit: prop.EmitTrue},
			"CanGoPrevious":  {Value: true, Writable: false, Emit: prop.EmitTrue},
			"CanPlay":        {Value: true, Writable: false, Emit: prop.EmitTrue},
			"CanPause":       {Value: true, Writable: false, Emit: prop.EmitTrue},
			"CanControl":     {Value: true, Writable: false, Emit: prop.EmitTrue},
		},
	}

	props, err := prop.Export(conn, "/org/mpris/MediaPlayer2", propsSpec)
	if err != nil {
		return
	}

	conn.Export(mp, "/org/mpris/MediaPlayer2", "org.mpris.MediaPlayer2.Player")

	node := introspect.Node{
		Name: "/org/mpris/MediaPlayer2",
		Interfaces: []introspect.Interface{
			{
				Name:    "org.freedesktop.DBus.Introspectable",
				Methods: introspect.Methods(introspect.NewIntrospectable(&introspect.Node{})),
			},
			{
				Name:    "org.freedesktop.DBus.Properties",
				Methods: introspect.Methods(props),
			},
			{
				Name:    "org.mpris.MediaPlayer2.Player",
				Methods: introspect.Methods(mp),
			},
		},
	}
	conn.Export(introspect.NewIntrospectable(&node), "/org/mpris/MediaPlayer2", "org.freedesktop.DBus.Introspectable")
}

type mprisPlayer struct {
	m *model
}

func (p *mprisPlayer) PlayPause() *dbus.Error {
	if p.m.terminalProgram != nil {
		p.m.terminalProgram.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	}
	return nil
}

func (p *mprisPlayer) Next() *dbus.Error {
	if p.m.terminalProgram != nil {
		p.m.terminalProgram.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	}
	return nil
}

func (p *mprisPlayer) Previous() *dbus.Error {
	if p.m.terminalProgram != nil {
		p.m.terminalProgram.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	}
	return nil
}

func (p *mprisPlayer) Play() *dbus.Error  { return p.PlayPause() }
func (p *mprisPlayer) Pause() *dbus.Error { return p.PlayPause() }
func (p *mprisPlayer) Stop() *dbus.Error  { return nil }

// These are typically required for full MPRIS spec compliance
func (p *mprisPlayer) SetPosition(o dbus.ObjectPath, pos int64) *dbus.Error { return nil }
func (p *mprisPlayer) OpenUri(uri string) *dbus.Error                      { return nil }
func (p *mprisPlayer) Seek(offset int64) *dbus.Error                      { return nil }
