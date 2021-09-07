package markdown

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/pgavlin/goldmark/ast"
	"github.com/pgavlin/goldmark/renderer"
	"github.com/pgavlin/goldmark/text"
	"github.com/pgavlin/goldmark/util"
)

type blockState struct {
	node  ast.Node
	fresh bool
}

type listState struct {
	marker  byte
	ordered bool
	index   int
}

// Renderer is a goldmark renderer that produces Markdown output. Due to information loss in goldmark, its output may
// not be textually identical to the source that produced the AST to be rendered, but the structure should match.
//
// NodeRenderers that want to override rendering of particular node types should write through the Write* functions
// provided by Renderer in order to retain proper indentation and prefices inside of lists and block quotes.
type Renderer struct {
	listStack []listState

	openBlocks []blockState

	prefixStack []string
	prefix      []byte
	atNewline   bool
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// blocks
	reg.Register(ast.KindDocument, r.RenderDocument)
	reg.Register(ast.KindHeading, r.RenderHeading)
	reg.Register(ast.KindBlockquote, r.RenderBlockquote)
	reg.Register(ast.KindCodeBlock, r.RenderCodeBlock)
	reg.Register(ast.KindFencedCodeBlock, r.RenderFencedCodeBlock)
	reg.Register(ast.KindHTMLBlock, r.RenderHTMLBlock)
	reg.Register(ast.KindLinkReferenceDefinition, r.RenderLinkReferenceDefinition)
	reg.Register(ast.KindList, r.RenderList)
	reg.Register(ast.KindListItem, r.RenderListItem)
	reg.Register(ast.KindParagraph, r.RenderParagraph)
	reg.Register(ast.KindTextBlock, r.RenderTextBlock)
	reg.Register(ast.KindThematicBreak, r.RenderThematicBreak)

	// inlines
	reg.Register(ast.KindAutoLink, r.RenderAutoLink)
	reg.Register(ast.KindCodeSpan, r.RenderCodeSpan)
	reg.Register(ast.KindEmphasis, r.RenderEmphasis)
	reg.Register(ast.KindImage, r.RenderImage)
	reg.Register(ast.KindLink, r.RenderLink)
	reg.Register(ast.KindRawHTML, r.RenderRawHTML)
	reg.Register(ast.KindText, r.RenderText)
	reg.Register(ast.KindString, r.RenderString)
	reg.Register(ast.KindWhitespace, r.RenderWhitespace)
}

func (r *Renderer) beginLine(w io.Writer) error {
	if len(r.openBlocks) != 0 {
		current := r.openBlocks[len(r.openBlocks)-1]
		if current.node.Kind() == ast.KindParagraph && !current.fresh {
			return nil
		}
	}

	n, err := w.Write(r.prefix)
	if n != 0 {
		r.atNewline = r.prefix[len(r.prefix)-1] == '\n'
	}
	return err
}

func (r *Renderer) writeLines(w util.BufWriter, source []byte, lines *text.Segments) error {
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		if _, err := r.Write(w, line.Value(source)); err != nil {
			return err
		}
	}
	return nil
}

type writer struct {
	r *Renderer
	w io.Writer
}

func (w *writer) Write(b []byte) (int, error) {
	return w.r.Write(w.w, b)
}

// Writer returns an io.Writer that uses the Renderer's Write method to ensure appropriate indentation and prefices
// are added at the beginning of each line.
func (r *Renderer) Writer(w io.Writer) io.Writer {
	return &writer{r: r, w: w}
}

// Write writes a slice of bytes to an io.Writer, ensuring that appropriate indentation and prefices
// are added at the beginning of each line.
func (r *Renderer) Write(w io.Writer, buf []byte) (int, error) {
	written := 0
	for len(buf) > 0 {
		if r.atNewline {
			if err := r.beginLine(w); err != nil {
				return 0, err
			}
		}

		atNewline := false
		newline := bytes.IndexByte(buf, '\n')
		if newline == -1 {
			newline = len(buf) - 1
		} else {
			atNewline = true
		}

		n, err := w.Write(buf[:newline+1])
		written += n
		r.atNewline = n > 0 && atNewline && n == newline+1
		if len(r.openBlocks) != 0 {
			r.openBlocks[len(r.openBlocks)-1].fresh = false
		}
		if err != nil {
			return written, err
		}
		buf = buf[n:]
	}
	return written, nil
}

// WriteByte writes a byte to an io.Writer, ensuring that appropriate indentation and prefices are added at the beginning
// of each line.
func (r *Renderer) WriteByte(w io.Writer, c byte) error {
	_, err := r.Write(w, []byte{c})
	return err
}

// WriteRune writes a rune to an io.Writer, ensuring that appropriate indentation and prefices are added at the beginning
// of each line.
func (r *Renderer) WriteRune(w io.Writer, c rune) (int, error) {
	buf := make([]byte, utf8.UTFMax)
	sz := utf8.EncodeRune(buf, c)
	return r.Write(w, buf[:sz])
}

// WriteString writes a string to an io.Writer, ensuring that appropriate indentation and prefices are added at the
// beginning of each line.
func (r *Renderer) WriteString(w io.Writer, s string) (int, error) {
	return r.Write(w, []byte(s))
}

// PushIndent adds the specified amount of indentation to the current line prefix.
func (r *Renderer) PushIndent(amount int) {
	r.PushPrefix(strings.Repeat(" ", amount))
}

// PushPrefix adds the specified string to the current line prefix.
func (r *Renderer) PushPrefix(prefix string) {
	r.prefixStack = append(r.prefixStack, prefix)
	r.prefix = append(r.prefix, []byte(prefix)...)
}

// PopPrefix removes the last piece added by a call to PushIndent or PushPrefix from the current line prefix.
func (r *Renderer) PopPrefix() {
	r.prefix = r.prefix[:len(r.prefix)-len(r.prefixStack[len(r.prefixStack)-1])]
	r.prefixStack = r.prefixStack[:len(r.prefixStack)-1]
}

// OpenBlock ensures that each block begins on a new line, and that blank lines are inserted before blocks as
// indicated by node.HasPreviousBlankLines.
func (r *Renderer) OpenBlock(w util.BufWriter, source []byte, node ast.Node) error {
	r.openBlocks = append(r.openBlocks, blockState{
		node:  node,
		fresh: true,
	})

	// Work around the fact that the first child of a node notices the same set of preceding blank lines as its parent.
	hasBlankPreviousLines := node.HasBlankPreviousLines()
	if p := node.Parent(); p != nil && p.FirstChild() == node {
		if p.Kind() == ast.KindDocument || p.Kind() == ast.KindListItem || p.HasBlankPreviousLines() {
			hasBlankPreviousLines = false
		}
	}

	if hasBlankPreviousLines {
		if err := r.WriteByte(w, '\n'); err != nil {
			return err
		}
	}

	if ws := node.LeadingWhitespace(); ws.Len() != 0 {
		if _, err := r.Write(w, ws.Value(source)); err != nil {
			return err
		}
	}

	r.openBlocks[len(r.openBlocks)-1].fresh = true

	return nil
}

// CloseBlock marks the current block as closed.
func (r *Renderer) CloseBlock(w io.Writer) error {
	if !r.atNewline {
		if err := r.WriteByte(w, '\n'); err != nil {
			return err
		}
	}

	r.openBlocks = r.openBlocks[:len(r.openBlocks)-1]
	return nil
}

// RenderDocument renders an *ast.Document node to the given BufWriter.
func (r *Renderer) RenderDocument(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	r.listStack, r.prefixStack, r.prefix, r.atNewline = nil, nil, nil, false
	return ast.WalkContinue, nil
}

// RenderHeading renders an *ast.Heading node to the given BufWriter.
func (r *Renderer) RenderHeading(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.OpenBlock(w, source, node); err != nil {
			return ast.WalkStop, err
		}

		if !node.(*ast.Heading).IsSetext {
			if _, err := r.WriteString(w, strings.Repeat("#", node.(*ast.Heading).Level)); err != nil {
				return ast.WalkStop, err
			}
			if err := r.WriteByte(w, ' '); err != nil {
				return ast.WalkStop, err
			}
		}
	} else {
		if node.(*ast.Heading).IsSetext {
			s := "==="
			if node.(*ast.Heading).Level == 2 {
				s = "---"
			}
			if !r.atNewline {
				if err := r.WriteByte(w, '\n'); err != nil {
					return ast.WalkStop, err
				}
			}
			if _, err := r.WriteString(w, s); err != nil {
				return ast.WalkStop, err
			}
		}

		if err := r.WriteByte(w, '\n'); err != nil {
			return ast.WalkStop, err
		}

		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderBlockquote renders an *ast.Blockquote node to the given BufWriter.
func (r *Renderer) RenderBlockquote(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.OpenBlock(w, source, node); err != nil {
			return ast.WalkStop, err
		}

		// TODO:
		// - case 63, an setext heading in a lazy blockquote
		// - case 208, a list item in a lazy blockquote
		// - cases 262 and 263, a blockquote in a list item

		if _, err := r.WriteString(w, "> "); err != nil {
			return ast.WalkStop, err
		}
		r.PushPrefix("> ")
	} else {
		r.PopPrefix()

		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderCodeBlock renders an *ast.CodeBlock node to the given BufWriter.
func (r *Renderer) RenderCodeBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}

	if err := r.OpenBlock(w, source, node); err != nil {
		return ast.WalkStop, err
	}

	// Each line of a code block needs to be aligned at the same offset, and a code block must start with at least four
	// spaces. To achieve this, we unconditionally add four spaces to the first line of the code block and indent the
	// rest as necessary.
	if _, err := r.WriteString(w, "    "); err != nil {
		return ast.WalkStop, err
	}

	r.PushIndent(4)
	defer r.PopPrefix()

	if err := r.writeLines(w, source, node.Lines()); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

// RenderFencedCodeBlock renders an *ast.FencedCodeBlock node to the given BufWriter.
func (r *Renderer) RenderFencedCodeBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}

	if err := r.OpenBlock(w, source, node); err != nil {
		return ast.WalkStop, err
	}

	code := node.(*ast.FencedCodeBlock)

	// Write the start of the fenced code block.
	fence := code.Fence
	if _, err := r.Write(w, fence); err != nil {
		return ast.WalkStop, err
	}
	language := code.Language(source)
	if _, err := r.Write(w, language); err != nil {
		return ast.WalkStop, err
	}
	if err := r.WriteByte(w, '\n'); err != nil {
		return ast.WalkStop, nil
	}

	// Write the contents of the fenced code block.
	if err := r.writeLines(w, source, node.Lines()); err != nil {
		return ast.WalkStop, err
	}

	// Write the end of the fenced code block.
	if err := r.beginLine(w); err != nil {
		return ast.WalkStop, err
	}
	if _, err := r.Write(w, fence); err != nil {
		return ast.WalkStop, err
	}
	if err := r.WriteByte(w, '\n'); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

// RenderHTMLBlock renders an *ast.HTMLBlock node to the given BufWriter.
func (r *Renderer) RenderHTMLBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}

	if err := r.OpenBlock(w, source, node); err != nil {
		return ast.WalkStop, err
	}

	// Write the contents of the HTML block.
	if err := r.writeLines(w, source, node.Lines()); err != nil {
		return ast.WalkStop, err
	}

	// Write the closure line, if any.
	html := node.(*ast.HTMLBlock)
	if html.HasClosure() {
		if _, err := r.Write(w, html.ClosureLine.Value(source)); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderLinkReferenceDefinition renders an *ast.LinkReferenceDefinition node to the given BufWriter.
func (r *Renderer) RenderLinkReferenceDefinition(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}

	if err := r.OpenBlock(w, source, node); err != nil {
		return ast.WalkStop, err
	}

	// Write the contents of the link reference definition.
	if err := r.writeLines(w, source, node.Lines()); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

// RenderList renders an *ast.List node to the given BufWriter.
func (r *Renderer) RenderList(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.OpenBlock(w, source, node); err != nil {
			return ast.WalkStop, err
		}

		list := node.(*ast.List)
		r.listStack = append(r.listStack, listState{
			marker:  list.Marker,
			ordered: list.IsOrdered(),
			index:   list.Start,
		})
	} else {
		r.listStack = r.listStack[:len(r.listStack)-1]
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderListItem renders an *ast.ListItem node to the given BufWriter.
func (r *Renderer) RenderListItem(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.OpenBlock(w, source, node); err != nil {
			return ast.WalkStop, err
		}

		// TODO:
		// - case 227, a code block following a list item

		markerWidth := 2
		state := &r.listStack[len(r.listStack)-1]
		if state.ordered {
			width, err := r.WriteString(w, strconv.FormatInt(int64(state.index), 10))
			if err != nil {
				return ast.WalkStop, err
			}
			state.index++
			markerWidth += width
		}
		if _, err := r.Write(w, []byte{state.marker, ' '}); err != nil {
			return ast.WalkStop, err
		}

		ws := node.LeadingWhitespace()
		offset := markerWidth + ws.Len()
		if o := node.(*ast.ListItem).Offset; offset < o {
			if _, err := r.Write(w, bytes.Repeat([]byte{' '}, o-offset)); err != nil {
				return ast.WalkStop, err
			}
			offset = o
		}
		r.PushIndent(offset)
	} else {
		r.PopPrefix()
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderParagraph renders an *ast.Paragraph node to the given BufWriter.
func (r *Renderer) RenderParagraph(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		// A paragraph that follows another paragraph or a blockquote must be preceded by a blank line.
		if !node.HasBlankPreviousLines() {
			if prev := node.PreviousSibling(); prev != nil && (prev.Kind() == ast.KindParagraph || prev.Kind() == ast.KindBlockquote) {
				if err := r.WriteByte(w, '\n'); err != nil {
					return ast.WalkStop, err
				}
			}
		}

		if err := r.OpenBlock(w, source, node); err != nil {
			return ast.WalkStop, err
		}
	} else {
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderTextBlock renders an *ast.TextBlock node to the given BufWriter.
func (r *Renderer) RenderTextBlock(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if enter {
		if err := r.OpenBlock(w, source, node); err != nil {
			return ast.WalkStop, err
		}
	} else {
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderThematicBreak renders an *ast.ThematicBreak node to the given BufWriter.
func (r *Renderer) RenderThematicBreak(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		if err := r.CloseBlock(w); err != nil {
			return ast.WalkStop, err
		}
		return ast.WalkContinue, nil
	}

	if err := r.OpenBlock(w, source, node); err != nil {
		return ast.WalkStop, err
	}

	if _, err := r.WriteString(w, "***\n"); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

// RenderAutoLink renders an *ast.AutoLink node to the given BufWriter.
func (r *Renderer) RenderAutoLink(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	if err := r.WriteByte(w, '<'); err != nil {
		return ast.WalkStop, err
	}
	if _, err := r.Write(w, node.(*ast.AutoLink).Label(source)); err != nil {
		return ast.WalkStop, err
	}
	if err := r.WriteByte(w, '>'); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

func (r *Renderer) shouldPadCodeSpan(source []byte, node *ast.CodeSpan) bool {
	c := node.FirstChild()
	if c == nil {
		return false
	}

	segment := c.(*ast.Text).Segment
	text := segment.Value(source)

	var firstChar byte
	if len(text) > 0 {
		firstChar = text[0]
	}

	allWhitespace := true
	for {
		if util.FirstNonSpacePosition(text) != -1 {
			allWhitespace = false
			break
		}
		c = c.NextSibling()
		if c == nil {
			break
		}
		segment = c.(*ast.Text).Segment
		text = segment.Value(source)
	}
	if allWhitespace {
		return false
	}

	var lastChar byte
	if len(text) > 0 {
		lastChar = text[len(text)-1]
	}

	return firstChar == '`' || firstChar == ' ' || lastChar == '`' || lastChar == ' '
}

// RenderCodeSpan renders an *ast.CodeSpan node to the given BufWriter.
func (r *Renderer) RenderCodeSpan(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	// TODO:
	// - case 330, 331, single space stripping -> contents need an additional leading and trailing space
	// - case 339, backtick inside text -> start/end need additional backtick

	code := node.(*ast.CodeSpan)
	delimiter := bytes.Repeat([]byte{'`'}, code.Backticks)
	pad := r.shouldPadCodeSpan(source, code)

	if _, err := r.Write(w, delimiter); err != nil {
		return ast.WalkStop, err
	}
	if pad {
		if err := r.WriteByte(w, ' '); err != nil {
			return ast.WalkStop, err
		}
	}
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		text := c.(*ast.Text).Segment
		if _, err := r.Write(w, text.Value(source)); err != nil {
			return ast.WalkStop, err
		}
	}
	if pad {
		if err := r.WriteByte(w, ' '); err != nil {
			return ast.WalkStop, err
		}
	}
	if _, err := r.Write(w, delimiter); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkSkipChildren, nil
}

// RenderEmphasis renders an *ast.Emphasis node to the given BufWriter.
func (r *Renderer) RenderEmphasis(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	em := node.(*ast.Emphasis)
	if _, err := r.WriteString(w, strings.Repeat(string([]byte{em.Marker}), em.Level)); err != nil {
		return ast.WalkStop, err
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) escapeLinkDest(dest []byte) []byte {
	requiresEscaping := false
	for _, c := range dest {
		if c <= 32 || c == '(' || c == ')' || c == 127 {
			requiresEscaping = true
			break
		}
	}
	if !requiresEscaping {
		return dest
	}

	escaped := make([]byte, 0, len(dest)+2)
	escaped = append(escaped, '<')
	for _, c := range dest {
		if c == '<' || c == '>' {
			escaped = append(escaped, '\\')
		}
		escaped = append(escaped, c)
	}
	escaped = append(escaped, '>')
	return escaped
}

func (r *Renderer) linkTitleDelimiter(title []byte) byte {
	for i, c := range title {
		if c == '"' && (i == 0 || title[i-1] != '\\') {
			return '\''
		}
	}
	return '"'
}

func (r *Renderer) renderLinkOrImage(w util.BufWriter, open string, refType ast.LinkReferenceType, label, dest, title []byte, enter bool) error {
	if enter {
		if _, err := r.WriteString(w, open); err != nil {
			return err
		}
	} else {
		switch refType {
		case ast.LinkFullReference:
			if _, err := r.WriteString(w, "]["); err != nil {
				return err
			}
			if _, err := r.Write(w, label); err != nil {
				return err
			}
			if err := r.WriteByte(w, ']'); err != nil {
				return err
			}
		case ast.LinkCollapsedReference:
			if _, err := r.WriteString(w, "][]"); err != nil {
				return err
			}
		case ast.LinkShortcutReference:
			if err := r.WriteByte(w, ']'); err != nil {
				return err
			}
		default:
			if _, err := r.WriteString(w, "]("); err != nil {
				return err
			}

			if _, err := r.Write(w, r.escapeLinkDest(dest)); err != nil {
				return err
			}
			if len(title) != 0 {
				delimiter := r.linkTitleDelimiter(title)
				if _, err := fmt.Fprintf(w, ` %c%s%c`, delimiter, string(title), delimiter); err != nil {
					return err
				}
			}

			if err := r.WriteByte(w, ')'); err != nil {
				return err
			}
		}
	}
	return nil
}

// RenderImage renders an *ast.Image node to the given BufWriter.
func (r *Renderer) RenderImage(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	img := node.(*ast.Image)
	if err := r.renderLinkOrImage(w, "![", img.ReferenceType, img.Label, img.Destination, img.Title, enter); err != nil {
		return ast.WalkStop, err
	}
	return ast.WalkContinue, nil
}

// RenderLink renders an *ast.Link node to the given BufWriter.
func (r *Renderer) RenderLink(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	link := node.(*ast.Link)
	if err := r.renderLinkOrImage(w, "[", link.ReferenceType, link.Label, link.Destination, link.Title, enter); err != nil {
		return ast.WalkStop, err
	}
	return ast.WalkContinue, nil
}

// RenderRawHTML renders an *ast.RawHTML node to the given BufWriter.
func (r *Renderer) RenderRawHTML(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkSkipChildren, nil
	}

	raw := node.(*ast.RawHTML)
	for i := 0; i < raw.Segments.Len(); i++ {
		segment := raw.Segments.At(i)
		if _, err := r.Write(w, segment.Value(source)); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkSkipChildren, nil
}

func isBlank(bytes []byte) bool {
	for _, b := range bytes {
		if b != ' ' {
			return false
		}
	}
	return true
}

// RenderText renders an *ast.Text node to the given BufWriter.
func (r *Renderer) RenderText(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	text := node.(*ast.Text)
	value := text.Segment.Value(source)

	if _, err := r.Write(w, value); err != nil {
		return ast.WalkStop, err
	}
	switch {
	case text.HardLineBreak():
		if _, err := r.WriteString(w, "\\\n"); err != nil {
			return ast.WalkStop, err
		}
	case text.SoftLineBreak():
		if err := r.WriteByte(w, '\n'); err != nil {
			return ast.WalkStop, err
		}
	}

	return ast.WalkContinue, nil
}

// RenderString renders an *ast.String node to the given BufWriter.
func (r *Renderer) RenderString(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	str := node.(*ast.String)
	if _, err := r.Write(w, str.Value); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}

// RenderWhitespace renders an *ast.Text node to the given BufWriter.
func (r *Renderer) RenderWhitespace(w util.BufWriter, source []byte, node ast.Node, enter bool) (ast.WalkStatus, error) {
	if !enter {
		return ast.WalkContinue, nil
	}

	if _, err := r.Write(w, node.(*ast.Whitespace).Segment.Value(source)); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}
