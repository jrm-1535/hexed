package main

import (
    "log"
    "fmt"
    "bytes"
    "math/big"
    "encoding/binary"

    "internal/layout"

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

    if err != nil {
        log.Fatalf( "openFileName error: %v\n", err )
    }
    res := dialog.Run( )
    if res == gtk.RESPONSE_ACCEPT {
        name = dialog.GetFilename( )
    }
    dialog.Destroy()
    return
}

func saveFileName() (name string) {
    dialog, err := gtk.FileChooserDialogNewWith2Buttons(
                    "Save File", window, gtk.FILE_CHOOSER_ACTION_SAVE,
                    "_Cancel", gtk.RESPONSE_CANCEL,
                    "_Save", gtk.RESPONSE_ACCEPT )

    if err != nil {
        log.Fatalf( "saveFileName error: %v\n", err )
    }
    res := dialog.Run( )
    if res == gtk.RESPONSE_ACCEPT {
        name = dialog.GetFilename( )
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

var gotoHistory *layout.History

func appendGotoHistory( lo *layout.Layout ) {
    value, err := lo.GetItemValue( "gotoInp" )
    if err != nil {
        log.Fatalf("appendGotoHistory: can't get goto input\n")
    }
    choices := gotoHistory.Update( value.(string) )
    if len( choices ) > 0 {
        searchArea.SetItemChoices( "gotoInp", choices, 0, nil )
    }
}

func getAddress( s string ) (pos int64) {
    for i := 0; i < len(s); i++ {
        pos <<= 4
        pos += int64(getNibbleFromHexDigit( s[i] ))
    }
    return
}

const (
    MAX_ADDRESS_SIZE = 64                   // 2^64-1 adddress in bits
    GOTO_INPUT_SIZE = MAX_ADDRESS_SIZE / 4  // in nibbles
)

func getGotoDialogDef( origin int64 ) interface{} {

    incrementalGoto := func( name string, val interface{} ) bool {
        text := val.(string)
        gotoPos( getAddress( text ) << 1 )
        return false
    }

    promptFmt := layout.TextFmt{ layout.REGULAR, layout.CENTER, 0, false, nil }
    gotoPrm := layout.ConstDef{ "gotoPrm", 0,
                                localizeText(gotoPrompt), "", &promptFmt }
    gotoCtl := layout.StrList{ gotoHistory.Get(), true, GOTO_INPUT_SIZE,
                               nil, keyPress }
    gotoInp := layout.InputDef{ "gotoInp", 0, "", localizeText(tooltipGoto),
                                incrementalGoto, &gotoCtl }
    bd := layout.BoxDef{ "", 0, 15, 0, "", false, layout.VERTICAL,
                        []interface{}{ &gotoPrm, &gotoInp } }
    return &bd
}

func gotoDialog( ) {
    var err error
    if nil == gotoHistory {
        if gotoHistory, err = layout.NewHistory( MAX_HISTORY_DEPTH ); err != nil {
            log.Fatalf("gotoDialog: could not create goto history: %v", err)
        }
    }
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
    origin := getCurrentPos()
    lo, err := layout.NewLayout( getGotoDialogDef( origin ) )
    if err != nil {
        log.Fatal("gotoDialog: could not make layout:", err)
    }
    carea.Container.Add( lo.GetRootWidget() )
    carea.ShowAll()
    response := gd.Run()
    switch response {
    case gtk.RESPONSE_ACCEPT:
        appendGotoHistory( lo )
    case gtk.RESPONSE_NONE, gtk.RESPONSE_CANCEL:
        gotoPos( origin )
    }
    gd.Destroy()
    return
}

// ---- preferences dialog

func changed( name string, val interface{} ) bool {
    pref := preferences{}
    pref[name] = val
    updatePreferences( pref )
    return false
}

const (
    PREF_COL_SPACING uint = 0
    PREF_ROW_SPACING uint = 5

    PREF_HEAD_PADDING uint = 5
    PREF_BODY_PADDING uint = 20
)

const (
    DISPLAY_GRID = "displayGrid"

    DISPLAY_HEADER = "displayHeader"
    MIN_NBL_PROMPT = "minblPrompt"
    MAX_NBL_PROMPT = "maxblPrompt"
    NBI_PROMPT = "nbiPrompt"
    CSEP_PROMPT = "csepPrompt"
    RSEP_PROMPT = "rsepPrompt"

    FONT_HEADER = "fontHeader"
    FONT_NAME_PROMPT = "fontNamePrompt"
    FONT_SIZE_PROMPT = "fontSizePrompt"

    THEME_HEADER = "themeHeader"
    THEME_NAME_PROMPT = "themeNamePrompt"
)

func makePreferenceDialogDisplayDef( ) interface{} {

    headerFmt := layout.TextFmt{ layout.BOLD, layout.LEFT, 20, false, nil }
    bodyFmt := layout.TextFmt{ layout.REGULAR, layout.LEFT, 30, false, nil }

    displayHeader := layout.ConstDef{
                    DISPLAY_HEADER, PREF_HEAD_PADDING,
                    localizeText(dialogPreferencesDisplay), "", &headerFmt }

    tooltipSB := localizeText(tooltipSpinButton)
    minblPrompt := layout.ConstDef{
                    MIN_NBL_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesDisplayMinBytesLine), "", &bodyFmt }
    minblValCtl := layout.IntCtl{ MIN_BYTES_PER_LINE_LOWER_BOUNDARY,
                                  MIN_BYTES_PER_LINE_HIGHER_BOUNDARY,
                                  MIN_BYTES_PER_LINE_INCREMENT_STEP }
    minblVal := layout.InputDef{
                    MIN_BYTES_LINE, 0, getIntPreference(MIN_BYTES_LINE),
                    tooltipSB, changed, &minblValCtl }

    maxblPrompt := layout.ConstDef{
                    MAX_NBL_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesDisplayMaxBytesLine), "", &bodyFmt }
    maxblValCtl := layout.IntCtl{ MAX_BYTES_PER_LINE_LOWER_BOUNDARY,
                                  MAX_BYTES_PER_LINE_HIGHER_BOUNDARY,
                                  MAX_BYTES_PER_LINE_INCREMENT_STEP }
    maxblVal := layout.InputDef{
                    MAX_BYTES_LINE, 0, getIntPreference(MAX_BYTES_LINE),
                    tooltipSB, changed, &maxblValCtl }

    nbiPrompt := layout.ConstDef{
                    NBI_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesDisplayLineIncrement), "", &bodyFmt }
    nbiValCtl := layout.IntCtl{ BYTES_INCREMENT_LOWER_BOUNDARY,
                                BYTES_INCREMENT_HIGHER_BOUNDARY,
                                BYTES_INCREMENT_INCREMENT_STEP }
    nbiVal := layout.InputDef{
                    LINE_BYTE_INC, 0, getIntPreference(LINE_BYTE_INC),
                    tooltipSB, changed, &nbiValCtl }

    csepPrompt := layout.ConstDef{
                    CSEP_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesDisplayBytesSeparator), "", &bodyFmt }
    csepValCtl := layout.IntCtl{ BYTES_SEPARATOR_LOWER_BOUNDARY,
                                 BYTES_SEPARATOR_HIGHER_BOUNDARY,
                                 BYTES_SEPARATOR_INCREMENT_STEP }
    csepVal := layout.InputDef{
                    HOR_SEP_SPAN, 0, getIntPreference(HOR_SEP_SPAN),
                    tooltipSB, changed, &csepValCtl }

    rsepPrompt := layout.ConstDef{
                    RSEP_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesDisplayLinesSeparator), "", &bodyFmt }
    rsepValCtl := layout.IntCtl{ LINES_SEPARATOR_LOWER_BOUNDARY,
                                 LINES_SEPARATOR_HIGHER_BOUNDARY,
                                 LINES_SEPARATOR_INCREMENT_STEP }
    rsepVal := layout.InputDef{
                    VER_SEP_SPAN, 0, getIntPreference(VER_SEP_SPAN),
                    tooltipSB, changed, &rsepValCtl }

    fontHeader := layout.ConstDef{
                    FONT_HEADER, PREF_HEAD_PADDING,
                    localizeText(dialogPreferencesFont), "", &headerFmt }

    tooltipSL := localizeText( tooltipSelList )
    fontNamePrompt := layout.ConstDef{
                    FONT_NAME_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesFontName), "", &bodyFmt }
    fontNames := getFontNames()
    fontNameValCtl := layout.StrList{ fontNames,false, 0, nil, nil }
    fontNameVal := layout.InputDef{
                    FONT_NAME, 0, getStringPreference(FONT_NAME),
                    tooltipSL, changed, &fontNameValCtl }

    fontSizePrompt := layout.ConstDef{
                    FONT_SIZE_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesFontSize), "", &bodyFmt }
    fontSizeValCtl := layout.IntCtl{ MIN_FONT_SIZE, MAX_FONT_SIZE, FONT_SIZE_INC }
    fontSizeVal := layout.InputDef{
                    FONT_SIZE, 0, getIntPreference(FONT_SIZE),
                    tooltipSB, changed, &fontSizeValCtl }

    themeHeader := layout.ConstDef{
                    THEME_HEADER, PREF_HEAD_PADDING,
                    localizeText(dialogPreferencesTheme), "", &headerFmt }

    themeNamePrompt := layout.ConstDef{
                    THEME_NAME_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesThemeName), "", &bodyFmt }
    themeNames, err := getThemeNames( )
    if err != nil {
        log.Fatalf( "Unable to get any theme name: %v\n", err )
    }

    themeNameValCtl := layout.StrList{ themeNames, false, 0, nil, nil }
    themeNameVal := layout.InputDef{
                    COLOR_THEME_NAME, 0,
                    getStringPreference(COLOR_THEME_NAME),
                    tooltipSL, changed, &themeNameValCtl }

    gd := layout.GridDef{ DISPLAY_GRID, 10, layout.HorizontalDef{ PREF_COL_SPACING,
                                                                []layout.ColDef{
                                                                    { true },
                                                                    { false },
                                                                            },
                                                              },
                            layout.VerticalDef{ PREF_ROW_SPACING, []layout.RowDef{
                                                { false, []interface{}{
                                                          &displayHeader, nil },
                                                },
                                                { false, []interface{}{
                                                          &minblPrompt, &minblVal },
                                                },
                                                { false, []interface{}{
                                                          &maxblPrompt, &maxblVal },
                                                },
                                                { false, []interface{}{
                                                          &nbiPrompt, &nbiVal },
                                                },
                                                { false, []interface{}{
                                                          &csepPrompt, &csepVal },
                                                },
                                                { false, []interface{}{
                                                          &rsepPrompt, &rsepVal },
                                                },
                                                { false, []interface{}{
                                                          &fontHeader, nil },
                                                },
                                                { false, []interface{}{
                                                          &fontNamePrompt, &fontNameVal },
                                                },
                                                { false, []interface{}{
                                                          &fontSizePrompt, &fontSizeVal },
                                                },
                                                { false, []interface{}{
                                                          &themeHeader, nil },
                                                },
                                                { false, []interface{}{
                                                          &themeNamePrompt, &themeNameVal },
                                                },
                                            },
                                },
                 }
    bd := layout.BoxDef{ "", 0, 15, 0, "", false, layout.VERTICAL, []interface{}{ &gd } }
    return &bd
}


func updatePreferenceDialogDisplayLanguage( lo *layout.Layout ) {
    lo.SetItemValue( DISPLAY_HEADER, localizeText(dialogPreferencesDisplay) )
    lo.SetItemValue( MIN_NBL_PROMPT, localizeText(dialogPreferencesDisplayMinBytesLine) )
    lo.SetItemTooltip( MIN_BYTES_LINE, localizeText(tooltipSpinButton) )
    lo.SetItemValue( MAX_NBL_PROMPT, localizeText(dialogPreferencesDisplayMaxBytesLine) )
    lo.SetItemTooltip( MAX_BYTES_LINE, localizeText(tooltipSpinButton) )
    lo.SetItemValue( NBI_PROMPT, localizeText(dialogPreferencesDisplayLineIncrement) )
    lo.SetItemTooltip( LINE_BYTE_INC, localizeText(tooltipSpinButton) )
    lo.SetItemValue( CSEP_PROMPT, localizeText(dialogPreferencesDisplayBytesSeparator) )
    lo.SetItemTooltip( HOR_SEP_SPAN, localizeText(tooltipSpinButton) )
    lo.SetItemValue( RSEP_PROMPT, localizeText(dialogPreferencesDisplayLinesSeparator) )
    lo.SetItemTooltip( VER_SEP_SPAN, localizeText(tooltipSpinButton) )

    lo.SetItemValue( FONT_HEADER, localizeText(dialogPreferencesFont) )
    lo.SetItemValue( FONT_NAME_PROMPT, localizeText(dialogPreferencesFontName) )
    lo.SetItemTooltip( FONT_NAME, localizeText(tooltipSelList) )
    lo.SetItemValue( FONT_SIZE_PROMPT, localizeText(dialogPreferencesFontSize) )
    lo.SetItemTooltip( FONT_SIZE, localizeText(tooltipSpinButton) )

    lo.SetItemValue( THEME_HEADER, localizeText(dialogPreferencesTheme) )
    lo.SetItemValue( THEME_NAME_PROMPT, localizeText(dialogPreferencesThemeName) )
    lo.SetItemTooltip( COLOR_THEME_NAME, localizeText(tooltipSelList) )
}

const (
    EDITOR_GRID = "editorGrid"

    LANGUAGE_HEADER = "languageHeader"
    LANGUAGE_NAME_PROMPT = "languageNamePrompt"

    EDITOR_HEADER = "editorHeader"
    START_READONLY_PROMPT = "startReadOnlyPrompt"
    START_REPLACE_PROMPT = "startReplacePrompt"

    SAVE_HEADER = "saveHeader"
    SAVE_PROMPT = "savePrompt"

    SEARCH_HEADER = "searchHeader"
    SEARCH_WRAP_PROMPT = "searchWrapPrompt"
)

func makePreferenceDialogEditorDef( ) interface{} {

    headerFmt := layout.TextFmt{ layout.BOLD, layout.LEFT, 20, false, nil }
    bodyFmt := layout.TextFmt{ layout.REGULAR, layout.LEFT, 30, false, nil }

    languageHeader := layout.ConstDef{
                    LANGUAGE_HEADER, PREF_HEAD_PADDING,
                    localizeText(dialogPreferencesLanguage), "", &headerFmt }

    tooltipSL := localizeText(tooltipSelList)
    languageNamePrompt := layout.ConstDef{
                    LANGUAGE_NAME_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesLanguageName), "", &bodyFmt }
    languageNames := getLanguageNames( )
    languageNameValCtl := layout.StrList{ languageNames, false, 0, nil, nil }
    languageNameVal := layout.InputDef{
                    LANGUAGE_NAME, 0,
                    getStringPreference(LANGUAGE_NAME),
                    tooltipSL, changed, &languageNameValCtl }

    editorHeader := layout.ConstDef{
                    EDITOR_HEADER, PREF_HEAD_PADDING,
                    localizeText(dialogPreferencesEditor), "", &headerFmt }

    tooltipCK := localizeText(tooltipSetMark)
    startReadOnlyPrompt := layout.ConstDef{
                    START_READONLY_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesEditorReadOnly), "", &bodyFmt }
    startReadOnlyVal := layout.InputDef{
                    START_READ_ONLY, 0, getBoolPreference(START_READ_ONLY),
                    tooltipCK, changed, nil }

    startReplacePrompt := layout.ConstDef{
                    START_REPLACE_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesEditorReplaceMode), "", &bodyFmt }
    startReplaceVal := layout.InputDef{
                    START_REPLACE_MODE, 0, getBoolPreference(START_REPLACE_MODE),
                    tooltipCK, changed, nil }


    saveHeader := layout.ConstDef{
                    SAVE_HEADER, PREF_HEAD_PADDING,
                    localizeText(dialogPreferencesSave), "", &headerFmt }

    savePrompt := layout.ConstDef{
                    SAVE_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesSaveBackup), "", &bodyFmt }
    saveVal := layout.InputDef{
                    CREATE_BACKUP_FILES, 0, getBoolPreference(CREATE_BACKUP_FILES),
                    tooltipCK, changed, nil }

    searchHeader := layout.ConstDef{
                    SEARCH_HEADER, PREF_HEAD_PADDING,
                    localizeText(dialogPreferencesSearch), "", &headerFmt }
    searchPrompt := layout.ConstDef{
                    SEARCH_WRAP_PROMPT, PREF_BODY_PADDING,
                    localizeText(dialogPreferencesSearchWrapAround), "", &bodyFmt }
    searchVal := layout.InputDef{
                     WRAP_MATCHES, 0, getBoolPreference(WRAP_MATCHES),
                     tooltipCK, changed, nil }

    gd := layout.GridDef{ EDITOR_GRID, 10, layout.HorizontalDef{ PREF_COL_SPACING,
                                                                  []layout.ColDef{
                                                                    { true },
                                                                    { false }, },
                                                                },
                            layout.VerticalDef{ PREF_ROW_SPACING, []layout.RowDef{
                                                { false, []interface{}{
                                                          &languageHeader, nil },
                                                },
                                                { false, []interface{}{
                                                          &languageNamePrompt,
                                                          &languageNameVal },
                                                },
                                                { false, []interface{}{
                                                          &editorHeader, nil },
                                                },
                                                { false, []interface{}{
                                                          &startReadOnlyPrompt,
                                                          &startReadOnlyVal },
                                                },
                                                { false, []interface{}{
                                                          &startReplacePrompt,
                                                          &startReplaceVal },
                                                },
                                                { false, []interface{}{
                                                          &saveHeader, nil },
                                                },
                                                { false, []interface{}{
                                                          &savePrompt, &saveVal },
                                                },
                                                { false, []interface{}{
                                                          &searchHeader, nil },
                                                },
                                                { false, []interface{}{
                                                          &searchPrompt, &searchVal },
                                                }, }, }, }

    bd := layout.BoxDef{ "", 0, 15, 0, "", false, layout.VERTICAL, []interface{}{ &gd } }
    return &bd
}

func updatePreferenceDialogEditorLanguage( lo *layout.Layout ) {
    lo.SetItemValue( LANGUAGE_HEADER, localizeText(dialogPreferencesLanguage) )
    lo.SetItemValue( LANGUAGE_NAME_PROMPT, localizeText(dialogPreferencesLanguageName) )
    lo.SetItemTooltip( LANGUAGE_NAME, localizeText(tooltipSelList) )

    lo.SetItemValue( EDITOR_HEADER, localizeText(dialogPreferencesEditor) )
    lo.SetItemValue( START_READONLY_PROMPT, localizeText(dialogPreferencesEditorReadOnly) )
    lo.SetItemTooltip( START_READ_ONLY, localizeText(tooltipSetMark) )
    lo.SetItemValue( START_REPLACE_PROMPT, localizeText(dialogPreferencesEditorReplaceMode) )
    lo.SetItemTooltip( START_REPLACE_MODE, localizeText(tooltipSetMark) )

    lo.SetItemValue( SAVE_HEADER, localizeText(dialogPreferencesSave) )
    lo.SetItemValue( SAVE_PROMPT, localizeText(dialogPreferencesSaveBackup) )
    lo.SetItemTooltip( CREATE_BACKUP_FILES, localizeText(tooltipSetMark) )

    lo.SetItemValue( SEARCH_HEADER, localizeText(dialogPreferencesSearch) )
    lo.SetItemValue( SEARCH_WRAP_PROMPT, localizeText(dialogPreferencesSearchWrapAround) )
    lo.SetItemTooltip( WRAP_MATCHES, localizeText(tooltipSetMark) )
}

var preferencesDialog *layout.Dialog

func cleanPrederencesDialog( dg *layout.Dialog ) {
    enablePreferences( true )
    preferencesDialog = nil
}

func showPreferencesDialog( ) {
    display := layout.DialogPage{ localizeText( dialogPreferencesDisplayTab ),
                                  makePreferenceDialogDisplayDef( ) }
    editor := layout.DialogPage{ localizeText( dialogPreferencesEditorTab ),
                                  makePreferenceDialogEditorDef( ) }
    var err error
    preferencesDialog, err = layout.NewDialog(
                                  localizeText(windowTitlePreferences),
                                  window, nil,
                                  layout.AT_PARENT_CENTER, layout.LEFT_POS,
                                  []layout.DialogPage{ display, editor },
                                  cleanPrederencesDialog, 1, 1 )
    if err != nil {
        log.Fatalf( "showPreferencesDialog: error creating dialog: %v", err )
    }
    enablePreferences( false )
}

func refreshPreferencesLanguage( pageNumber int, page *layout.Layout ) bool {
    var tabNameId int
    if pageNumber == 0 {
        updatePreferenceDialogDisplayLanguage( page )
        tabNameId = dialogPreferencesDisplayTab
    } else {
        updatePreferenceDialogEditorLanguage( page )
        tabNameId = dialogPreferencesEditorTab
    }
    preferencesDialog.SetPageName( pageNumber, localizeText( tabNameId ) )
    return false
}

func refreshPreferencesDialogLanguage( ) {
    if preferencesDialog != nil {
        preferencesDialog.VisitContent( refreshPreferencesLanguage )
        preferencesDialog.SetTitle( localizeText(windowTitlePreferences) )
    }
}

func updatePreferencesDialogFontSize( ) {
    if preferencesDialog != nil {
        display, err := preferencesDialog.GetPage( 0 )
        if err == nil {
            display.SetItemValue( FONT_SIZE, getIntPreference( FONT_SIZE ) )
        }
    }
}

// --- explore dialog

type explore struct {
    dialog      *layout.Dialog
    lo          *layout.Layout
    data        []byte
    offset      int64
    firstBit,
    nBits       int
    msbFirst    bool
    bitStream   string
    endian      binary.ByteOrder
}

func (exp *explore)setDialogTitle( ) {
    title := fmt.Sprintf( "%s @%#x bit %d", localizeText(windowTitleExplore),
                          exp.offset, exp.firstBit )
    exp.dialog.SetTitle( title )
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

func (exp *explore)updateValue( name string, base int, signed bool ) {
    if text, ok := exp.getBitStreamString( base, signed ); ok {
        exp.lo.SetItemValue( name, text )
    }
}

const (
    BITSTREAM_HEADER = "bitstream"

    FIRST_BIT_PRM = "firstBitPrm"
    FIRST_BIT = "firstBit"

    NUMBER_BITS_PRM = "numBitsPrm"
    NUMBER_BITS = "numBits"

    MSBF_PRM = "msbFirstPrm"

    BINARY_PRM = "binPrm"
    BINARY_VAL = "binVal"

    OCTAL_PRM = "octalPrm"
    OCTAL_VAL = "octalVal"

    HEXA_PRM = "hexaPrm"
    HEXA_VAL = "hexaVal"

    UNSIGNED_DEC_PRM = "uDecPrm"
    UNSIGNED_DEC = "uDec"

    SIGNED_DEC_PRM = "sDecPrm"
    SIGNED_DEC = "sDec"
)

func (exp *explore)updateBitStream( ) bool {

    exp.makeBitStream( )
    exp.lo.SetItemValue( BINARY_VAL, exp.bitStream )

    exp.updateValue( OCTAL_VAL, 8, false )
    exp.updateValue( HEXA_VAL, 16, false )

    exp.updateValue( SIGNED_DEC, 10, true )
    exp.updateValue( UNSIGNED_DEC, 10, false )
    return false
}

func (exp *explore)updateFirstBit( firstBit int ) bool {
    bitLen := len(exp.data) << 3
    if firstBit + exp.nBits > bitLen {
        exp.nBits = bitLen - firstBit
        exp.lo.SetItemValue( NUMBER_BITS, exp.nBits )
    }
    exp.firstBit = firstBit
    exp.setDialogTitle()
    return exp.updateBitStream()
}

func (exp *explore)updateNBits( nBits int ) bool {
    bitLen := len(exp.data) << 3
    if exp.firstBit + nBits > bitLen {
        exp.firstBit = bitLen - nBits
        exp.lo.SetItemValue( FIRST_BIT, exp.firstBit )
    }
    exp.nBits = nBits
    return exp.updateBitStream()
}

func (exp *explore)getBitOrderControl( ) (bitOrderNames []string,
                                          bitOrder int,
                                          bitOrderChanged func(
                                              string, interface{} ) bool) {
    bitOrderNames = make( []string, 2 )
    bitOrderNames[0] = localizeText(dialogExploreBitStreamMSBFirst)
    bitOrderNames[1] = localizeText(dialogExploreBitStreamMSBLast)

    if getBoolPreference(BITSTREAM_MSBF) == true {
        exp.msbFirst = true
        bitOrder = 0
    } else {
        exp.msbFirst = false
        bitOrder = 1
    }

    bitOrderChanged = func( name string, val interface{} ) bool {
        pref := preferences{}
        orderName := val.(string)
        // localize in case language has changed in the meantime
        if orderName == localizeText(dialogExploreBitStreamMSBFirst) {
            exp.msbFirst = true
            pref[name] = true
        } else {
            exp.msbFirst = false
            pref[name] = false
        }
        updatePreferences( pref )
        return exp.updateBitStream( )
    }
    return
}

func (exp *explore) setBitstreamCoeff( firstBit int ) (maxNBits,
                                                       maxFirstBit int) {

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
        log.Fatalf("setBitstreamCoeff: not enough bits from first bit %d nBits %d within %d bits\n",
                    firstBit, n, bitLen )
    }

    maxNBits = bitLen
    if maxNBits > 128 {
        maxNBits = 128
    }
    maxFirstBit = bitLen - 1
    if maxFirstBit > 127 {
        maxFirstBit = 127
    }
    return
}

// explore dialog spacing & padding
const (
    EXP_COL_SPACING uint = 20
    EXP_ROW_SPACING uint = 5
    EXP_BODY_PADDING uint = 10
)

func getBitstreamBoxDef( exp *explore, firstBit int,
                         tooltipSP, tooltipSL, tooltipCC string,
                         cc func( string, *gdk.Event ) bool ) *layout.BoxDef {

    bodyFmt := layout.TextFmt{ layout.REGULAR, layout.LEFT, 0, false, nil }
    numberFmt := layout.TextFmt{ layout.MONOSPACE, layout.LEFT, 0, true, cc }

    maxNBits, maxFirstBit := exp.setBitstreamCoeff( firstBit )
    firstBitPrm := layout.ConstDef{
                    FIRST_BIT_PRM, EXP_BODY_PADDING,
                    localizeText(dialogExploreBitStreamFirstBit), "", &bodyFmt }
    firstBitCtl := layout.IntCtl{ 0, maxFirstBit, 1 }

    shiftChanged := func( name string, val interface{} ) bool {
        return exp.updateFirstBit( int(val.(float64)) )
    }
    firstBitVal := layout.InputDef{
                    FIRST_BIT, 0, exp.firstBit, tooltipSP, shiftChanged, &firstBitCtl }

    numBitsPrm := layout.ConstDef{
                    NUMBER_BITS_PRM, EXP_BODY_PADDING,
                    localizeText(dialogExploreBitStreamNumberBits), "", &bodyFmt }
    numBitsCtl := layout.IntCtl{ 1, maxNBits, 1 }

    nBitsChanged := func( name string, val interface{} ) bool {
        exp.updateNBits( int(val.(float64)) )
        return false
    }
    numBitsVal := layout.InputDef{
                    NUMBER_BITS, 0, exp.nBits, tooltipSP, nBitsChanged, &numBitsCtl }

    msbFirstPrm := layout.ConstDef{
                    MSBF_PRM, EXP_BODY_PADDING,
                    localizeText(dialogExploreBitStreamMSB), "", &bodyFmt }

    orderNames,order, orderChanged := exp.getBitOrderControl( )
    msbFirstCtl := layout.StrList{ orderNames, false, 0, nil, nil }
    msbFirstVal := layout.InputDef{
                    BITSTREAM_MSBF, 0, orderNames[order], tooltipSL,
                    orderChanged, &msbFirstCtl }

    bitControl := layout.BoxDef{ "", 5, 0, 0, "", false, layout.HORIZONTAL,
                                 []interface{} { &firstBitPrm, &firstBitVal,
                                                 &numBitsPrm, &numBitsVal,
                                                 &msbFirstPrm, &msbFirstVal } }

    binaryPrm := layout.ConstDef{
                    BINARY_PRM, EXP_BODY_PADDING,
                    localizeText(dialogExploreBitStreamBinary), "", &bodyFmt }

    exp.makeBitStream( )
    binaryVal := layout.ConstDef{
                    BINARY_VAL, 0,
                    exp.bitStream, tooltipCC, &numberFmt }
    binary := layout.BoxDef{ "", 5, 0, 0, "", false, layout.HORIZONTAL,
                             []interface{} {
                                        &binaryPrm,
                                        &binaryVal } }

    octal, ok := exp.getBitStreamString( 8, false )
    var hexa, signed, unsigned string
    if ok {
        if hexa, ok = exp.getBitStreamString( 16, false ); ok {
            if signed, ok = exp.getBitStreamString( 10, true ); ok {
                unsigned, ok = exp.getBitStreamString( 10, false )
            }
        }
    }
    var bitBox layout.BoxDef
    if ok {
        octalPrm := layout.ConstDef{
                    OCTAL_PRM, EXP_BODY_PADDING,
                    localizeText(dialogExploreOctal), "", &bodyFmt }
        octalVal := layout.ConstDef{
                    OCTAL_VAL, 0, octal, tooltipCC, &numberFmt }
        hexaPrm := layout.ConstDef{
                    HEXA_PRM, EXP_BODY_PADDING,
                    localizeText(dialogExploreHexa), "", &bodyFmt }
        hexaVal := layout.ConstDef{
                    HEXA_VAL, 0, hexa, tooltipCC, &numberFmt }

        bases8_16 := layout.BoxDef{ "", 5, 0, 0, "", false, layout.HORIZONTAL,
                                []interface{} { &octalPrm, &octalVal,
                                                &hexaPrm, &hexaVal } }

        unsignedPrm := layout.ConstDef{
                        UNSIGNED_DEC_PRM, EXP_BODY_PADDING,
                        localizeText(dialogExploreUnsigned), "", &bodyFmt }
        unsignedVal := layout.ConstDef{
                        UNSIGNED_DEC, 0, unsigned, tooltipCC, &numberFmt }
        signedPrm := layout.ConstDef{
                        SIGNED_DEC_PRM, EXP_BODY_PADDING,
                        localizeText(dialogExploreSigned), "", &bodyFmt }
        signedVal := layout.ConstDef{
                        SIGNED_DEC, 0, signed, tooltipCC, &numberFmt }

        decimal := layout.BoxDef{ "", 5, 0, 0, "", false, layout.HORIZONTAL,
                                  []interface{} { &unsignedPrm, &unsignedVal,
                                                  &signedPrm, &signedVal } }

        bitBox = layout.BoxDef{ "", 10, 20, 10, "", false, layout.VERTICAL,
                             []interface{} { &bitControl, &binary,
                                             &bases8_16, &decimal } }
    } else {
        bitBox = layout.BoxDef{ "", 10, 20, 10, "", false, layout.VERTICAL,
                             []interface{} { &bitControl, &binary } }
    }
    return  &layout.BoxDef{ BITSTREAM_HEADER, 5, 5, 0,
                            localizeText(dialogExploreBitStream),
                            true, layout.VERTICAL, []interface{}{ &bitBox } }
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
    if size <= bitLen {
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
        default:
            log.Panicf( "getExploreIntValue: unsupported int size %d\n", size )
        }
    }
    return ""
}

func (exp *explore)getExploreFloatValue( size int ) string {
    bitLen := len(exp.data) << 3
    if size <= bitLen {
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
        default:
            log.Panicf( "getExploreIntValue: unsupported float size %d\n", size )
        }
    }
    return ""
}

const (
    VALUE_HEADER = "values"     // box header

    ENDIAN_PRM = "endianPrm"    // endian prompt

    INT_HEADER = "int"          // sub-box int header
    SIGNED_INT = "signed"       // content value
    UNSIGNED_INT = "unsigned"   // content value
    HEXA_INT = "hexa"           // content value
    OCTAL_INT = "octal"         // content value

    INT8 = "int8"
    S8 = "signed8"              // content value
    U8 = "unsigned8"            // content value
    H8 = "hexa8"                // content value
    O8 = "octal8"               // content value

    INT16 = "int16"
    S16 = "signed16"            // content value
    U16 = "unsigned16"          // content value
    H16 = "hexa16"              // content value
    O16 = "octal16"             // content value

    INT32 = "int32"
    S32 = "signed32"            // content value
    U32 = "unsigned32"          // content value
    H32 = "hexa32"              // content value
    O32 = "octal32"             // content value

    INT64 = "int64"
    S64 = "signed64"            // content value
    U64 = "unsigned64"          // content value
    H64 = "hexa64"              // content value
    O64 = "octal64"             // content value

    REAL_HEADER = "real"        // sub-box real header

    REAL32 = "real32"
    F32 = "float32"             // content value

    REAL64 = "real64"
    F64 = "float64"             // content value
)

// Make sure constant content value names above match the names below
func (exp *explore) updateValuesWithEndianness( ) {
    names := [...]string{ "signed", "unsigned", "hexa", "octal" }
    sizes := [...]int{ 16, 32, 64 }
    suffixes := [...]string{ "16", "32", "64" }
    for j := 0; j < len(sizes); j++ {
        for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
            textVal := exp.getExploreIntValue( sizes[j], i )
            exp.lo.SetItemValue( names[i]+suffixes[j], textVal )
        }
    }
    exp.lo.SetItemValue( F32, exp.getExploreFloatValue( 32 ) )
    exp.lo.SetItemValue( F64, exp.getExploreFloatValue( 64 ) )
}

func (exp *explore)getEndianessControl( ) (endianNames []string, endian int,
                                           changed func(string, interface{}) bool) {
    endianNames = make( []string, 2 )
    endianNames[0] = localizeText(dialogExploreEndianBig)
    endianNames[1] = localizeText(dialogExploreEndianLittle)

    if getBoolPreference(BIG_ENDIAN_NAME) == true {
        endian = 0
        exp.endian = binary.BigEndian
    } else {
        endian = 1
        exp.endian = binary.LittleEndian
    }

    changed = func( name string, val interface{} ) bool {
        pref := preferences{}
        endianName := val.(string)
        // localize in case language has changed in the meantime
        if endianName == localizeText(dialogExploreEndianBig) {
            exp.endian = binary.BigEndian
            pref[name] = true
        } else {
            exp.endian = binary.LittleEndian
            pref[name] = false
        }
        updatePreferences( pref )
        exp.updateValuesWithEndianness( )
        return false
    }
    return
}

func getValueBoxDef( exp *explore,
                     tooltipSP, tooltipSL, tooltipCC string,
                     cc func( string, *gdk.Event ) bool ) *layout.BoxDef {
    const (
        INT_SIZE = 25
        FLOAT_SIZE = 23
    )

    bodyFmt := layout.TextFmt{ layout.REGULAR, layout.LEFT, 0, false, nil }
    headerFmt := layout.TextFmt{ layout.MONOSPACE, layout.RIGHT, 8, false, nil }
    intHeaderFmt := layout.TextFmt{ layout.MONOSPACE, layout.CENTER, INT_SIZE, false, nil }
    valueFmt := layout.TextFmt{ layout.MONOSPACE, layout.RIGHT, 0, true, cc }
    realHeaderFmt := layout.TextFmt{ layout.MONOSPACE, layout.CENTER, FLOAT_SIZE, false, nil }

    endianNames, endian, eChanged := exp.getEndianessControl( )
    endianPrm := layout.ConstDef{
                    ENDIAN_PRM, EXP_BODY_PADDING,
                    localizeText(dialogExploreEndian), "", &bodyFmt }

    endianCtl := layout.StrList{ endianNames, false, 0, nil, nil }
    endianVal := layout.InputDef{
                    BIG_ENDIAN_NAME, 0, endianNames[endian],
                    tooltipSL, eChanged, &endianCtl }

    endianBox := layout.BoxDef{ "", 5, 0, 0, "", false, layout.HORIZONTAL,
                                []interface{} { &endianPrm, &endianVal } }

    intHeader := layout.ConstDef{
                    INT_HEADER, EXP_BODY_PADDING,
                    localizeText(dialogExploreInt), "", &headerFmt }
    signed := layout.ConstDef{
                    SIGNED_INT, 0,
                    localizeText(dialogExploreSigned), "", &intHeaderFmt }
    unsigned := layout.ConstDef{
                    UNSIGNED_INT, 0,
                    localizeText(dialogExploreUnsigned), "", &intHeaderFmt }
    hexa := layout.ConstDef{
                    HEXA_INT, 0,
                    localizeText(dialogExploreHexa), "", &intHeaderFmt }
    octal := layout.ConstDef{
                    OCTAL_INT, 0,
                    localizeText(dialogExploreOctal), "", &intHeaderFmt }


    int8Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int8Vals[i] = exp.getExploreIntValue( 8, i )
    }
    byteHeader := layout.ConstDef{
                    INT8, EXP_BODY_PADDING,
                    localizeText(dialogExploreInt8), "", &headerFmt }

    signedByte := layout.ConstDef{ S8, 0, int8Vals[0], tooltipCC, &valueFmt }
    unsignedByte := layout.ConstDef{ U8, 0, int8Vals[1], tooltipCC, &valueFmt }
    hexaByte := layout.ConstDef{ H8, 0, int8Vals[2], tooltipCC, &valueFmt }
    octalByte := layout.ConstDef{ O8, 0, int8Vals[3], tooltipCC, &valueFmt }

    int16Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int16Vals[i] = exp.getExploreIntValue( 16, i )
    }

    wordHeader := layout.ConstDef{
                    INT16, EXP_BODY_PADDING,
                    localizeText(dialogExploreInt16), "", &headerFmt }
    signedWord := layout.ConstDef{ S16, 0, int16Vals[0], tooltipCC, &valueFmt }
    unsignedWord := layout.ConstDef{ U16, 0, int16Vals[1], tooltipCC, &valueFmt }
    hexaWord := layout.ConstDef{ H16, 0, int16Vals[2], tooltipCC, &valueFmt }
    octalWord := layout.ConstDef{ O16, 0, int16Vals[3], tooltipCC, &valueFmt }

    int32Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int32Vals[i] = exp.getExploreIntValue( 32, i )
    }
    longHeader := layout.ConstDef{
                    INT32, EXP_BODY_PADDING,
                    localizeText(dialogExploreInt32), "", &headerFmt }
    signedLong := layout.ConstDef{ S32, 0, int32Vals[0], tooltipCC, &valueFmt }
    unsignedLong := layout.ConstDef{ U32, 0, int32Vals[1], tooltipCC, &valueFmt }
    hexaLong := layout.ConstDef{ H32, 0, int32Vals[2], tooltipCC, &valueFmt }
    octalLong := layout.ConstDef{ O32, 0, int32Vals[3], tooltipCC, &valueFmt }

    int64Vals := make( []string, N_FORMATS )
    for i:= SIGNED_DECIMAL_FORMAT; i < N_FORMATS; i++ {
        int64Vals[i] = exp.getExploreIntValue( 64, i )
    }
    llongHeader := layout.ConstDef{
                    INT64, EXP_BODY_PADDING,
                    localizeText(dialogExploreInt64), "", &headerFmt }
    signedLlong := layout.ConstDef{ S64, 0, int64Vals[0], tooltipCC, &valueFmt }
    unsignedLlong := layout.ConstDef{ U64, 0, int64Vals[1], tooltipCC, &valueFmt }
    hexaLlong := layout.ConstDef{ H64, 0, int64Vals[2], tooltipCC, &valueFmt }
    octalLlong := layout.ConstDef{ O64, 0, int64Vals[0], tooltipCC, &valueFmt }

    float32Val := exp.getExploreFloatValue( 32 )
    float64Val := exp.getExploreFloatValue( 64 )
    realHeader := layout.ConstDef{
                    REAL_HEADER, EXP_BODY_PADDING,
                    localizeText(dialogExploreReal), "", &headerFmt }
    real32 := layout.ConstDef{
                    REAL32, 0,
                    localizeText(dialogExploreFloat32), "", &realHeaderFmt }
    real32Val := layout.ConstDef{ F32, 0, float32Val, tooltipCC, &valueFmt }
    real64 := layout.ConstDef{
                    REAL64, 0,
                    localizeText(dialogExploreFloat64), "", &realHeaderFmt }
    real64Val := layout.ConstDef{ F64, 0, float64Val, tooltipCC, &valueFmt }

    intValGrid := layout.GridDef{ "", 10, layout.HorizontalDef{ EXP_COL_SPACING,
                                                                []layout.ColDef{
                                                                    { false },
                                                                    { false },
                                                                    { false },
                                                                    { false },
                                                                    { false },
                                                                        },
                                                              },
                                  layout.VerticalDef{ EXP_ROW_SPACING, []layout.RowDef{
                                                        { false, []interface{} {
                                                                  &intHeader,
                                                                  &signed,
                                                                  &unsigned,
                                                                  &hexa,
                                                                  &octal } },
                                                        { false, []interface{} {
                                                                  &byteHeader,
                                                                  &signedByte,
                                                                  &unsignedByte,
                                                                  &hexaByte,
                                                                  &octalByte } },
                                                        { false, []interface{} {
                                                                  &wordHeader,
                                                                  &signedWord,
                                                                  &unsignedWord,
                                                                  &hexaWord,
                                                                  &octalWord } },
                                                        { false, []interface{} {
                                                                  &longHeader,
                                                                  &signedLong,
                                                                  &unsignedLong,
                                                                  &hexaLong,
                                                                  &octalLong } },
                                                        { false, []interface{} {
                                                                  &llongHeader,
                                                                  &signedLlong,
                                                                  &unsignedLlong,
                                                                  &hexaLlong,
                                                                  &octalLlong } },
                                                                            },
                                                },
                                }

    realValGrid := layout.GridDef{ "", 10, layout.HorizontalDef{ EXP_COL_SPACING,
                                   []layout.ColDef{
                                            { false },
                                            { false },
                                            { false },
                                                },
                                    },
                                  layout.VerticalDef{ EXP_ROW_SPACING,
                                    []layout.RowDef{
                                        { false, []interface{} {
                                                  &realHeader,
                                                  &real32,
                                                  &real64 } },
                                        { false, []interface{} {
                                                  nil,
                                                  &real32Val,
                                                  &real64Val } },
                                                   },
                                                },
                                }

    // add vertical margin for first and last item
    valBox := layout.BoxDef{ "", 10, 10, 15, "", false, layout.VERTICAL,
                             []interface{} {
                                        &endianBox,
                                        &intValGrid,
                                        &realValGrid } }

    return &layout.BoxDef{ VALUE_HEADER, 5, 5, 15,
                           localizeText(dialogExploreValues),
                           true, layout.VERTICAL, []interface{} { &valBox } }
}

func makeExploreDialogDef( exp *explore, firstBit int ) interface{} {

    tooltipSP := localizeText(tooltipSpinButton)
    tooltipSL := localizeText(tooltipSelList)
    tooltipCC := localizeText(tooltipCopyValue)

    copyContent := func( name string, event *gdk.Event ) bool {
        log.Println( "copyContent called" )
        copy2Clipboard := func( ) {
            val, err := exp.lo.GetItemValue( name )
            if err == nil {
                t, ok := val.(string)
                if ok {
                    setClipboardAscii( t )
                }
            }
        }
        layout.AddPopupMenuItem( "copyValue", localizeText(actionCopyValue), copy2Clipboard )
        aNames := []string{ "copyValue" }
        layout.PopupContextMenu( aNames, event )
        layout.DelPopupMenuItem( "copyValue" )
        return true
    }

    bits := getBitstreamBoxDef( exp, firstBit,
                                tooltipSP, tooltipSL, tooltipCC, copyContent )
    values := getValueBoxDef( exp, tooltipSP, tooltipSL, tooltipCC, copyContent )
    bd := layout.BoxDef{ "", 0, 5, 5, "", false, layout.VERTICAL,
                         []interface{} {
                                bits, values,
                                 } }
    return &bd
}

func showExploreDialog( data []byte, nibblePos int64 ) {

    exp := new( explore )
    exp.data = data
    exp.offset = nibblePos >> 1

    var firstBit int
    if nibblePos & 1 == 1 {
        firstBit = 4
    } else {
        firstBit = 0
    }

    expDef := layout.DialogPage{ "", makeExploreDialogDef( exp, firstBit ) }
    dg, err := layout.NewDialog( "", window, exp,
                                 layout.AT_PARENT_CENTER,layout.LEFT_POS,
                                 []layout.DialogPage{ expDef }, nil, 300, 300 )
    if err != nil {
        log.Fatalf( "showExploreDialog: error creating dialog: %v", err )
    }

    exp.lo, err = dg.GetPage(0)
    if err != nil {
        log.Fatalf( "showExploreDialog: error getting page: %v", err )
    }
    exp.dialog = dg
    exp.setDialogTitle( )
}

func refreshExploreLanguage( dg *layout.Dialog ) bool {
    if exp, ok := dg.GetUserData().(*explore); ok {
        tooltipSP := localizeText(tooltipSpinButton)
        tooltipSL := localizeText(tooltipSelList)
        tooltipCC := localizeText(tooltipCopyValue)

        exp.lo.SetItemValue( BITSTREAM_HEADER, localizeText(dialogExploreBitStream) )

        exp.lo.SetItemValue( FIRST_BIT_PRM, localizeText(dialogExploreBitStreamFirstBit) )
        exp.lo.SetItemValue( NUMBER_BITS_PRM, localizeText(dialogExploreBitStreamNumberBits) )
        exp.lo.SetItemValue( MSBF_PRM, localizeText(dialogExploreBitStreamMSB) )
                             orderNames, order, orderChanged := exp.getBitOrderControl( )
        exp.lo.SetItemChoices( BITSTREAM_MSBF, orderNames, order, orderChanged )

        exp.lo.SetItemTooltip( FIRST_BIT, tooltipSP )
        exp.lo.SetItemTooltip( NUMBER_BITS, tooltipSP )
        exp.lo.SetItemTooltip( BITSTREAM_MSBF, tooltipSL )

        exp.lo.SetItemValue( BINARY_PRM, localizeText(dialogExploreBitStreamBinary) )
        exp.lo.SetItemValue( OCTAL_PRM, localizeText(dialogExploreOctal) )
        exp.lo.SetItemValue( HEXA_PRM, localizeText(dialogExploreHexa) )
        exp.lo.SetItemValue( UNSIGNED_DEC_PRM, localizeText(dialogExploreUnsigned) )
        exp.lo.SetItemValue( SIGNED_DEC_PRM, localizeText(dialogExploreSigned) )

        exp.lo.SetItemTooltip( BINARY_VAL, tooltipCC )
        exp.lo.SetItemTooltip( OCTAL_VAL, tooltipCC )
        exp.lo.SetItemTooltip( HEXA_VAL, tooltipCC )
        exp.lo.SetItemTooltip( UNSIGNED_DEC, tooltipCC )
        exp.lo.SetItemTooltip( SIGNED_DEC, tooltipCC )

        exp.lo.SetItemValue( VALUE_HEADER, localizeText(dialogExploreValues) )

        exp.lo.SetItemValue( ENDIAN_PRM, localizeText(dialogExploreEndian) )

        endianNames, endian, endianChanged := exp.getEndianessControl( )
        exp.lo.SetItemChoices( BIG_ENDIAN_NAME, endianNames, endian, endianChanged )

        exp.lo.SetItemTooltip( BIG_ENDIAN_NAME, tooltipSL )

        exp.lo.SetItemValue( INT_HEADER, localizeText(dialogExploreInt) )
        exp.lo.SetItemValue( SIGNED_INT, localizeText(dialogExploreSigned) )
        exp.lo.SetItemValue( UNSIGNED_INT, localizeText(dialogExploreUnsigned) )
        exp.lo.SetItemValue( HEXA_INT, localizeText(dialogExploreHexa) )
        exp.lo.SetItemValue( OCTAL_INT, localizeText(dialogExploreOctal) )

        exp.lo.SetItemValue( INT8, localizeText(dialogExploreInt8) )
        exp.lo.SetItemTooltip( S8, tooltipCC )
        exp.lo.SetItemTooltip( U8, tooltipCC )
        exp.lo.SetItemTooltip( H8, tooltipCC )
        exp.lo.SetItemTooltip( O8, tooltipCC )

        exp.lo.SetItemValue( INT16, localizeText(dialogExploreInt16) )
        exp.lo.SetItemTooltip( S16, tooltipCC )
        exp.lo.SetItemTooltip( U16, tooltipCC )
        exp.lo.SetItemTooltip( H16, tooltipCC )
        exp.lo.SetItemTooltip( O16, tooltipCC )

        exp.lo.SetItemValue( INT32, localizeText(dialogExploreInt32) )
        exp.lo.SetItemTooltip( S32, tooltipCC )
        exp.lo.SetItemTooltip( U32, tooltipCC )
        exp.lo.SetItemTooltip( H32, tooltipCC )
        exp.lo.SetItemTooltip( O32, tooltipCC )

        exp.lo.SetItemValue( INT64, localizeText(dialogExploreInt64) )
        exp.lo.SetItemTooltip( S64, tooltipCC )
        exp.lo.SetItemTooltip( U64, tooltipCC )
        exp.lo.SetItemTooltip( H64, tooltipCC )
        exp.lo.SetItemTooltip( O64, tooltipCC )

        exp.lo.SetItemValue( REAL_HEADER, localizeText(dialogExploreReal) )
        exp.lo.SetItemValue( REAL32, localizeText(dialogExploreFloat32) )
        exp.lo.SetItemTooltip( F32, tooltipCC )

        exp.lo.SetItemValue( REAL64, localizeText(dialogExploreFloat64) )
        exp.lo.SetItemTooltip( F64, tooltipCC )
    }
    return false
}

func refreshExploreDialogsLanguage( ) {
    layout.VisitDialogs( refreshExploreLanguage )
}

func refreshDialogs( ) {
    refreshPreferencesDialogLanguage( )
    refreshExploreDialogsLanguage( )
}
