package main

import (
    "log"

    "internal/layout"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

type menuTextIds struct {
    title, hint int
}

var menuResIds map[string]menuTextIds
var protectState bool

func setProtectedState( state bool ) {
    protectState = state
    refreshProtectMenuItem( state )
}

// initially enabled actions (when no file is opened yet)
const (
    ENABLE_NEW = true
    ENABLE_OPEN = true
    ENABLE_SAVE = false
    ENABLE_SAVE_AS = false
    ENABLE_REVERT = false
    ENABLE_RECENT = false
    ENABLE_CLOSE = false
    ENABLE_EXIT = true

    ENABLE_UNDO = false
    ENABLE_REDO = false
    ENABLE_PROTECT = false
    ENABLE_CUT = false
    ENABLE_COPY = false
    ENABLE_PASTE = false
    ENABLE_DELETE = false
    ENABLE_SELECT_ALL = false
    ENABLE_EXPLORE = false
    ENABLE_PREFERENCES = true

    ENABLE_FIND = false
    ENABLE_REPLACE = false
    ENABLE_GOTO = false

    ENABLE_CONTENTS = true
    ENABLE_ABOUT = true
)

func getMenuDefs( ) ( int, *[]layout.MenuItemDef ) {

    menuResIds = make( map[string]menuTextIds )
    noAccel := layout.AccelCode{ 0, 0, 0 }

    separator := layout.MenuItemDef{ "", "", "", nil, nil, noAccel, false }

    np := func( ) { newPage( "", false ) }
    op := func( ) { fileName := openFileName( )
                        if fileName != "" {
                            newPage( fileName, false )
                        }
                  }
    xit := func( ) { exitApplication( nil ) }

    menuResIds["new"] = menuTextIds{ menuFileNew, menuFileNewHelp }
    menuResIds["open"] = menuTextIds{ menuFileOpen, menuFileOpenHelp }
    menuResIds["save"] = menuTextIds{ menuFileSave, menuFileSaveHelp }
    menuResIds["saveAs"] = menuTextIds{ menuFileSaveAs, menuFileSaveAsHelp }
    menuResIds["revert"] = menuTextIds{ menuFileRevert, menuFileRevertHelp }
    menuResIds["recent"] = menuTextIds{ menuFileRecent, menuFileRecentHelp }
    menuResIds["close"] = menuTextIds{ menuFileClose, menuFileCloseHelp }
    menuResIds["exit"] = menuTextIds{ menuFileQuit, menuFileQuitHelp }

    var fileMenuDef = []layout.MenuItemDef {
        { "new", localizeText(menuFileNew), localizeText(menuFileNewHelp),
          nil, np, layout.AccelCode{ 'n', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
          ENABLE_NEW },
        { "open", localizeText(menuFileOpen), localizeText(menuFileOpenHelp),
          nil, op, layout.AccelCode{ 'o', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
          ENABLE_OPEN },
        separator,
        { "save", localizeText(menuFileSave), localizeText(menuFileSaveHelp),
          nil, saveCurrentPage, layout.AccelCode{ 's', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_SAVE },
        { "saveAs", localizeText(menuFileSaveAs), localizeText(menuFileSaveAsHelp),
          nil, saveCurrentPageAs, layout.AccelCode{ 's', gdk.CONTROL_MASK |
          gdk.SHIFT_MASK, gtk.ACCEL_VISIBLE }, ENABLE_SAVE_AS },
        { "revert", localizeText(menuFileRevert), localizeText(menuFileRevertHelp),
          nil, revertCurrentPage, noAccel, ENABLE_REVERT },
        separator,
        { "recent", localizeText(menuFileRecent), localizeText(menuFileRecentHelp),
          nil, nil, noAccel, ENABLE_RECENT },
        { "close", localizeText(menuFileClose), localizeText(menuFileCloseHelp),
          nil, closeCurrentPage, layout.AccelCode{ 'w', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_CLOSE },
        { "exit", localizeText(menuFileQuit), localizeText(menuFileQuitHelp),
          nil, xit, layout.AccelCode{ 'q', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
          ENABLE_EXIT },
    }

    prtct := func( ) {
        setProtectedState( ! protectState )     // toggle state
        temporarilySetReadOnly( protectState )
    }

    xpl := func () {
        showExploreDialog( getBytesAtCaret( 0 ) )
    }

    menuResIds["undo"] = menuTextIds{ menuEditUndo, menuEditUndoHelp }
    menuResIds["redo"] = menuTextIds{ menuEditRedo, menuEditRedoHelp }
// "protect" is treated as a special case
    menuResIds["cut"] = menuTextIds{ menuEditCut, menuEditCutHelp }
    menuResIds["copy"] = menuTextIds{ menuEditCopy, menuEditCopyHelp }
    menuResIds["paste"] = menuTextIds{ menuEditPaste, menuEditPasteHelp }
    menuResIds["delete"] = menuTextIds{ menuEditDelete, menuEditDeleteHelp }
    menuResIds["selectAll"] = menuTextIds{ menuEditSelect, menuEditSelectHelp }
    menuResIds["explore"] = menuTextIds{ menuEditExplore, menuEditExploreHelp }
    menuResIds["preferences"] = menuTextIds{ menuEditPreferences, menuEditPreferencesHelp }

    var editMenuDef = []layout.MenuItemDef {
        { "undo", localizeText(menuEditUndo), localizeText(menuEditUndoHelp),
          nil, undoLast, layout.AccelCode{ 'z', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
          ENABLE_UNDO },
        { "redo", localizeText(menuEditRedo), localizeText(menuEditRedoHelp),
          nil, redoLast, layout.AccelCode{ 'z', gdk.CONTROL_MASK | gdk.SHIFT_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_REDO },
        separator,
        { "protect", localizeText(menuEditModify), localizeText(menuEditModifyHelp),
          nil, prtct, layout.AccelCode{ 0, 0, 0 }, ENABLE_PROTECT },
        separator,
        { "cut", localizeText(menuEditCut), localizeText(menuEditCutHelp),
          nil, cutSelection, layout.AccelCode{ 'x', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_CUT },
        { "copy", localizeText(menuEditCopy), localizeText(menuEditCopyHelp),
          nil, copySelection, layout.AccelCode{ 'c', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_COPY },
        { "paste", localizeText(menuEditPaste), localizeText(menuEditPasteHelp),
          nil, pasteClipboard, layout.AccelCode{ 'v', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_PASTE },
        { "delete", localizeText(menuEditDelete), localizeText(menuEditDeleteHelp),
          nil, deleteSelection, layout.AccelCode{ gdk.KEY_Delete, 0,
          gtk.ACCEL_VISIBLE }, ENABLE_DELETE },
        separator,
        { "selectAll", localizeText(menuEditSelect), localizeText(menuEditSelectHelp),
          nil, selectAll, layout.AccelCode{ 'a', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_SELECT_ALL },
        { "explore", localizeText(menuEditExplore), localizeText(menuEditExploreHelp),
          nil, xpl, layout.AccelCode{ 'e', gdk.CONTROL_MASK | gdk.MOD1_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_EXPLORE },
        separator,
        { "preferences", localizeText(menuEditPreferences),
          localizeText(menuEditPreferencesHelp), nil, showPreferencesDialog,
          layout.AccelCode{ 0, 0, 0 }, ENABLE_PREFERENCES  },
    }

    menuResIds["find"] = menuTextIds{ menuSearchFind, menuSearchFindHelp }
    menuResIds["replace"] = menuTextIds{ menuSearchReplace, menuSearchReplaceHelp }
    menuResIds["goto"] = menuTextIds{ menuSearchGoto, menuSearchGotoHelp }

    var searchMenuDef = []layout.MenuItemDef {
        { "find", localizeText(menuSearchFind), localizeText(menuSearchFindHelp),
          nil, searchDialog, layout.AccelCode{ 'f', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_FIND },
        { "replace", localizeText(menuSearchReplace),
          localizeText(menuSearchReplaceHelp), nil, replaceDialog,
          layout.AccelCode{ 'h', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
          ENABLE_REPLACE },
        { "goto", localizeText(menuSearchGoto), localizeText(menuSearchGotoHelp),
          nil, gotoDialog, layout.AccelCode{ 'j', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, ENABLE_GOTO },
    }

    menuResIds["contents"] = menuTextIds{ menuHelpContent, menuHelpContentHelp }
    menuResIds["about"] = menuTextIds{ menuHelpAbout, menuHelpAboutHelp }

    var helpMenuDef = []layout.MenuItemDef {
        { "contents", localizeText(menuHelpContent), localizeText(menuHelpContentHelp),
          nil, hexedHelp, layout.AccelCode{ 0, 0, 0 }, ENABLE_CONTENTS },
        { "about", localizeText(menuHelpAbout), localizeText(menuHelpAboutHelp),
          nil, aboutDialog, layout.AccelCode{ 0, 0, 0 }, ENABLE_ABOUT },
    }

    menuResIds["file"] = menuTextIds{ menuFile, -1 }
    menuResIds["edit"] = menuTextIds{ menuEdit, -1 }
    menuResIds["search"] = menuTextIds{ menuSearch, -1 }
    menuResIds["help"] = menuTextIds{ menuHelp, -1 }

    var menuDef = []layout.MenuItemDef {
        { "file", localizeText(menuFile), "", &fileMenuDef, nil,
          layout.AccelCode{ 0, 0, 0 }, true },
        { "edit", localizeText(menuEdit), "", &editMenuDef, nil,
          layout.AccelCode{ 0, 0, 0 }, true },
        { "search", localizeText(menuSearch), "", &searchMenuDef, nil,
          layout.AccelCode{ 0, 0, 0 }, true },
        { "help", localizeText(menuHelp), "", &helpMenuDef, nil,
          layout.AccelCode{ 0, 0, 0 }, true },
    }
    return len(menuResIds), &menuDef
}

const MAX_RECENT_FILES = 10
var fileHistory *layout.History

func initFileHistory( ) {
    var err error
    fileHistory, err = layout.NewHistory( MAX_RECENT_FILES )
    if err != nil {
        log.Fatalf( "initFileHistory: unable to create recent file history: %v",
                    err )
    }
    err = fileHistory.Set( getStringSlicePreference( RECENT_FILES ) )
    if err != nil {
        log.Fatalf( "initFileHistory: unable to set history content: %v", err )
    }
}

func updateRecentFiles( ) {
    layout.AttachHistoryMenuToMenuItem( "recent", fileHistory,
                                        localizeText(menuFileRecentHelp),
                                        func( path string ) {
                                            newPage( path, false )
                                        } )
}

func addFileToHistory( filePath string ) {
    v := fileHistory.Update( filePath )
    if len ( v ) != 0 {
        log.Printf("addFileToHistory: recent files %v\n", v )

        pref := preferences{}
        pref[RECENT_FILES] = v
        updatePreferences( pref )

        updateRecentFiles()
    }
}

type menuHint int
func (mh * menuHint)Show( hint string ) {
    showMenuHint( hint )
}

func (mh * menuHint)Remove( ) {
    removeMenuHint()
}

func (mh * menuHint)Clear( ) {
    clearMenuHint()
}

// refresh menus after language change
func refreshMenus( ) {
    refresh := func( name string ) {
        if name == "protect" {  // special case
            refreshProtectMenuItem( protectState )
            return
        }
        nameIds, ok := menuResIds[name]
        if ! ok {
            log.Fatalf("refresh menu %s: no such item\n", name)
        }
        var title, hint string
        if nameIds.title != -1 {
            title = localizeText( nameIds.title )
        }
        if nameIds.hint != -1 {
            hint = localizeText( nameIds.hint )
        }
        layout.SetMenuItemTexts( name, title, hint )
    }
    layout.ForEachMenuItemDo( refresh )
    if fileHistory.Depth() > 0 {
        updateRecentFiles()
    }
}

// initialize menu bar and file history
func initMenus( protect bool ) (*gtk.AccelGroup, *gtk.MenuBar) {
    protectState = protect
    nItems, menuTreeDef := getMenuDefs()
    accel, menubar := layout.InitMenuBar( nItems, menuTreeDef, (*menuHint)(nil) )
    initFileHistory()
    if fileHistory.Depth() > 0 {
        updateRecentFiles()
    }
    return accel, menubar
}


// A toolbar is just a simplified form of layout, with a horizontal or vertical
// box containing buttons, which are defined as buttons in a layout, using the
// same InputDef structure but limited to the button case.

const (
    OPEN_ICON_NAME = "document-open-symbolic"
    SAVE_ICON_NAME = "document-save-symbolic"
    SAVEAS_ICON_NAME = "document-save-as-symbolic"

    UNDO_ICON_NAME = "edit-undo-symbolic"
    REDO_ICON_NAME = "edit-redo-symbolic"

    PROTECTED_ICON_NAME = "changes-prevent-symbolic"
    WRITEABLE_ICON_NAME = "changes-allow-symbolic"

    CUT_ICON_NAME = "edit-cut-symbolic"
    COPY_ICON_NAME = "edit-copy-symbolic"
    PASTE_ICON_NAME = "edit-paste-symbolic"
    SELECT_ALL_ICON_NAME = "edit-select-all-symbolic"

    EXPLORE_ICON_NAME = "applications-utilities-symbolic"
//    EXPLORE_ICON_NAME = "find-location-symbolic"

    PREFERENCES_ICON_NAME = "preferences-system-symbolic"
//    PREFERENCES_ICON_NAME = "preferences-other-symbolic"

    FIND_ICON_NAME = "system-search-symbolic"
//    SEARCH_ICON_NAME = "edit-find-symbolic"
//    SEARCH_ICON_NAME = "folder-saved-search-symbolic"
    REPLACE_ICON_NAME = "edit-find-replace-symbolic"
)

var toolLayout *layout.Layout

func refreshProtectMenuItem( state bool ) {
    var title, icon, hint string
    if state {  // protected => item allows switching to non-protected
        title = localizeText(menuEditModify)
        icon = WRITEABLE_ICON_NAME
        hint = localizeText(menuEditModifyHelp)
    } else {    // non-protected => item allows switching to protected
        title = localizeText(menuEditFreeze)
        icon = PROTECTED_ICON_NAME
        hint = localizeText(menuEditFreezeHelp)
    }
    layout.SetMenuItemTexts( "protect", title, hint )
    toolLayout.SetButtonIcon( "protect", icon )
    toolLayout.SetItemTooltip( "protect", hint )
}

func action( name string, val interface{} ) bool {
    layout.GetMenuItemAction( name )()
    return true
}

func initToolbar( ) *gtk.Widget {

    if toolLayout != nil {
        panic("initToolbar: toolbar was already created\n")
    }

    openButCtl := layout.ButtonCtl{ ENABLE_OPEN, false, false }
    openLabel := layout.IconDef{ OPEN_ICON_NAME }
    // all too;bar items must ahve the same name as corresponding menuItems
    // in order to be able to invoke the same actions
    openb := layout.InputDef{ "open", 0, &openLabel,
                             localizeText(menuFileOpenHelp),
                             action, &openButCtl }

    saveButCtl := layout.ButtonCtl{ ENABLE_SAVE, false, false }
    saveLabel := layout.IconDef{ SAVE_ICON_NAME }
    saveb := layout.InputDef{ "save", 0, &saveLabel,
                             localizeText(menuFileSaveHelp),
                             action, &saveButCtl }

    sep := layout.ConstDef{ "", 0, "    ", "", nil }

    undoButCtl := layout.ButtonCtl{ ENABLE_UNDO, false, false }
    undoLabel := layout.IconDef{ UNDO_ICON_NAME }
    undob := layout.InputDef{ "undo", 0, &undoLabel,
                             localizeText(menuEditUndoHelp),
                             action, &undoButCtl }

    redoButCtl := layout.ButtonCtl{ ENABLE_REDO, false, false }
    redoLabel := layout.IconDef{ REDO_ICON_NAME }
    redob := layout.InputDef{ "redo", 0, &redoLabel,
                             localizeText(menuEditRedoHelp),
                             action, &redoButCtl }

    protectButCtl := layout.ButtonCtl{ ENABLE_PROTECT, false, false }
    protectLabel := layout.IconDef{ WRITEABLE_ICON_NAME }
    protectb := layout.InputDef{ "protect", 0, &protectLabel,
                                localizeText(menuEditModifyHelp),
                                action, &protectButCtl }

    cutButCtl := layout.ButtonCtl{ ENABLE_CUT, false, false }
    cutLabel := layout.IconDef{ CUT_ICON_NAME }
    cutb := layout.InputDef{ "cut", 0, &cutLabel,
                                localizeText(menuEditCutHelp),
                                action, &cutButCtl }
    copyButCtl := layout.ButtonCtl{ ENABLE_COPY, false, false }
    copyLabel := layout.IconDef{ COPY_ICON_NAME }
    copyb := layout.InputDef{ "copy", 0, &copyLabel,
                                localizeText(menuEditCopyHelp),
                                action, &copyButCtl }
    pasteButCtl := layout.ButtonCtl{ ENABLE_PASTE, false, false }
    pasteLabel := layout.IconDef{ PASTE_ICON_NAME }
    pasteb := layout.InputDef{ "paste", 0, &pasteLabel,
                                localizeText(menuEditPasteHelp),
                                action, &pasteButCtl }

    exploreButCtl := layout.ButtonCtl{ ENABLE_EXPLORE, false, false }
    exploreLabel := layout.IconDef{ EXPLORE_ICON_NAME }
    exploreb := layout.InputDef{ "explore", 0, &exploreLabel,
                                localizeText(menuEditExploreHelp),
                                action, &exploreButCtl }

    findButCtl := layout.ButtonCtl{ ENABLE_FIND, false, false }
    findLabel := layout.IconDef{ FIND_ICON_NAME }
    findb := layout.InputDef{ "find", 0, &findLabel,
                                localizeText(menuSearchFindHelp),
                                action, &findButCtl }
    replaceButCtl := layout.ButtonCtl{ ENABLE_REPLACE, false, false }
    replaceLabel := layout.IconDef{ REPLACE_ICON_NAME }
    replaceb := layout.InputDef{ "replace", 0, &replaceLabel,
                                localizeText(menuSearchReplaceHelp),
                                action, &replaceButCtl }

    lo, err := layout.NewLayout( &layout.BoxDef{ "", 5, 5, 0, "", false,
                                 layout.HORIZONTAL, []interface{}{
                                                         &openb, &saveb,
                                                         &sep,
                                                         &undob, &redob,
                                                         &sep,
                                                         &protectb,
                                                         &sep,
                                                         &cutb, &copyb, &pasteb,
                                                         &sep,
                                                         &exploreb,
                                                         &sep,
                                                         &findb, &replaceb,
                                                       } } )
    if err != nil {
        log.Fatalf( "initToolbar: unable to create the toolbar layout: %v", err )
    }
    toolLayout = lo
    return lo.GetRootWidget()
}

// The following functions update menus when switching between pages
func enablePreferences( state bool ) {
    layout.EnableMenuItem( "preferences", state )
}

func pageExists( state bool ) {
    layout.EnableMenuItem( "close", state )
    if state == false {
        fileExists( false ) // must be first to get correct protect state
        dataExists( false )
        selectionDataExists( false, false )
        undoRedoUpdate( false, false )
        modificationAllowed( false, false )
        explorePossible( false )
    }
}

func fileExists( state bool ) {
    if layout.IsMenuItemEnabled( "protect" ) {
        layout.EnableMenuItem( "save", state )
        toolLayout.SetButtonActive( "save", state )
    }
    layout.EnableMenuItem( "revert", state )
}

func explorePossible( state bool ) {
    layout.EnableMenuItem( "explore", state )
    toolLayout.SetButtonActive( "explore", state )
}

func dataExists( state bool ) {
    layout.EnableMenuItem( "selectAll", state )
    if layout.IsMenuItemEnabled( "protect" ) {
        layout.EnableMenuItem( "saveAs", state )
    }
    layout.EnableMenuItem( "find", state )
    toolLayout.SetButtonActive( "find", state )
    layout.EnableMenuItem( "replace", state )
    toolLayout.SetButtonActive( "replace", state )
    layout.EnableMenuItem( "goto", state )
}

func pasteDataExists( state bool ) {
    layout.EnableMenuItem( "paste", state &&
                    hasPageFocus() && isCurrentPageWritable() )
    toolLayout.SetButtonActive( "paste", state &&
                    hasPageFocus() && isCurrentPageWritable() )
}

func selectionDataExists( enableState bool, readOnly bool ) {
    layout.EnableMenuItem( "copy", enableState && hasPageFocus() )
    toolLayout.SetButtonActive( "copy", enableState && hasPageFocus() )
    layout.EnableMenuItem( "cut", enableState && ! readOnly && hasPageFocus() )
    toolLayout.SetButtonActive( "cut", enableState && ! readOnly && hasPageFocus() )
    layout.EnableMenuItem( "delete", enableState && ! readOnly &&hasPageFocus() )
}

func undoRedoUpdate( undo, redo bool ) {
    layout.EnableMenuItem( "undo", undo )
    toolLayout.SetButtonActive( "undo", undo )
    layout.EnableMenuItem( "redo", redo )
    toolLayout.SetButtonActive( "redo", redo )
}

func modificationAllowed( enableState, modificationState bool ) {
    setProtectedState( ! modificationState )
    layout.EnableMenuItem( "protect", enableState )
    toolLayout.SetButtonActive( "protect", enableState )
}
