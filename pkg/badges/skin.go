package badges

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"strings"

	"github.com/antzucaro/qstr"
	"github.com/ungerik/go-cairo"
)

// Position is an (x,y) coordinate
type Position struct {
	X float64
	Y float64
}

// TextConfig provides information used to place text on a canvas
type TextConfig struct {
	// The full path to the TTF font file to be loaded
	Font string

	// The size in points the font will be
	FontSize float64

	// The X and Y location where the bottom-left of the text will begin
	Pos Position

	// The color the text will be
	Color []qstr.RGBColor

	// Rotation
	Angle int

	// Dimensions
	Height   int
	Width    int
	MaxWidth int

	// Text alignment
	Align string
}

// SkinParams represents parameters given to Skin objects
type SkinParams struct {
	Background         string
	BackgroundColor    qstr.RGBColor
	Overlay            string
	Font               string
	Width              int
	Height             int
	NumGameTypes       int
	NickConfig         TextConfig
	NoStatsConfig      TextConfig
	GameTypeConfig     []TextConfig
	GameCountsConfig   []TextConfig
	WinPctLabelConfig  TextConfig
	WinPctConfig       TextConfig
	WinConfig          TextConfig
	LossConfig         TextConfig
	KDRatioLabelConfig TextConfig
	KDRatio            TextConfig
	KillsConfig        TextConfig
	DeathsConfig       TextConfig
	PlayingTimeConfig  TextConfig
}

// Renderer is what actually places the graphical elements on the canvas
type Renderer interface {
	placeText(text string, config TextConfig)
	placeQStr(text qstr.QStr, config TextConfig, lightnessFloor float64, lightnessCeiling float64)
}

// CairoRenderer is a Renderer that uses the cairo C library under the hood.
type CairoRenderer struct {
	surface *cairo.Surface
}

// setFont sets the font properties on a surface
func (c *CairoRenderer) setFont(config TextConfig) error {
	var color qstr.RGBColor
	if len(config.Color) == 0 {
		color = qstr.RGBColor{1.0, 1.0, 1.0}
	} else {
		color = config.Color[0]
	}
	c.surface.SelectFontFace(config.Font, cairo.FONT_SLANT_NORMAL, cairo.FONT_WEIGHT_NORMAL)
	c.surface.SetFontSize(config.FontSize)
	c.surface.SetSourceRGB(color.R, color.G, color.B)
	return nil
}

// placeText "writes" text on the drawing canvas
func (c *CairoRenderer) placeText(text string, config TextConfig) {
	if config.FontSize == 0 {
		return
	}

	c.setFont(config)

	te := c.surface.TextExtents(text)

	if config.Align == "" || config.Align == "left" {
		c.surface.MoveTo(config.Pos.X, config.Pos.Y)
	} else if config.Align == "center" {
		c.surface.MoveTo(config.Pos.X-te.Xbearing-te.Width/2, config.Pos.Y-te.Ybearing)
	} else {
		c.surface.MoveTo(config.Pos.X-te.Xbearing-te.Width, config.Pos.Y)
	}

	c.surface.Save()

	if config.Angle > 0 {
		c.surface.Rotate(float64(config.Angle) * math.Pi / 180.0)
	}

	c.surface.ShowText(text)
	c.surface.Restore()
}

// placeQStr does the same thing as placeText does, but with potentially
// colorized QStrs
func (c *CairoRenderer) placeQStr(text qstr.QStr, config TextConfig, lightnessFloor float64, lightnessCeiling float64) {
	c.setFont(config)

	// shrink the nick until it fits within the allotted space
	stripped := text.Stripped()
	te := c.surface.TextExtents(stripped)
	for te.Width > float64(config.MaxWidth) {
		// decrease the fontsize by two points and try again
		config.FontSize -= 2

		c.setFont(config)

		te = c.surface.TextExtents(stripped)
	}

	x := config.Pos.X
	var cappedColor qstr.RGBColor
	for _, colorPart := range text.ColorParts() {
		// allow capping the lightness at the high or low end depending on the background
		if lightnessFloor > 0 || lightnessCeiling < 1 {
			cappedColor = colorPart.Color.CapLightness(lightnessFloor, lightnessCeiling)
		} else {
			cappedColor = colorPart.Color
		}

		c.surface.SetSourceRGB(cappedColor.R, cappedColor.G, cappedColor.B)
		c.surface.MoveTo(x, config.Pos.Y)
		c.surface.Save()
		c.surface.ShowText(colorPart.Part)
		c.surface.Restore()

		// the starting point for the next part is the end of the last one
		te := c.surface.TextExtents(colorPart.Part)
		x += te.Width + te.Xbearing
	}
}

// Skin represents the look and feel of a XonStat badge
type Skin struct {
	Name   string
	Params SkinParams

	// TODO: this is renderer specific - build this into the interface?
	CairoRenderer
}

// String representation of a Skin
func (s *Skin) String() string {
	return s.Name
}

func (s *Skin) ShadeKDRatio(kdRatio float64, hiColor, midColor, loColor *qstr.RGBColor) qstr.RGBColor {
	var nr float64
	var c1, c2 *qstr.RGBColor
	if kdRatio >= 1.0 {
		nr = kdRatio - 1.0
		if nr > 1 {
			nr = 1.0
		}
		c1 = hiColor
		c2 = midColor
	} else {
		nr = kdRatio
		c1 = midColor
		c2 = loColor
	}

	// shade the KDRatio according to how good it is
	r := nr*c1.R + (1-nr)*c2.R
	g := nr*c1.G + (1-nr)*c2.G
	b := nr*c1.B + (1-nr)*c2.B

	return qstr.RGBColor{r, g, b}
}

func (s *Skin) ShadeWinPct(winPct float64, hiColor, midColor, loColor *qstr.RGBColor) qstr.RGBColor {
	var nr float64
	var c1, c2 *qstr.RGBColor

	if winPct > 50.0 {
		nr = 2 * (winPct/100 - 0.5)
		c1 = &s.Params.WinPctConfig.Color[0]
		c2 = &s.Params.WinPctConfig.Color[1]
	} else {
		nr = 2 * (winPct / 100)
		c1 = &s.Params.WinPctConfig.Color[1]
		c2 = &s.Params.WinPctConfig.Color[2]
	}

	// shade the WinPct according to how good it is
	r := nr*c1.R + (1-nr)*c2.R
	g := nr*c1.G + (1-nr)*c2.G
	b := nr*c1.B + (1-nr)*c2.B

	return qstr.RGBColor{r, g, b}
}

// Render the provided PlayerData using this Skin
func (s *Skin) Render(pd *PlayerData, filename string, surfaceCache map[string]*cairo.Surface) {
	// find the base surface for this skin (if not found, we've initialized things incorrectly)
	baseSurface := surfaceCache[s.Name]

	// paint that base onto a new surface
	s.surface = cairo.NewSurface(cairo.FORMAT_ARGB32, s.Params.Width, s.Params.Height)
	s.surface.SetSourceSurface(baseSurface, 0.0, 0.0)
	s.surface.Paint()

	// Nick
	s.placeQStr(pd.Nick, s.Params.NickConfig, 0.4, 1)

	// Game type labels along with Elos for those game types
	for i, gc := range pd.GameCounts[:3] {
		s.placeText(gc.GameTypeCd, s.Params.GameTypeConfig[i])
		s.placeText(fmt.Sprintf("%d", gc.GameCount), s.Params.GameCountsConfig[i])
	}

	// Kill Ratio and its details
	s.placeText("Kill Ratio", s.Params.KDRatioLabelConfig)

	kdRatio := pd.KDRatio()
	s.Params.KDRatio.Color[0] = s.ShadeKDRatio(kdRatio, &s.Params.KDRatio.Color[0], &s.Params.KDRatio.Color[1],
		&s.Params.KDRatio.Color[2])
	s.placeText(fmt.Sprintf("%.3f", kdRatio), s.Params.KDRatio)

	s.placeText(fmt.Sprintf("%d kills", pd.Kills), s.Params.KillsConfig)
	s.placeText(fmt.Sprintf("%d deaths", pd.Deaths), s.Params.DeathsConfig)

	// Win Percentage and its details
	s.placeText("Win Percentage", s.Params.WinPctLabelConfig)

	winPct := pd.WinPct()
	s.Params.WinPctConfig.Color[0] = s.ShadeWinPct(winPct, &s.Params.WinPctConfig.Color[0],
		&s.Params.WinPctConfig.Color[1], &s.Params.WinPctConfig.Color[2])
	s.placeText(fmt.Sprintf("%.2f%%", winPct), s.Params.WinPctConfig)
	s.placeText(fmt.Sprintf("%d wins", pd.Wins), s.Params.WinConfig)
	s.placeText(fmt.Sprintf("%d losses", pd.Losses), s.Params.LossConfig)

	// Playing time
	s.placeText(fmt.Sprintf("Playing Time: %s", pd.PlayingTimeString()), s.Params.PlayingTimeConfig)

	s.surface.WriteToPNG(filename)
	s.surface.Finish()
	s.surface.Destroy()
}

// LoadSkins loads up skin parameters from JSON files in dir, then constructs Skins from them
func LoadSkins(dir string) map[string]Skin {
	skins := make(map[string]Skin, 0)

	searchDir := fmt.Sprintf("%s/*json", dir)
	jsonFiles, err := filepath.Glob(searchDir)
	if err != nil {
		fmt.Println(err)
		return skins
	}

	for _, fileName := range jsonFiles {
		jsonFile, err := ioutil.ReadFile(fileName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		var s Skin
		err = json.Unmarshal(jsonFile, &s.Params)
		if err != nil {
			fmt.Println(err)
			continue
		}

		// "skins/default.json" -> "default"
		base := filepath.Base(fileName)
		name := base[0:strings.Index(base, ".json")]
		s.Name = name

		skins[name] = s
	}

	return skins
}

// LoadSurfaces constructs a map[skin name] -> *cairo.Surface for that name
func LoadSurfaces(skins map[string]Skin) map[string]*cairo.Surface {
	surfaceMap := make(map[string]*cairo.Surface)

	for name, skin := range skins {
		// the base surface
		surface := cairo.NewSurface(cairo.FORMAT_ARGB32, skin.Params.Width, skin.Params.Height)

		// load the background
		if skin.Params.Background != "" {
			bg, _ := cairo.NewSurfaceFromPNG(skin.Params.Background)

			bgW := bg.GetWidth()
			bgH := bg.GetHeight()

			bgX := 0
			bgY := 0
			for bgX < skin.Params.Width {
				bgY = 0
				for bgY < skin.Params.Height {
					surface.SetSourceSurface(bg, float64(bgX), float64(bgY))
					surface.Paint()
					bgY += bgH
				}
				bgX += bgW
			}
		}

		// load the overlay
		if skin.Params.Overlay != "" {
			overlay, _ := cairo.NewSurfaceFromPNG(skin.Params.Overlay)
			surface.SetSourceSurface(overlay, 0.0, 0.0)
			surface.Paint()
		}
		surfaceMap[name] = surface
	}

	return surfaceMap
}
