package display

type Window struct {
}

func WindowNew() *Window {
	return &Window{}
}

func (w *Window) Init() error {
	return nil
}

func (w *Window) End() error {
	return nil
}

// Ensure current buffer contents are still visible on screen
// if not scroll down/up to prefered % (or center)
func (w *Window) Frame() {

}

// Executes incremental redisplay stopping if
// user is still typing, then picking up
func (w *Window) Redisplay() {
}

// Works as Redisplay but also centers point to prefered screen %
func (w *Window) Recenter() {
}

// Forced full display
func (w *Window) Refresh() {
}

// SetPreferredPercentage(perc)
// GetPointRow() - on screeen
// GetPointCol() - taking in account line wrap + ui
// WindowCreate(w)
// WindowDestroy(w)
// WindowGrow(w, amount)
// GetWindowTopLine(w) - on screen
// GetWindowBottomLine(w)
// GetWindowTop(w) Position - buffer position at visible top left
// GetWindowBottom(w) Position - buffer position at visible bottom right
