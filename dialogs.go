package main

import (
    "log"
    "fmt"
    "bytes"
    "math/big"
    "encoding/binary"

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

// --- open, close and save file dialogs

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
// ---- goto dialog

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

// ---- preferences dialog

func changed( name string, val interface{} ) {
    pref := preferences{}
    pref[name] = val
    update( pref )
}

func getPreferenceDialogEditorDef( ) *boxDef {
    separator := headerDef { " ", 0, 0 }
    fontHeader := headerDef { localizeText(dialogPreferencesFont), 5, 10 }
    var fontNames = [...]string{ "Courier 10 Pitch", "Monospace" }
    fontName := contentDef { localizeText(dialogPreferencesFontName),
                             nil, FONT_NAME,
                             getStringPreference(FONT_NAME), fontNames[0:],
                             changed, 20 }
    fontSize := contentDef { localizeText(dialogPreferencesFontSize),
                             nil, FONT_SIZE,
                             getIntPreference(FONT_SIZE), inputCtl{ 9, 25, 2 },
                             changed, 20 }

    displayHeader := headerDef { localizeText(dialogPreferencesDisplay), 5, 10 }
    minBytesLine := contentDef { localizeText(dialogPreferencesDisplayMinBytesLine),
                                nil, MIN_BYTES_LINE,
                                getIntPreference(MIN_BYTES_LINE), inputCtl{ 8, 32, 4 },
                                changed, 20 }
    maxBytesLine := contentDef { localizeText(dialogPreferencesDisplayMaxBytesLine),
                                nil, MAX_BYTES_LINE,
                                getIntPreference(MAX_BYTES_LINE), inputCtl{ 32, 64, 4 },
                                changed, 20 }
    BytesLineInc := contentDef { localizeText(dialogPreferencesDisplayLineIncrement),
                                nil, LINE_BYTE_INC,
                                getIntPreference(LINE_BYTE_INC), inputCtl{ 4, 16, 2 },
                                changed, 20 }
    ByteSeparator := contentDef { localizeText(dialogPreferencesDisplayBytesSeparator),
                                nil, HOR_SEP_SPAN,
                                getIntPreference(HOR_SEP_SPAN), inputCtl{ 0, 16, 4 },
                                changed, 20 }
    LineSeparator := contentDef { localizeText(dialogPreferencesDisplayLinesSeparator),
                                nil, VER_SEP_SPAN,
                                getIntPreference(VER_SEP_SPAN), inputCtl{ 0, 32, 8 },
                                changed, 20 }

    editorHeader := headerDef { localizeText(dialogPreferencesEditor), 5, 10 }
    startReadOnly := contentDef { localizeText(dialogPreferencesEditorReadOnly),
                                nil, START_READ_ONLY,
                                getBoolPreference(START_READ_ONLY), nil, changed, 20 }
    startReplace := contentDef { localizeText(dialogPreferencesEditorReplaceNode),
                                nil, START_REPLACE_MODE,
                                getBoolPreference(START_REPLACE_MODE), nil,
                                changed, 20 }

    searchHeader := headerDef { localizeText(dialogPreferencesSearch), 5, 10 }
    wrapAround := contentDef { localizeText(dialogPreferencesSearchWrapAround),
                                nil, WRAP_MATCHES,
                                getBoolPreference(WRAP_MATCHES), nil,
                                changed, 20 }

    return &boxDef { 5, 0, false, "", VERTICAL, []interface{} {
                     &displayHeader, &minBytesLine, &maxBytesLine,
                     &BytesLineInc, &ByteSeparator, &LineSeparator,
                     &fontHeader, &fontName, &fontSize,
                     &editorHeader, &startReadOnly, &startReplace,
                     &searchHeader, &wrapAround, &separator } }
}

func getPreferenceDialogSaveDef( ) *boxDef {

    saveHeader := headerDef { localizeText(dialogPreferencesSave), 5, 10 }
    backup := contentDef { localizeText(dialogPreferencesSaveBackup),
                           nil, CREATE_BACKUP_FILES,
                           getBoolPreference(CREATE_BACKUP_FILES), nil,
                           changed, 20 }
    return &boxDef { 15, 5, false, "", VERTICAL, []interface{} {
                                                    &saveHeader, &backup } }
}

func getPreferenceDialogTheme( ) *boxDef {

    themeNames, err := getThemeNames( )
    if err != nil {
        log.Fatalf( "Unable to get any theme name: %v\n", err )
    }

    themeHeader := headerDef { localizeText(dialogPreferencesTheme), 5, 10 }
    themeName := contentDef { localizeText(dialogPreferencesThemeName),
                              nil, COLOR_THEME_NAME,
                              getStringPreference(COLOR_THEME_NAME), themeNames,
                              changed, 20 }
    return &boxDef { 15, 5, false, "", VERTICAL, []interface{} {
                                                &themeHeader, &themeName } }
}

type preferenceNotebook struct {
    *gtk.Notebook
}

func (pn *preferenceNotebook) appendPreferencePage( bdef *boxDef,
                                                    tabId int ) {

    dbox := makeDataBox( bdef )
    tab, err := gtk.LabelNew( localizeText( tabId ) )
    if err != nil {
        log.Fatalf( "appendPreferencePage: Unable to create label: %v\n", err )
    }
    if pageIndex := pn.AppendPage( dbox.Frame, tab ); -1 == pageIndex {
        log.Fatalf( "appendPreferencePage: Unable to append page\n" )
    }
}

func showPreferencesDialog( ) {
    dialog, err := gtk.WindowNew( gtk.WINDOW_TOPLEVEL )
    if err != nil {
        log.Fatalf( "showPreferencesDialog: unable to create top-level window: %v\n", err )
    }

    prefNB := new(preferenceNotebook)
    prefNB.Notebook, err = gtk.NotebookNew( )
    if err != nil {
        log.Fatalf( "showPreferencesDialog: unable to create notebook: %v\n", err )
    }

    prefNB.Notebook.SetTabPos( gtk.POS_LEFT )
    dialog.Add( prefNB.Notebook )

    editorDef := getPreferenceDialogEditorDef( )
    prefNB.appendPreferencePage( editorDef, dialogPreferencesEditorTab )
    saveDef := getPreferenceDialogSaveDef( )
    prefNB.appendPreferencePage( saveDef, dialogPreferencesSaveTab )
    themeDef := getPreferenceDialogTheme( )
    prefNB.appendPreferencePage( themeDef, dialogPreferencesThemeTab )

    dialog.SetTransientFor( window )
    dialog.SetTypeHint( gdk.WINDOW_TYPE_HINT_DIALOG )
    dialog.SetPosition( gtk.WIN_POS_CENTER_ON_PARENT )
    dialog.SetTitle(localizeText(windowTitlePreferences))
    dialog.SetDefaultSize(300, 300)

    dialog.Connect( "delete-event", cleanPreferencesDialog )
    enablePreferences( false )
    dialog.ShowAll()
}

func cleanPreferencesDialog( pd *gtk.Window ) bool {
    enablePreferences( true )
    return false
}

// --- explore dialog

type explore struct {
    dialog *gtk.Window
    dataBox *dataBox
    data []byte
    firstBit, nBits int
    msbFirst bool
    bitStream string
    endian binary.ByteOrder
}

func (exp *explore)makeBitStream( ) {
    bPos := int(exp.firstBit/8)
    fBit := exp.firstBit % 8

    buf := make( []byte, exp.nBits )

    getBit := func( ) byte {
        if fBit == 8 {
            bPos ++
            fBit = 0
        }
        v := exp.data[bPos] << fBit
        fBit ++
        if (v & 0x80) == 0x80 {
            return '1'
        } else {
           return '0'
        }
    }

    if exp.msbFirst {
        for i := 0; i < exp.nBits; i++ {
            buf[i] = getBit()
        }
    } else {
        for i := exp.nBits-1; i >= 0 ; i-- {
            buf[i] = getBit()
        }
    }
    exp.bitStream = string(buf)
}

func (exp *explore)getBitStreamString( base int, signed bool ) (t string, ok bool) {
    v := new(big.Int)
    _, ok = v.SetString( exp.bitStream, 2 )
    if ok {
        if signed && exp.bitStream[0] == '1' {
            l := len(exp.bitStream)
            c := big.NewInt( 1 )
            c.Lsh( c, uint(l) )
            s := c.Sub( v, c )
            t = s.Text( base )
        } else {
           t = v.Text( base )
        }
    }
    return
}

func (exp *explore)updateValue( base int, signed bool, name string ) {
    if text, ok := exp.getBitStreamString( base, signed ); ok {
        setTextContent( exp.dataBox, name, text )
    }
}

func (exp *explore)updateBitStream( ) {

    exp.makeBitStream( )
    setTextContent( exp.dataBox, "BINARY_VALUE", exp.bitStream )

    exp.updateValue( 8, false, "OCTAL_VALUE" )
    exp.updateValue( 16, false, "HEXA_VALUE" )

    exp.updateValue( 10, true, "SIGNED_VALUE" )
    exp.updateValue( 10, false, "UNSIGNED_VALUE" )
}

func (exp *explore)updateFirstBit( firstBit int ) {
    bitLen := len(exp.data) << 3
    if firstBit + exp.nBits > bitLen {
        exp.nBits = bitLen - firstBit
        setIntContent( exp.dataBox, "NUMBER_BITS", exp.nBits )
    }
    exp.firstBit = firstBit
    exp.updateBitStream()
}

func (exp *explore)updateNBits( nBits int ) {
    bitLen := len(exp.data) << 3
    if exp.firstBit + nBits > bitLen {
        exp.firstBit = bitLen - nBits
        setIntContent( exp.dataBox, "FIRST_BIT", exp.firstBit )
    }
    exp.nBits = nBits
    exp.updateBitStream()
}

func getBitStreamCtlBox( exp *explore, firstBit int ) *boxDef {

    var n int
    if firstBit == 0 {
        n = 8
    } else {
        n = 4
    }

    exp.firstBit = firstBit
    exp.nBits = n

    bitLen := len(exp.data) << 3
    if firstBit + n > bitLen {
        log.Fatalf("getBitStreamCtlBox: not enough bits from first bit %d nBits %d within %d bits\n",
                    firstBit, n, bitLen )
    }

    maxNBits := bitLen
    if maxNBits > 128 {
        maxNBits = 128
    }
    maxFirstBit := bitLen - 1
    if maxFirstBit > 127 {
        maxFirstBit = 127
    }

    shiftChanged := func( name string, val interface{} ) {
        exp.updateFirstBit( int(val.(float64)) )
    }

    shift := contentDef { localizeText(dialogExploreBitStreamFirstBit),
                          nil, "FIRST_BIT", firstBit,
                          inputCtl{ 0, maxFirstBit, 1 }, shiftChanged, 0 }

    nBitsChanged := func( name string, val interface{} ) {
        exp.updateNBits( int(val.(float64)) )
    }
    nBits := contentDef { localizeText(dialogExploreBitStreamNumberBits),
                          nil, "NUMBER_BITS", n,
                          inputCtl{ 1, maxNBits, 1 }, nBitsChanged, 0 }

    var bitOrderNames [2]string
    bitOrderNames[0] = localizeText(dialogExploreBitStreamMSBFirst)
    bitOrderNames[1] = localizeText(dialogExploreBitStreamMSBLast)

    bitOrderChanged := func( name string, val interface{} ) {
        bitOrderName := val.(string)
fmt.Printf("Setting bit order to %s\b", bitOrderName)
        if bitOrderName == bitOrderNames[0] {
            exp.msbFirst = true
        } else {
            exp.msbFirst = false
        }
        exp.updateBitStream( )
    }

    exp.msbFirst = true     // default
    bitOrder := contentDef { localizeText(dialogExploreBitStreamMSB), nil,
                             "BIT_ORDER", bitOrderNames[0],
                             bitOrderNames[0:], bitOrderChanged, 0 }

    return &boxDef { 18, 20, false, "", HORIZONTAL, []interface{} {
                     &shift, &nBits, &bitOrder } }
}

func getBitStreamBox( exp *explore, firstBit int ) *boxDef {

    bitStreamTitle := localizeText(dialogExploreBitStream)
    bitStreamCtlBox := getBitStreamCtlBox( exp, firstBit )
    exp.makeBitStream( )

    octal, ok := exp.getBitStreamString( 8, false )
    var hexa, signed, unsigned string
    if ok {
        if hexa, ok = exp.getBitStreamString( 16, false ); ok {
            if signed, ok = exp.getBitStreamString( 10, true ); ok {
                unsigned, ok = exp.getBitStreamString( 10, false )
            }
        }
    }

    bitStreamCtl := constCtl{ LEFT_ALIGN, 0, true, true, true }
    bitStreamBinary := contentDef { localizeText(dialogExploreBitStreamBinary),
                                    nil, "BINARY_VALUE",
                                    exp.bitStream, bitStreamCtl, nil, 0 }
    bitStreamBinaryBox := boxDef { 18, 20, false, "", HORIZONTAL, []interface{} {
                                                     &bitStreamBinary } }

    if ok {
        bitStreamOctal := contentDef { localizeText(dialogExploreOctal),
                                       nil, "OCTAL_VALUE", octal, bitStreamCtl, nil, 0 }
        bitStreamHexa := contentDef { localizeText(dialogExploreHexa),
                                       nil, "HEXA_VALUE", hexa, bitStreamCtl, nil, 0 }
        bitStreamOH := boxDef{ 10, 20, false, "", HORIZONTAL, []interface{} {
                               &bitStreamOctal, &bitStreamHexa } }

        bitStreamUnsigned := contentDef { localizeText(dialogExploreUnsigned),
                                        nil, "UNSIGNED_VALUE", unsigned, bitStreamCtl, nil, 0 }
        bitStreamSigned := contentDef { localizeText(dialogExploreSigned),
                                        nil, "SIGNED_VALUE", signed, bitStreamCtl, nil, 0 }
        bitStreamDecimal := boxDef{ 10, 20, false, "", HORIZONTAL, []interface{} {
                                    &bitStreamUnsigned, &bitStreamSigned } }

        return &boxDef { 10, 10, true, bitStreamTitle, VERTICAL, []interface{} {
                         bitStreamCtlBox, &bitStreamBinaryBox, &bitStreamOH,
                         &bitStreamDecimal, &headerDef { " ", 0, 0 } } }
    } else  {
        return &boxDef { 10, 10, true, bitStreamTitle, VERTICAL, []interface{} {
                         bitStreamCtlBox, &bitStreamBinary } }
    }
}

const (
    SIGNED_DECIMAL_FORMAT = iota
    UNSIGNED_DECIMAL_FORMAT
    HEXADECIMAL_FORMAT
    OCTAL_FORMAT

    N_FORMATS
)

func (exp *explore)getExploreIntValue( size, format int ) string {
    formatString := [...]string{ "%d", "%d", "%x", "%o" }
    bitLen := len(exp.data) << 3
    if size > bitLen {
        return ""
    }
    switch size {
    case 8:
        if SIGNED_DECIMAL_FORMAT == format {
            return fmt.Sprintf( formatString[format], int8(exp.data[0]) )
        }
        return fmt.Sprintf( formatString[format], exp.data[0] )
    case 16:
        v := exp.endian.Uint16( exp.data[0:] )
        if SIGNED_DECIMAL_FORMAT == format {
            return fmt.Sprintf( formatString[format], int16(v) )
        }
        return fmt.Sprintf( formatString[format], v )
    case 32:
        v := exp.endian.Uint32( exp.data[0:] )
        if SIGNED_DECIMAL_FORMAT == format {
            return fmt.Sprintf( formatString[format], int32(v) )
        }
        return fmt.Sprintf( formatString[format], v )
    case 64:
        v := exp.endian.Uint64( exp.data[0:] )
        if SIGNED_DECIMAL_FORMAT == format {
            return fmt.Sprintf( formatString[format], int64(v) )
        }
        return fmt.Sprintf( formatString[format], v )
    }
    panic("Unknown int size\n")
}

func (exp *explore)getExploreFloatValue( size int ) string {
    bitLen := len(exp.data) << 3
    if size > bitLen {
        return ""
    }
    buf := bytes.NewReader( exp.data )
    switch size {
    case 32:
        var v float32
        err := binary.Read( buf, exp.endian, &v )
        if err != nil {
            log.Fatal( "getExploreFloatValue: failed to read float32:", err )
        }
        return fmt.Sprintf( "%g", v )

    case 64:
        var v float64
        err := binary.Read( buf, exp.endian, &v )
        if err != nil {
            log.Fatal( "getExploreFloatValue: failed to read float64:", err )
        }
        return fmt.Sprintf( "%g", v )
    }
    panic("Unknown float size\n")
}

func (exp *explore) updateValuesWithEndianness( ) {
    names := [...]string{ "SIGNED", "UNSIGNED", "HEXA", "OCTAL" }
    sizes := [...]int{ 16, 32, 64 }
    suffixes := [...]string{ "16", "32", "64" }
    for j := 0; j < len(sizes); j++ {
        for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
            textVal := exp.getExploreIntValue( sizes[j], i )
            setTextContent( exp.dataBox, names[i]+suffixes[j], textVal )
        }
    }

    textVal := exp.getExploreFloatValue( 32 )
    setTextContent( exp.dataBox, "FLOAT32", textVal )
    textVal = exp.getExploreFloatValue( 64 )
    setTextContent( exp.dataBox, "FLOAT64", textVal )
}

func getEndianessDef( exp *explore ) *contentDef {
    var endianNames [2]string
    endianNames[0] = localizeText(dialogExploreEndianBig)
    endianNames[1] = localizeText(dialogExploreEndianLittle)

    endiannesChanged := func( name string, val interface{} ) {
        pref := preferences{}
        endianName := val.(string)
fmt.Printf("Setting endianess to %s\b", endianName)
        if endianName == endianNames[0] {
            exp.endian = binary.BigEndian
            pref[name] = true
        } else {
            exp.endian = binary.LittleEndian
            pref[name] = false
        }
        update( pref )
        exp.updateValuesWithEndianness( )
    }

    var initialVal string
    if getBoolPreference(BIG_ENDIAN_NAME) == true {
        initialVal = endianNames[0]
        exp.endian = binary.BigEndian
    } else {
        initialVal = endianNames[1]
        exp.endian = binary.LittleEndian
    }
    return &contentDef { localizeText(dialogExploreEndian), nil,
                         BIG_ENDIAN_NAME, initialVal,
                         endianNames[0:], endiannesChanged, 20 }
}

const (
    TITLE_ALIGN = CENTER_ALIGN
    LABEL_ALIGN = RIGHT_ALIGN
    VALUE_ALIGN = RIGHT_ALIGN

    SIGNED_SIZE = 23
    UNSIGNED_SIZE = 23
    HEXDECIMAL_SIZE = 16
    OCTAL_SIZE = 25
    FLOAT_SIZE = 23
)

func getValueBox( exp *explore ) *boxDef {

    endianRow := boxDef{ 10, 10, false, "", HORIZONTAL, []interface{} {
                               getEndianessDef( exp ) } }

    valueTitle := localizeText(dialogExploreValues)
    headerStyle := &constCtl{ LABEL_ALIGN, 10, false, true, false }
    headerRow :=  boxDef{ 5, 20, false, "", HORIZONTAL, []interface{} {
                          &contentDef { localizeText(dialogExploreInt), headerStyle, "",
                                        localizeText(dialogExploreSigned),
                                        constCtl{ TITLE_ALIGN, SIGNED_SIZE, false, true, false }, nil, 0 },
                          &contentDef { "", nil, "",
                                        localizeText(dialogExploreUnsigned),
                                        constCtl{ TITLE_ALIGN, UNSIGNED_SIZE, false, true, false }, nil, 0 },
                          &contentDef { "", nil, "", localizeText(dialogExploreHexa),
                                        constCtl{ TITLE_ALIGN, HEXDECIMAL_SIZE, false, true, false }, nil, 0 },
                          &contentDef { "", nil, "", localizeText(dialogExploreOctal),
                                        constCtl{ TITLE_ALIGN, OCTAL_SIZE, false, true, false }, nil, 0 } } }

    signedCtl := constCtl{ VALUE_ALIGN, SIGNED_SIZE, true, true, true }
    unsignedCtl := constCtl{ VALUE_ALIGN, UNSIGNED_SIZE, true, true, true }
    hexaCtl := constCtl{ VALUE_ALIGN, HEXDECIMAL_SIZE, true, true, true }
    octalCtl := constCtl{ VALUE_ALIGN, OCTAL_SIZE, true, true, true }

    int8Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int8Vals[i] = exp.getExploreIntValue( 8, i )
    }

    int8Row := boxDef{ 5, 20, false, "", HORIZONTAL, []interface{} {
                &contentDef { localizeText(dialogExploreInt8), headerStyle, "SIGNED8",
                              int8Vals[SIGNED_DECIMAL_FORMAT], signedCtl, nil, 0 },
                &contentDef { "", nil, "UNSIGNED8", int8Vals[UNSIGNED_DECIMAL_FORMAT],
                              unsignedCtl, nil, 0 },
                &contentDef { "", nil, "HEXA8", int8Vals[HEXADECIMAL_FORMAT],
                              hexaCtl, nil, 0 },
                &contentDef { "", nil, "OCTAL8", int8Vals[OCTAL_FORMAT],
                              octalCtl, nil, 0 } } }

    int16Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int16Vals[i] = exp.getExploreIntValue( 16, i )
    }

    int16Row := boxDef{ 5, 20, false, "", HORIZONTAL, []interface{} {
                &contentDef { localizeText(dialogExploreInt16), headerStyle, "SIGNED16",
                              int16Vals[SIGNED_DECIMAL_FORMAT], signedCtl, nil, 0 },
                &contentDef { "", nil, "UNSIGNED16", int16Vals[UNSIGNED_DECIMAL_FORMAT],
                              unsignedCtl, nil, 0 },
                &contentDef { "", nil, "HEXA16", int16Vals[HEXADECIMAL_FORMAT],
                              hexaCtl, nil, 0 },
                &contentDef { "", nil, "OCTAL16", int16Vals[OCTAL_FORMAT],
                              octalCtl, nil, 0 } } }

    int32Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int32Vals[i] = exp.getExploreIntValue( 32, i )
    }

    int32Row := boxDef{ 5, 20, false, "", HORIZONTAL, []interface{} {
                &contentDef { localizeText(dialogExploreInt32), headerStyle, "SIGNED32",
                              int32Vals[SIGNED_DECIMAL_FORMAT], signedCtl, nil, 0 },
                &contentDef { "", nil, "UNSIGNED32", int32Vals[UNSIGNED_DECIMAL_FORMAT],
                              unsignedCtl, nil, 0 },
                &contentDef { "", nil, "HEXA32", int32Vals[HEXADECIMAL_FORMAT],
                              hexaCtl, nil, 0 },
                &contentDef { "", nil, "OCTAL32", int32Vals[OCTAL_FORMAT],
                              octalCtl, nil, 0 } } }

    int64Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int64Vals[i] = exp.getExploreIntValue( 64, i )
        fmt.Printf( "int64Vals[%d]=%s\n", i, int64Vals[i] )
    }

    int64Row := boxDef{ 5, 20, false, "", HORIZONTAL, []interface{} {
                &contentDef { localizeText(dialogExploreInt64), headerStyle, "SIGNED64",
                              int64Vals[SIGNED_DECIMAL_FORMAT], signedCtl, nil, 0 },
                &contentDef { "", nil, "UNSIGNED64", int64Vals[UNSIGNED_DECIMAL_FORMAT],
                              unsignedCtl, nil, 0 },
                &contentDef { "", nil, "HEXA64", int64Vals[HEXADECIMAL_FORMAT],
                              hexaCtl, nil, 0 },
                &contentDef { "", nil, "OCTAL64", int64Vals[OCTAL_FORMAT],
                              octalCtl, nil, 0 } } }

    floatHeader := boxDef{ 5, 20, false, "", HORIZONTAL, []interface{} {
                   &contentDef { localizeText(dialogExploreReal), headerStyle, "",
                                 localizeText(dialogExploreFloat32),
                                 constCtl{ TITLE_ALIGN, FLOAT_SIZE, false, true, false }, nil, 0 },
                   &contentDef { "", nil, "",
                                 localizeText(dialogExploreFloat64),
                                 constCtl{ TITLE_ALIGN, FLOAT_SIZE, false, true, false }, nil, 0 } } }

    float32Val := exp.getExploreFloatValue( 32 )
    float64Val := exp.getExploreFloatValue( 64 )

    floatRow := boxDef{ 5, 20, false, "", HORIZONTAL, []interface{} {
                &contentDef { "", headerStyle, "FLOAT32", float32Val,
                              constCtl{ VALUE_ALIGN, FLOAT_SIZE, true, true, true }, nil, 0 },
                &contentDef { "", nil, "FLOAT64", float64Val,
                              constCtl{ VALUE_ALIGN, FLOAT_SIZE, true, true, true }, nil, 0 } } }

    separator := headerDef { " ", 0, 0 }
    return &boxDef { 10, 10, true, valueTitle, VERTICAL, []interface{} {
                     &endianRow, &headerRow,
                     &int8Row, &int16Row, &int32Row, &int64Row,
                     &separator, &floatHeader, &floatRow, &separator } }
}

func getExploreDialogDef( exp *explore, bitOffset int ) *boxDef {

    bitStream := getBitStreamBox( exp, bitOffset )
    values := getValueBox( exp )

    exploreBox := boxDef { 15, 0, false, "", VERTICAL, []interface{} {
                           bitStream, values } }
    return &exploreBox
}

func showExploreDialog( data []byte, bitOffset int ) {

    exp := new( explore )

    var err error
    exp.dialog, err = gtk.WindowNew( gtk.WINDOW_TOPLEVEL )
    if err != nil {
        log.Fatalf( "showExploreDialog: unable to create top-level window: %v\n", err )
    }
    exp.data = data
    def := getExploreDialogDef( exp, bitOffset )
    exp.dataBox = makeDataBox( def )

    for k, v := range exp.dataBox.access {
        fmt.Printf( "explore DataBox: k=%s => %T value\n", k, v )
    }

    exp.dialog.Add( exp.dataBox.Frame )
    exp.dialog.SetTransientFor( window )
    exp.dialog.SetTypeHint( gdk.WINDOW_TYPE_HINT_DIALOG )
    exp.dialog.SetPosition( gtk.WIN_POS_CENTER_ON_PARENT )
    exp.dialog.SetTitle(localizeText(windowTitleExplore))
    exp.dialog.SetDefaultSize(300, 300)

    exp.dialog.ShowAll()
}
