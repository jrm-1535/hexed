package main

import (
    "fmt"
    "log"
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

// The top level widget is always a box with a frame (which may not be visible)
// that contains vertically aligned children: other boxes, simple headers or
// contents. A box can be vertical or horizontal, where children are packed
// starting from the top or from the left, respectively. The optional frame
// title can be accessible through the box name and the methods getLabel
// and setLabel

// Header is just a label (string) within an invisible horizontal frame with
// padding on top and on left sides. Header name is a string that can be used
// to get or set the header label programmatically while the dialog is visible,
// through getLabel and setlabel

// Content is an horizontal box for a label (string) and an associated value.
// content-value types are limited to boolean, numbers or strings. Content
// value can be constant (not modifiable from UI) or variable. Content name is
// a string that can be used to modify both its label and value. The label is
// accessible through methods getLabel and setLabel. The value is retrieved
// either through a callback after the value has been modified by UI or
// programmatically through get<value-type>Value( contentName ). value can be
// modified using set<value-type>Value( contentName, newValue).
// Content label is packed starting from the left and value is packed starting
// from the right, in order to fill up the whole available space.
type boxDef struct {
    spacing     int             // space between box chidren
    padding     uint            // padding in parent box, if any

    frame       bool            // visible frame around box
    title       string          // optional frame or box title
    name        string          // used to get or set title (maybe empty)

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
    label       string      // visible text in UI
    name        string      // internal name to get or set the label
    top, left   uint
}

// box content values can be header, content(bool, int or string) or box
// Horizontal box is made of a series of header and content 
// [ header label] [ label 1  value 1 entry ] ... [ label n value n entry ]
// Vertical box is made of a series of horizontal boxes with header or content
// [ header label                     ] label alone
// [  content label 1   value 1 entry ] label + value per content
// [[ content label 1   value 1 entry ]...[ content label n   value n entry]]

type itemReference struct {
    monosize    bool        // true if label is in monospace
    label       interface{} // reference to the item label
    value       interface{} // reference to the item value
}

// nokey prevents entering data otherwise than by spinner (with min, max  & inc)
func noKey( ) bool {
    return true
}

func (db *dataBox)newLabel( name string, mono bool, label interface{} ) {
    if name != "" {
        ref, ok := db.access[name]
        if ! ok {
            ref = itemReference{ mono, label, nil }
        } else {
            ref.label = label
        }
        db.access[name] = ref
    }
}

func (db *dataBox)newValue( name string, mono bool, value interface{} ) {
// FIXME: ignore mono for the time being: add later
    if name != "" {
        ref, ok := db.access[name]
        if ! ok {
            ref = itemReference{ false, nil, value }
        } else {
            ref.value = value
        }
        db.access[name] = ref
    }
}

func (db *dataBox)addBoolContent( innerBox *gtk.Box, content *contentDef ) {

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
            content.changed( content.name, v )
        }
        input.Connect( "toggled", toggled )
    }
    db.newValue( content.name, false, input )
    innerBox.PackEnd( input, false, false, content.padding )
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
        return false

    default:
        state := ev.State() & 0x0f
//        printDebug("decimalKey: key %#x state=%#x *CTL=%#x SHIFT=%#x ALT=%#x\n",
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

func (db *dataBox)addIntContent( innerBox *gtk.Box, content *contentDef ) {

    intVal := content.initVal.(int)

    if content.changed == nil {
        constant, err := gtk.LabelNew( fmt.Sprintf( "%d", intVal ) )
        if nil != err {
            log.Fatalf("addIntContent: Could not create constant int %d: %v",
                        intVal, err)
        }
        db.newValue( content.name, false, constant )
        innerBox.PackEnd( constant, false, false, content.padding )

    } else  if valCtl, ok := content.valueCtl.(inputCtl); ok {
        input, err := gtk.SpinButtonNewWithRange( float64(valCtl.inputMin),
                                                  float64(valCtl.inputMax),
                                                  float64(valCtl.inputInc) )
        if nil != err {
            log.Fatal("addIntContent: Could not create input button:", err)
        }
        input.SetValue( float64(intVal) )
        input.Entry.Connect( "key-press-event", noKey )
        valueChanged := func( button *gtk.SpinButton ) {
            v := button.GetValue()
            content.changed( content.name, v )
        }
        input.Connect( "value-changed", valueChanged )
        db.newValue( content.name, false, input )
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
        db.newValue( content.name, false, input )
        innerBox.PackEnd( input, false, false, content.padding )

    } else {    // no or unexpected control, use default
        input, err := gtk.EntryNew()
        if err != nil {
            log.Fatalf( "addIntContent: cannot create entry: %v\n", err )
        }
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
        db.newValue( content.name, false, input )
        innerBox.PackEnd( input, false, false, content.padding )
    }
}

func (db *dataBox)wrapInFrame( w gtk.IWidget, title, name string,
                               visible bool ) (*gtk.Frame) {
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
    db.newLabel( name, false, frame )
    return frame
}

func copyContent( label *gtk.Label, event *gdk.Event ) bool {
    buttonEvent := gdk.EventButtonNewFromEvent( event )
    evButton := buttonEvent.Button()

    if evButton != gdk.BUTTON_PRIMARY {
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

func makeMonosizeMarkUp( t string ) string {

const (
    MONOSIZE_PREFIX = "<span face=\"monospace\">"
    MONOSIZE_SUFFIX = "</span>"
)

   return MONOSIZE_PREFIX+t+MONOSIZE_SUFFIX
}

func (db *dataBox)makeConstantText( text string, name string, isValue bool,
                                    cCtl *constCtl ) *gtk.Frame {
    constant, err := gtk.LabelNew( "" )
    if nil != err {
        log.Fatalf("makeConstantText: Could not create label %s: %v", text, err)
    }
    var monosize bool
    if cCtl != nil && cCtl.monosize {
        monosize = true
        text = makeMonosizeMarkUp( text )
    }

    constant.SetMarkup(  text )
    if isValue {
        db.newValue( name, monosize, constant )
    } else {
        db.newLabel( name, monosize, constant )
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
                log.Fatalf("makeConstantText: could not create event box: %v", err)
            }
            eb.SetAboveChild( true )
            eb.Add( constant )
            cc := func( eventbox *gtk.EventBox, event *gdk.Event ) bool {
                return copyContent( constant, event )
            }
            eb.Connect( "button_press_event", cc )
            eb.SetTooltipText( localizeText( tooltipCopyValue ) )
            return db.wrapInFrame( eb, "", "", cCtl.frame )
        }
        return db.wrapInFrame( constant, "", "", cCtl.frame )
    }
    return db.wrapInFrame( constant, "", "", false )
}

func (db *dataBox)addTextContent( innerBox *gtk.Box, content *contentDef ) {

    textVal := content.initVal.(string)

    if content.changed == nil {
        if cCtl, ok := content.valueCtl.(constCtl); ok {
            frame := db.makeConstantText( textVal, content.name, true, &cCtl )
            innerBox.PackEnd( frame, false, false, content.padding )
        } else {
            frame := db.makeConstantText( textVal, content.name, true, nil )
            innerBox.PackEnd( frame, false, false, content.padding )
        }
        return
    }
    lenCtl, ok := content.valueCtl.(lengthCtl)
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
        db.newValue( content.name, false, input )
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
        db.newValue( content.name, false, input )
        innerBox.PackEnd( input, false, false, content.padding )
    }
}

func (db *dataBox)addContentToBox( box *gtk.Box, content *contentDef ) {
    label := db.makeConstantText( content.label, content.name,
                                  false, content.labelCtl )
    innerBox := wrapChildInHorizontalBox( label, content.padding )

    switch content.initVal.(type) {
    case nil:   // nothing else to add
    case bool:
        db.addBoolContent( innerBox, content )
    case int:
        db.addIntContent( innerBox, content )
    case string:
        db.addTextContent( innerBox, content )
    default:
        printDebug( "addContentToBox: unsupported type %T\n", content.initVal )
        return
    }
    box.PackStart( innerBox, false, false, 0 )    // in parent box
}

func (db *dataBox)addHeaderToBox( box *gtk.Box, header *headerDef ) {
    label, err := gtk.LabelNew( "" )
    if nil != err {
        log.Fatal("addHeaderToBox: Could not create label %s:", header, err)
    }
    label.SetMarkup( header.label )
    db.newLabel( header.name, false, label )
    innerBox := wrapChildInHorizontalBox( label, header.left )
    box.PackStart( innerBox, false, false, header.top )     // in parent box
}

func (db *dataBox)addBoxToBox( box *gtk.Box, item *boxDef ) {
    child := db.makeBox( item )
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

func (db *dataBox)makeBox( def * boxDef ) *gtk.Frame {

    box, err := gtk.BoxNew( gtk.Orientation(def.direction), def.spacing )
    if nil != err {
        log.Fatal("makeBox: Could not create box", err)
    }
    for _, item := range def.content {
        switch item := item.(type) {
        case *headerDef:
            db.addHeaderToBox( box, item )
        case *contentDef:
            db.addContentToBox( box, item )
        case *boxDef:
            db.addBoxToBox( box, item )
        default:
            printDebug( "makeBox: unsupported type %T\n", item )
        }
    }
    return db.wrapInFrame( box, def.title, def.name, def.frame )
}

type dataBox struct {
    *gtk.Frame
    access map[string]itemReference
}

func makeDataBox( def * boxDef ) (db *dataBox) {
    db = new(dataBox)
    db.access = make( map[string]itemReference )
    db.Frame = db.makeBox( def )
    return
}

func (db *dataBox) getLabel( name string ) (label string) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "getLabel: item %s does not exist\n", name )
    }
    switch lb := ref.label.(type) {
    case *gtk.Label :
        label = lb.GetLabel( )
    case *gtk.Frame :
        label = lb.GetLabel( )
    default:
        log.Panicf( "getLabel: item %s has no label\n", name )
    }
    return
}

func (db *dataBox) setLabel( name string, label string ) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "setLabel: item %s does not exist\n", name )
    }

    if ref.monosize {
        label = makeMonosizeMarkUp( label )
    }
    switch lb := ref.label.(type) {
    case *gtk.Label :
        lb.SetLabel( label )
    case *gtk.Frame :
        lb.SetLabel( label )
    default:
        log.Panicf( "setLabel: unexpected label type %T for item %s\n", lb, name )
    }
}

func (db *dataBox) getBoolValue( name string ) bool {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "getBoolValue: item %s does not exist\n", name )
    }
    input, ok := ref.value.(*gtk.CheckButton)
    if ! ok {
        log.Panicf("getBoolValue: item %s is not a bool %s\n", name )
    }
    return input.ToggleButton.GetActive( )
}

func (db *dataBox) setBoolValue( name string, val bool ) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "setBoolValue: item %s does not exist\n", name )
    }
    input, ok := ref.value.(*gtk.CheckButton)
    if ! ok {
        log.Panicf("setBoolValue: item %s is not a bool\n", name )
    }
    input.ToggleButton.SetActive( val )
}

func convertTextToInt64( t string ) int64 {
    v, e := strconv.ParseInt( t, 10, 64)
    if e != nil {
        log.Fatalf( "getIntValue: cannot convert constant text %s to int\n", t )
    }
    return v
}

func (db *dataBox)getIntValue( name string ) (v int64) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "getIntValue: item %s does not exist\n", name )
    }
    intValue := ref.value
    switch intVal := intValue.(type) {
    case *gtk.Label: // constant value in label
        t, e := intVal.GetText( )
        if e != nil {
            log.Fatalf( "getIntValue: cannot get label text for item %s\n", name )
        }
        v = convertTextToInt64( t )
    case *gtk.SpinButton:
        v = int64(intVal.GetValue( ) )
    case *gtk.ComboBoxText:
        v = convertTextToInt64( intVal.GetActiveText( ) )
    case *gtk.Entry:
        t, e := intVal.GetText( )
        if e != nil {
            log.Fatalf( "getIntValue: cannot get entry text for item %s\n", name )
        }
        v = convertTextToInt64( t )
    default:
        log.Panicf( "getIntValue: unexpected item type %T\n", intVal )
    }
    return
}

func (db *dataBox)setIntValue( name string, val int ) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "setIntValue: item %s does not exist\n", name )
    }
    intValue := ref.value
    t := fmt.Sprintf( "%d", val )
    switch intVal := intValue.(type) {
    case *gtk.Label: // constant label
        intVal.SetText( t )
    case *gtk.SpinButton:
        intVal.SetValue( float64(val) )
    case *gtk.ComboBoxText:
        model, err := intVal.GetModel()
        if err != nil {
            log.Fatalf( "setIntValue: unable to get model: %v\n", err )
        }
        list := model.ToTreeModel()
        var index = 0
        if iter, nonEmpty := list.GetIterFirst( ); nonEmpty {
            for {
                v, err := list.GetValue( iter, 0 )
                if err != nil {
                    log.Fatalf( "setIntValue: unable to get list value: %v\n",
                                err )
                }
                var ls string
                ls, err = v.GetString()
                if err != nil {
                    log.Fatalf( "setIntValue: unable to get list string: %v\n",
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
        log.Panicf( "setIntValue: unexpected item type %T\n", intVal )
    }
}

func (db *dataBox)getTextValue( name string ) (t string) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "getTextValue: item %s does not exist\n", name )
    }
    var err error
    textValue := ref.value
    switch textVal := textValue.(type) {
    case *gtk.Label:
        t, err = textVal.GetText( )
        if err != nil {
            log.Fatalf( "getTextValue: cannot get label text for item %s\n", name )
        }
    case *gtk.ComboBoxText:
        t = textVal.GetActiveText( )
    default:
        log.Panicf( "getTextValue: unexpected item type %T\n", textVal )
    }
    return
}

func (db *dataBox) setTextValue( name string, text string ) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "setTextValue: item %s does not exist\n", name )
    }
    textValue := ref.value

    switch textVal := textValue.(type) {
    case *gtk.Label: // constant label
        if ref.monosize {
            text = makeMonosizeMarkUp( text )
        }
        textVal.SetMarkup( text )
    case *gtk.Entry:
        textVal.SetText( text )
    case *gtk.ComboBoxText:
        model, err := textVal.GetModel()
        if err != nil {
            log.Fatalf( "setTextValue: unable to get model: %v\n", err )
        }
        list := model.ToTreeModel()
        var index = 0
        if iter, nonEmpty := list.GetIterFirst( ); nonEmpty {
            for {
                v, err := list.GetValue( iter, 0 )
                if err != nil {
                    log.Fatalf( "setTextValue: unable to get list value: %v\n",
                                err )
                }
                var ls string
                ls, err = v.GetString()
                if err != nil {
                    log.Fatalf( "setTextValue: unable to get list string: %v\n",
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
        log.Panicf( "setTextValue: unexpected item type %T\n", textVal )
    }
}

func (db *dataBox) setTextChoices( name string, choices []string, active int,
                                   changed func( name string, val interface{} ) ) {
    ref, ok := db.access[name]
    if ! ok {
        log.Panicf( "setTextChoices: item %s does not exist\n", name )
    }
    textChoices := ref.value
    switch textValues := textChoices.(type) {
    case *gtk.ComboBoxText:
        textValues.RemoveAll()
        for _, v := range choices {
            textValues.AppendText( v )
        }
        if active >= 0 && active < len(choices) {
            textValues.SetActive( active )
        }
        textChanged := func( b *gtk.ComboBoxText ) {
            t := b.GetActiveText()
            if t != "" {
                changed( name, t )
            }
        }
        textValues.Connect( "changed", textChanged )

    default:
        log.Panicf( "setTextChoices: unexpected itemm type %T\n", textValues )
    }
}
