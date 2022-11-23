package main

import (
    "fmt"
    "os"
)

const (
    defFileSize = 1024
    defUndoSize =  128
    defSliceSize = 32       // must be 2 power
)

type command int
const (
    DELETE  = iota + 1
    INSERT
    REPLACE
)

func getCommandString( c command ) string {
    var cs string
    switch c {
    case DELETE:    cs = "DELETE"
    case INSERT:    cs = "INSERT"
    case REPLACE:   cs = "REPLACE"
    default:        cs = "Not a Command"
    }
    return cs
}
    
type operation struct {
    cmd                 command
    position,                           // where in data cmd takes place
    tag,                                // caller's command tag (returned at undo/redo)
    delLen,                             // length of replaced or deleted data
    insLen              int64           // length of replacing or inserted data
    backup, cut         int64           // where data is saved for undo
}

type storage struct {
    curData             []byte          // always current data
    newData             []byte          // inserted/replacing data (for redo)
    cutData             []byte          // deleted/replaced data (for redo)
    stack               []operation     // UNDO-REDO stack
    top                 int             // stack pointer
    notifyDataChange    func( )
    notifyLenChange     func( l int64 )
    notifyUndoRedo      func( u, r bool )
}

// some debug functions

func (s *storage) print(  ) {
    fmt.Printf("Storage %v\n", s.curData )
}

func (s *storage) printInternal( ) {
    fmt.Printf( "newData storage %v\n", s.newData )
    fmt.Printf( "cutData storage %v\n", s.cutData )
//    fmt.Printf( "pasteData storage %v\n", s.pasteData )
}

func (s *storage) printStack( ) {
    fmt.Printf( "Stack length %d, top=%d\n", len(s.stack), s.top )
    for i, op := range s.stack {
        if i == s.top {
            fmt.Printf( "  ------------------------------------\n" )
        }
        fmt.Printf( "  @%d cmd %v, pos %d, delLen %d, insLen %d, back %d, cut %d\n",
                    i, op.cmd, op.position, op.delLen, op.insLen, op.backup, op.cut )
    }
}

// return current storage length
func (s *storage) length() int64 {
    return int64(len(s.curData))
}

func (s *storage) setNotifyLenChange( f func( int64 ) ) {
    s.notifyLenChange = f
}

func (s *storage) setNotifyUndoRedoAble( f func( u, r bool ) ) {
    s.notifyUndoRedo = f
}

func (s *storage) setNotifyDataChange( f func() ) {
    s.notifyDataChange = f
}

// return current storage content
// Beware, it is only a shallow copy: DO NOT MODIFY the returned data
func (s *storage) getData( start, beyond int64 ) []byte {
    if start < 0 || start > beyond || beyond > int64(len(s.curData)) {
fmt.Printf("getData: start %d, beyond %d out of range [0-%d]\n", start, beyond, len(s.curData))
        return nil
    }
    return s.curData[start:beyond]
}

// default or specific storage initialization
func initStorage( path string ) ( *storage, error ) {

    var initialData []byte
    var err error

    if path == "" {
        initialData = make( []byte, 0, defFileSize )
    } else {
        initialData, err = os.ReadFile( path )
        if err != nil {
            return nil, err
        }
    }

    var result storage
    result.stack = make( []operation, 0, defUndoSize )
    result.curData = initialData
    return &result, nil
}

func (s *storage)reload( path string ) error {
    // check if file can still be opened
    _, err := os.Open( path )
    if err != nil {
        return err
    }
    // Assume everything will be OK
    s.curData = nil
    s.newData = nil
    s.cutData = nil
    s.stack = make( []operation, 0, defUndoSize )
    s.top = 0
    if s.curData, err = os.ReadFile( path ); err == nil {
        if s.notifyDataChange != nil {
            s.notifyDataChange()
        }
        if s.notifyUndoRedo != nil {
            s.notifyUndoRedo( s.top > 0, true )
        }
        if s.notifyLenChange != nil {
            s.notifyLenChange( int64(len(s.curData)) )
        }
    }
    return err
}

// The undo redo stack is a slice of previous commands.
// The oldest command that can be undone is at the bottom of the stack (0)
// The latest command that can be undone is immediately below the top of the
// stack (top-1)
//
// Undo just decrements top, takes the command at top, undo it (insert > delete,
// delete > insert saved and replace > opposite replace), but leaves the undone
// command in the slice (now at top+1). Undo works until top is 0.
//
// Redo just looks at the command at top, if it exists in the slice. It then
// re-executes the command and increments top, It works until top reaches the
// length of the slice.

// Undo the command just below the top of the undo/redo stack.
// Undo reverses DELETE by executing INSERT with the saved data; it reverses
// INSERT by executing DELETE at the position and with the length of the
// inserted data. REPLACE, which is a combination of DELETE and INSERT is
// reversed similarly. In all cases, undo does not push the command and does
// not save the restored data again, since the command slice, backup and cut
// storages are still valid: the top of the stack is just decremented.
func (s *storage) undo( ) (pos, tag int64, err error) {
    if s.top <= 0 {
        err = fmt.Errorf( "undo stack empty\n" )
        return
    }

    s.top --                // pop undo stack (and push into redo stack)

    if s.notifyUndoRedo != nil {
        s.notifyUndoRedo( s.top > 0, true )
    }

    op := s.stack[s.top]

    cmd := op.cmd
    pos = op.position
    tag = op.tag
    dl := op.delLen
    il := op.insLen

    switch cmd {
    default:
        panic("operation not implemented\n" )
    case DELETE:
        c := op.cut
        fmt.Printf( "UNDO DELETE @position %d, length %d\n", pos, dl )
        err = s.insertInCurData( pos, tag, s.cutData[c:c+dl], false )

    case INSERT:
        fmt.Printf( "UNDO INSERT @position %d, length %d\n", pos, il )
        err = s.cutInCurData( pos, tag, il, false )

    case REPLACE:
        fmt.Printf( "UNDO REPLACE @position %d, length (del %d, ins %d)\n", pos, dl, il )
        c := op.cut
        err = s.replaceInCurData( pos, tag, il, s.cutData[c:c+dl], false )
    }
    return
}

// redo the command et the top of the undo/redo stack.
// if top is the length of the slice, no command can be redone.
// Otherwise, redo takes the command at top. increments top, and re-executes
// the command as if it was requested again. Redo does not push the command
// and does not save the restored data again, since the command slice, backup
// and cut storages are still valid: the top of the stack is just incremented.
func (s *storage) redo( ) (pos, tag int64, err error) {
    //fmt.Printf( "REDO stack pointer %d\n", s.top )
    if s.top >= len( s.stack ) {
        err = fmt.Errorf( "redo stack empty\n" )
        return
    }
    op := s.stack[s.top]
    s.top ++                // pop redo stack (and push into undo stack)

    if s.notifyUndoRedo != nil {
        s.notifyUndoRedo( true, s.top < len( s.stack) )
    }

    cmd := op.cmd
    pos = op.position
    tag = op.tag
    dl := op.delLen
    il := op.insLen

    switch cmd {
    default:
        panic("operation not implemented\n" )
    case DELETE:
        fmt.Printf( "REDO DELETE @position %d, length %d\n", pos, dl )
        err = s.cutInCurData( pos, tag, dl, false )

    case INSERT:
        b := op.backup
        fmt.Printf( "REDO INSERT @position %d, length %d. backup %d\n", pos, il, b )
        err = s.insertInCurData( pos, tag, s.newData[b:b+il], false )

    case REPLACE:
        b := op.backup
        c := op.cut
        fmt.Printf( "REDO REPLACE @position %d, length (del %d, ins %d), cut %d, backup %d\n",
                    pos, dl, il, c, b )
        err = s.replaceInCurData( pos, tag, dl, s.newData[b:b+il], false )
    }
    return
}

func (s *storage) areUndoRedoPossible( ) (u, r bool) {
    if s.top < len( s.stack ) {
        r = true
    }
    if s.top > 0 {
        u = true
    }
    return
}

// push command {operation (DELETE, INSERT or REPLACE), position, current
// and new length} onto the undo/redo stack. Usually this means just appending
// a command to the stack, but if operations were previously undone, the top
// of the stack is not at the end of the slice anymore. In that case, all
// operations that could be redone until now are just removed from the slice
// and the backup (newData) and cut (cutData) storage are cleaned: it is now
// impossible to redo them, and previously saved data must be forgotten.

// because it might change newData and/or cutData length, push must be called
// before updating newData or cutData (see insertInCurData and cutInCurData).
func (s *storage) push( c command, p, t, cl, nl int64 ) {
    if s.top >= len( s.stack ) {
        s.stack = append( s.stack, operation{ c, p, t, cl, nl,
                                              int64(len(s.newData)),
                                              int64(len(s.cutData)) } )
        s.top ++
    } else { // previous commands go out of scope, clean up newData or cutData
        for i := len( s.stack)-1; i >= s.top; i-- {
            op := s.stack[i]
            fmt.Printf( "UNDO/REDO command @%d cmd=%s, pos=%d, delLen=%d insLen=%d lost\n",
                        s.top, getCommandString(op.cmd), op.position, op.delLen, op.insLen )

            switch op.cmd {
            case INSERT:
                if int64(len(s.newData)) <= op.backup {
                    panic( "push: newData inconsistent\n" );
                }
                s.newData = s.newData[:op.backup]
            case REPLACE:
                if int64(len(s.newData)) <= op.backup {
                    panic( "push: newData inconsistent\n" );
                }
                if int64(len(s.cutData)) <= op.cut {
                    panic( "push: cutData inconsistent\n" );
                }
                s.newData = s.newData[:op.backup]
                s.cutData = s.cutData[:op.cut]
            case DELETE:
                if int64(len(s.cutData)) <= op.cut {
                    panic( "push: cutData inconsistent\n" );
                }
                s.cutData = s.cutData[:op.cut]
            default:
                panic( "push: unkown command in stack\n" )
            }
        }
        fmt.Printf( "newData storage %v\n", s.newData )
        fmt.Printf( "cutData storage %v\n", s.cutData )
        s.stack[s.top] = operation{ c, p, t, cl, nl,
                                    int64(len(s.newData)),
                                    int64(len(s.cutData)) }
        s.top ++
        s.stack = s.stack[:s.top]
    }
    s.notifyUndoRedo( true, s.top < len( s.stack) )
}

// insert data bytes at position pos in main storage.
// if save is true, push the INSERT command onto the undo/redo stack and
//                  save the inserted data bytes into the backup storage.
func (s *storage) insertInCurData( pos, tag int64, data []byte, save bool ) error {
    cl := int64(len(s.curData))
    if cl < pos || pos < 0 {
        return fmt.Errorf("INSERT outside data boundaries\n")
    }
    il := int64(len(data))
    if save {
        s.push( INSERT, pos, tag, 0, il ) // no byte deleted, il inserted
        s.newData = append( s.newData, data... )
    }
    if cl == pos {   // just append bytes
        s.curData = append( s.curData, data... )
    } else {
        if pos + il > cl {
            //          +++++++++xx  data
            //          0         il
            // =========:::::::::  s.curData
            //         pos      cl

            // =========:::::::::xx
            //         pos     cl pos+il
            s.curData = append( s.curData, data[cl - pos:]... )

            // =========:::::::::xx:::::::::
            //         pos     cl        cl+il
            s.curData = append( s.curData, s.curData[pos:cl]... )

            //          +++++++++xx  data
            // =========+++++++++xx:::::::::
            //         pos     cl pos+il
            copy( s.curData[pos:], data[:cl-pos] )
        } else {
            //          ++++++  data
            //          0    il
            // =========::xxxxxx  s.curData
            //         pos     cl

            // =========::xxxxxxxxxxxx
            s.curData = append( s.curData, s.curData[cl-il:]...)

            //          ++++++  data
            //          0    il
            // =========::xxxxxxxxxxxx
            //         pos     cl

            // =========::xxxx::xxxxxx
            copy( s.curData[pos+il:cl], s.curData[pos:] )     

            // =========++++++::xxxxxx
            copy( s.curData[pos:], data )
        }
    }
    if il != int64(len( s.curData )) - cl {
        panic("INSERT: Wrong length after insertion\n")
    }
    if s.notifyLenChange != nil {
        s.notifyLenChange( s.length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

// interface used in insertPastedBytesAt and replaceWithPastedBytesAt
type byteGetter interface {
    get( ) byte
    size( ) int64
}

func (s *storage) insertClipboardAt( pos, tag int64, bg byteGetter ) error {
    cl := int64(len(s.curData))
    if cl < pos || pos < 0 {
        return fmt.Errorf("INSERT outside data boundaries\n")
    }
    il := bg.size()
    s.push( INSERT, pos, tag, 0, il ) // no byte deleted, il inserted
    nPos := int64(len(s.newData))
    for i:= int64(0); i < il; i++ {
        s.newData = append( s.newData, bg.get( ) )
    }
    // newData[nPos:nPos+il] now contains the data to insert
    if cl == pos {   // just append bytes
        s.curData = append( s.curData, s.newData[nPos:]... )
    } else {
        if pos + il > cl {
            //          +++++++++xx  data
            //        nPos     nPos+il
            // =========:::::::::  s.curData
            //         pos      cl

            // =========:::::::::xx
            //         pos     cl pos+il
            s.curData = append( s.curData, s.newData[nPos+(cl-pos):]... )

            // =========:::::::::xx:::::::::
            //         pos     cl        cl+il
            s.curData = append( s.curData, s.curData[pos:cl]... )

            //          +++++++++xx  data
            // =========+++++++++xx:::::::::
            //         pos     cl pos+il
            copy( s.curData[pos:], s.newData[nPos:nPos+(cl-pos)] )
        } else {
            //          ++++++  data
            //          0    il
            // =========::xxxxxx  s.curData
            //         pos     cl

            // =========::xxxxxxxxxxxx
            s.curData = append( s.curData, s.curData[cl-il:]...)

            //          ++++++  data
            //          0    il
            // =========::xxxxxxxxxxxx
            //         pos     cl

            // =========::xxxx::xxxxxx
            copy( s.curData[pos+il:cl], s.curData[pos:] )     

            // =========++++++::xxxxxx
            copy( s.curData[pos:], s.newData[nPos:] )
        }
    }
    if s.notifyLenChange != nil {
        s.notifyLenChange( s.length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

// insert a single byte at position pos
func (s *storage) insertByteAt( pos, tag int64, v byte ) error {
    return s.insertInCurData( pos, tag, []byte{ v }, true )
}

// insert multiple bytes at position pos
func (s *storage) insertBytesAt( pos, tag int64, v []byte ) error {
    return s.insertInCurData( pos, tag, v, true )
}

// delete or cut data starting at pos, for the given length.
// if save is true, push the DELETE command onto the undo/redo stack and
//                  save the deleted data into the cut storage.
func (s *storage) cutInCurData( pos, tag, length int64, save bool ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || length < 0 || (pos + length) > curLen {
        return fmt.Errorf("DELETE outside data boundaries\n")
    }

    if save {
        s.push( DELETE, pos, tag, length, 0 )    // length bytes removed, 0 inserted
        s.cutData = append( s.cutData, s.curData[pos:pos+length]... )
    }
    if pos == curLen-length {
        s.curData = s.curData[:pos]
    } else {
        copy( s.curData[pos:], s.curData[pos+length:] )
        s.curData = s.curData[:curLen-length]
    }
    if s.notifyLenChange != nil {
        s.notifyLenChange( s.length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

// delete one byte at position pos.
func (s *storage) deleteByteAt( pos, tag int64 ) error {
    return s.cutInCurData( pos, tag, 1, true )
}

// delete n bytes at position pos.
func (s *storage) deleteBytesAt( pos, tag, n int64 ) error {
    return s.cutInCurData( pos, tag, n, true )
}

// save n bytes starting at position pos into the paste buffer
// this is used for cut and copy
func (s *storage) saveClipboardData( pos, n int64 ) error {
    if pos < 0 || pos + n > int64(len(s.curData)) {
        return fmt.Errorf("CUT/COPY outside data boundaries\n")
    }
//    s.pasteData = s.pasteData[0:0]
//    s.pasteData = append( s.pasteData, s.curData[pos:pos+n]... )
    setClipboardData( s.curData[pos:pos+n] )
    return nil
}

// cut n bytes starting at position pos
func (s *storage) cutBytesAt( pos, tag, n int64 ) error {
    if err := s.saveClipboardData( pos, n ); err != nil {
        return err
    }
    return s.cutInCurData( pos, tag, n, true )
}

// copy n bytes starting at position pos
func (s *storage) copyBytesAt( pos, n int64 ) error {
    return s.saveClipboardData( pos, n )
}

// replace existing data[pos:pos+dl] in the main buffer with new data.
// if save is true, push REPLACE command onto the undo/redo stack,
//                  save the existing data in the cut storage and
//                  save the new data into the backup storage.
func (s *storage) replaceInCurData( pos, tag, dl int64, data []byte, save bool ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || dl < 0 || (pos + dl) > curLen {
        return fmt.Errorf("REPLACE outside data boundaries\n")
    }
    il := int64(len(data))
    if save {
        s.push( REPLACE, pos, tag, dl, il ) // dl removed, il inserted
        s.cutData = append( s.cutData, s.curData[pos:pos+dl]... )
        s.newData = append( s.newData, data... )
    }
    if il <= dl {
        // replace first il bytes in [pos:pos+il]
        copy( s.curData[pos:], data )
        // move remaining bytes beyond pos+dl to pos+il        
        copy( s.curData[pos+il:], s.curData[pos+dl:] )
        // shrink the slice to pos+(il-dl)
        s.curData = s.curData[:curLen-dl+il]
    } else {
        // ==--------=========::::  s.curData
        //   ^       ^            ^
        //  pos     +dl        curLen
        //   xxxxxxxx++++           data
        //   ^           ^
        //   0          il

        // ==xxxxxxxx=========::::  s.curData
        copy( s.curData[pos:pos+dl], data )

        // ==xxxxxxxx=========::::  s.curData
        // ==xxxxxxxx=========::::::::  s.curData
        s.curData = append( s.curData, s.curData[curLen-(il-dl):]... )

        // ==xxxxxxxx=========::::::::  s.curData
        // ==xxxxxxxx=============::::  s.curData
        copy( s.curData[pos+il:curLen], s.curData[pos+dl:] )

        // ==xxxxxxxx=============::::  s.curData
        //   xxxxxxxx++++
        // ==xxxxxxxx++++=========::::
        copy( s.curData[pos+dl:], data[dl:] )
    }
    if curLen != int64(len(s.curData)) && s.notifyLenChange != nil {
        s.notifyLenChange( s.length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

func (s *storage) replaceWithClipboardAt( pos, tag, dl int64, bg byteGetter ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || dl < 0 || (pos + dl) > curLen {
        return fmt.Errorf("REPLACE outside data boundaries\n")
    }
    il := bg.size()
    s.push( REPLACE, pos, tag, dl, il ) // dl removed, il inserted
    s.cutData = append( s.cutData, s.curData[pos:pos+dl]... )
    nPos := int64(len(s.newData))
    for i := int64(0); i < il; i++ {
        s.newData = append( s.newData, bg.get( ) )
    }

    if il <= dl {
        // replace first il bytes in [pos:pos+il]
        copy( s.curData[pos:], s.newData[nPos:] )
        // move remaining bytes beyond pos+dl to pos+il        
        copy( s.curData[pos+il:], s.curData[pos+dl:] )
        // shrink the slice to pos+(il-dl)
        s.curData = s.curData[:curLen-dl+il]
    } else {
        // ==--------=========::::  s.curData
        //   ^       ^            ^
        //  pos     +dl        curLen
        //   xxxxxxxx++++           s.newData
        //   ^           ^
        //  nPos        +il

        // ==xxxxxxxx=========::::  s.curData
        copy( s.curData[pos:pos+dl], s.newData[nPos:] )

        // ==xxxxxxxx=========::::  s.curData
        // ==xxxxxxxx=========::::::::  s.curData
        s.curData = append( s.curData, s.curData[curLen-(il-dl):]... )

        // ==xxxxxxxx=========::::::::  s.curData
        // ==xxxxxxxx=============::::  s.curData
        copy( s.curData[pos+il:curLen], s.curData[pos+dl:] )

        // ==xxxxxxxx=============::::  s.curData
        //   xxxxxxxx++++
        // ==xxxxxxxx++++=========::::
        copy( s.curData[pos+dl:], s.newData[nPos+dl:] )
    }
    if curLen != int64(len(s.curData)) && s.notifyLenChange != nil {
        s.notifyLenChange( s.length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

// replace one byte at position pos
func (s *storage) replaceByteAt( pos, tag int64, v byte ) error {
    return s.replaceInCurData( pos, tag, 1, []byte{v}, true )
}

// replace multiple bytes at position pos
func (s *storage) replaceBytesAt( pos, tag, l int64, v []byte ) error {
    return s.replaceInCurData( pos, tag, l, v, true )
}

// special operations to avoid creating extra slices when deleting bytes

// replaceByteAtAndEraseFollowingBytes replaces the first byte at pos with the
// given byte v, and erases the following n bytes . This is used in case of
// replace editing mode when a selection is replaced by a single byte. This is
// treated as a replaceBytesAt in order to allow undo/redo as usual, but the
// caller does not have to create an extra, potentially large, slice full of
// zeros.
func (s *storage) replaceByteAtAndEraseFollowingBytes( pos, tag int64, v byte,
                                                       n int64 ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || n < 0 || (pos + n + 1) > curLen {
        return fmt.Errorf("REPLACE outside data boundaries\n")
    }
    il := int64(n+1)

    s.push( REPLACE, pos, tag, il, il ) // il removed, il inserted
    s.cutData = append( s.cutData, s.curData[pos:pos+il]... )
    nPos := int64(len(s.newData))
    s.newData = append( s.newData, v )

    for i := int64(0); i < n; i++ {
        s.newData = append( s.newData, 0 )
    }
    copy( s.curData[pos:pos+il], s.newData[nPos:nPos+il] )
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}
