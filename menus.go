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

func refreshProtectMenuItem( state bool ) {
    var title, hint string
    if state {  // protected => item allows switching to non-protected
        title = localizeText(menuEditModify)
        hint = localizeText(menuEditModifyHelp)
    } else {    // non-protected => item allows switching to protected
        title = localizeText(menuEditFreeze)
        hint = localizeText(menuEditFreezeHelp)
    }
    layout.SetMenuItemTexts( "protect", title, hint )
}

func setProtectedState( state bool ) {
    protectState = state
    refreshProtectMenuItem( state )
}

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
          true },
        { "open", localizeText(menuFileOpen), localizeText(menuFileOpenHelp),
          nil, op, layout.AccelCode{ 'o', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
          true },
        separator,
        { "save", localizeText(menuFileSave), localizeText(menuFileSaveHelp),
          nil, saveCurrentPage, layout.AccelCode{ 's', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
        { "saveAs", localizeText(menuFileSaveAs), localizeText(menuFileSaveAsHelp),
          nil, saveCurrentPageAs, layout.AccelCode{ 's', gdk.CONTROL_MASK |
          gdk.SHIFT_MASK, gtk.ACCEL_VISIBLE }, false },
        { "revert", localizeText(menuFileRevert), localizeText(menuFileRevertHelp),
          nil, revertCurrentPage, noAccel, false },
        separator,
        { "recent", localizeText(menuFileRecent), localizeText(menuFileRecentHelp),
          nil, nil, noAccel, false },
        { "close", localizeText(menuFileClose), localizeText(menuFileCloseHelp),
          nil, closeCurrentPage, layout.AccelCode{ 'w', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
        { "exit", localizeText(menuFileQuit), localizeText(menuFileQuitHelp),
          nil, xit, layout.AccelCode{ 'q', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE },
          true },
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
          false },
        { "redo", localizeText(menuEditRedo), localizeText(menuEditRedoHelp),
          nil, redoLast, layout.AccelCode{ 'z', gdk.CONTROL_MASK | gdk.SHIFT_MASK,
          gtk.ACCEL_VISIBLE }, false },
        separator,
        { "protect", localizeText(menuEditModify), localizeText(menuEditModifyHelp),
          nil, prtct, layout.AccelCode{ 0, 0, 0 }, false },
        separator,
        { "cut", localizeText(menuEditCut), localizeText(menuEditCutHelp),
          nil, cutSelection, layout.AccelCode{ 'x', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
        { "copy", localizeText(menuEditCopy), localizeText(menuEditCopyHelp),
          nil, copySelection, layout.AccelCode{ 'c', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
        { "paste", localizeText(menuEditPaste), localizeText(menuEditPasteHelp),
          nil, pasteClipboard, layout.AccelCode{ 'v', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
        { "delete", localizeText(menuEditDelete), localizeText(menuEditDeleteHelp),
          nil, deleteSelection, layout.AccelCode{ gdk.KEY_Delete, 0,
          gtk.ACCEL_VISIBLE }, false },
        separator,
        { "selectAll", localizeText(menuEditSelect), localizeText(menuEditSelectHelp),
          nil, selectAll, layout.AccelCode{ 'a', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
        { "explore", localizeText(menuEditExplore), localizeText(menuEditExploreHelp),
          nil, xpl, layout.AccelCode{ 'e', gdk.CONTROL_MASK | gdk.MOD1_MASK,
          gtk.ACCEL_VISIBLE }, false },
        separator,
        { "preferences", localizeText(menuEditPreferences),
          localizeText(menuEditPreferencesHelp), nil, showPreferencesDialog,
          layout.AccelCode{ 0, 0, 0 }, true  },
    }

    menuResIds["find"] = menuTextIds{ menuSearchFind, menuSearchFindHelp }
    menuResIds["replace"] = menuTextIds{ menuSearchReplace, menuSearchReplaceHelp }
    menuResIds["goto"] = menuTextIds{ menuSearchGoto, menuSearchGotoHelp }

    var searchMenuDef = []layout.MenuItemDef {
        { "find", localizeText(menuSearchFind), localizeText(menuSearchFindHelp),
          nil, searchDialog, layout.AccelCode{ 'f', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
        { "replace", localizeText(menuSearchReplace),
          localizeText(menuSearchReplaceHelp), nil, replaceDialog,
          layout.AccelCode{ 'h', gdk.CONTROL_MASK, gtk.ACCEL_VISIBLE }, false },
        { "goto", localizeText(menuSearchGoto), localizeText(menuSearchGotoHelp),
          nil, gotoDialog, layout.AccelCode{ 'j', gdk.CONTROL_MASK,
          gtk.ACCEL_VISIBLE }, false },
    }

    menuResIds["contents"] = menuTextIds{ menuHelpContent, menuHelpContentHelp }
    menuResIds["about"] = menuTextIds{ menuHelpAbout, menuHelpAboutHelp }

    var helpMenuDef = []layout.MenuItemDef {
        { "contents", localizeText(menuHelpContent), localizeText(menuHelpContentHelp),
          nil, hexedHelp, layout.AccelCode{ 0, 0, 0 }, true },
        { "about", localizeText(menuHelpAbout), localizeText(menuHelpAboutHelp),
          nil, aboutDialog, layout.AccelCode{ 0, 0, 0 }, true },
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
        update( pref )

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

// The following functions update menus when switching between pages
func enablePreferences( state bool ) {
    layout.EnableMenuItem( "preferences", state )
}

func pageExists( state bool ) {
    layout.EnableMenuItem( "close", state )
    if state == false {
        dataExists( false )
        selectionDataExists( false, false )
        undoRedoUpdate( false, false )
        modificationAllowed( false, false )
        fileExists( false )
    }
}

func fileExists( state bool ) {
    if layout.IsMenuItemEnabled( "protect" ) {
        layout.EnableMenuItem( "save", state )
    }
    layout.EnableMenuItem( "revert", state )
}

func explorePossible( state bool ) {
    layout.EnableMenuItem( "explore", state )
}

func dataExists( state bool ) {
    layout.EnableMenuItem( "selectAll", state )
    if layout.IsMenuItemEnabled( "protect" ) {
        layout.EnableMenuItem( "saveAs", state )
    }
    layout.EnableMenuItem( "find", state )
    layout.EnableMenuItem( "replace", state )
    layout.EnableMenuItem( "goto", state )
}

func pasteDataExists( state bool ) {
    layout.EnableMenuItem( "paste", state &&
                    hasPageFocus() && isCurrentPageWritable() )
}

func selectionDataExists( enableState bool, readOnly bool ) {
    layout.EnableMenuItem( "copy", enableState && hasPageFocus() )
    layout.EnableMenuItem( "cut", enableState && ! readOnly && hasPageFocus() )
    layout.EnableMenuItem( "delete", enableState && ! readOnly &&hasPageFocus() )
}

func undoRedoUpdate( undo, redo bool ) {
    layout.EnableMenuItem( "undo", undo )
    layout.EnableMenuItem( "redo", redo )
}

func modificationAllowed( enableState, modificationState bool ) {
    setProtectedState( ! modificationState )
    layout.EnableMenuItem( "protect", enableState )
}
