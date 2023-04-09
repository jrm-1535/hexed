package layout

import (
    "log"

	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

// Accelerator key code
type AccelCode struct {
    Key     uint                // Ascii code
    Mod     gdk.ModifierType    // key modifier (Shift, Control, Alt)
    Flag    gtk.AccelFlags      // Visible or not in menu text
}

// Menus and MenuBars are containers of MenuItems
// A top level Menu or Menubar is just a single MenuItem pointing to a subMenu.
// They have "" as Name and they just have a slice of MenuItems as subMenu. A
// separator line in a Menu is a MenuItemDef without Title.
type MenuItemDef struct {
    Name        string          // item name (must be unique across menus)
    Title       string          // item text
    Hint        string          // hint text
    SubMenu     *[]MenuItemDef  // sub menu definition (nil if not a submenu
                                // or if the submenu is dynamically managed)
    Action      func( )         // action associated with the item, if any
    Accel       AccelCode       // shortcut key

    Enable      bool            // initial enable state
    Check       bool            // true if it is a check mark
    Initial     bool            // if check, initial check state
}

type menuItem struct {
//    gtkMenuItem *gtk.MenuItem
    gtkMenuItem gtk.IMenuItem
    title       string          // needed by popup menus
    action      func ( )
}

var (
    menuItems   map[string]*menuItem    // menu items mapped by item name
    help        HintControl
)

func locateMenuItemByName( name string ) (mi *menuItem) {
    mi = menuItems[ name ]
    if mi == nil {
        log.Panicf( "Unable to locate menu Item for action %s\n", name )
    }
    return
}

func (mi *menuItem) getGtkMenuItem( ) (gmi *gtk.MenuItem) {
    switch gtkMi := mi.gtkMenuItem.(type) {
    case *gtk.MenuItem:
        gmi = gtkMi
    case *gtk.CheckMenuItem:
        gmi = &(gtkMi.MenuItem)
    default:
        log.Panicf( "Menu Item type %T is not supported\n", gmi )
    }
    return
}

func getGtkMenuItemByName( name string ) (gmi *gtk.MenuItem) {
    mi := locateMenuItemByName( name )
    return mi.getGtkMenuItem( )
}

// enable or disable individual menu item. The argument aName  uniquely
// identifies the menu item across all menus.
func EnableMenuItem( name string, enable bool ) {
    gmi := getGtkMenuItemByName( name )
    gmi.SetSensitive( enable )
}

func IsMenuItemEnabled( name string ) bool {
    gmi := getGtkMenuItemByName( name )
    return gmi.GetSensitive()
}

func IsMenuItemChecked( name string ) bool {
    mi := locateMenuItemByName( name )
    if gcmi, ok := mi.gtkMenuItem.(*gtk.CheckMenuItem); ok {
        return gcmi.GetActive()
    }
    return false
}


func SetMenuItemTexts( name, title, hint string ) {
    mi := locateMenuItemByName( name )
    gmi := mi.getGtkMenuItem( )
    if title != "" {
        gmi.SetLabel( title )
        mi.title = title
    }
    if hint != "" {
        gmi.Connect( "enter-notify-event",
                       func ( gmi *gtk.MenuItem ) {
                                    help.Show( hint )
                       }  )
    }
}

func GetMenuItemAction( name string ) func() {
    mi := locateMenuItemByName( name )
    return mi.action
}

func ForEachMenuItemDo( f func(name string) ) {
    for name, _ := range menuItems {
        f( name )
    }
}

func AddPopupMenuItem( name string, title string, action func() ) {
    if _, ok := menuItems[name]; ok {
        log.Panicf( "addPopupMenuItem: item already exists for action name <%s>\n",
                    name )
    }
    menuItems[ name ] = &menuItem{ nil, title, action }
}

func DelPopupMenuItem( name string ) {
    delete( menuItems, name )
}

func PopupContextMenu( names []string, event *gdk.Event ) {
    popUpMenu, err := gtk.MenuNew()
	if err != nil {
		log.Fatal("Unable to create context menu:", err)
	}

    for _, name := range names {
        mi := locateMenuItemByName( name )

        menuItem, err := gtk.MenuItemNewWithLabel( mi.title )
        if err != nil {
            log.Fatal("Unable to create context menu item:", err)
        }
        menuItem.Show()
        menuItem.Connect( "activate", mi.action )
        popUpMenu.Append( menuItem )
    }
    popUpMenu.PopupAtPointer( event )
}

func newMenuItem( def *MenuItemDef, shortCuts *gtk.AccelGroup ) (mi *menuItem) {
    if def.Name == "" {
        log.Panicf("newMenuItem: got an item definition without name: %v\n", def)
    }
    var ( gmi *gtk.MenuItem; gcmi *gtk.CheckMenuItem; err error )
    if def.Check {
        if def.Title == "" {
            gcmi, err = gtk.CheckMenuItemNew( )
        } else {
            gcmi, err = gtk.CheckMenuItemNewWithMnemonic( def.Title )
        }
        gcmi.SetActive( def.Initial )
        gmi = &(gcmi.MenuItem)
    } else {
        if def.Title == "" {
            gmi, err = gtk.MenuItemNew( )
        } else {
            gmi, err = gtk.MenuItemNewWithMnemonic( def.Title )
        }
    }
    action := def.Action
    if action != nil {
        menuAction := func( ) {
            help.Clear()
            action( )
        }
        gmi.Connect( "activate", menuAction )
    }

    hint := def.Hint
    if hint != "" {
        gmi.Connect( "enter-notify-event",
                    func ( ) { help.Show( hint ) }  )
        gmi.Connect( "leave-notify-event", help.Remove )
    }

    if err != nil {
        log.Fatalf( "Unable to create a GTK MenuItem: %v\n", err )
    }

    if def.Accel.Key != 0 {
        gmi.AddAccelerator( "activate", shortCuts, def.Accel.Key,
                            def.Accel.Mod, def.Accel.Flag )
    }

    if def.SubMenu != nil {
        gmi.SetSubmenu( newGtkMenu( def.SubMenu, shortCuts ) )
    }

    gmi.SetSensitive( def.Enable )

    mi = new( menuItem )
    if gcmi != nil {
        mi.gtkMenuItem = gcmi
    } else {
        mi.gtkMenuItem = gmi
    }
    mi.title = def.Title
    mi.action = action

    menuItems[def.Name] = mi
    return
}

func newGtkMenu( itemDefs *[]MenuItemDef, shortCuts *gtk.AccelGroup ) *gtk.Menu {
    gmn, err := gtk.MenuNew( )
    if err != nil {
        log.Fatalf( "Unable to create a GTK Menu %v\n", err )
    }

    for _, smiDef := range *itemDefs {

        if smiDef.Name != "" {                    // a real item
            smi := newMenuItem( &smiDef, shortCuts )
            gmn.Add( smi.getGtkMenuItem() )
        } else {                                // a separation line
            gms, err := gtk.SeparatorMenuItemNew( )
            if err != nil {
                log.Fatalf( "Unable to create a GTK SeparatorMenuItem: %v\n", err )
            }
            gmn.Add( gms )
        }
    }
    return gmn
}

func newMenuBar( defRow *[]MenuItemDef ) (*gtk.AccelGroup, *gtk.MenuBar) {

    shortCuts, err := gtk.AccelGroupNew()
    if err != nil {
        log.Fatalf( "Unable to create a GTK accelerator group: %v\n", err )
    }

    menuBar, err := gtk.MenuBarNew()
    if err != nil {
        log.Fatalf( "Unable to create a GTK MenuBar: %v\n", err )
    }

    for _, defEl := range(*defRow) {
        m := newMenuItem( &defEl, shortCuts )
        menuBar.Append( m.gtkMenuItem )
    }

    return shortCuts, menuBar
}

// HintControl is the interface that controls the display of hints
type HintControl interface {
    Show( hint string )     // show a hint about a menu item
    Remove( )               // remove the last hint
    Clear( )                // clear all possible menu hints
}

// A MenuBar is a row of MenuItems, each with a title and a subMenu containing
// the list of actual menu items. It is assumed here that there is only one
// top-level menu bar.
func InitMenuBar( nItems int, defRow *[]MenuItemDef,
                  hint HintControl ) (*gtk.AccelGroup, *gtk.MenuBar) {
    menuItems = make( map[string]*menuItem, nItems )
    help = hint
    return newMenuBar( defRow )
}

func newGtkHistoryMenu( h *History, leading string,
                        action func( path string ) ) *gtk.Menu {
    recentFiles := h.Get()
    if len( recentFiles ) == 0 {
        return nil
    }

    gmn, err := gtk.MenuNew( )
    if err != nil {
        log.Fatalf( "Unable to create a GTK Menu %v\n", err )
    }

    for _, filePath := range recentFiles {
        if filePath != "" {
            gmi, err := gtk.MenuItemNewWithLabel( filePath)
            if err != nil {
                log.Fatalf( "Unable to create a GTK MenuItem: %v\n", err )
            }
            gmi.SetSensitive( true )

            path := filePath
            gmi.Connect( "activate", func( ) { action( path ) } )

            gmi.Connect( "enter-notify-event",
                        func ( gmi *gtk.MenuItem ) {
                          help.Show(leading + path) } )
            gmi.Connect( "leave-notify-event", help.Remove )
            gmi.Show()
            gmn.Add( gmi )
        }
    }
    return gmn
}

func AttachHistoryMenuToMenuItem( name string, h *History, leading string,
                                  action func( path string ) ) {
    gmi := getGtkMenuItemByName( name )
    if h.Depth() > 0 {
        gmi.SetSubmenu( newGtkHistoryMenu( h, leading, action ) )
        gmi.SetSensitive( true )
    } else {
        gmi.SetSensitive( true )
    }
}
