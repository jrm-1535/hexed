package main

import (
    "log"
    "fmt"
    "strconv"
	"github.com/gotk3/gotk3/gtk"
//	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gdk"
)

func errorDisplay( format string, args ...interface{} ) {

    dialog := gtk.MessageDialogNew ( window, gtk.DIALOG_DESTROY_WITH_PARENT,
                                     gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE,
                                     format, args )
    dialog.Run( )
    dialog.Destroy( )
}

func openFileName( ) (name string) {
    dialog, err := gtk.FileChooserDialogNewWith2Buttons(
                    "Open File", window, gtk.FILE_CHOOSER_ACTION_OPEN,
                    "_Cancel", gtk.RESPONSE_CANCEL,
                    "_Open", gtk.RESPONSE_ACCEPT )

    if err == nil {
        res := dialog.Run( )
        if res == gtk.RESPONSE_ACCEPT {
            name = dialog.GetFilename( )
        }
    } else {
        fmt.Printf( "openFileName error: %v\n", err )
    }
    dialog.Destroy()
    return
}

func saveFileName() (name string) {
    dialog, err := gtk.FileChooserDialogNewWith2Buttons(
                    "Save File", window, gtk.FILE_CHOOSER_ACTION_SAVE,
                    "_Cancel", gtk.RESPONSE_CANCEL,
                    "_Save", gtk.RESPONSE_ACCEPT )

    if err == nil {
        res := dialog.Run( )
        if res == gtk.RESPONSE_ACCEPT {
            name = dialog.GetFilename( )
        }
    } else {
        fmt.Printf( "saveFileName error: %v\n", err )
    }
    dialog.Destroy()
    return
}

const (
    CANCEL = iota
    DO
    SAVE_THEN_DO
)

func closeFileDialog( ) (op int) {
    cd, err := gtk.DialogNewWithButtons( localizeText(dialogCloseTitle), window,
                    gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT,
                    []interface{} { localizeText(buttonSave), gtk.RESPONSE_ACCEPT },
                    []interface{} { localizeText(buttonCloseWithoutSave), gtk.RESPONSE_REJECT },
                    []interface{} { localizeText(buttonCancel), gtk.RESPONSE_CANCEL } )
    if err != nil {
        log.Fatal("closeFileDialog: could not create gtk dialog:", err)
    }
    cd.SetDefaultResponse( gtk.RESPONSE_ACCEPT )
    carea, err := cd.GetContentArea()
    if err != nil {
        log.Fatal("closeFileDialog: could not get content area:", err)
    }
    label, err := gtk.LabelNew( localizeText(warningCloseFile) )
    if err != nil {
        log.Fatal("closeFileDialog: could not create content label:", err)
    }
    carea.Container.Add( label )
    carea.ShowAll()
    response := cd.Run()
    switch response {
    case gtk.RESPONSE_ACCEPT:
        op = SAVE_THEN_DO
    case gtk.RESPONSE_REJECT:
        op = DO
    case gtk.RESPONSE_NONE, gtk.RESPONSE_CANCEL:
        op = CANCEL
    }
    cd.Destroy()
    return
}

func hexFilter( entry *gtk.Entry, event *gdk.Event ) bool {
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
        if hex, _ := getNibbleFromKey( key ); hex {
            return false
        }
        return true
    }
}

func gotoDialog( ) (op int, pos int64) {
    gd, err := gtk.DialogNewWithButtons( localizeText(dialogGotoTitle), window,
                    gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT,
                    []interface{} { localizeText(buttonGo), gtk.RESPONSE_ACCEPT },
                    []interface{} { localizeText(buttonCancel), gtk.RESPONSE_CANCEL } )
    if err != nil {
        log.Fatal("gotoDialog: could not create gtk dialog:", err)
    }
    gd.SetDefaultResponse( gtk.RESPONSE_ACCEPT )
    carea, err := gd.GetContentArea()
    if err != nil {
        log.Fatal("gotoDialog: could not get content area:", err)
    }
    label, err := gtk.LabelNew( localizeText( gotoPrompt ) )
    if err != nil {
        log.Fatal("gotoDialog: could not create content label:", err)
    }
    carea.Container.Add( label )

    entry, err := gtk.EntryNew( )
    if err != nil {
        log.Fatal("gotoDialog: could not create content entry:", err)
    }

    entry.SetActivatesDefault( true )
    entry.Connect( "key-press-event", hexFilter )
    carea.Container.Add( entry )

    carea.ShowAll()
    response := gd.Run()
    switch response {
    case gtk.RESPONSE_ACCEPT:
        text, err := entry.GetText()
        if err != nil {
            panic("Cannot get entry text\n")
        }
        if _, err = fmt.Sscanf( text, "%x", &pos ); err != nil {
            panic( err )
        }
        pos = pos << 1
        op = DO
    case gtk.RESPONSE_NONE, gtk.RESPONSE_CANCEL:
        op = CANCEL
    }
    gd.Destroy()
    return
}

// generic dialog management

type boxDef struct {
    spacing     int
    padding     uint

    content     []interface{}   // boxDef, contentDef or string (header)
}

// content values can be bool, int or string
// initVal gives the initial input value (current preference value)
// valueCtl restricts what can be the value
// a nil valueCtl means no restictions
// If initVal is a string, valueCtl can be:
//   - [...]string for a list of possible inputs
//   - inputCtl for the min and max length of the input
// If initVal is an int, valueCtl can be:
//   - [...]int for a list of possible inputs
//   - inputCtl for the min and max input value
// Not used for boolean
type contentDef struct {
    label       string
    name        string
    initVal     interface{}
    valueCtl    interface{}
    changed     func( name string, val interface{} )
}

type inputCtl struct {
    inputMin    int
    inputMax    int
    inputInc    int
}

type lengthCtl struct {
    inputMax    uint
}

func noKey( ) bool {
    return true
}

func addBoolContent( label *gtk.Label, content *contentDef ) *gtk.Box {
//    fmt.Printf( "got input type bool\n" )
    boolVal := content.initVal.(bool)
    innerBox, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 0 )
    if nil != err {
        log.Fatal("addBoolContent: Could not create inner box:", err)
    }
    input, err := gtk.CheckButtonNew( )
    if nil != err {
        log.Fatal("addBoolContent: Could not create bool bool input button:", err)
    }
    input.ToggleButton.SetActive( boolVal )
    toggled := func( button *gtk.CheckButton ) {
        v := button.ToggleButton.GetActive()
        //fmt.Printf("Notifying content changed: %s=%v\n", content.name, v )
        content.changed( content.name, v )
    }
    input.Connect( "toggled", toggled )

    innerBox.PackStart( gtk.IWidget(label), true, true, 20 ) // in horizontal box
    innerBox.PackStart( gtk.IWidget(input), false, false, 60 ) // in horizontal box

    return innerBox
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

func addIntContent( label *gtk.Label, content *contentDef ) *gtk.Box {
//    fmt.Printf( "got input type int, min %d max %d\n", content.inputMin, content.inputMax )
    intVal := content.initVal.(int)
    innerBox, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 10 )
    if nil != err {
        log.Fatal("addIntContent: Could not create inner box:", err)
    }
    innerBox.PackStart( gtk.IWidget(label), false, false, 20 ) // in horizontal box

    if valCtl, ok := content.valueCtl.(inputCtl); ok {
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
        innerBox.PackEnd( gtk.IWidget(input), false, false, 5 )    // in horizontal box

    } else if valCtl, ok := content.valueCtl.([]int); ok {
        input, err := gtk.ComboBoxTextNew( )
        if err != nil {
            log.Fatalf( "addTextContent: cannot create comboText: %v\n", err )
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
        innerBox.PackEnd( gtk.IWidget(input), false, false, 5 )    // in horizontal box

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
        innerBox.PackEnd( gtk.IWidget(input), false, false, 5 )    // in horizontal box
    }
    return innerBox
}

func addTextContent( label *gtk.Label, content *contentDef ) *gtk.Box {
//    fmt.Printf( "got input type text, min %d max %d\n", content.inputMin, content.inputMax )
    textVal := content.initVal.(string)
    innerBox, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 10 )
    if nil != err {
        log.Fatal("addTextContent: Could not create inner box:", err)
    }
    innerBox.PackStart( gtk.IWidget(label), false, false, 20 ) // in horizontal box

    lenCtl, ok := content.valueCtl.(lengthCtl)
//fmt.Printf( "valCtl type=%T, ok=%t\n", valCtl, ok )
    if nil == content.valueCtl || ok {  // accept nil valueClt
        input, err := gtk.EntryNew( )
        if nil != err {
            log.Fatal("addTextContent: Could not create text input:", err)
        }
        if ok {
            input.SetMaxLength( int(lenCtl.inputMax) )
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
        innerBox.PackEnd( gtk.IWidget(input), false, false, 5 )    // in horizontal box
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
        innerBox.PackEnd( input, false, false, 5 )    // in horizontal box
    }
    return innerBox
}

func addContentToBox( box *gtk.Box, content *contentDef ) {
    label, err := gtk.LabelNew( "" )
    if nil != err {
        log.Fatal("Could not create label %s:", content.label, err)
    }
    label.SetMarkup( content.label )
    label.SetXAlign( 0.03 )

    switch content.initVal.(type) {
    case nil:
        box.PackStart( gtk.IWidget(label), false, false, 0 )      // in vertical box

    case bool:
        innerBox := addBoolContent( label, content )
        box.PackStart( gtk.IWidget(innerBox), false, false, 0 )    // in vertical box (parent)

    case int:
        innerBox:= addIntContent( label, content )
        box.PackStart( gtk.IWidget(innerBox), false, false, 0 )    // in vertical box (parent)

    case string:
        innerBox := addTextContent( label, content )
        box.PackStart( gtk.IWidget(innerBox), false, false, 0 )    // in vertical box (parent)
    }
}

func addHeaderToBox( box *gtk.Box, header string ) {
    label, err := gtk.LabelNew( "" )
    if nil != err {
        log.Fatal("Could not create label %s:", header, err)
    }
    label.SetMarkup( header )
    label.SetXAlign( 0.03 )
    box.PackStart( gtk.IWidget(label), false, false, 0 )      // in vertical box
}

func makeBox( def * boxDef ) *gtk.Box {
//    fmt.Printf( "BoxDefT spacing %d padding %d, nb args %d\n",
//                def.spacing, def.padding, len(def.content) )
    box, err := gtk.BoxNew( gtk.ORIENTATION_VERTICAL, def.spacing )
    if nil != err {
        log.Fatal("Could not create box", err)
    }

    for _, item := range def.content {
        switch item := item.(type) {
        case string:
            addHeaderToBox( box, item )
        case *contentDef:
//            fmt.Printf( "got contentDef with label %s\n", item.label )
            addContentToBox( box, item )
        case *boxDef:
//            fmt.Printf( "got boxDef\n" )
            child := makeBox( item )
            box.PackStart( gtk.IWidget(child), false, false, def.padding )
        default:
            fmt.Printf( "makeBox: got something else %T\n", item )
        }
    }
    return box
}

func changed( name string, val interface{} ) {
    pref := preferences{}
    pref[name] = val
    update( pref )
}

func getPreferenceDialogEditorDef( ) *boxDef {
    fontHeader := localizeText(dialogPreferencesFont)
    var fontNames = [...]string{ "Courier 10 Pitch", "Monospace" }
    fontName := contentDef { localizeText(dialogPreferencesFontName), FONT_NAME,
                             getStringPreference(FONT_NAME), fontNames[0:], changed }
    fontSize := contentDef { localizeText(dialogPreferencesFontSize), FONT_SIZE,
                                   getIntPreference(FONT_SIZE), inputCtl{ 9, 25, 2 }, changed }
    fontData := boxDef { 0, 2, []interface{} { &fontName, &fontSize } }
    fontBox := boxDef { 5, 0, []interface{} { fontHeader, &fontData } }

    displayHeader := localizeText(dialogPreferencesDisplay)
    minBytesLine := contentDef { localizeText(dialogPreferencesDisplayMinBytesLine),
                                MIN_BYTES_LINE,
                                getIntPreference(MIN_BYTES_LINE), inputCtl{ 8, 32, 4 }, changed }
    maxBytesLine := contentDef { localizeText(dialogPreferencesDisplayMaxBytesLine),
                                MAX_BYTES_LINE,
                                getIntPreference(MAX_BYTES_LINE), inputCtl{ 32, 64, 4 }, changed }
    BytesLineInc := contentDef { localizeText(dialogPreferencesDisplayLineIncrement),
                                LINE_BYTE_INC,
                                getIntPreference(LINE_BYTE_INC), inputCtl{ 4, 16, 2 }, changed }
    BytesSeparator := contentDef { localizeText(dialogPreferencesDisplayBytesSeparator),
                                HOR_SEP_SPAN,
                                getIntPreference(HOR_SEP_SPAN), inputCtl{ 0, 16, 4 }, changed }
    LInesSeparator := contentDef { localizeText(dialogPreferencesDisplayLinesSeparator),
                                VER_SEP_SPAN,
                                getIntPreference(VER_SEP_SPAN), inputCtl{ 0, 32, 8 }, changed }
    displayData := boxDef { 0, 2, []interface{} { &minBytesLine, &maxBytesLine, &BytesLineInc,
                                                  &BytesSeparator, &LInesSeparator } }
    displayBox := boxDef { 5, 0, []interface{} { displayHeader, &displayData } }

    editorHeader := localizeText(dialogPreferencesEditor)
    startReadOnly := contentDef { localizeText(dialogPreferencesEditorReadOnly),
                                START_READ_ONLY,
                                getBoolPreference(START_READ_ONLY), nil, changed }
    startReplace := contentDef { localizeText(dialogPreferencesEditorReplaceNode),
                                START_REPLACE_MODE,
                                getBoolPreference(START_REPLACE_MODE), nil, changed }
    editorData := boxDef { 0, 2, []interface{} { &startReadOnly, &startReplace } }
    editorBox := boxDef { 5, 0, []interface{} { editorHeader, &editorData } }

    searchHeader := localizeText(dialogPreferencesSearch)
    wrapAround := contentDef { localizeText(dialogPreferencesSearchWrapAround),
                                WRAP_MATCHES,
                                getBoolPreference(WRAP_MATCHES), nil, changed }
    searchData := boxDef { 0, 2, []interface{} { &wrapAround } }
    searchBox := boxDef { 5, 0, []interface{} { searchHeader, &searchData } }

    wholeBox := boxDef { 0, 5, []interface{} { &displayBox, &fontBox,
                                               &editorBox, &searchBox } }
    return &wholeBox
}

func getPreferenceDialogSaveDef( ) *boxDef {

    saveHeader := localizeText(dialogPreferencesSave)
    backup := contentDef { localizeText(dialogPreferencesSaveBackup),
                           CREATE_BACKUP_FILES,
                           getBoolPreference(CREATE_BACKUP_FILES), nil, changed }
    saveData := boxDef { 0, 2, []interface{} { &backup } }
    saveBox := boxDef { 5, 0, []interface{} { saveHeader, &saveData } }

    wholeBox := boxDef { 0, 5, []interface{} { &saveBox } }
    return &wholeBox
}

func getPreferenceDialogTheme( ) *boxDef {

    themeNames, err := getThemeNames( )
    if err != nil {
        log.Fatalf( "Unable to get any theme name: %v\n", err )
    }

    themeHeader := localizeText(dialogPreferencesTheme)
    themeName := contentDef { localizeText(dialogPreferencesThemeName), COLOR_THEME_NAME,
                             getStringPreference(COLOR_THEME_NAME), themeNames, changed }
    themeData := boxDef { 0, 2, []interface{} { &themeName } }
    themeBox := boxDef { 5, 0, []interface{} { themeHeader, &themeData } }

    wholeBox := boxDef { 0, 5, []interface{} { &themeBox } }
    return &wholeBox
}

var preferencesDialog *gtk.Window

type preferenceNotebook struct {
    *gtk.Notebook
}

func (pn *preferenceNotebook) appendPreferencePage( bdef *boxDef,
                                                    tabId int ) error {

    box := makeBox( bdef )
    tab, err := gtk.LabelNew( localizeText( tabId ) )
    if err != nil {
        return err
    }
    if pageIndex := pn.AppendPage( box, tab ); -1 == pageIndex {
        log.Fatalf( "appendPage: Unable to append page\n" )
    }
    return nil
}

func showPreferencesDialog( ) (err error) {
    preferencesDialog, err = gtk.WindowNew( gtk.WINDOW_TOPLEVEL )
    if err != nil {
        return err
    }

    preferencesDialog.SetTypeHint( gdk.WINDOW_TYPE_HINT_DIALOG )

    prefNB := new(preferenceNotebook)
    prefNB.Notebook, err = gtk.NotebookNew( )
//    ntbk, err := gtk.NotebookNew( )
    if err != nil {
        return err
    }

    prefNB.Notebook.SetTabPos( gtk.POS_LEFT )
//    ntbk.ConnectAfter( "switch-page", switchPage )

    preferencesDialog.Add( prefNB.Notebook )

    editorDef := getPreferenceDialogEditorDef( )
    prefNB.appendPreferencePage( editorDef, dialogPreferencesEditorTab )
    saveDef := getPreferenceDialogSaveDef( )
    prefNB.appendPreferencePage( saveDef, dialogPreferencesSaveTab )
    themeDef := getPreferenceDialogTheme( )
    prefNB.appendPreferencePage( themeDef, dialogPreferencesThemeTab )

    preferencesDialog.SetTransientFor( window )
    preferencesDialog.SetTypeHint( gdk.WINDOW_TYPE_HINT_DIALOG )

    preferencesDialog.SetTitle(localizeText(windowTitlePreferences))
    preferencesDialog.SetDefaultSize(300, 300)

    preferencesDialog.ShowAll()

    return nil
}
