package larkutils

import (
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
)

type TextBuilder struct {
	builder strings.Builder
}

func NewTextMsgBuilder() *TextBuilder {
	return &TextBuilder{
		builder: strings.Builder{},
	}
}

func (t *TextBuilder) Text(text string) *TextBuilder {
	t.builder.WriteString(text)
	return t
}

func (t *TextBuilder) AtUser(userId, name string) *TextBuilder {
	t.builder.WriteString("<at user_id=\"")
	t.builder.WriteString(userId)
	t.builder.WriteString("\">")
	t.builder.WriteString(name)
	t.builder.WriteString("</at>")
	return t
}

func AtUserString(openID string) string {
	return fmt.Sprintf("<at id=%s>某个用户</at>", openID)
}

func (t *TextBuilder) Build() string {
	tmpStruct := struct {
		Text string `json:"text"`
	}{
		Text: t.builder.String(),
	}
	s, err := sonic.MarshalString(tmpStruct)
	if err != nil {
		panic(err)
	}
	return s
}
