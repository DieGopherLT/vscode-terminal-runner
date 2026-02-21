package workspace

// workspaceSaveResultMsg is sent after async workspace save operations.
type workspaceSaveResultMsg struct {
	err error
}
