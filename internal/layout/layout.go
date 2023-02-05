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

type Layout struct {
    root   gtk.IWidget
    access map[string]*itemReference
}

func (lo *Layout)GetRootWidget() *gtk.Widget {
    return lo.root.ToWidget()
}

func (lo *Layout)GetItemNames() (names []string) {
    names = make( []string, len(lo.access) )
    i := 0
    for name := range lo.access {
        names[i] = name
        i++
    }
    return
}

func (lo *Layout)View( status bool ) {
    if lo.root != nil {
        lo.root.ToWidget().SetVisible( status )
    }
}

func MakeLayout( def interface{} ) (layout *Layout, err error) {

    layout = new(Layout)
    layout.access = make( map[string]*itemReference )

    var itemRef *itemReference
    switch def := def.(type) {
    case *GridDef:
        itemRef, err = layout.addGridItem( def )
    case *BoxDef:
        itemRef, err = layout.addBoxItem( def )
    case *ConstDef:
        itemRef, err = layout.addConstItem( def )
    case *InputDef:
        itemRef, err = layout.addInputItem( def )
    default:
        return nil, fmt.Errorf( "makeLayout: unsupported type %T\n", def )
    }
    if err != nil {
        return nil, fmt.Errorf( "makeLayout: unable to make root: %v", err )
    }
    layout.root = itemRef.item.(gtk.IWidget)
    return
}

func (dg *DataGrid)SetRowVisible( rowIndex int, visible bool ) error {
    if rowIndex < 0 || rowIndex >= len(dg.rowItems) {
        return fmt.Errorf( "setRowVisible: row index %d out of range [0:%d[\n",
                           rowIndex, len(dg.rowItems) )
    }
    rowItems := dg.rowItems[rowIndex]
    for _, colItemRef := range rowItems {
        wdg := colItemRef.item.(gtk.IWidget).ToWidget()
        wdg.SetVisible( visible )
    }
    return nil
}

func (dg *DataGrid)SetColVisible( colIndex int, visible bool ) error {
    if colIndex < 0 || colIndex >= len(dg.colItems) {
        return fmt.Errorf( "setRowVisible: col index %d out of range [0:%d[\n",
                           colIndex, len(dg.colItems) )
    }
    colItems := dg.colItems[colIndex]
    for _, rowItemRef := range colItems {
        wdg := rowItemRef.item.(gtk.IWidget).ToWidget()
        wdg.SetVisible( visible )
    }
    return nil
}

func (lo *Layout)getButton( name string ) (*TextFmt, *gtk.Button, error) {
    ref, ok := lo.access[name]
    if ! ok {
        return nil, nil, fmt.Errorf("item %s does not exist\n", name )
    }
    switch item := ref.item.(type) {
    case *gtk.ToggleButton:
        return ref.format, &item.Button, nil
    case *gtk.Button:
        return ref.format, item, nil
    default:
        break
    }
    return nil, nil, fmt.Errorf( "item %s is not a button\n", name )
}

func (lo *Layout)GetButtonLabel( name string ) (string, error) {
    _, button, err := lo.getButton( name )
    if err != nil {
        return "", fmt.Errorf("GetButtonLabel: %v\n", err )
    }
    return button.GetLabel()
}

func (lo *Layout)SetButtonLabel( name string, label string ) error {
    format, button, err := lo.getButton( name )
    if err != nil {
        return fmt.Errorf("SetButtonLabel: %v\n", err )
    }
    markup := makeMarkup( format, label )
    button.SetLabel( markup )
    return nil
}

func (lo *Layout) GetButtonActive( name string ) (bool, error) {
    _, button, err := lo.getButton( name )
    if err != nil {
        return false, fmt.Errorf("GetButtonActive: %v\n", err )
    }
    return button.GetSensitive( ), nil
}

func (lo *Layout) SetButtonActive( name string, state bool ) error {
    _, button, err := lo.getButton( name )
    if err != nil {
        return fmt.Errorf("SetButtonActive: %v\n", err )
    }
    button.SetSensitive( state )
    return nil
}

func (lo *Layout)GetItemTooltip( name string ) (string, error) {
    ref, ok := lo.access[name]
    if ! ok {
        return "", fmt.Errorf("GetItemTooltip: item %s does not exist\n", name )
    }
    var wg *gtk.Widget
    switch item := ref.item.(type) {
    case *gtk.CheckButton:
        wg = item.ToWidget()
    case *gtk.ToggleButton:
        wg = item.ToWidget()
    case *gtk.Button:
        wg = item.ToWidget()
    case *gtk.Label:
        wg = item.ToWidget()
    case *gtk.SpinButton:
        wg = item.ToWidget()
    case *gtk.ComboBoxText:
        wg = item.ToWidget()
    case *gtk.Entry:
        wg = item.ToWidget()
    default:
        break
    }
    if wg != nil {
        return wg.GetTooltipMarkup()
    }
    return "", fmt.Errorf( "GetItemTooltip: item %s is not a button\n", name )
}

func (lo *Layout)SetItemTooltip( name string, tooltip string ) error {
    ref, ok := lo.access[name]
    if ! ok {
        return fmt.Errorf("SetItemTooltip: item %s does not exist\n", name )
    }
    var wg *gtk.Widget
    switch item := ref.item.(type) {
    case *gtk.CheckButton:
        wg = item.ToWidget()
    case *gtk.ToggleButton:
        wg = item.ToWidget()
    case *gtk.Button:
        wg = item.ToWidget()
    case *gtk.Label:
        wg = item.ToWidget()
    case *gtk.SpinButton:
        wg = item.ToWidget()
    case *gtk.ComboBoxText:
        wg = item.ToWidget()
    case *gtk.Entry:
        wg = item.ToWidget()
    default:
        break
    }
    if wg != nil {
        wg.SetTooltipMarkup( tooltip )
        return nil
    }
    return fmt.Errorf( "SetItemTooltip: item %s is not a button\n", name )
}

// getItemValue returns the current value associated with the given item name
// Depending on item type, returned values can be:
//   - bool for constant or input bool and for toogle button
//   - int64 for constant or input int
//   - string for constant or input string
// Error is returned if the given name does not match any known item or if the
// value type does not match the expected item value. Since press button have no
// value, getting its value returns an error.
func (lo *Layout) GetItemValue( name string ) (interface{}, error) {
    ref, ok := lo.access[name]
    if ! ok {
        return nil, fmt.Errorf("getItemValue: item %s does not exist\n", name )
    }
    switch item := ref.item.(type) {
    case *DataGrid:
        return item, nil
    case *gtk.CheckButton:
        return item.ToggleButton.GetActive(), nil

    case *gtk.ToggleButton:
        return item.GetActive(), nil

    case *gtk.Button:
        return nil, fmt.Errorf("getItemValue: item %s does not have a value\n", name )

    case *gtk.Label:
        t, err := item.GetText( )
        if err == nil {
            if ref.strIsInt {
                return convertTextToInt64( t ), nil
            }
            return t, nil
        }

        return nil, fmt.Errorf( "getItemValue: cannot get label for item %s\n",
                                name)
    case *gtk.SpinButton:
        return int64(item.GetValue( ) ), nil

    case *gtk.ComboBoxText:
        t := item.GetActiveText( )
        if ref.strIsInt {
            return convertTextToInt64( t ), nil
        }
        return t, nil

    case *gtk.Entry:
        t, err := item.GetText()
        if err == nil {
            if ref.strIsInt {
                return convertTextToInt64( t ), nil
            }
            return t, nil
        }
        return nil, fmt.Errorf("getItemValue: cannot get entry for item %s\n",
                               name)

    default:
        log.Printf( "getItemValue: unexpected internal type %T\n", item )
        panic( "FIX IT" )
    }
}

func setComboBoxTextValue( item *gtk.ComboBoxText, text string ) error {
//fmt.Printf("setItemValue: updating ComboBoxText with text %s\n", text )

    bin := item.ComboBox.Bin
    entry, err := bin.GetChild()
    if err != nil {
        return fmt.Errorf("unable to get child entry: %v", err)
    }
    if entry != nil {               // if with entry, accept text
        entry := entry.(*gtk.Entry)
        entry.SetText( text )
        return nil
    }

    model, err := item.GetModel()   // else accept if text is in model
    if err != nil {
        return err
    }
    list := model.ToTreeModel()
    var index = 0
    if iter, nonEmpty := list.GetIterFirst( ); nonEmpty {
        for {
            v, err := list.GetValue( iter, 0 )
            if err != nil {
                return err
            }
            var ls string
            ls, err = v.GetString()
            if err != nil {
                return err
            }
            if ls == text {
                item.SetActive( index )
                return nil
            }
            index++
            if false == list.IterNext( iter ) {
                break
            }
        }
    }
    return fmt.Errorf( "\"%s\" is not in combo box list\n", text )
}

func setToggleButtonValue( button *gtk.ToggleButton, value interface{} ) bool {
    if v, ok := value.(bool); ok {
        button.SetActive( v )
        return true
    }
    return false
}

func convertTextToInt64( t string ) int64 {
    v, e := strconv.ParseInt( t, 10, 64)
    if e != nil {
        log.Fatalf( "getIntValue: cannot convert constant text %s to int\n", t )
    }
    return v
}

func (lo *Layout) SetItemValue( name string, value interface{} ) error {
    ref, ok := lo.access[name]
    if ! ok {
        return fmt.Errorf("setItemValue: item %s does not exist\n", name )
    }
    switch item := ref.item.(type) {
    case *gtk.CheckButton:
        if setToggleButtonValue( &item.ToggleButton, value ) {
            return nil
        }
        return fmt.Errorf("setItemValue: wrong value type %T for bool item %s\n",
                          value, name )

    case *gtk.ToggleButton:
        if setToggleButtonValue( item, value ) {
            return nil
        }
        return fmt.Errorf("setItemValue: wrong value type %T for toggle button item %s\n",
                          value, name )

    case *gtk.Button:
        return fmt.Errorf("setItemValue: press button have no value\n")

    case *gtk.Label:
        if ref.strIsInt {
            if v, ok := value.(int); ok {
                item.SetText( fmt.Sprintf("%d", v) )
                return nil
            }
        } else {
            if t, ok := value.(string); ok {
                v := makeMarkup( ref.format, t )
                item.SetMarkup( v )
                return nil
            }
        }
        return fmt.Errorf("setItemValue: wrong value type %T for item %s\n",
                          value, name )

    case *gtk.ComboBoxText:
        var err error = fmt.Errorf("wrong value type %T\n", value )
        if ref.strIsInt {
            if v, ok := value.(int); ok {
                err = setComboBoxTextValue( item, fmt.Sprintf("%d", v) )
            }
        } else if v, ok := value.(string); ok {
            err = setComboBoxTextValue( item, v )
        }

        if err != nil {
            return fmt.Errorf("setItemValue: failed for item %s: %v", name, err)
        }
        return nil

    case *gtk.Entry:
        if ref.strIsInt {
            if v, ok := value.(int); ok {
                item.SetText( fmt.Sprintf("%d", v) )
                return nil
            }
        } else if v, ok := value.(string); ok {
            item.SetText( v )
            return nil
        }
        return fmt.Errorf("setItemValue: wrong value type %T for item %s\n",
                          value, name )

    case *gtk.Frame:
        if t, ok := value.(string); ok {
            v := makeMarkup( ref.format, t )
            item.SetLabel( v )
            return nil
        }
        return fmt.Errorf("setItemValue: wrong value type %T for item %s\n",
                          value, name )

    default:
        log.Printf( "setItemValue: unexpected internal type %T\n", item )
        panic( "FIX IT" )
    }
}

func (lo *Layout) getItemEntry( name string ) (*gtk.Entry, error) {
    ref, ok := lo.access[name]
    if ! ok {
        return nil, fmt.Errorf("item %s does not exist\n", name )
    }
    switch item := ref.item.(type) {
    case *gtk.ComboBoxText:
        bin := item.ComboBox.Bin
        entry, err := bin.GetChild()
        if err != nil {
            return nil, fmt.Errorf("unable to get entry: %v", err)
        }
        if entry != nil {
            return entry.(*gtk.Entry), nil
        }

    case *gtk.Entry:
        return item, nil

    default:
        break
    }
    return nil, fmt.Errorf("item %d has no entry", name)
}

func (lo *Layout) SetEntryFocus( name string, noSelection bool ) error {
    entry, err := lo.getItemEntry( name )
    if err != nil {
        return fmt.Errorf("SetEntryFocus: %v", err )
    }
    if noSelection {
        entry.GrabFocusWithoutSelecting()
    } else  {
        entry.Widget.GrabFocus()
    }
    return nil
}

func (lo *Layout) SetEntrySelection( name string, start, beyond int ) error {
    entry, err := lo.getItemEntry( name )
    if err != nil {
        return fmt.Errorf("setEntrySelection: %v", err )
    }
    entry.SelectRegion( start, beyond )
    return nil
}

func (lo *Layout) SetEntryCursor( name string, position int ) error {
    entry, err := lo.getItemEntry( name )
    if err != nil {
        return fmt.Errorf("setEntrySelection: %v", err )
    }
    entry.SetPosition( position )
    return nil
}

func (lo *Layout) SetItemChoices( name string,
                                  choices interface{},
                                  active int,
                                  changed func( name string,
                                                val interface{} ) bool ) error {
    ref, ok := lo.access[name]
    if ! ok {
        return fmt.Errorf( "setItemChoices: item %s does not exist\n", name )
    }
    var (
        l          int
        isInt      bool
        intChoices []int
        strChoices []string
    )
    switch c := choices.(type) {
    case []string:
        strChoices = c
    case []int:
        intChoices = c
        isInt = true
    default:
        return fmt.Errorf( "setItemChoices: choices %T are not supported\n", choices )
    }

    switch item := ref.item.(type) {
    case *gtk.ComboBoxText:
        item.RemoveAll()
        if isInt {
            l = len(intChoices)
            for _, choice := range intChoices {
                t := fmt.Sprintf("%d", choice )
                item.AppendText( t )
            }
        } else {
            l = len(strChoices)
            for _, choice := range strChoices {
                item.AppendText( choice )
            }
        }
        if active >= 0 && active < l {
            item.SetActive( active )
        }
        if changed != nil {
            var textChanged func( b *gtk.ComboBoxText ) bool
            if isInt {
                textChanged = func( b *gtk.ComboBoxText ) bool {
                    t := b.GetActiveText()
                    if t != "" {
                        v, err := strconv.Atoi(t)
                        if err == nil {
                            return changed( name, v )
                        }
                    }
                    return false
                }
            } else {
                textChanged = func( b *gtk.ComboBoxText ) bool {
                    t := b.GetActiveText()
                    if t != "" {
                        return changed( name, t )
                    }
                    return false
                }
            }
            item.Connect( "changed", textChanged )
        }
        if isInt {      // filter out non decimal keys
            child, _ := item.Bin.GetChild( )
            if child != nil {
                entry := child.(*gtk.Entry)
                entry.Connect( "key-press-event", decimalKey )
            }
        }
        return nil

    default:
        return fmt.Errorf( "setItemChoices: unsupported item type %T\n", item )
    }
}

