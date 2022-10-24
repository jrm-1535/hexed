package main

import (
    "log"
    "fmt"
	"github.com/gotk3/gotk3/gtk"
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
        fmt.Printf("Got Special key %#x\n", key)
        return false

    default:
        if hex, _ := getNibbleFromKey( key ); hex {
            fmt.Printf("Got key %#x\n", key)
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

func findDialog( ) (op int, target string) {
    gd, err := gtk.DialogNewWithButtons( localizeText(dialogFindTitle), window,
                    gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT,
                    []interface{} { localizeText(buttonFind), gtk.RESPONSE_ACCEPT },
                    []interface{} { localizeText(buttonCancel), gtk.RESPONSE_CANCEL } )
    if err != nil {
        log.Fatal("findDialog: could not create gtk dialog:", err)
    }
    gd.SetDefaultResponse( gtk.RESPONSE_ACCEPT )
    carea, err := gd.GetContentArea()
    if err != nil {
        log.Fatal("findDialog: could not get content area:", err)
    }
    label, err := gtk.LabelNew( localizeText( findPrompt ) )
    if err != nil {
        log.Fatal("findDialog: could not create content label:", err)
    }
    carea.Container.Add( label )

    entry, err := gtk.EntryNew( )
    if err != nil {
        log.Fatal("findDialog: could not create content entry:", err)
    }

    entry.SetActivatesDefault( true )
    entry.Connect( "key-press-event", hexFilter )
    carea.Container.Add( entry )

    carea.ShowAll()
    response := gd.Run()
    switch response {
    case gtk.RESPONSE_ACCEPT:
        var err error
        target, err = entry.GetText()
        if err != nil {
            panic("Cannot get entry text\n")
        }
        op = DO
    case gtk.RESPONSE_NONE, gtk.RESPONSE_CANCEL:
        op = CANCEL
    }
    gd.Destroy()
    return
}

/*
// generic dialog management

type boxDef struct {
    spacing     int
    padding     uint

    content     []interface{}
}

const (
    inputTypeNone = 0
    inputTypeBool = 1
    inputTypeInt  = 2
    inputTypeText = 3
)

type contentDef struct {
    label       string
    name        string
    initVal     interface{}

    inputMin    int     // unused if inputType is not inputTypeInt
    inputMax    int     // unused if inputType is not inputTypeInt
}

func addBoolContent( label *gtk.Label, content *contentDef,
                     getters map[string]getInputValue ) *gtk.Box {
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
    getters[content.name] = func( ) interface{} { return input.ToggleButton.GetActive() }

    innerBox.PackStart( gtk.IWidget(input), false, false, 20 ) // in horizontal box
    innerBox.PackStart( gtk.IWidget(label), false, false, 0 )  // in horizontal box
    return innerBox
}

func addIntContent( label *gtk.Label, content *contentDef,
                    getters map[string]getInputValue ) *gtk.Box {
//    fmt.Printf( "got input type int, min %d max %d\n", content.inputMin, content.inputMax )
    intVal := content.initVal.(int)
    innerBox, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 10 )
    if nil != err {
        log.Fatal("addIntContent: Could not create inner box:", err)
    }
    input, err := gtk.SpinButtonNewWithRange( float64(content.inputMin), float64(content.inputMax), 1.0 )
    if nil != err {
        log.Fatal("addIntContent: Could not create int input button:", err)
    }

    input.SetValue( float64(intVal) )
    getters[content.name] = func( ) interface{} { return input.GetValue() }

    innerBox.PackStart( gtk.IWidget(label), false, false, 20 ) // in horizontal box
    innerBox.PackEnd( gtk.IWidget(input), false, false, 5 )    // in horizontal box
    return innerBox
}

func addTextContent( label *gtk.Label, content *contentDef,
                     getters map[string]getInputValue ) *gtk.Box {
//    fmt.Printf( "got input type text, min %d max %d\n", content.inputMin, content.inputMax )
    textVal := content.initVal.(string)
    innerBox, err := gtk.BoxNew( gtk.ORIENTATION_HORIZONTAL, 10 )
    if nil != err {
        log.Fatal("addTextContent: Could not create inner box:", err)
    }
    input, err := gtk.EntryNew( )
    if nil != err {
        log.Fatal("addTextContent: Could not create text input:", err)
    }
    input.SetMaxLength( content.inputMax )
    input.SetText( textVal )

    getTextValue := func( ) interface{} {
        getText := func() string {
            text, e := input.GetText()
            if nil != e {
                text = ""
            }
            return text
        }
        return getText()
    }
    getters[content.name] = getTextValue

    innerBox.PackStart( gtk.IWidget(label), false, false, 20 ) // in horizontal box
    innerBox.PackEnd( gtk.IWidget(input), false, false, 5 )    // in horizontal box
    return innerBox
}

func addContentToBox( box *gtk.Box, content *contentDef, padding uint,
                      getters map[string]getInputValue ) {
    label, err := gtk.LabelNew( "" )
    if nil != err {
        log.Fatal("Could not create label %s:", content.label, err)
    }
    label.SetMarkup( content.label )
    label.SetXAlign( 0.03 )

    switch content.initVal.(type) {
    case nil:
//        fmt.Printf( "got input type none\n" )
        box.PackStart( gtk.IWidget(label), false, false, 0 )      // in vertical box

    case bool:
        innerBox := addBoolContent( label, content, getters )
        box.PackStart( gtk.IWidget(innerBox), false, false, 0 )    // in vertical box (parent)

    case int:
        innerBox:= addIntContent( label, content, getters )
        box.PackStart( gtk.IWidget(innerBox), false, false, 0 )    // in vertical box (parent)

    case string:
        innerBox := addTextContent( label, content, getters )
        box.PackStart( gtk.IWidget(innerBox), false, false, 0 )    // in vertical box (parent)
    }
}

func makeBox( def * boxDef, getters map[string]getInputValue ) *gtk.Box {
//    fmt.Printf( "BoxDefT spacing %d padding %d, nb args %d\n",
//                def.spacing, def.padding, len(def.content) )
    box, err := gtk.BoxNew( gtk.ORIENTATION_VERTICAL, def.spacing )
    if nil != err {
        log.Fatal("Could not create box", err)
    }

    for _, item := range def.content {
        switch item := item.(type) {
        case *contentDef:
//            fmt.Printf( "got contentDef with label %s\n", item.label )
            addContentToBox( box, item, def.padding, getters )
        case *boxDef:
//            fmt.Printf( "got boxDef\n" )
            child := makeBox( item, getters )
            box.PackStart( gtk.IWidget(child), false, false, def.padding )
        default:
            fmt.Printf( "makeBox: got something else %v\n", item )
        }
    }
    return box
}

type getInputValue func ( ) interface{}

func dialog( name string, dialogDef *boxDef ) map[string]interface{} {
    pd, err := gtk.DialogNewWithButtons( name, window, gtk.DIALOG_MODAL | gtk.DIALOG_DESTROY_WITH_PARENT,
                                         []interface{} { localizeText(DialogButtonOk), gtk.RESPONSE_OK },
                                         []interface{} { localizeText(DialogButtonCancel), gtk.RESPONSE_CANCEL } )
    if err != nil {
        log.Fatal("dialog: could not create gtk dialog:", err)
    }
    pd.SetDefaultResponse( gtk.RESPONSE_CANCEL )
//    pd.SetResponseSensitive( gtk.RESPONSE_OK, true )

    carea, err := pd.GetContentArea()
    if err != nil {
        log.Fatal("dialog: could not get preference content area:", err)
    }

    getters := make(map[string]getInputValue)
    box := makeBox( dialogDef, getters )
    carea.Container.Add( box )
    carea.ShowAll()

    response := pd.Run()
    var result map[string]interface{}
    switch response {
    case gtk.RESPONSE_OK:
        fmt.Printf("dialog: OK\n")
        result = make(map[string]interface{})
        for pName, pGetter := range( getters ) {
            result[pName] = pGetter()
        }
        fmt.Printf("result: %v\n", result )

    case gtk.RESPONSE_CANCEL:
        fmt.Printf("dialog: CANCEL\n")
    case gtk.RESPONSE_NONE:
        fmt.Printf("dialog: NONE\n")
    }
    pd.Destroy()
    return result
}

/*
func getPreferenceDialogDef( ) *boxDef {
    channelTitle := contentDef { model.LocalizeText(model.DialogPreferencesPodcastTitle), "", nil, 0, 0 }
    maxChannels := contentDef { model.LocalizeText(model.DialogPreferencesPodcastMaximum), "max_podcasts",
                                settings.GetInt("max_podcasts"), 2, 199 }
    updateChannels := contentDef { model.LocalizeText(model.DialogPreferencesPodcastPeriod),
                                   "rss_poll_period_in_minutes",
                                   settings.GetInt("rss_poll_period_in_minutes"), 1, 60 }
    channelData := boxDef { 0, 2, []interface{} { &maxChannels, &updateChannels } }
    channelBox := boxDef { 5, 0, []interface{} { &channelTitle, &channelData } }

    videoTitle := contentDef { model.LocalizeText(model.DialogPreferencesVideoTitle), "", nil, 0, 0 }
    trashVideos := contentDef { model.LocalizeText(model.DialogPreferencesVideoTrashDelay),
                                "move_to_trash_after_period_in_days",
                                settings.GetInt("move_to_trash_after_period_in_days"), 1, 99 }
    videoData := boxDef { 0, 2, []interface{} { &trashVideos } }
    videoBox := boxDef { 5, 0, []interface{} { &videoTitle, &videoData } }

    downloadTitle := contentDef { model.LocalizeText(model.DialogPreferencesDownloadTitle), "", nil, 0, 0 }
    startWithLatest := contentDef { model.LocalizeText(model.DialogPreferencesDownloadPriority), "download_latest",
                                    settings.GetBool("download_latest"), 0, 1 }
    maxDownloads := contentDef { model.LocalizeText(model.DialogPreferencesDownloadMaximum),
                                 "max_downloads_in_parallel",
                                 settings.GetInt("max_downloads_in_parallel"), 1, 12 }
    maxDonwloadsPerChannel := contentDef { model.LocalizeText(model.DialogPreferencesDownloadMaxPerPodcast),
                                           "max_downloads_per_podcast",
                                           settings.GetInt("max_downloads_per_podcast"), 1, 9 }
    downloadData := boxDef { 0, 0, []interface{} { &startWithLatest, &maxDownloads, &maxDonwloadsPerChannel } }
    downloadBox := boxDef { 5, 0, []interface{} { &downloadTitle, &downloadData } }

    trashTitle := contentDef { model.LocalizeText(model.DialogPreferencesTrashTitle), "", nil, 0, 0 }
    emptyTrash := contentDef { model.LocalizeText(model.DialogPreferencesTrashEmptyAuto), "empty_trash_on_exit",
                               settings.GetBool("empty_trash_on_exit"), 0, 1 }
    gracePeriod := contentDef { model.LocalizeText(model.DialogPreferencesTrashEmptyDelay),
                                "empty_trash_grace_period_in_days",
                                settings.GetInt("empty_trash_grace_period_in_days"), 1, 59 }
    trashData := boxDef { 0, 0, []interface{} { &emptyTrash, &gracePeriod } }
    trashBox := boxDef { 5, 0, []interface{} { &trashTitle, &trashData } }

    historyTitle := contentDef { model.LocalizeText(model.DialogPreferencesHistoryTitle), "", nil, 0, 0 }
    timeWindow := contentDef { model.LocalizeText(model.DialogPreferencesHistoryRemember),
                               "months_before_forgetting_history",
                               settings.GetInt("months_before_forgetting_history"), 1, 12 }
    historyData := boxDef { 0, 0, []interface{} { &timeWindow } }
    historyBox := boxDef { 5, 0, []interface{} { &historyTitle, &historyData } }

    settingsTitle := contentDef { model.LocalizeText(model.DialogPreferencesSettingTitle), "", nil, 0, 0 }
    player := contentDef { model.LocalizeText(model.DialogPreferencesSettingPlayer), "player_cmd",
                           settings.GetString("player_cmd"), 2, 32 }
    torrent := contentDef { model.LocalizeText(model.DialogPreferencesSettingTorrent), "torrent_cmd",
                            settings.GetString("torrent_cmd"), 2, 32 }
    settingsData := boxDef { 0, 0, []interface{} { &player, &torrent } }
    settingsBox := boxDef { 5, 0, []interface{} { &settingsTitle, &settingsData } }

    wholeBox := boxDef { 0, 5, []interface{} { &channelBox, &videoBox, &downloadBox,
                                               &trashBox, &historyBox, &settingsBox } }
    return &wholeBox
}

func preferencesDialog( ) {
    result := dialog( model.LocalizeText(model.DialogPreferencesWindowTitle), getPreferenceDialogDef( ) )
    if result != nil {
        settings.Update( settings.Preferences(result) )
    }
}
*/

