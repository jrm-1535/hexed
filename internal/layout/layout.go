package layout

import (
    "fmt"
    "log"
    "strings"
    "strconv"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

/*
    Box:

        Linear area of cells, such that all cells are in a single column or in
        a single row. Each cell share the same size on the other dimension, i.e.
        same width for a column or same height for a row, but each can have a
        different size in the box dimension, i.e. height in a column or width
        in a row.
*/
type BoxDef struct {
    Name        string          // used to manipulate item after creation

    Padding     uint            // left padding in parent box or cell

    Margin      int             // first/last item and box (within direction)
    Spacing     uint            // spacing between items
    Title       string          // optional frame title
    Border      bool            // visible frame around box

    Direction   Orientation     // box orientation (HORIZONTAL/VERTICAL)
    ItemDefs    []interface{}   // *boxDef, *gridDef,  *constDef or *inputDef
}

type Orientation gtk.Orientation
const (
    HORIZONTAL  Orientation = Orientation(gtk.ORIENTATION_HORIZONTAL)
    VERTICAL                = Orientation(gtk.ORIENTATION_VERTICAL)
)

/*
    Grid:

        Rectangular area of cells, such that all cells in the same column have
        the same width, which is either the width of the widest item in the
        column or the remaining width in the row if the column has the expand
        property set, and all cells in the same row have the same height, which
        is the height of the tallest item in the row or the remaining height i

        Although the number of cells is fixed in a given grid, cells can be
        empty, leaving holes in the grid.
*/
type GridDef struct { 
    Name        string          // used to manipulate item after creation

    Padding     uint            // left padding in parent box or cell

    H           HorizontalDef   // defines the columns
    V           VerticalDef     // defines the rows
}

type HorizontalDef struct {
    Spacing     uint            // space between columns
    Columns     []ColDef        // list of column definitions
}

type VerticalDef struct {
    Spacing     uint            // space between rows
    Rows        []RowDef        // list of row definitions
}

type ColDef struct {
    Expand      bool            // whether to expand if extra room available
}

type RowDef struct {
    Expand      bool            // whether to expand if extra room available
    Items       []interface{}   // row items
}

/*
    Item:

        Items can be empty (nil), a single constant value (*constDef) that
        cannot be modified by the user, a single input value or input button
        that can modified or pressed by the user (*inputDef), an horizontal or
        vertical box of items (*boxDef), or a grid of items (*gridDef).

        Constant or input values can be of type boolean, integer or string.
        Input buttons can be with a text or an icon label, and can behave as
        a press button or a toggle button. The button value is alaways a bool,
        always true for a press button, the toggle state for a toggle button.
*/
type ConstDef struct {          // constant field (ouput only)
    Name        string          // internal name to get or set the text

    Padding     uint            // left padding in parent box or cell

    Value       interface{}     // bool, int, string

    ToolTip     string          // possible constant description
    Format      *TextFmt        // presentation format (int & string)
}

type InputDef struct {          // user modifiable field
    Name        string          // use to manipulate item after creation

    Padding     uint            // left padding in parent box or cell


    Value       interface{}     // initial value (bool, int, string, or
                                // initial button label (textDef or iconDef)

    ToolTip     string          // input help for users
    Changed     func( name string, val interface{} ) bool // change notification
                                // val is same type as value or in case of
                                // button, a bool (state for toggle button).
    Control     interface{}     // input control or button control:
                                //   intCtl or intList, strCtl or strList, or
                                //   buttonCtl.
}

// Text formatting
type TextFmt struct {           // text formatted in a rectangular frame
    Attributes  FontAttr        // monospace, bold, italic, etc.
    Align       AlignAttr       // how to align text in frame
    FrameSize   int             // max number of chars within frame (0 no limit)
    Border      bool            // whether the frame has a border
    Copy        func( name string, event *gdk.Event ) bool // nil if not allowed
}

type FontAttr uint
const (
    REGULAR FontAttr = 0
    MONOSPACE FontAttr = 1 << iota
    BOLD
    ITALIC
// TODO: Add basic color?
)

type AlignAttr uint
const (
    LEFT AlignAttr = iota
    CENTER
    RIGHT
)

// Label for buttons
type TextDef    struct {
    Text        string          // button label
    Format      *TextFmt
}

type IconDef    struct {
    Name        string          // stock icon name
}

const (
    MAX_SELECTION_LENGTH = 63                   // in bytes
    MAX_TEXT_LENGTH = 2 * MAX_SELECTION_LENGTH  // in nibbles
    MAX_STORE_ROW = 9                           // 10 entries  (0-9)
)

// input control types (no control for boolean)
type IntCtl struct {            // for integer input (from Min to Max by Inc)
    InputMin    int             // minimum value acceptable
    InputMax    int             // maximum value acceptable
    InputInc    int             // increment between acceptable values
}

type IntList struct {           // for integer input (from list w/wo entry)
    List        []int           // initial list of integer
    FreeEntry   bool            // whether non-list entry is accepted
}

type StrCtl struct {            // for string input (entry with length control)
    InputMax    int
}

type KeyModifier uint           // Key-modifier bitmask
const (
    SHIFT = KeyModifier(gdk.SHIFT_MASK)     // 1
    LOCK = KeyModifier(gdk.LOCK_MASK)       // 2
    CONTROL = KeyModifier(gdk.CONTROL_MASK) // 4
    ALT = KeyModifier(gdk.MOD1_MASK)        // 8
)

type StrList struct {           // for string input (from list w/wo entry)
    List        []string        // initial list of strings
    FreeEntry   bool            // whether non-list entry is accepted
    InputMax    int             // maximum free entry length
    MouseBut    func( name string, but gdk.Button ) bool
    KeyPress    func( name string, key uint, mod KeyModifier) bool
}

type ButtonCtl struct {
    Toggle      bool            // whether button press toggles its state
    Initial     bool            // if toggle, initial toggle state
}

// internal item reference
type itemReference struct {
    format      *TextFmt        // item format
    strIsInt    bool            // true if internal string is actually an int
    item        interface{}     // reference to the internal item, either
                                // *gtk.IWidget or *DataGrid
}

func setTooltip( wg gtk.IWidget, tooltip string ) {
    if tooltip != "" {
        wg.ToWidget().SetTooltipText( tooltip )
    }
}

func makeButtonInput( def *InputDef ) (*itemReference, error ) {

    control := def.Control.(*ButtonCtl)
    var (
        tb *gtk.ToggleButton
        b *gtk.Button
        err error
    )
    switch v := def.Value.(type) {
    case *TextDef:
        if control.Toggle  {
            tb, err = gtk.ToggleButtonNewWithLabel( v.Text )
            if err != nil {
                return nil,
                fmt.Errorf("makeButtonInput: can't create toggle button with label: %v", err)
            }
            tb.SetActive( control.Initial )

        } else {
            b, err = gtk.ButtonNewWithLabel( v.Text )
            if err != nil {
                return nil, fmt.Errorf("makeButtonInput: can't create button with label: %v", err)
            }
        }

    case *IconDef:
        if control.Toggle {
            tb, err = gtk.ToggleButtonNew( )
            if err != nil {
                return nil, fmt.Errorf("makeButtonInput: can't create toggle button: %v", err)
            }
            tb.SetActive( control.Initial )
            b = &tb.Button
        } else {
            b, err = gtk.ButtonNew( )
            if err != nil {
                return nil, fmt.Errorf("makeButtonInput: can't create button: %v", err)
            }
        }
        icon, err := gtk.ImageNewFromIconName( v.Name, gtk.ICON_SIZE_BUTTON )
        if err != nil {
            return nil, fmt.Errorf("addButtonItem: could not create icon image: %v", err)
        }
        b.SetImage( icon )

    default:
        return nil, fmt.Errorf( "addButtonItem: unknown button type %T\n", v )
    }
    if def.Changed != nil {
        if control.Toggle {
            toggled := func( button *gtk.ToggleButton ) bool {
                state := button.GetActive( )
                return def.Changed( def.Name, state )
            }
            tb.Connect( "toggled", toggled )
        } else {
            clicked := func( button *gtk.Button ) bool {
                return def.Changed( def.Name, true )
            }
            b.Connect( "clicked", clicked )
        }
    }
    if control.Toggle {
        return &itemReference{ nil, false, tb }, nil
    }
    return &itemReference{ nil, false, b }, nil
}

func makeTextInput( def *InputDef ) (*itemReference, error) {

    textVal := def.Value.(string)

    if ctl, ok := def.Control.(*StrList); ok {
        var ( input *gtk.ComboBoxText; err error )
        if ctl.FreeEntry {
            input, err = gtk.ComboBoxTextNewWithEntry( )
        } else {
            input, err = gtk.ComboBoxTextNew( )
        }
        if err != nil {
            return nil, fmt.Errorf( "addTextInput: cannot create comboText: %v\n", err )
        }
        for i, v := range ctl.List {
            input.AppendText( v )
            if textVal == v {
                input.SetActive( i )
            }
        }
        if ctl.FreeEntry {
            child, err := input.Bin.GetChild()
            if err != nil {
                return nil,
                       fmt.Errorf( "addTextInput: unable to get entry: %v\n",
                                   err )
            }
            entry := child.(*gtk.Entry)
            entry.SetMaxLength( ctl.InputMax )
            entry.SetCanFocus( true )

            if ctl.MouseBut != nil {
                mouseNotify := func( wg *gtk.Entry, ev *gdk.Event ) bool {
                    evb := gdk.EventButtonNewFromEvent( ev )
                    id := evb.Button()
                    return ctl.MouseBut( def.Name, id )
                }
                entry.Connect( "button-press-event", mouseNotify )
            }
            if ctl.KeyPress != nil {
                keyNotify := func( wg *gtk.Entry, ev *gdk.Event ) bool {
                    evk := gdk.EventKeyNewFromEvent(ev)
                    return ctl.KeyPress( def.Name, evk.KeyVal(),
                                         KeyModifier(evk.State() & 0x0f) )
                }
                entry.Connect( "key-press-event", keyNotify )
            }
        }
        if def.Changed != nil {
            textChanged := func( b *gtk.ComboBoxText ) bool {
                t := b.GetActiveText()
                if t != "" {
                    return def.Changed( def.Name, t )
                }
                return false
            }
            input.Connect( "changed", textChanged )
        }
        return &itemReference{ nil, false, input }, nil
    }
    var ( lenCtl *StrCtl; ok bool )
    if def.Control != nil {
        lenCtl, ok = def.Control.(*StrCtl)
        if ! ok {
            return nil, fmt.Errorf("addTextInput: unsupported control %T\n",
                                   def.Control)
        }
    }   // accept nil control as no restrictions

    input, err := gtk.EntryNew( )
    if nil != err {
        return nil, fmt.Errorf("addTextInput: Could not create text input:", err)
    }
    if ok {
        input.SetMaxLength( lenCtl.InputMax )
    }
    input.SetText( textVal )
    textChanged := func( e *gtk.Entry ) bool {
        t, err := e.GetText( )
        if err != nil {
            log.Fatal("addTextInput: can't get entry text after change:", err )
        }
        if t != "" {
            return def.Changed( def.Name, t )
        }
        return false
    }
    input.Connect( "activate", textChanged )
    return &itemReference{ nil, false, input }, nil
}

func isKeyHexa( keyVal uint ) (hexa bool) {
    if keyVal & 0xff00 == 0 {                       // regular keys
        b := byte(keyVal & 0xff)
        if b < '0' || (b < 'A' && b > '9') || (b < 'a' && b > 'F') || b > 'f' {
            return
        }
    } else if keyVal & 0xfff0 != 0xffb0 {           // not num from keypad
        return
    }
    return true
}

func HexaFilter( key uint, mod KeyModifier ) bool {
    switch key {
    case gdk.KEY_Home, gdk.KEY_End, gdk.KEY_Left, gdk.KEY_Right,
         gdk.KEY_Insert, gdk.KEY_BackSpace, gdk.KEY_Delete,
         gdk.KEY_Return, gdk.KEY_KP_Enter:
        return false

    default:
        if mod & gdk.CONTROL_MASK != 0 {
            return false
        }
        if hexa := isKeyHexa( key ); hexa {
            return false
        }
        return true
    }
}

func isKeyDecimal( keyVal uint ) (decimal bool) {
    if keyVal & 0xff00 == 0 {                       // regular keys
        b := byte(keyVal & 0xff)
        if b < '0' || b > '9' {                     // decimal digits
            return
        }
    } else if keyVal & 0xfff0 != 0xffb0 {           // ! num from keypad
        return
    }
    return true
}

func deciFilter( key uint, mod KeyModifier ) bool {
    switch key {
    case gdk.KEY_Home, gdk.KEY_End, gdk.KEY_Left, gdk.KEY_Right,
         gdk.KEY_Insert, gdk.KEY_BackSpace, gdk.KEY_Delete,
         gdk.KEY_Return, gdk.KEY_KP_Enter:
        return false

    default:
        if mod & gdk.CONTROL_MASK != 0 {
            return false
        }
        if isKeyDecimal( key ) {
            return false
        }
        return true
    }
}

// decimalKey prevents entering non decimal integer values
func decimalKey( entry *gtk.Entry, event *gdk.Event ) bool {
    ev := gdk.EventKeyNewFromEvent(event)
    return deciFilter( ev.KeyVal(), KeyModifier(ev.State() & 0x0f) )
}

// nokey prevents entering data otherwise than by spinner (with min, max  & inc)
func noKey( ) bool {
    return true
}

func makeIntInput( def *InputDef ) (*itemReference, error) {

    intVal := def.Value.(int)

    if ctl, ok := def.Control.(*IntCtl); ok {
        input, err := gtk.SpinButtonNewWithRange( float64(ctl.InputMin),
                                                  float64(ctl.InputMax),
                                                  float64(ctl.InputInc) )
        if nil != err {
            return nil, fmt.Errorf("addIntInput: Could not create input button:", err)
        }
        input.SetValue( float64(intVal) )
        input.Entry.Connect( "key-press-event", noKey )
        valueChanged := func( button *gtk.SpinButton ) bool {
            v := button.GetValue()
            return def.Changed( def.Name, v )
        }
        input.Connect( "value-changed", valueChanged )
        return &itemReference{ nil, true, input }, nil

    } else if ctl, ok := def.Control.(*IntList); ok {
        var ( input *gtk.ComboBoxText; err error )
        if ctl.FreeEntry {
            input, err = gtk.ComboBoxTextNewWithEntry( )
        } else {
            input, err = gtk.ComboBoxTextNew( )
        }
        if err != nil {
            return nil, fmt.Errorf( "addIntInput: cannot create comboText: %v\n", err )
        }
        for i, v := range ctl.List {
            input.AppendText( fmt.Sprintf("%d", v) )
            if intVal == v {
                input.SetActive( i )
            }
        }
        if ctl.FreeEntry {
            child, err := input.Bin.GetChild( )
            if err != nil {
                return nil,
                       fmt.Errorf( "addTextInput: unable to get entry: %v\n",
                                   err )
            }
            entry := child.(*gtk.Entry)
            entry.Connect( "key-press-event", decimalKey )
        }
        valueChanged := func( b *gtk.ComboBoxText ) bool {
            t := b.GetActiveText()
            if t != "" {
                v, err := strconv.Atoi( t )
                if err != nil {
                    log.Fatal("addIntInput: can't get number after change:", err )
                }
                return def.Changed( def.Name, float64(v) )
            }
            return false
        }
        input.Connect( "changed", valueChanged )
        return &itemReference{ nil, true, input }, nil

    } else if def.Control != nil {
        return nil, fmt.Errorf( "addIntInput: unsupported control: %T\n",
                                def.Control )
    }
    input, err := gtk.EntryNew()        // no control, use default ENTRY
    if err != nil {
        return nil, fmt.Errorf( "addIntInput: cannot create entry: %v\n", err )
    }
    input.Connect( "key-press-event", decimalKey )
    input.SetText( fmt.Sprintf("%d", intVal ) )
    valueChanged := func( e *gtk.Entry ) bool {
        t, err := e.GetText( )
        if err != nil {
            log.Fatal("addIntInput: can't get entry after change:", err )
        }
        if t != "" {
            v, err := strconv.Atoi( t )
            if err != nil {
                log.Fatal("addIntInput: can't get number after change:", err )
            }
            return def.Changed( def.Name, float64(v) )
        }
        return false
    }
    input.Connect( "activate", valueChanged )
    return &itemReference{ nil, true, input }, nil
}

func newBoolObject( val bool ) (input *gtk.CheckButton, err error) {

    input, err = gtk.CheckButtonNew( )
    if nil == err {
        input.ToggleButton.SetActive( val )
    }
    return
}

func makeBoolInput( def *InputDef ) (*itemReference, error) {

    input, err := newBoolObject( def.Value.(bool) )
    if err != nil {
        return nil, fmt.Errorf("addBoolInput: cannot create boolean: %v", err)
    }
    if def.Changed == nil {
        input.SetSensitive( false )
    } else {
        toggled := func( button *gtk.CheckButton ) bool {
            v := button.ToggleButton.GetActive()
            return def.Changed( def.Name, v )
        }
        input.Connect( "toggled", toggled )
    }
    return &itemReference{ nil, false, input }, nil
}

func wrapChildInHorizontalBox( child gtk.IWidget, padding uint ) *gtk.Box {
    innerBox, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 10 )
    if nil != err {
        log.Fatal("wrapChidInHorizontalBox: Could not create inner box:", err)
    }
    innerBox.PackStart( child, true, true, padding )
    return innerBox
}

func (lo *Layout)addInputItem( def *InputDef ) (*itemReference, error) {

    var ( err error; itemRef *itemReference )
    switch def.Value.(type) {
    case bool:
        itemRef, err = makeBoolInput( def )
    case int:
        itemRef, err = makeIntInput( def )
    case string:
        itemRef, err = makeTextInput( def )
    case *TextDef, *IconDef:
        itemRef, err = makeButtonInput( def )
    default:
        fmt.Printf( "addInputItem: unsupported type %T\n", def.Value )
    }
    if err != nil {
        return nil, fmt.Errorf("addInputItem: unable to create input: %v", err)
    }
    lo.addItem( itemRef, def.Name, def.ToolTip )
    if def.Padding > 0 {
        return &itemReference{ itemRef.format, itemRef.strIsInt,
                        wrapChildInHorizontalBox( itemRef.item.(gtk.IWidget),
                                                  def.Padding )}, nil
    }
    return itemRef, nil
}

func (lo *Layout)addItem( ir *itemReference, name string, tooltip string ) {
    if name != "" {
        lo.access[name] = ir
    }
    if tooltip != "" {
        wg := ir.item.(gtk.IWidget).ToWidget()
        wg.SetTooltipText( tooltip )
    }
}

func makeMarkup( format *TextFmt, value string ) string {

    if format == nil || format.Attributes == 0 {
        return value
    }

    const (
        MARKUP_PREFIX = "<span"
        MONOSPACE_MARKUP = " face=\"monospace\""
        BOLD_MARKUP = " weight=\"bold\""
        ITALIC_MARKUP = " style=\"italic\""
        MARKUP_STOP = ">"
        MARKUP_SUFFIX = "</span>"
    )
    var b strings.Builder
    b.WriteString( MARKUP_PREFIX )
    if format.Attributes & MONOSPACE == MONOSPACE {
        b.WriteString( MONOSPACE_MARKUP )
    }
    if format.Attributes & BOLD == BOLD {
        b.WriteString( BOLD_MARKUP )
    }
    if format.Attributes & ITALIC == ITALIC {
        b.WriteString( ITALIC_MARKUP )
    }
    b.WriteString( MARKUP_STOP )
    b.WriteString( value )
    b.WriteString( MARKUP_SUFFIX )
    return b.String()
}

func (lo *Layout)addConstText( text string,
                               def *ConstDef ) (*itemReference, error) {

    constant, err := gtk.LabelNew( "" )
    if nil != err {
       return nil, 
              fmt.Errorf("addConstText: Could not create label %s: %v",
                         text, err)
    }
    format := def.Format
    value := makeMarkup( format, text )
    constant.SetMarkup( value )

    if format == nil {
        item, err := wrapInFrame( constant, "", false )
        if err != nil {
            return nil,
                   fmt.Errorf("addConstText: Could not create frame %s: %v",
                              text, err)
        }
        return &itemReference{ nil, false, item }, nil
    }

    if format.FrameSize > 0 {
        constant.SetWidthChars( format.FrameSize )
        constant.SetMaxWidthChars( format.FrameSize )
    }

    switch format.Align {
    case LEFT:
        constant.SetXAlign( 0.0 )
    case CENTER:
        constant.SetXAlign( 0.5 )
    case RIGHT:
         constant.SetXAlign( 1.0 )
    }
    lo.addItem( &itemReference{ format, false, constant }, def.Name, def.ToolTip )
    var item interface{}
    if format.Copy != nil {
        eb, err := gtk.EventBoxNew( )
        if err != nil {
            return nil,
                   fmt.Errorf("addConstText: could not create event box: %v",
                               err)
        }
        eb.SetAboveChild( true )
        eb.Add( constant )
        cc := func( eventbox *gtk.EventBox, event *gdk.Event ) bool {
            buttonEvent := gdk.EventButtonNewFromEvent( event )
            evButton := buttonEvent.Button()

            if evButton != gdk.BUTTON_PRIMARY {
                return format.Copy( def.Name, event )
            }
            return false
        }
        eb.Connect( "button_press_event", cc )
        item, err = wrapInFrame( eb, "", format.Border )
    } else {
        item, err = wrapInFrame( constant, "", format.Border )
    }
    if err != nil {
        return nil, fmt.Errorf("addConstText: could not create frame: %v",
                                err )
    }
    return &itemReference{ format, false, item }, nil
}

func (lo *Layout) addConstBool( def *ConstDef ) (*itemReference, error) {
    bob, err := newBoolObject( def.Value.(bool) )
    if err != nil {
        return nil, fmt.Errorf("addConstBool: cannot create boolean: %v", err)
    }
    bob.SetSensitive( false )
    iRef := itemReference{ nil, false, bob }
    lo.addItem( &iRef, def.Name, def.ToolTip )
    return &iRef, nil
}

func (lo *Layout) addConstItem( def *ConstDef ) (*itemReference, error) {
    var ( err error; itemRef *itemReference )
    switch val := def.Value.(type) {
    case bool:
        itemRef, err = lo.addConstBool( def )
    case int:
        itemRef, err = lo.addConstText( fmt.Sprintf("%d", val), def )
    case string:
        itemRef, err = lo.addConstText( val, def )
    default:
        return nil, fmt.Errorf( "addConstantItem: unsupported type %T\n", def )
    }
    if err != nil {
        return nil, fmt.Errorf("addConstantItem: unable to create item: %v", err)
    }
    if def.Padding > 0 {
        item := itemRef.item.(gtk.IWidget)
        return &itemReference{ itemRef.format, itemRef.strIsInt,
                            wrapChildInHorizontalBox( item, def.Padding )}, nil
    }
    return itemRef, nil
}

func wrapInFrame( w gtk.IWidget,
                  title string, border bool ) (*gtk.Frame, error) {

    frame, err := gtk.FrameNew( title )
    if err != nil {
        return nil, fmt.Errorf("wrapInFrame: Could not create frame", err)
    }
    frame.Add( w )
    if border {
        frame.SetShadowType( gtk.SHADOW_IN )
    } else {
        frame.SetShadowType( gtk.SHADOW_NONE )
    }
    return frame, nil
}

func (lo *Layout)addBoxItem( def * BoxDef ) (*itemReference, error) {

    box, err := gtk.BoxNew( gtk.Orientation(def.Direction), int(def.Spacing) )
    if nil != err {
        return nil, fmt.Errorf("addBoxItem: Could not create box: %v", err)
    }

    var itemRef *itemReference
    for i, itemDef := range def.ItemDefs {

        switch itemDef := itemDef.(type) {
        case *BoxDef:
            itemRef, err = lo.addBoxItem( itemDef)
        case *GridDef:
            itemRef, err = lo.addGridItem( itemDef )
        case *ConstDef:
            itemRef, err = lo.addConstItem( itemDef )
        case *InputDef:
            itemRef, err = lo.addInputItem( itemDef )
        default:
            return nil, fmt.Errorf("addBoxItem: unsupported type %T\n", itemDef)
        }
        if err != nil {
            return nil, fmt.Errorf("addBoxItem: could not create item: %v", err)
        }
        widget := itemRef.item.(gtk.IWidget).ToWidget()
        if i == 0 {
            if def.Direction == HORIZONTAL {
                widget.SetMarginStart( def.Margin )
            } else {
                widget.SetMarginTop( def.Margin )
            }
        }
        if i == len(def.ItemDefs) - 1 {
            if def.Direction == HORIZONTAL {
                widget.SetMarginEnd( def.Margin )
            } else {
                widget.SetMarginBottom( def.Margin )
            }
        }
        box.PackStart( widget, false, false, 0 )
    }
    frame, err := wrapInFrame( box, def.Title, def.Border )
    if err != nil {
        return nil, fmt.Errorf("addBoxItem: Could not create frame: %v", err)
    }
    if def.Name != "" {
        lo.access[def.Name] = &itemReference{ nil, false, frame }
    }
    if def.Padding > 0 {
        return &itemReference{ nil, false,
                        wrapChildInHorizontalBox( frame, def.Padding )}, nil
    }
    return &itemReference{ nil, false, frame }, nil
}

func (lo *Layout)addItemToGrid( dg *DataGrid, c, r int, 
                                 itemDef interface{} ) (err error) {
    var itemRef *itemReference
    switch itemDef := itemDef.(type) {
    case nil:
        return
    case *BoxDef:
        itemRef, err = lo.addBoxItem( itemDef )
    case *GridDef:
        itemRef, err = lo.addGridItem( itemDef )
    case *ConstDef:
        itemRef, err = lo.addConstItem( itemDef )
    case *InputDef:
        itemRef, err = lo.addInputItem( itemDef )
    default:
        return fmt.Errorf( "addItem: unsupported type %T\n", itemDef )
        
    }
    if err == nil {
        item := itemRef.item

        item.(gtk.IWidget).ToWidget().SetHExpand( dg.colExpand[c] )
        item.(gtk.IWidget).ToWidget().SetVExpand( dg.rowExpand[r] )

        dg.grid.Attach( item.(gtk.IWidget), c, r, 1, 1 )

        /// keep track of the item in the column c, at row r
        dg.colItems[c] = append( dg.colItems[c], itemRef )
        // add item to grid rows and columns
        dg.rowItems[r] = append( dg.rowItems[r], itemRef )
    }
    return
}

type DataGrid struct {
    grid        *gtk.Grid

    colExpand,
    rowExpand   []bool
    colItems,
    rowItems    [][]*itemReference
}

type Layout struct {
    root   gtk.IWidget
    access map[string]*itemReference
}

func (lo *Layout) addGridItem( def *GridDef ) (ir *itemReference, err error) {
    dg := new(DataGrid)
    dg.grid, err = gtk.GridNew( )
    if err != nil {
        return
    }
    dg.grid.SetColumnSpacing( def.H.Spacing )
    dg.grid.SetRowSpacing( def.V.Spacing )

    dg.colItems = make( [][]*itemReference, len(def.H.Columns) )
    dg.rowItems = make( [][]*itemReference, len(def.V.Rows) )

    dg.colExpand = make( []bool, len(def.H.Columns) )
    dg.rowExpand = make( []bool, len(def.V.Rows) )

    for c, colDef := range def.H.Columns {
        dg.colExpand[c] = colDef.Expand
    }
    for r, rowDef := range def.V.Rows {
        dg.rowExpand[r] = rowDef.Expand
        for c, itemDef := range rowDef.Items {
            if err = lo.addItemToGrid( dg, c, r, itemDef ); err != nil {
                return
            }
        }
    }
    if def.Name != "" {
        lo.access[def.Name] = &itemReference{ nil, false, dg }
    }
    if def.Padding > 0 {
        ir = &itemReference{ nil, false,
                             wrapChildInHorizontalBox( dg.grid, def.Padding ) }
    } else {
        ir = &itemReference{ nil, false, dg.grid }
    }
    return
}

