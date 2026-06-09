package fg

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
	"google.golang.org/protobuf/proto"
)

// TextRun is one styled run of text inside a RichText. Build one with Span.
type TextRun struct {
	text     string
	size     float64
	color    Color
	colorSet bool
	bold     bool
}

// Span starts a styled text run with the given text.
func Span(text string) *TextRun {
	return &TextRun{text: text}
}

// Bold makes the run bold and returns it for chaining.
func (s *TextRun) Bold() *TextRun {
	s.bold = true

	return s
}

// Color sets the run color and returns it for chaining.
func (s *TextRun) Color(c Color) *TextRun {
	s.color = c
	s.colorSet = true

	return s
}

// Size sets the run font size in logical pixels and returns it for chaining.
func (s *TextRun) Size(v float64) *TextRun {
	s.size = v

	return s
}

func (s *TextRun) marshal() *fugov1.TextSpan {
	color := ""
	if s.colorSet {
		color = s.color.String()
	}

	return &fugov1.TextSpan{
		Text:     s.text,
		Color:    color,
		FontSize: s.size,
		Bold:     s.bold,
	}
}

// RichTextWidget renders text with mixed per-run styles. Build one with
// RichText, passing styled runs from Span.
type RichTextWidget struct {
	spans []*TextRun
	baseWidget
}

// RichText composes a paragraph from styled runs, e.g.
// fg.RichText(fg.Span("Hello ").Bold(), fg.Span("world").Color(fg.Colors.Red)).
func RichText(spans ...*TextRun) *RichTextWidget {
	return &RichTextWidget{spans: spans}
}

func (r *RichTextWidget) isWidget()                {}
func (r *RichTextWidget) widgetChildren() []Widget { return nil }

func (r *RichTextWidget) walkNodes(counter *uint32) []*fugov1.WidgetNode {
	*counter++
	r.id = *counter

	spans := make([]*fugov1.TextSpan, 0, len(r.spans))
	for _, s := range r.spans {
		spans = append(spans, s.marshal())
	}

	props, _ := proto.Marshal(&fugov1.RichTextProps{Spans: spans})

	return []*fugov1.WidgetNode{{
		Id:    r.id,
		Key:   r.key,
		Type:  fugov1.WidgetType_RICHTEXT,
		Props: props,
	}}
}
