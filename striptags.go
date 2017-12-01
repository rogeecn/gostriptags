package striptags

import (
	"bytes"
	"golang.org/x/net/html"
	//"code.google.com/p/go.net/html"
	"io"
	"reflect"
	"strings"
)

const escapedChars = "&'<>\"\r"

func escape(w *bytes.Buffer, s string) error {
	i := strings.IndexAny(s, escapedChars)
	for i != -1 {
		if _, err := w.WriteString(s[:i]); err != nil {
			return err
		}
		var esc string
		switch s[i] {
		case '&':
			esc = "&amp;"
		case '\'':
			// "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
			esc = "&#39;"
		case '<':
			esc = "&lt;"
		case '>':
			esc = "&gt;"
		case '"':
			// "&#34;" is shorter than "&quot;".
			esc = "&#34;"
		case '\r':
			esc = "&#13;"
		default:
			panic("unrecognized escape character")
		}
		s = s[i+1:]
		if _, err := w.WriteString(esc); err != nil {
			return err
		}
		i = strings.IndexAny(s, escapedChars)
	}
	_, err := w.WriteString(s)
	return err
}

type StripTags struct {
	EscapeInValid bool                   // escape invalid tag
	TrimSpace     bool                   // trim html space
	ValidTags     map[string]interface{} // default validation tag, see also this.Init method
	ValidAttrs    map[string]bool        // default validation attribute
	DisableAttrs  map[string]bool        // default disabled attribute
	buf           *bytes.Buffer
}

func (this *StripTags) Init() {
	if this.ValidTags == nil {
		this.ValidTags = map[string]interface{}{
			"a": map[string]interface{}{
				"href": func(v string) bool {
					// if return true,this attr will be deleted
					// false will be kept
					return strings.HasPrefix(v, "javascript:")
				},
			},
			"abbr":    true,
			"address": true,
			"article": true,

			"audio":      true,
			"b":          true,
			"blockquote": true,
			"br":         true,
			"button":     true,
			"caption":    true,
			"code":       true,
			"cite":       true,
			"div":        true,
			"dl":         true,
			"dt":         true,
			"dd":         true,
			"del":        true,
			"em":         true,

			"h1": true,
			"h2": true,
			"h3": true,
			"h4": true,
			"h5": true,
			"h6": true,

			"hr":     true,
			"i":      true,
			"kbd":    true,
			"li":     true,
			"ol":     true,
			"p":      true,
			"pre":    true,
			"small":  true,
			"span":   true,
			"strong": true,
			"sub":    true,

			"table": true,
			"thead": true,
			"tbody": true,
			"tfoot": true,
			"tr":    true,
			"th":    true,
			"td":    true,

			"time":  true,
			"u":     true,
			"ul":    true,
			"video": true,
			"img":   true,
		}
	}

	if this.ValidAttrs == nil {
		this.ValidAttrs = map[string]bool{
			"title": true, "id": true,
			"class": true, "alt": true,
			"rel": true, "valign": true,
			"align": true, "rowspan": true,
			"colspan": true,
		}
	}

	if this.DisableAttrs == nil {
		this.DisableAttrs = map[string]bool{
			"onclick": true, "onerror": true,
		}
	}
	if this.buf == nil {
		this.buf = bytes.NewBuffer([]byte{})
	}
}

func (this *StripTags) handleAttr(token *html.Token) {
	if len(token.Attr) == 0 {
		return
	}
	tag_name := token.DataAtom.String()
	ref := reflect.ValueOf(this.ValidTags[tag_name])
	var (
		attrs      []html.Attribute
		valid_attr bool
	)

	for _, attr := range token.Attr {

		if _, ok := this.DisableAttrs[attr.Key]; ok {
			continue
		}
		if _, valid_attr = this.ValidAttrs[attr.Key]; valid_attr {

		} else {
			attr_config_type := ref.Type().Kind()

			if attr_config_type == reflect.Map {
				attr_config := ref.Interface().(map[string]interface{})
				switch attr_config[attr.Key].(type) {
				case bool:
					valid_attr = attr_config[attr.Key].(bool)
					break
				case func(string) bool:
					f := attr_config[attr.Key].(func(string) bool)
					valid_attr = !f(attr.Val)
					break
				}
			} else if attr_config_type == reflect.Bool {
				valid_attr = ref.Interface() == true
			}
		}
		if valid_attr {
			attrs = append(attrs, attr)
		}
	}
	token.Attr = attrs
}
func (this *StripTags) handleTag(token *html.Token) {
	_, ok := this.ValidTags[token.Data]
	if ok {
		this.handleAttr(token)
		this.buf.WriteString(token.String())
		return
	}

	if !this.EscapeInValid {
		return
	}
	this.buf.WriteString("&lt;")
	if token.Type == html.EndTagToken {
		this.buf.WriteString("/")
	}
	this.buf.WriteString(token.DataAtom.String())
	if len(token.Attr) > 0 {
		for _, attr := range token.Attr {
			this.buf.WriteByte(' ')
			this.buf.WriteString(attr.Key)
			this.buf.WriteString(`="`)
			escape(this.buf, attr.Val)
			this.buf.WriteByte('"')
		}
	}
	if token.Type == html.SelfClosingTagToken {
		this.buf.WriteString("/")
	}
	this.buf.WriteString("&gt;")

}

func (this *StripTags) Fetch(html_str string) (string, error) {
	this.Init()

	reader := bytes.NewReader([]byte(strings.TrimSpace(html_str)))
	tokenizer := html.NewTokenizer(reader)

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return "", err
		}

		token := tokenizer.Token()

		switch token.Type {
		case html.StartTagToken, html.SelfClosingTagToken:
			this.handleTag(&token)
			break
		case html.TextToken:
			var clean_text string = token.String()
			if this.TrimSpace {
				clean_text = strings.TrimSpace(clean_text)
			}
			if clean_text != "" {
				this.buf.WriteString(clean_text)
			}
			break
		case html.EndTagToken:
			this.handleTag(&token)
			break
		}
	}
	return this.buf.String(), nil
}

func NewStripTags() *StripTags {
	return new(StripTags)
}
