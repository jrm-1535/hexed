package main

import (
    "fmt"
    "log"
    "os"
    "io"
    "strings"
    "bytes"
    "math"
	"strconv"
    "errors"
    "path/filepath"
	"encoding/xml"

	"github.com/gotk3/gotk3/cairo"
)

const (
    HEXED_HOME = ".hexed"
    HEXED_THEMES_DIR = "themes"
    HEXED_DEFAULT_THEME = "default.xml"
)

const THEME_DIRECTORY = "/usr/share/gtksourceview-3.0/styles/"

var hexedHome   string
func setHexedHome( ) {
    home := os.Getenv( "HOME" )
    fmt.Printf("Home is \"%s\"\n", home)
    if home == "" {
        panic( "Unable to get $HOME\n" )
    }
    hexedHome = filepath.Join( home, HEXED_HOME )
}

func appendAvailableThemes( path string, paths *[]string ) (err error) {

    var files []os.DirEntry
    if files, err = os.ReadDir( path ); err != nil {
        return
    }

    fmt.Printf( "Themes:\n" )
    for _, file := range files {
        if ! file.IsDir() {
            name := file.Name()
            if strings.HasSuffix( name, ".xml" ) {
                fmt.Printf( " %s\n", name )
                *paths = append( *paths, filepath.Join( path, name ) )
            }
        }
    }
    return
}

func appendHomeThemes( paths *[]string ) (err error) {

    home := hexedHome
    if home == "" {
        panic("Hexed Home not set\n")
    }

    var files []os.DirEntry
    files, err = os.ReadDir( home )
    if err != nil && errors.Is( err, os.ErrNotExist ) {
        // try to create directory .hexed inside home directory
        err = os.Mkdir( home, 0750 )
        if err != nil {
            return
        }
    }
    theme_path := filepath.Join( home, HEXED_THEMES_DIR )
    for _, file := range( files ) {
        if file.IsDir() && file.Name() == HEXED_THEMES_DIR {
            return appendAvailableThemes( theme_path, paths )
        }
    }
    // try to create the default hexed theme

    err = os.Mkdir( theme_path, 0750 )
    if err != nil {
        return
    }

    var f *os.File
    f, err = os.Create( filepath.Join( theme_path, HEXED_DEFAULT_THEME ) )
    if err != nil {
        return
    }
    _, err = f.WriteString( `
<?xml version="1.0" encoding="UTF-8"?>

<style-scheme id="hexed-default" _name="Hexed Dark" version="1.0">

  <author>jrm</author>
  <_description>Default dark hexed color scheme</_description>

  <color name="black"                       value="#000000"/>
  <color name="white"                       value="#ffffff"/>
  <color name="grey"                        value="#808080"/>

  <color name="orange1"                     value="#a89400"/>
  <color name="orange2"                     value="#f5a800"/>
  <color name="orange3"                     value="#f59400"/>

  <color name="dark-blue"                   value="#191999"/>
  <color name="dark-purple"                 value="#770099"/>

  <color name="light-green"                 value="#336600"/>

  <!-- hexed Settings -->
  <style name="address"                     foreground="orange1" background="black"/>
  <style name="hexadecimal"                 foreground="orange2" background="black"/>
  <style name="ascii"                       foreground="orange3" background="black"/>

  <!-- Search Matching -->
  <style name="current-match"               background="dark-purple"/>
  <style name="search-match"                foreground="white" background="dark-blue"/>

  <style name="selection"                   foreground="black"   background="light-green"/>

  <style name="separator"                   foreground="grey"/>

  <style name="cursor"                      foreground="white"/>

</style-scheme>
` )
    if err == nil {
        if err = f.Close(); err == nil {
            return appendAvailableThemes( theme_path, paths )
        }
        return      // return error
    }
    f.Close()
    return
}

func getThemePaths( ) (paths []string, err error ) {
    appendHomeThemes( &paths )
    err = appendAvailableThemes( THEME_DIRECTORY, &paths )
    return
}

const (
    THEME_ELEMENT = "style-scheme"
    COLOR_ELEMENT = "color"
    STYLE_ELEMENT = "style"

    THEME_ID = "id"
    THEME_NAME = "_name"
    THEME_VERSION = "version"

    COLOR_NAME = "name"
    COLOR_VALUE = "value"

    STYLE_NAME = "name"
    STYLE_FOREGROUND_COLOR = "foreground"
    STYLE_BACKGROUND_COLOR = "background"
)

// colorPatterns order
const (
    ADDR_AREA_FOREGROUND = iota
    ADDR_AREA_BACKGROUND
    HEXA_AREA_FOREGROUND
    HEXA_AREA_BACKGROUND
    ASCI_AREA_FOREGROUND
    ASCI_AREA_BACKGROUND

    CURRENT_MATCH_BACKGROUND
    OTHER_MATCHES_BACKGROUND
    SELECTION_BACKGROUND

    SEPARATOR_FOREGROUND

    CARET_FOREGROUND

    N_PATTERNS
)

/* Hexed makes use of the following colors:

    address area foreground (text) and background
    hexadecimal area foreground and background
    ascii area foreground and background
    search match where the caret is located, background only
    search match other locations, background only
    text selection, background only
    rows and columns separator, foreground only
    caret, foreground only

    Theme styles define the color palette used in hexed. A specific theme
    style is defined for hexed, but regular gtksourceview themes can be used
    as well (with mixed results, since they are defined for different editors)
    The hexed specific theme  defines the following style names:

    ADDR_AREA:      "address" (foreground and background)
    HEXA_AREA:      "hexadecimal" (foreground and background)
    ASCI_AREA:      "ascii" (foreground and background)

    CURRENT_MATCH:  "current-match" (background)
    OTHER_MATCHES:  "search-match" (background)

    SELECTION:      "selection" (background)
    SEPARATOR:      "separator" (foreground)
    CARET:          "cursor" (foreground)

    In case of most of the gtksourceview themes, hexed uses instead:

    ADDR_AREA:      "line-numbers" (foreground and background)
    HEXA_AREA:      "text" (foreground and background)
    ASCI_AREA:      "right-margin" (foreground and background)

    CURRENT_MATCH:  "bracket-mismatch" (background)
    OTHER_MATCHES:  "search-match" (background)

    SELECTION:      "selection" (background)
    SEPARATOR:      "bracket-match" (foreground)
    CARET:          "cursor" (foreground)

    With classic or tango styles it uses the following styles, based not only
    on their apparent semantics but also on their values.

    ADDR_AREA:      "def:statement" (foreground),
                        "current-line-number" (background)
    HEXA_AREA       "def:statement" (foreground),
                        "background-pattern" (background)
    ASCI_AREA       "right-margin" (foreground and background)

    CURRENT_MATCH   "def:warning" (background)
    OTHER_MATCHES   "search-match" (background)

    SELECTION       "current-line" (background)
    SEPARATOR       "bracket-match" (background as foreground)
    CARET           "def-special-char" (foreground)
*/

type choice struct {
    index, priority int    // -1, -1 if not used
}

type alternate struct {
    name        string
    foreground,
    background  choice
}

var keyMapping = [...]alternate{
    { "address",                choice{ ADDR_AREA_FOREGROUND, 3 },
                                choice{ ADDR_AREA_BACKGROUND, 3 } },
    { "hexadecimal",            choice{ HEXA_AREA_FOREGROUND, 3 },
                                choice{ HEXA_AREA_FOREGROUND, 3 } },
    { "ascii",                  choice{ ASCI_AREA_FOREGROUND, 3 },
                                choice{ ASCI_AREA_FOREGROUND, 3 } },

    { "current-match",          choice{ -1, -1 },
                                choice{ CURRENT_MATCH_BACKGROUND, 3 } },
    { "search-match",           choice{ -1, -1 },
                                choice{ OTHER_MATCHES_BACKGROUND, 3 } },

    { "selection",              choice{ -1, -1 },
                                choice{ SELECTION_BACKGROUND, 3 } },
    { "separator",              choice{ SEPARATOR_FOREGROUND, 3 },
                                choice{ -1, -1 } },
    { "cursor",                 choice{ CARET_FOREGROUND, 3 },
                                choice{ -1, -1 } },

    { "line-numbers",           choice{ ADDR_AREA_FOREGROUND, 2 },
                                choice{ ADDR_AREA_BACKGROUND, 2 } },
    { "text",                   choice{ HEXA_AREA_FOREGROUND, 2 },
                                choice{ HEXA_AREA_BACKGROUND, 2 } },
    { "right-margin",           choice{ ASCI_AREA_FOREGROUND, 2 },
                                choice{ ASCI_AREA_BACKGROUND, 2 } },

    { "bracket-match",          choice{ -1, -1 },
                                choice{ OTHER_MATCHES_BACKGROUND, 2 } },

    { "current-line",           choice{ -1, -1 },
                                choice{ SELECTION_BACKGROUND, 1 } },

    { "def:preprocessor",       choice{ ADDR_AREA_FOREGROUND, 1 },
                                choice{ -1, -1 } },
    { "current-line-number",    choice{ -1, -1 },
                                choice{ ADDR_AREA_BACKGROUND, 1 } },
    { "def:statement",          choice{ HEXA_AREA_FOREGROUND, 1 },
                                choice{ -1, -1 } },
    { "background-pattern",     choice{ -1, -1 },
                                choice{ HEXA_AREA_BACKGROUND, 1 } },

    { "bracket-mismatch",       choice{ -1, -1 },
                                choice{ CURRENT_MATCH_BACKGROUND, 1 } },

    { "def:warning",            choice{ -1, -1 },
                                choice{ SELECTION_BACKGROUND, 1 } },
    { "draw-spaces",            choice{ -1, -1 },
                                choice{ SEPARATOR_FOREGROUND, 1 } },
    { "def-special-char",       choice{ CARET_FOREGROUND, 1 },
                                choice{ -1, -1 } },
}

func getAlternateMapping( n string ) *alternate {
    // hexed mapping has priority
    for i := 0; i < len(keyMapping); i++ {
        if keyMapping[i].name == n {
//fmt.Printf("Found alternate %d (%s)\n", i, n)
            return &keyMapping[i]
        }
    }
    return nil
}

type theme struct {
    version         string
    name            string
    id              string

    colors          map[string]uint32
    colorPatterns   [N_PATTERNS][4]byte  // [0] is priority, [1][2][3]=rgb
}

const (
    colP = iota
    colR
    colG
    colB
)

func (t *theme) setPattern( index int, priority int, col uint32 ) {
    if index < 0 || index >= N_PATTERNS {
        panic("Invalid pattern index\n")
    }

//fmt.Printf( " set uint32 col=%#08x, b=%#02x, g=%#x, r=%#02x @index %d, priority %d\n",
//            col, byte( col ), byte( col >> 8 ), byte( col >> 16 ), index, priority )
    if t.colorPatterns[index][colP] < byte(priority) {
        t.colorPatterns[index][colB] = byte( col )
        t.colorPatterns[index][colG] = byte( col >> 8 )
        t.colorPatterns[index][colR] = byte( col >> 16 )
        t.colorPatterns[index][colP] = byte(priority)
    }
}

func makeUint32( s string ) (uint32, error) {
    if s[0] != '#' {
        return 0, fmt.Errorf( "Not a valid color definition (%s)\n", s )
    }
    v, err := strconv.ParseUint( s[1:], 16, 32 )
    if err != nil {
        // not a number, may be a well-known color name
//fmt.Printf("makeUint32: not a number : #%s\n", s[1:])
        if s[1:] == "black" {
            v = 0
        } else if s[1:] == "white" {
            v = 0xffffff
        } else {
            return 0, fmt.Errorf( "Not a valid color definition (%s)\n", s )
        }
    }
//fmt.Printf( " makeUint32: string=\"%s\" val=%#08x\n", s, v )
    return uint32(v), nil
}

func (t *theme) getUint32ColorFromRef( ref string ) (col uint32, err error) {

    if ref[0] == '#' {  // immediate value
        col, err = makeUint32( ref )
    } else {            // a color name
        var ok bool
        if col, ok = t.colors[ref]; ! ok {
            err = fmt.Errorf( "Undefined color \"%s\"\n", ref )
        }
    }
    return
}

// XML schema:
//  THEME_ELEMENT{id, name, version}
//    ( COLOR_ELEMENT{name, value} || STYLE_ELEMENT{foreground, background} )*
func (t *theme) storeElement( e xml.StartElement, s []string ) error {
    switch e.Name.Local {
    case THEME_ELEMENT:
        if len(s) != 0 {    // not a top-level element => error
            return fmt.Errorf( "theme: wrong structure (style-scheme)\n" )
        }
        for _, attr := range e.Attr {
            switch attr.Name.Local {
            case THEME_ID:
                t.id = attr.Value
            case THEME_NAME:
                t.name = attr.Value
            case THEME_VERSION:
                t.version = attr.Value
            }
        }
    case COLOR_ELEMENT:
        if len(s) != 1 {
            return fmt.Errorf( "theme: wrong structure (color)\n" )
        }
        var name, val string
        for _ , attr:= range e.Attr {
            switch attr.Name.Local {
            case COLOR_NAME:
                name = attr.Value
            case COLOR_VALUE:
                val = attr.Value
            }
        }
        if name != "" && val != "" {
            v, err := t.getUint32ColorFromRef( val )
            if err != nil {
                return err
            }
            t.colors[name] = v
        }
    case STYLE_ELEMENT:
        if len(s) != 1 {
            return fmt.Errorf( "theme: wrong structure (style)\n" )
        }
        var name, fg, bg string
        for _, attr := range e.Attr {
            switch attr.Name.Local {
            case STYLE_NAME:
                name = attr.Value
            case STYLE_FOREGROUND_COLOR:
                fg = attr.Value
            case STYLE_BACKGROUND_COLOR:
                bg = attr.Value
            }
        }
        alt := getAlternateMapping( name )
        if nil != alt {
            if fg != "" && alt.foreground.index != -1 {
                if col, err := t.getUint32ColorFromRef( fg ); err == nil {
                    t.setPattern( alt.foreground.index,
                                  alt.foreground.priority, col )
                }
            }
            if bg != "" && alt.background.index != -1 {
                if col, err := t.getUint32ColorFromRef( bg ); err == nil {
                    t.setPattern( alt.background.index,
                                  alt.background.priority, col )
                }
            }
        }
    }
    return nil
}

func initTheme( ) {

    paths, err := getThemePaths( )
    if err != nil {
        log.Fatalf("No theme available: %v\n", err)
    }
    name :=  getStringPreference(COLOR_THEME_NAME)
    path, err :=  findThemePathFromName( paths, name )
    if err != nil {
        log.Fatalf("Unknown path for theme %s: %v\n", name, err)
    }
    th, err := readTheme( path )
    if err != nil {
        log.Fatalf( "Theme file %s is not valid\n", path)
    }
    th.fixColorPatterns( )
    th.setThemePatterns( )
}

var cairoPatterns [N_PATTERNS]*cairo.Pattern

func setAddForegroundColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[ADDR_AREA_FOREGROUND] )
}

func setAddBackgroundColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[ADDR_AREA_BACKGROUND] )
}

func setHexForegroundColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[HEXA_AREA_FOREGROUND] )
}

func setHexBackgroundColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[HEXA_AREA_BACKGROUND] )
}

func setAscForegroundColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[ASCI_AREA_FOREGROUND] )
}

func setAscBackgroundColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[ASCI_AREA_BACKGROUND] )
}

func setCurrentMatchColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[CURRENT_MATCH_BACKGROUND] )
}

func setOtherMatchesColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[OTHER_MATCHES_BACKGROUND] )
}

func setSelectionColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[SELECTION_BACKGROUND] )
}

func setSeparatorColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[SEPARATOR_FOREGROUND] )
}

func setCaretColor( cr *cairo.Context ) {
    cr.SetSource( cairoPatterns[CARET_FOREGROUND] )
}

func setOppositeRGB( dst, src *[4]byte ) {
    (*dst)[colR] = 255 - (*src)[colR]
    (*dst)[colG] = 255 - (*src)[colG]
    (*dst)[colB] = 255 - (*src)[colB]
}

func getRelativeLUminance( col *[4]byte ) float64 {

    RsRGB := float64((*col)[colR]) / 255.0
    GsRGB := float64((*col)[colG]) / 255.0
    BsRGB := float64((*col)[colB]) / 255.0

    var r, g, b float64
    if RsRGB <= 0.03928 {
        r = RsRGB / 12.92
    } else {
        r = math.Pow( ((RsRGB+0.055)/1.055), 2.4 )
    }

    if GsRGB <= 0.03928 {
        g = GsRGB/12.92
    } else {
        g = math.Pow( ((GsRGB+0.055)/1.055), 2.4 )
    }

    if BsRGB <= 0.03928 {
        b = BsRGB/12.92
    } else {
        b = math.Pow( ((BsRGB+0.055)/1.055), 2.4 )
    }
    return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

func isContrastSufficient( col1, col2 *[4]byte ) bool {
    l1 := getRelativeLUminance( col1 )
    l2 := getRelativeLUminance( col2 )

    contrast := (l1 + 0.05) / (l2 + 0.05)
    if contrast < 1.0 {
        contrast = 1.0 / contrast
    }
fmt.Printf("Contrast = %f\n", contrast )
    return contrast >= 2.0
}

// Fix color pattern if a pattern was not set (missing in one theme file) or if
// contrast between background and foregrpund color is not sufficient.
// - for ADDR, force foreground to be opposite of ADDR background
// - for HEXA chars, if background was never specified (priority == 0), use
//                   instead ADDR background
//                   then, if contrast between foreground and background colors
//                         is insufficient, use ADDR foreground if contrast with
//                         HEXA background is sufficient, else take the opposite
//                         of HEXA background
// - for ASCI chars, if background was never specified (priority == 0), use
//                   instead HEXA background
//                   then, if contrast between foreground and background colors
//                         is insufficient, use HEXA foreground if contrast with
//                         ASCI background is sufficient, else take the opposite
//                         of ASCI background
// - for separator, if contrast with HEXA background is insufficient, use HEXA
//                  foreground instead
// - for cursor, if missing or insufficient contrast with HEXA background, force
//               to opposite of HEXA background
func (t *theme) fixColorPatterns( ) {
    if ! isContrastSufficient( &t.colorPatterns[ADDR_AREA_FOREGROUND],
                                 &t.colorPatterns[ADDR_AREA_BACKGROUND] ) {
//fmt.Printf(" ADDR insufficient contrast between F & B, setting ADDR F= ~ADDR B\n")
        setOppositeRGB( &t.colorPatterns[ADDR_AREA_FOREGROUND],
                            &t.colorPatterns[ADDR_AREA_BACKGROUND] )
    }
    if t.colorPatterns[HEXA_AREA_BACKGROUND][colP] == 0 {
//fmt.Printf(" HEXA undefined B, setting HEXA B= ADDR B\n")
        t.colorPatterns[HEXA_AREA_BACKGROUND] =
                                t.colorPatterns[ADDR_AREA_BACKGROUND]
    }
    if ! isContrastSufficient( &t.colorPatterns[HEXA_AREA_FOREGROUND],
                                 &t.colorPatterns[HEXA_AREA_BACKGROUND] ) {
//fmt.Printf(" HEXA insufficient contrast between F & B, checking ADDR F contrast\n")
        if isContrastSufficient( &t.colorPatterns[ADDR_AREA_FOREGROUND],
                                   &t.colorPatterns[HEXA_AREA_BACKGROUND] ) {
//fmt.Printf(" >> sufficient contrast between ADDR F & HEXA B, setting HEXA F= ADDR F\n")
            t.colorPatterns[HEXA_AREA_FOREGROUND] =
                                t.colorPatterns[ADDR_AREA_FOREGROUND]
        } else  {
//fmt.Printf(" >> INsufficient contrast between ADDR F & HEXA B, setting HEXA F=~HEXA B\n")
            setOppositeRGB( &t.colorPatterns[HEXA_AREA_FOREGROUND],
                                &t.colorPatterns[HEXA_AREA_BACKGROUND] )
        }
    }
    if t.colorPatterns[ASCI_AREA_BACKGROUND][colP] == 0 {
//fmt.Printf(" ASCI undefined B, setting ASCI B= HEXA B\n")
        t.colorPatterns[ASCI_AREA_BACKGROUND] =
                                t.colorPatterns[HEXA_AREA_BACKGROUND]
    }
    if ! isContrastSufficient( &t.colorPatterns[ASCI_AREA_FOREGROUND],
                                &t.colorPatterns[ASCI_AREA_BACKGROUND] ) {
//fmt.Printf(" ASCI insufficient contrast between F & B, checking HEXA F contrast\n")
        if isContrastSufficient( &t.colorPatterns[ASCI_AREA_FOREGROUND],
                                  &t.colorPatterns[HEXA_AREA_FOREGROUND] ) {
//fmt.Printf(" >> sufficient contrast between HEXA F & ASCI B, setting ASCI F= HEXA F\n")
            t.colorPatterns[ASCI_AREA_FOREGROUND] =
                                t.colorPatterns[HEXA_AREA_FOREGROUND]
        } else  {
//fmt.Printf(" >> INsufficient contrast between HEXA F & ASCI B, setting ASCI F=~ASCI B\n")
            setOppositeRGB( &t.colorPatterns[ASCI_AREA_FOREGROUND],
                                &t.colorPatterns[ASCI_AREA_BACKGROUND] )
        }
    }
    if ! isContrastSufficient( &t.colorPatterns[SEPARATOR_FOREGROUND],
                                &t.colorPatterns[HEXA_AREA_BACKGROUND] ) {
//fmt.Printf(" insufficient contrast between SEPA F & HEXA B, setting SEPA F=HEXA F\n")
        t.colorPatterns[SEPARATOR_FOREGROUND] =
                                t.colorPatterns[HEXA_AREA_FOREGROUND]
    }
    if ! isContrastSufficient( &t.colorPatterns[CARET_FOREGROUND],
                                &t.colorPatterns[HEXA_AREA_BACKGROUND] ) {
//fmt.Printf(" insufficient contrast between CARET F & HEXA B, setting CARET F=~HEXA B\n")
        setOppositeRGB( &t.colorPatterns[CARET_FOREGROUND],
                                &t.colorPatterns[HEXA_AREA_BACKGROUND] )
    }
}

func (t *theme) setThemePatterns( ) error {
fmt.Printf("Setting Theme Color Patterns:\n")
    var err error
    for i := 0; i < N_PATTERNS; i++ {
        r := float64(t.colorPatterns[i][1]) / float64(255)
        g := float64(t.colorPatterns[i][2]) / float64(255)
        b := float64(t.colorPatterns[i][3]) / float64(255)
fmt.Printf( "  pattern %d, cairo=(%f,%f,%f)\n", i, r, g, b )
        cairoPatterns[i], err = cairo.NewPatternFromRGB( r, g, b )
        if err != nil {
            return err
        }
    }
    return nil
}

func findThemePathFromName( paths []string,
                            name string ) (path string, err error ) {

    for _, path = range paths {
        var themeName string
        if themeName, err = getThemeName( path ); err == nil {
            if themeName == name {
                return
            }
        }
    }
    path = ""
    err = fmt.Errorf( "can't find theme path\n" )
    return
}

func getThemeNames( ) (names []string, err error) {

    paths, err := getThemePaths( )
    if err != nil {
        err = fmt.Errorf("getThemeNames: no theme available: %v\n", err)
    }

    for _, path := range paths {
        name, err := getThemeName( path )
        //fmt.Printf( "Adding name=%s err=%v @ %d\n", name, err, i )
        if err == nil {
            names = append( names, name )
        }   // else ignore files with errors
    }
    if len(names) == 0 {
        err = fmt.Errorf("getThemeNames: no theme name available\n" )
    }
    return
}

func getThemeName( path string ) (name string, err error) {
    var file *os.File
    if file, err = os.Open( path ); err != nil {
        return
    }
//fmt.Printf( "Opened path=%s err=%v\n", path, err )
    var data []byte
    data, err = io.ReadAll( file )

    dec := xml.NewDecoder( bytes.NewBuffer( data ) )

    var tok interface{}
    for {
        tok, err = dec.Token()
//fmt.Printf( "got token=%T err=%v\n", tok, err )
        if err != nil {
            return
        }
        switch tok := tok.(type) {
        case xml.StartElement:
            if tok.Name.Local == THEME_ELEMENT {
                for _, attr := range tok.Attr {
                    if attr.Name.Local == THEME_NAME {
                        name = attr.Value
                        return
                    }
                }
                err = fmt.Errorf( "getThemeName: missing scheme name\n" )
                return
            } else {
                break
            }
        default:    // ignore chatData (white char) and comments
        }
    }
    err = fmt.Errorf( "getThemeName: wrong structure (style-scheme)\n" )
    return 
}

func readTheme( path string ) (th *theme, err error) {
    var file *os.File
    if file, err = os.Open( path ); err != nil {
        return
    }

    var data []byte
    data, err = io.ReadAll( file )
    th, err = getThemeData( data )
    return
}

func getThemeData( data []byte ) (* theme, error) {

    dec := xml.NewDecoder( bytes.NewBuffer( data ) )
    var stack []string
    var th = new(theme)
    th.colors = make( map[string]uint32 )

    for {
        tok, err := dec.Token()
        if err != nil {
            if err == io.EOF {
                break
            }
            return nil, fmt.Errorf( "getThemeData: %v\n", err )
        }

        switch tok := tok.(type) {
        case xml.StartElement:
            err = th.storeElement( tok, stack )
            if err != nil {
                return nil, fmt.Errorf("getThemeData: %v\n", err )
            }
            stack = append( stack, tok.Name.Local )

        case xml.EndElement:
            stack = stack[:len(stack)-1]
//        case xml.CharData:  // just ignore
        }
    }
    return th, nil
}
