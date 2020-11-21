package main

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gookit/color"
	"golang.org/x/text/unicode/norm"
)

type formats struct {
	single      string
	full        string
	transparent string
	linechange  string
	left        string
	right       string
	rANSI       string
	title       string
	creset      string
	clearOEL    string
}

var (
	// Map for color names and their respective foreground [0] or background [1] color codes
	colorMap map[string][2]string = map[string][2]string{
		"black":        {"30", "40"},
		"red":          {"31", "41"},
		"green":        {"32", "42"},
		"yellow":       {"33", "43"},
		"blue":         {"34", "44"},
		"magenta":      {"35", "45"},
		"cyan":         {"36", "46"},
		"white":        {"37", "47"},
		"default":      {"39", "49"},
		"darkGray":     {"90", "100"},
		"lightRed":     {"91", "101"},
		"lightGreen":   {"92", "102"},
		"lightYellow":  {"93", "103"},
		"lightBlue":    {"94", "104"},
		"lightMagenta": {"95", "105"},
		"lightCyan":    {"96", "106"},
		"lightWhite":   {"97", "107"},
	}
)

// Returns the color code for a given color name
func getColorFromName(colorName string, isBackground bool) (string, error) {
	colorMapOffset := 0
	if isBackground {
		colorMapOffset = 1
	}
	if colorCodes, found := colorMap[colorName]; found {
		return colorCodes[colorMapOffset], nil
	}
	return "", errors.New("color name does not exist")
}

// Renderer writes colorized strings
type Renderer struct {
	Buffer  *bytes.Buffer
	formats *formats
	shell   string
}

const (
	// Transparent implies a transparent color
	Transparent = "transparent"
	zsh         = "zsh"
	bash        = "bash"
)

func (r *Renderer) init(shell string) {
	r.shell = shell
	r.formats = &formats{
		rANSI: "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))",
	}
	switch shell {
	case zsh:
		r.formats.single = "%%{\x1b[%sm%%}%s%%{\x1b[0m%%}"
		r.formats.full = "%%{\x1b[%sm\x1b[%sm%%}%s%%{\x1b[0m%%}"
		r.formats.transparent = "%%{\x1b[%s;49m\x1b[7m%%}%s%%{\x1b[m\x1b[0m%%}"
		r.formats.linechange = "%%{\x1b[%d%s%%}"
		r.formats.left = "%%{\x1b[%dC%%}"
		r.formats.right = "%%{\x1b[%dD%%}"
		r.formats.title = "%%{\033]0;%s\007%%}"
		r.formats.creset = "%{\x1b[0m%}"
		r.formats.clearOEL = "%{\x1b[K%}"
	case bash:
		r.formats.single = "\\[\x1b[%sm\\]%s\\[\x1b[0m\\]"
		r.formats.full = "\\[\x1b[%sm\x1b[%sm\\]%s\\[\x1b[0m\\]"
		r.formats.transparent = "\\[\x1b[%s;49m\x1b[7m\\]%s\\[\x1b[m\x1b[0m\\]"
		r.formats.linechange = "\\[\x1b[%d%s\\]"
		r.formats.left = "\\[\x1b[%dC\\]"
		r.formats.right = "\\[\x1b[%dD\\]"
		r.formats.title = "\\[\033]0;%s\007\\]"
		r.formats.creset = "\\[\x1b[0m\\]"
		r.formats.clearOEL = "\\[\x1b[K\\]"
	default:
		r.formats.single = "\x1b[%sm%s\x1b[0m"
		r.formats.full = "\x1b[%sm\x1b[%sm%s\x1b[0m"
		r.formats.transparent = "\x1b[%s;49m\x1b[7m%s\x1b[m\x1b[0m"
		r.formats.linechange = "\x1b[%d%s"
		r.formats.left = "\x1b[%dC"
		r.formats.right = "\x1b[%dD"
		r.formats.title = "\033]0;%s\007"
		r.formats.creset = "\x1b[0m"
		r.formats.clearOEL = "\x1b[K"
	}
}

// Gets the ANSI color code for a given color string.
// This can include a valid hex color in the format `#FFFFFF`,
// but also a name of one of the first 16 ANSI colors like `lightBlue`.
func (r *Renderer) getAnsiFromColorString(colorString string, isBackground bool) string {
	colorFromName, err := getColorFromName(colorString, isBackground)
	if err == nil {
		return colorFromName
	}
	style := color.HEX(colorString, isBackground)
	return style.Code()
}

func (r *Renderer) writeColoredText(background, foreground, text string) {
	var coloredText string
	if foreground == Transparent && background != "" {
		ansiColor := r.getAnsiFromColorString(background, false)
		coloredText = fmt.Sprintf(r.formats.transparent, ansiColor, text)
	} else if background == "" || background == Transparent {
		ansiColor := r.getAnsiFromColorString(foreground, false)
		coloredText = fmt.Sprintf(r.formats.single, ansiColor, text)
	} else if foreground != "" && background != "" {
		bgAnsiColor := r.getAnsiFromColorString(background, true)
		fgAnsiColor := r.getAnsiFromColorString(foreground, false)
		coloredText = fmt.Sprintf(r.formats.full, bgAnsiColor, fgAnsiColor, text)
	}
	r.Buffer.WriteString(coloredText)
}

func (r *Renderer) writeAndRemoveText(background, foreground, text, textToRemove, parentText string) string {
	r.writeColoredText(background, foreground, text)
	return strings.Replace(parentText, textToRemove, "", 1)
}

func (r *Renderer) write(background, foreground, text string) {
	// first we match for any potentially valid colors enclosed in <>
	rex := regexp.MustCompile(`<([#A-Za-z0-9]+)?(?:,([#A-Za-z0-9]+))?>(.*?)<\/>`)
	match := rex.FindAllStringSubmatch(text, -1)
	for i := range match {
		extractedForegroundColor := match[i][1]
		extractedBackgroundColor := match[i][2]
		if col := r.getAnsiFromColorString(extractedForegroundColor, false); col == "" && extractedForegroundColor != Transparent && len(extractedBackgroundColor) == 0 {
			continue // we skip invalid colors
		}
		if col := r.getAnsiFromColorString(extractedBackgroundColor, false); col == "" && extractedBackgroundColor != Transparent && len(extractedForegroundColor) == 0 {
			continue // we skip invalid colors
		}

		// reuse function colors if only one was specified
		if len(extractedBackgroundColor) == 0 {
			extractedBackgroundColor = background
		}
		if len(extractedForegroundColor) == 0 {
			extractedForegroundColor = foreground
		}

		escapedTextSegment := match[i][0]
		innerText := match[i][3]
		textBeforeColorOverride := strings.Split(text, escapedTextSegment)[0]
		text = r.writeAndRemoveText(background, foreground, textBeforeColorOverride, textBeforeColorOverride, text)
		text = r.writeAndRemoveText(extractedBackgroundColor, extractedForegroundColor, innerText, escapedTextSegment, text)
	}
	// color the remaining part of text with background and foreground
	r.writeColoredText(background, foreground, text)
}

func (r *Renderer) lenWithoutANSI(str string) int {
	re := regexp.MustCompile(r.formats.rANSI)
	stripped := re.ReplaceAllString(str, "")
	switch r.shell {
	case zsh:
		stripped = strings.ReplaceAll(stripped, "%{", "")
		stripped = strings.ReplaceAll(stripped, "%}", "")
	case bash:
		stripped = strings.ReplaceAll(stripped, "\\[", "")
		stripped = strings.ReplaceAll(stripped, "\\]", "")
	}
	var i norm.Iter
	i.InitString(norm.NFD, stripped)
	var count int
	for !i.Done() {
		i.Next()
		count++
	}
	return count
}

func (r *Renderer) carriageForward() string {
	return fmt.Sprintf(r.formats.left, 1000)
}

func (r *Renderer) setCursorForRightWrite(text string, offset int) string {
	strippedLen := r.lenWithoutANSI(text) + -offset
	return fmt.Sprintf(r.formats.right, strippedLen)
}

func (r *Renderer) changeLine(numberOfLines int) string {
	position := "B"
	if numberOfLines < 0 {
		position = "F"
		numberOfLines = -numberOfLines
	}
	return fmt.Sprintf(r.formats.linechange, numberOfLines, position)
}

func (r *Renderer) setConsoleTitle(title string) {
	fmt.Printf(r.formats.title, title)
}

func (r *Renderer) string() string {
	return r.Buffer.String()
}

func (r *Renderer) reset() {
	r.Buffer.Reset()
}

func (r *Renderer) creset() {
	fmt.Print(r.formats.creset)
}

func (r *Renderer) print(text string) {
	fmt.Print(text)
	r.clearEOL()
}

func (r *Renderer) clearEOL() {
	fmt.Print(r.formats.clearOEL)
}
