package cardutil

type CardBuildHelper struct {
	Title    string
	SubTitle string
	Text     string
}

func NewCardBuildHelper() *CardBuildHelper {
	return &CardBuildHelper{}
}

func (h *CardBuildHelper) SetTitle(title string) *CardBuildHelper {
	h.Title = title
	return h
}

func (h *CardBuildHelper) SetSubTitle(subTitle string) *CardBuildHelper {
	h.SubTitle = subTitle
	return h
}

func (h *CardBuildHelper) SetText(text string) *CardBuildHelper {
	h.Text = text
	return h
}
