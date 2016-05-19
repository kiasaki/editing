package main

type ModeList struct {
	modes []*Mode
}

func NewModeList() *ModeList {
	return &ModeList{[]*Mode{}}
}

func (ml *ModeList) Add(m *Mode) {
	currentModes := ml.modes
	ml.modes = []*Mode{}
	for _, mode := range currentModes {
		if m.Kind == ModeEditing && mode.Kind == ModeEditing {
			continue
		}
		if m.Kind == ModeMajor && mode.Kind == ModeMajor {
			continue
		}
		if m.Name == mode.Name {
			// Skip same name mode so it's get replaced (not doubled)
			continue
		}
		ml.modes = append(ml.modes, mode)
	}
	ml.modes = append(ml.modes, m)
}

func (ml *ModeList) EditingMode() *Mode {
	for _, mode := range ml.modes {
		if mode.Kind == ModeEditing {
			return mode
		}
	}
	return nil
}

func (ml *ModeList) MajorMode() *Mode {
	for _, mode := range ml.modes {
		if mode.Kind == ModeMajor {
			return mode
		}
	}
	return nil
}

func (ml *ModeList) MinorModes() []*Mode {
	minorModes := []*Mode{}
	for _, mode := range ml.modes {
		if mode.Kind == ModeMinor {
			minorModes = append(minorModes, mode)
		}
	}
	return minorModes
}

// Passes key event to modes in order of priority
func (ml *ModeList) HandleEvent(w *World, b *Buffer, key *Key) bool {
	majorMode := ml.MajorMode()
	if majorMode != nil {
		if majorMode.HandleEvent(w, b, key) {
			// The major mode handled the key
			// return so that it doesn't get handled twice
			return true
		}
	}

	editingMode := ml.EditingMode()
	if editingMode != nil {
		if editingMode.HandleEvent(w, b, key) {
			return true
		}
	}

	for _, mode := range ml.MinorModes() {
		if mode.HandleEvent(w, b, key) {
			return true
		}
	}

	return false
}
