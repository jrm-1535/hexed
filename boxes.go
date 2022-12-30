package main

import (
    "log"
    "fmt"
    "strconv"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)


// generic content box management

type orientation gtk.Orientation
const (
    HORIZONTAL  orientation = orientation(gtk.ORIENTATION_HORIZONTAL)
    VERTICAL                = orientation(gtk.ORIENTATION_VERTICAL)
)

// The top level widget is always a vertical frame (which may not be visible)
// that contains a box, which contains children: other frames, simple header or
// content. A box can be vertical or horizontal, and children are packed
// starting from the top or left, respectively. Header is just a string within
// an invisible horizontal frame with padding on top and on left sides.
// Content is an horizontal frame for a name (string) and an associated value.
// content-value types are limited to boolean, numbers or strings. Content
// value can be constant (not modifiable from UI) or variable. Content can
// have a name used to identify its value, either through a callback after
// the value has been modified by UI or to programmatically change the value
// while the dialog is visible.
// The content name is packed starting from the left and value is packed
// starting from the end, in order to fill up the whole frame.
type boxDef struct {
    spacing     int             // space between box chidren
    padding     uint            // padding in parent box, if any

    frame       bool            // visible frame around box
    title       string          // optional frame title

    direction   orientation     // box orientation
    content     []interface{}   // boxDef, contentDef or header
}

// contentDef gives the initial value, what can be entered and what to do when
// a value has been modified. The function "changed" is called when an initial
// value is modified.
// if "changed" is given as nil, the initial value cannot be modified by the
// user, although it can still be programmatically modified.
// initVal gives the initial input value (usually a current preference value)
// valueCtl restricts what can be the value
// - a nil valueCtl means no restrictions (typical if changed is nil)
// - If initVal is a string, valueCtl can be:
//   - []string for a list of possible inputs
//   - lengthCtl for the max length of the input or
//     constCtl if changed is missing
// - If initVal is an int, valueCtl can be:
//   - []int for a list of possible inputs
//   - inputCtl for the min, max and increment of the input value
// - valueCtl is not used for boolean
type contentDef struct {
    label       string      // text on left hand  (may be empty)
    labelCtl    *constCtl   // optional left hand text styling
    name        string      // passed as 1st arg to changed and/or used to change
                            // value programmatically after the dialog is shown

    initVal     interface{} // initial value
    valueCtl    interface{} // value control
    changed     func( name string, val interface{} )    // change notification

    padding     uint        // padding in parent box
}

type inputCtl struct {      // for integer input
    inputMin    int
    inputMax    int
    inputInc    int
}

type lengthCtl struct {     // for text input
    inputMax    int
}

const (
    LEFT_ALIGN = -1
    CENTER_ALIGN = 0
    RIGHT_ALIGN = 1
)

type constCtl struct {   // for label or constant text value (if changed is nil)
    align       int         // -1 left, 0 center, +1 right
    size        int         // number of chars within frame
    frame       bool        // if frame around total space
    monosize    bool        // if all characters have same size
    canCopy     bool        // constant can be copied in clipboard
}

// header is a simple label with optional top & left padding
type headerDef  struct {
    label       string      //
    top, left   uint
}

// box content values can be header, content(bool, int or string) or box
// Horizontal box is made of a series of header and content 
// [ header label] [ label 1  value 1 entry ] ... [ label n value n entry ]
// Vertical box is made of a series of horizontal boxes with header or content
// [ header label                     ] label alone
// [  content label 1   value 1 entry ] label + value per content
// [[ content label 1   value 1 entry ]...[ content label n   value n entry]]

// nokey prevents entering data otherwise than by spinner (with min, max  & inc)
func noKey( ) bool {
    return true
}

func addBoolContent( innerBox *gtk.Box, content *contentDef,
                     access map[string]interface{} ) {
//    fmt.Printf( "got input type bool\n" )
    boolVal := content.initVal.(bool)

    input, err := gtk.CheckButtonNew( )
    if nil != err {
        log.Fatal("addBoolContent: Could not create bool bool input button:", err)
    }
    input.ToggleButton.SetActive( boolVal )
    if content.changed == nil {
        input.SetSensitive( false )
    } else {
        toggled := func( button *gtk.CheckButton ) {
            v := button.ToggleButton.GetActive()
            //fmt.Printf("Notifying content changed: %s=%v\n", content.name, v )
            content.changed( content.name, v )
        }
        input.Connect( "toggled", toggled )
    }
    if access != nil && content.name != "" {
        access[content.name] = input
    }
    innerBox.PackEnd( input, false, false, content.padding )
}

func setBoolContent( db *dataBox, name string, val bool ) {
    input, ok := db.access[name].(gtk.CheckButton)
    if ! ok {
        log.Fatalf("setBoolContent: no such bool value %s\n", name )
    }
    input.ToggleButton.SetActive( val )
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

// decimalKey prevents entering non decimal integer values
func decimalKey( entry *gtk.Entry, event *gdk.Event ) bool {
    ev := gdk.EventKeyNewFromEvent(event)
    key := ev.KeyVal()
    switch key {
    case HOME_KEY, END_KEY, LEFT_KEY, RIGHT_KEY,
         INSERT_KEY, BACKSPACE_KEY, DELETE_KEY,
         ENTER_KEY, KEYPAD_ENTER_KEY:
//        fmt.Printf("Got Special key %#x\n", key)
        return false

    default:
        state := ev.State() & 0x0f
//        fmt.Printf("Got key %#x state=%#x *CTL=%#x SHIFT=%#x ALT=%#x\n",
//                    key, state, gdk.CONTROL_MASK, gdk.SHIFT_MASK, gdk.MOD1_MASK)
        if state & gdk.CONTROL_MASK != 0 {
            return false
        }
        if isKeyDecimal( key ) {
            return false
        }
        return true
    }
}

func addIntContent( innerBox *gtk.Box, content *contentDef,
                    access map[string]interface{} ) {
//    fmt.Printf( "got input type int, min %d max %d\n", content.inputMin, content.inputMax )
    intVal := content.initVal.(int)

    if content.changed == nil {
        constant, err := gtk.LabelNew( fmt.Sprintf( "%d", intVal ) )
        if nil != err {
            log.Fatalf("addIntContent: Could not create constant int %d: %v",
                        intVal, err)
        }
        if access != nil && content.name != "" {
            access[content.name] = constant
        }
        innerBox.PackEnd( constant, false, false, content.padding )

    } else  if valCtl, ok := content.valueCtl.(inputCtl); ok {
        input, err := gtk.SpinButtonNewWithRange( float64(valCtl.inputMin),
                                                  float64(valCtl.inputMax),
                                                  float64(valCtl.inputInc) )
        if nil != err {
            log.Fatal("addIntContent: Could not create input button:", err)
        }
//        input.SetNumeric(true)
        input.SetValue( float64(intVal) )
        input.Entry.Connect( "key-press-event", noKey )
        valueChanged := func( button *gtk.SpinButton ) {
            v := button.GetValue()
            content.changed( content.name, v )
        }
        input.Connect( "value-changed", valueChanged )
        if access != nil && content.name != "" {
            access[content.name] = input
        }
        innerBox.PackEnd( input, false, false, content.padding )

    } else if valCtl, ok := content.valueCtl.([]int); ok {
        input, err := gtk.ComboBoxTextNew( )
        if err != nil {
            log.Fatalf( "addIntContent: cannot create comboText: %v\n", err )
        }
        for i, v := range valCtl {
            input.AppendText( fmt.Sprintf("%d", v) )
            if intVal == v {
                input.SetActive( i )
            }
        }
        valueChanged := func( b *gtk.ComboBoxText ) {
            t := b.GetActiveText()
            if t != "" {
                v, err := strconv.Atoi( t )
                if err != nil {
                    log.Fatal("addIntContent: can't get number:", err )
                }
                content.changed( content.name, float64(v) )
            }
        }
        input.Connect( "changed", valueChanged )
        if access != nil && content.name != "" {
            access[content.name] = input
        }
        innerBox.PackEnd( input, false, false, content.padding )

    } else {    // no or unexpected control, use default
        input, err := gtk.EntryNew()
        if err != nil {
            log.Fatalf( "addIntContent: cannot create entry: %v\n", err )
        }
//        input.SetInputPurpose( gtk.INPUT_PURPOSE_DIGITS )
        input.Connect( "key-press-event", decimalKey )
        input.SetText( fmt.Sprintf("%d", intVal ) )

        valueChanged := func( e *gtk.Entry ) {
            t, err := e.GetText( )
            if err != nil {
                log.Fatal("addIntContent: can't get entry text:", err )
            }
            if t != "" {
                v, err := strconv.Atoi( t )
                if err != nil {
                    log.Fatal("addIntContent: can't get number:", err )
                }
                content.changed( content.name, float64(v) )
            }
        }
        input.Connect( "activate", valueChanged )
        if access != nil && content.name != "" {
            access[content.name] = input
        }
        innerBox.PackEnd( input, false, false, content.padding )
    }
}

func setIntContent( db *dataBox, name string, val int ) {
    intContent, ok := db.access[name]
    if ! ok {
        log.Fatalf("setIntContent: no such int value %s\n", name )
    }
    t := fmt.Sprintf( "%d", val )
    switch intVal := intContent.(type) {
    case *gtk.Label: // constant label
        intVal.SetText( t )
    case *gtk.SpinButton:
        intVal.SetValue( float64(val) )
    case *gtk.ComboBoxText:
        model, err := intVal.GetModel()
        if err != nil {
            log.Fatalf( "setIntContent: unable to get model: %v\n", err )
        }
        list := model.ToTreeModel()
        var index = 0
        if iter, nonEmpty := list.GetIterFirst( ); nonEmpty {
            for {
                v, err := list.GetValue( iter, 0 )
                if err != nil {
                    log.Fatalf( "setIntContent: unable to get list value: %v\n",
                                err )
                }
                var ls string
                ls, err = v.GetString()
                if err != nil {
                    log.Fatalf( "setIntContent: unable to get list string: %v\n",
                                err )
                }
                if ls == t {
                    intVal.SetActive( index )
                    break
                }
                index++
                if false == list.IterNext( iter ) {
                    break
                }
            }
        } else {
            intVal.AppendText( t )
            intVal.SetActive(0)
        }
    case *gtk.Entry:
        intVal.SetText( t )
    default:
        log.Fatalf( "setIntContent: unexpected input type %T\n", intVal )
    }
}

func wrapInFrame( w gtk.IWidget, title string, visible bool ) (*gtk.Frame) {
    frame, err := gtk.FrameNew( title )
    if err != nil {
        log.Fatal("wrapInFrame: Could not create frame", err)
    }
    frame.Add( w )
    if visible {
        frame.SetShadowType( gtk.SHADOW_IN )
    } else {
        frame.SetShadowType( gtk.SHADOW_NONE )
    }
    return frame
}

func copyContent( label *gtk.Label, event *gdk.Event ) bool {
    buttonEvent := gdk.EventButtonNewFromEvent( event )
    evButton := buttonEvent.Button()
    fmt.Printf("Event button=%d\n", evButton)

    if evButton != gdk.BUTTON_PRIMARY {
        fmt.Printf("Pressed mouse button %d\n", evButton )
        copy2Clipboard := func( ) {
            t, err := label.GetText()
            if err == nil {
                setClipboardAscii( t )
            }
        }
        addPopupMenuItem( "copyValue", actionCopyValue, copy2Clipboard )
        aNames := []string{ "copyValue" }
        popupContextMenu( aNames, event )
        delPopupMenuItem( "copyValue" )
        return true
    }
    return false
}

func makeConstantLabel( text string, name string, cCtl *constCtl,
                        access map[string]interface{} ) *gtk.Frame {
    constant, err := gtk.LabelNew( "" )
    if nil != err {
        log.Fatalf("makeConstantLabel: Could not create label %s: %v", text, err)
    }
    if cCtl != nil && cCtl.monosize {
        constant.SetMarkup( "<span face=\"monospace\">"+text+"</span>" )
    } else {
        constant.SetMarkup( text )
    }
    if name != "" {
        access[name] = constant
    }
    if cCtl != nil {
        constant.SetWidthChars( cCtl.size )
        switch cCtl.align {
        case LEFT_ALIGN:
            constant.SetXAlign( 0.0 )
        case RIGHT_ALIGN:
            constant.SetXAlign( 1.0 )
        }
        if cCtl.canCopy {
            eb, err := gtk.EventBoxNew( )
            if err != nil {
                log.Fatalf("makeConstantLabel: could not create event box: %v", err)
            }
            eb.SetAboveChild( true )
            eb.Add( constant )
            cc := func( eventbox *gtk.EventBox, event *gdk.Event ) bool {
                return copyContent( constant, event )
            }
            eb.Connect( "button_press_event", cc )
            eb.SetTooltipText( "Right click to copy Value" )
            return wrapInFrame( eb, "", cCtl.frame )
        }
        return wrapInFrame( constant, "", cCtl.frame )
    }
    return wrapInFrame( constant, "", false )
}

func addTextContent( innerBox *gtk.Box, content *contentDef,
                     access map[string]interface{} ) {
//    fmt.Printf( "got input type text, min %d max %d\n", content.inputMin, content.inputMax )
    textVal := content.initVal.(string)

    if content.changed == nil {
        if cCtl, ok := content.valueCtl.(constCtl); ok {
            frame := makeConstantLabel( textVal, content.name, &cCtl, access )
            innerBox.PackEnd( frame, false, false, content.padding )
        } else {
            frame := makeConstantLabel( textVal, content.name, nil, access )
            innerBox.PackEnd( frame, false, false, content.padding )
        }
        return
    }
    lenCtl, ok := content.valueCtl.(lengthCtl)
//fmt.Printf( "valCtl type=%T, ok=%t\n", valCtl, ok )
    if nil == content.valueCtl || ok {  // accept nil valueClt as no restrictions
        input, err := gtk.EntryNew( )
        if nil != err {
            log.Fatal("addTextContent: Could not create text input:", err)
        }
        if ok {
            input.SetMaxLength( lenCtl.inputMax )
        }
        input.SetText( textVal )
        textChanged := func( e *gtk.Entry ) {
            t, err := e.GetText( )
            if err != nil {
                log.Fatal("addTextContent: can't get entry text:", err )
            }
            if t != "" {
                content.changed( content.name, t )
            }
        }
        input.Connect( "activate", textChanged )
        if access != nil && content.name != "" {
            access[content.name] = input
        }
        innerBox.PackEnd( input, false, false, content.padding )
    } else if valCtl, ok := content.valueCtl.([]string); ok {
        input, err := gtk.ComboBoxTextNew( )
        if err != nil {
            log.Fatalf( "addTextContent: cannot create comboText: %v\n", err )
        }
        for i, v := range valCtl {
            input.AppendText( v )
            if textVal == v {
                input.SetActive( i )
            }
        }
        textChanged := func( b *gtk.ComboBoxText ) {
            t := b.GetActiveText()
            if t != "" {
                content.changed( content.name, t )
            }
        }
        input.Connect( "changed", textChanged )
        if access != nil && content.name != "" {
            access[content.name] = input
        }
        innerBox.PackEnd( input, false, false, content.padding )
    }
}

func setTextContent( db *dataBox, name string, text string ) {
    textContent, ok := db.access[name]
    if ! ok {
        log.Fatalf("setIntContent: no such int value %s\n", name )
    }
    switch textVal := textContent.(type) {
    case *gtk.Label: // constant label
        textVal.SetMarkup( text )
    case *gtk.Entry:
        textVal.SetText( text )
    case *gtk.ComboBoxText:
        model, err := textVal.GetModel()
        if err != nil {
            log.Fatalf( "setIntContent: unable to get model: %v\n", err )
        }
        list := model.ToTreeModel()
        var index = 0
        if iter, nonEmpty := list.GetIterFirst( ); nonEmpty {
            for {
                v, err := list.GetValue( iter, 0 )
                if err != nil {
                    log.Fatalf( "setIntContent: unable to get list value: %v\n",
                                err )
                }
                var ls string
                ls, err = v.GetString()
                if err != nil {
                    log.Fatalf( "setIntContent: unable to get list string: %v\n",
                                err )
                }
                if ls == text {
                    textVal.SetActive( index )
                    break
                }
                index++
                if false == list.IterNext( iter ) {
                    break
                }
            }
        } else {
            textVal.AppendText( text )
            textVal.SetActive(0)
        }
    default:
        log.Fatalf( "setTextContent: unexpected input type %T\n", textVal )
    }
}

func addContentToBox( box *gtk.Box, content *contentDef,
                      access map[string]interface{} ) {
    label := makeConstantLabel( content.label, "", content.labelCtl, nil )
    innerBox := wrapChildInHorizontalBox( label, content.padding )

    switch content.initVal.(type) {
    case nil:   // nothing else to add
    case bool:
        addBoolContent( innerBox, content, access )
    case int:
        addIntContent( innerBox, content, access )
    case string:
        addTextContent( innerBox, content, access )
    default:
        fmt.Printf( "addContentToBox: got something else %T\n", content.initVal )
        return
    }
    box.PackStart( innerBox, false, false, 0 )    // in parent box
}

func addHeaderToBox( box *gtk.Box, header *headerDef ) {
    label, err := gtk.LabelNew( "" )
    if nil != err {
        log.Fatal("addHeaderToBox: Could not create label %s:", header, err)
    }
    label.SetMarkup( header.label )
    innerBox := wrapChildInHorizontalBox( label, header.left )
    box.PackStart( innerBox, false, false, header.top )     // in parent box
}

func addBoxToBox( box *gtk.Box, item *boxDef, access map[string]interface{} ) {
    child := makeBox( item, access )
    innerBox := wrapChildInHorizontalBox( child, item.padding )
    box.PackStart( innerBox, false, false, 0 )      // in parent box
}

func wrapChildInHorizontalBox( child gtk.IWidget, padding uint ) *gtk.Box {
    innerBox, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 10 )
    if nil != err {
        log.Fatal("wrapChidInHorizontalBox: Could not create inner box:", err)
    }
    innerBox.PackStart( child, false, false, padding )
    return innerBox
}

func makeBox( def * boxDef, access map[string]interface{} ) *gtk.Frame {

    box, err := gtk.BoxNew( gtk.Orientation(def.direction), def.spacing )
    if nil != err {
        log.Fatal("makeBox: Could not create box", err)
    }
    for _, item := range def.content {
        switch item := item.(type) {
        case *headerDef:
            addHeaderToBox( box, item )
        case *contentDef:
            addContentToBox( box, item, access )
        case *boxDef:
            addBoxToBox( box, item, access )
        default:
            fmt.Printf( "makeBox: got something else %T\n", item )
        }
    }
    return wrapInFrame( box, def.title, def.frame )
}

type dataBox struct {
    *gtk.Frame
    access map[string]interface{}
}

func makeDataBox( def * boxDef ) (db *dataBox) {
    db = new(dataBox)
    db.access = make( map[string]interface{} )
    db.Frame = makeBox( def, db.access )
    return
}
