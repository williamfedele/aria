package audio

type Track struct {
	TTitle, Path, Format string
}

func (t Track) Title() string       { return t.TTitle }
func (t Track) Description() string { return t.Path }
func (t Track) FilterValue() string { return t.TTitle + t.Path }
