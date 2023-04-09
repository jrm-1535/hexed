package layout

import (
    "fmt"
    "log"
    "strconv"

	"github.com/gotk3/gotk3/gtk"
)

// GetRootWidget returns the layout root widget that was created by newLayout
func (lo *Layout)GetRootWidget() *gtk.Widget {
    return lo.root.ToWidget()
}

// GetItemNames returns the list of all item names included in the Layout. That
// list is not sorted alphabetically.
func (lo *Layout)GetItemNames() (names []string) {
    names = make( []string, len(lo.access) )
    i := 0
    for name := range lo.access {
        names[i] = name
        i++
    }
    return
}

// SetVisible makes the Layout root widget visible or invisible depending on
// the argument visible.
func (lo *Layout)SetVisible( visible bool ) {
    if lo.root != nil {
        lo.root.ToWidget().SetVisible( visible )
    }
}

// SetRowVisible makes a row in the grid specified by the argument index
// visible or invisible depending on the argument visible.
func (dg *DataGrid)SetRowVisible( index int, visible bool ) error {
    if index < 0 || index >= len(dg.rowItems) {
        return fmt.Errorf( "setRowVisible: row index %d out of range [0:%d[\n",
                           index, len(dg.rowItems) )
    }
    rowItems := dg.rowItems[index]
    for _, colItemRef := range rowItems {
        wdg := colItemRef.item.(gtk.IWidget).ToWidget()
        wdg.SetVisible( visible )
    }
    return nil
}

// SetColVisible makes a column in the grid specified by the argument index
// visible or invisible depending on the argument visible.
func (dg *DataGrid)SetColVisible( index int, visible bool ) error {
    if index < 0 || index >= len(dg.colItems) {
        return fmt.Errorf( "setRowVisible: col index %d out of range [0:%d[\n",
                           index, len(dg.colItems) )
    }
    colItems := dg.colItems[index]
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

// GetButtonLabel returns the label of the button identified by its definition
// name, or an error if the given name does not match a button.
func (lo *Layout)GetButtonLabel( name string ) (string, error) {
    _, button, err := lo.getButton( name )
    if err != nil {
        return "", fmt.Errorf("GetButtonLabel: %v\n", err )
    }
    return button.GetLabel()
}

// SetButtonLabel sets the label of the button identified by its definition
// name, or returns an error if the given name does not match a button.
func (lo *Layout)SetButtonLabel( name string, label string ) error {
    format, button, err := lo.getButton( name )
    if err != nil {
        return fmt.Errorf("SetButtonLabel: %v\n", err )
    }
    markup := makeMarkup( format, label )
    button.SetLabel( markup )
    return nil
}

// GetButtonActive returns the sensitivity status of the button identified by
// its definition name, or an error if the given name does not match a button.
func (lo *Layout) GetButtonActive( name string ) (bool, error) {
    _, button, err := lo.getButton( name )
    if err != nil {
        return false, fmt.Errorf("GetButtonActive: %v\n", err )
    }
    return button.GetSensitive( ), nil
}

// SetButtonActive sets the sensitivity status of the button identified by its
// definition name, or returns an error if the given name does not match a
// button.
func (lo *Layout) SetButtonActive( name string, state bool ) error {
    _, button, err := lo.getButton( name )
    if err != nil {
        return fmt.Errorf("SetButtonActive: %v\n", err )
    }
    button.SetSensitive( state )
    return nil
}

func (lo *Layout)SetButtonIcon( name string, iconName string ) error {
    format, button, err := lo.getButton( name )
    if err != nil {
        return fmt.Errorf("SetButtonIcon: %v\n", err )
    }
    if format != nil {
        return fmt.Errorf("SetButtonIcon: button %s has a label not an icon\n", name)
    }
    icon, err := gtk.ImageNewFromIconName( iconName, gtk.ICON_SIZE_BUTTON )
    if err != nil {
        return fmt.Errorf("SetButtonIcon: could not create icon image: %v", err)
    }
    button.SetImage( icon )
    return nil
}

// GetItemTooltip returns the tooltip text associated with the item identified
// by its definition name, or an error if the given name does not match an item
// that can have a tooltip.
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
    return "", fmt.Errorf( "GetItemTooltip: item %s cannot have tooltip\n", name )
}

// SetItemTooltip sets the tooltip text associated with the item identified by
// its definition name, or returns an error if the given name does not match an
// item that can have a tooltip.
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
    return fmt.Errorf( "SetItemTooltip: item %s cannot have tooltip\n", name )
}

// GetItemValue returns the current value associated with the given item name
// Depending on item type, returned values can be:
//   - bool for constant or input bool and for a toogle button
//   - int64 for constant or input int
//   - string for constant or input string
// Error is returned if the given name does not match any known item or if the
// value type does not match the expected item value. Since press button have
// no value, getting its value returns an error.
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

// SetItemValue sets the value associated with the given item name
// Depending on item type, the passed value can be:
//   - bool for constant or input bool and for a toogle button
//   - int64 for constant or input int
//   - string for constant or input string
// Error is returned if the given name does not match any known item or if the
// value type does not match the expected item value. Since press button have
// no value, setting its value returns an error.
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

    case *gtk.SpinButton:
        if v, ok := value.(int); ok {
            item.SetValue( float64(v) )
            return nil
        }
        return fmt.Errorf("setItemValue: wrong value type %T for spin button item %s\n",
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

// SetEntryFocus sets focus on the text entry associated with the given item
// name. It returns an error if the item does not exist or if it has no text
// entry.
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

// SetEntrySelection selects text within the text entry associated with the
// given item name. It returns an error if the item does not exist or if it
// has no text entry.
func (lo *Layout) SetEntrySelection( name string, start, beyond int ) error {
    entry, err := lo.getItemEntry( name )
    if err != nil {
        return fmt.Errorf("setEntrySelection: %v", err )
    }
    entry.SelectRegion( start, beyond )
    return nil
}

// SetEntryCursor sets the cursor within the text entry associated with the
// given item name. It returns an error if the item does not exist or if it
// has no text entry.
func (lo *Layout) SetEntryCursor( name string, position int ) error {
    entry, err := lo.getItemEntry( name )
    if err != nil {
        return fmt.Errorf("setEntrySelection: %v", err )
    }
    entry.SetPosition( position )
    return nil
}

// SetItemChoices redefines the list of choices associated with the given item
// name. It returns an error if the item does not exist or if it does not
// support multiple choices.
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

// History is an opaque type that may be used as choices in a drop-down menu
// item. It has a maximum depth and is managed in a last recently used manner.
// It is commonly used to remember frequent previous choices.
type History struct {
    store           []string                // previous history entries
}

// NewHistory creates and returns a new history with a given depth. It only
// returns an error if the depth is negative or zero.
func NewHistory( maxDepth int ) (h *History, err error) {
    if maxDepth < 1 {
        err = fmt.Errorf("NewHistory: max depth %d is too small\n", maxDepth)
        return
    }
    h = new( History )
    h.store = make( []string, 0, maxDepth )
    return 
}

func (h *History) Set( content []string ) error {
    if len( content ) > cap(h.store) {
        return fmt.Errorf( "Set history: capacity overflow (%d in %d)\n",
                            len( content ), cap(h.store) )
    }
    h.store = h.store[0:0]
    h.store = append( h.store, content... )
    return nil
}

// Get returns the current history as a slice of strings.
func (h *History) Get( ) []string {
    return h.store[:]
}

// Depth returns the current history size.
func (h *History) Depth( ) int {
    return len(h.store)
}

// Update enters a new text entry in history if it does not exist yet. It
// returns the modified history slices, so that they can be set as choices.
// Update looks up existing entries for the new text to enter in history.
// if an existing history entry is found that matches the new text, that entry
// is just moved up to the first (most recent) entry and the history depth is
// left unchanged. The reordered history is returned. If that entry was already
// the first entry, the whole history is not modified and an empty history is
// returned to indicate that choices do not need to be updated.
func (h *History) Update( text string ) []string {
    l := len(h.store)
    if l > 0 {                          // history is not empty
        for i := 0; i < l; i++ {            // check if already in history
            if text == h.store[i] {
                if i == 0 {                 // same content as before, 
                    return h.store[0:0]     // no update needed
                }                           
                entry := h.store[i]         // move store[i] -> store[0]
                copy ( h.store[1:i+1], h.store[0:i] )
                h.store[0] = entry
                return h.store[0:l]         // return reordered content
            }
        }
    }
    if l < cap(h.store) {               // not found in history, make room
        l++                             // if possible, otherwise drop oldest
        h.store = h.store[0:l]
    }
    copy( h.store[1:l], h.store[0:] )
    h.store[0] = text                   // store new entry
    return h.store[0:l]                 // return updated content
}

