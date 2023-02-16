package edit

import (
    "fmt"
    "log"
    "os"
)

const (
    defFileSize = 1024
    defUndoSize =  128
    defSliceSize = 32       // must be 2 power
)

type singlePosOp struct {
    tag,                                // caller's command tag (returned at undo/redo)
    delLen,                             // length of replaced or deleted data
    insLen,                             // length of replacing or inserted data
    backup, cut,                        // where data is saved for undo
    position            int64           // where in data operation takes place
}

type multiPosOp struct {
    tag,                                // caller's command tag (returned at undo/redo)
    delLen,                             // length of replaced or deleted data
    insLen,                             // length of replacing or inserted data
    backup, cut         int64           // where data is saved for undo
    positions           []int64         // where in data operations take place
}

// clipboard interface, used in cutBytesAt, copyBytesAt, insertClipboardAt
// and replaceWithClipboardAt
type Clipboard interface{
    Set( data []byte )
    Get( ) byte
    Size( ) int64
}

type Storage struct {
    curData             []byte          // always current data
    newData             []byte          // inserted/replacing data (for redo)
    cutData             []byte          // deleted/replaced data (for redo)

    stack               []interface{}   // UNDO-REDO stack
    clip                Clipboard

    top                 int             // stack pointer
    notifyDataChange    func( )
    notifyLenChange     func( l int64 )
    notifyUndoRedo      func( u, r bool )

    lost                bool            // true if full undo history is unknown
}

// some debug functions

func (s *Storage) Print(  ) {
    fmt.Printf("Storage %v\n", s.curData )
}

func (s *Storage) PrintInternal( ) {
    fmt.Printf( "newData storage %v\n", s.newData )
    fmt.Printf( "cutData storage %v\n", s.cutData )
}

func (s *Storage) PrintStack( ) {
    fmt.Printf( "Stack length %d, top=%d\n", len(s.stack), s.top )
    for i, op := range s.stack {
        if i == s.top {
            fmt.Printf( "  ------------------------------------\n" )
        }

        switch op := op.(type) {
        case singlePosOp:
            fmt.Printf( "  @%d tag %d, delLen %d, insLen %d, back %d, cut %d, nPos %d pos[0] %d\n",
                        i, op.tag, op.delLen, op.insLen, op.backup, op.cut,
                        1, op.position )
        case multiPosOp:
            fmt.Printf( "  @%d tag %d, delLen %d, insLen %d, back %d, cut %d, nPos %d pos[0] %d\n",
                        i, op.tag, op.delLen, op.insLen, op.backup, op.cut,
                        len(op.positions), op.positions[0] )
        }
    }
}

// return current storage length
func (s *Storage) Length() int64 {
    return int64(len(s.curData))
}

// return whether storage is presumed modified since creation/load
func (s *Storage)IsDirty( ) bool {
    if s.lost {
        return true
    }
    return s.top > 0
}

func (s *Storage) SetNotifyLenChange( f func( int64 ) ) {
    s.notifyLenChange = f
}

func (s *Storage) SetNotifyUndoRedoAble( f func( u, r bool ) ) {
    s.notifyUndoRedo = f
}

func (s *Storage) SetNotifyDataChange( f func() ) {
    s.notifyDataChange = f
}

// return current storage content
// Beware, it is only a shallow copy: DO NOT MODIFY the returned data
func (s *Storage) GetData( start, beyond int64 ) []byte {
    if start < 0 || start > beyond || beyond > int64(len(s.curData)) {
        return nil
    }
    return s.curData[start:beyond]
}

// default or specific storage initialization
func InitStorage( path string, clip Clipboard ) ( *Storage, error ) {

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

    var result Storage
    result.stack = make( []interface{}, 0, defUndoSize )
    result.curData = initialData
    result.clip = clip
    return &result, nil
}

func (s *Storage)Reload( path string ) error {
    // check if file can still be opened
    _, err := os.Open( path )
    if err != nil {
        return err
    }
    // Assume everything will be OK
    s.curData = nil
    s.newData = nil
    s.cutData = nil
    s.stack = make( []interface{}, 0, defUndoSize )
    s.top = 0
    s.lost = false
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
func (s *Storage) Undo( ) (pos, tag int64, err error) {
    if s.top <= 0 {
        err = fmt.Errorf( "undo stack empty\n" )
        return
    }

    s.top --                // pop undo stack (and push into redo stack)

    if s.notifyUndoRedo != nil {
        s.notifyUndoRedo( s.top > 0, true )
    }

    op := s.stack[s.top]
    switch op := op.(type) {
    case singlePosOp:
        tag = op.tag
        pos = op.position
        c := op.cut
        s.replaceInCurDataNotify( pos, tag, op.insLen,
                                  s.cutData[c:c+op.delLen] )
    case multiPosOp:
        tag = op.tag
        pos = op.positions[0]
        c := op.cut
//fmt.Printf( "undo multiPosOp: tag=%d insLen=%d delLen=%d nPos=%d pos=%v\n",
//            tag, op.insLen, op.delLen, len(op.positions), op.positions )
//fmt.Printf("       backup=%d %v cut=%d %v\n", op.backup, s.newData, op.cut, s.cutData )
        s.replaceInCurDataAtMultipleLocations( op.positions, tag, op.insLen,
                                               s.cutData[c:c+op.delLen], false )
    }
    return
}

// redo the command et the top of the undo/redo stack.
// if top is the length of the slice, no command can be redone.
// Otherwise, redo takes the command at top. increments top, and re-executes
// the command as if it was requested again. Redo does not push the command
// and does not save the restored data again, since the command slice, backup
// and cut storages are still valid: the top of the stack is just incremented.
func (s *Storage) Redo( ) (pos, tag int64, err error) {

    if s.top >= len( s.stack ) {
        err = fmt.Errorf( "redo stack empty\n" )
        return
    }
    op := s.stack[s.top]
    s.top ++                // pop redo stack (and push into undo stack)

    if s.notifyUndoRedo != nil {
        s.notifyUndoRedo( true, s.top < len( s.stack) )
    }

    switch op := op.(type) {
    case singlePosOp:
        tag = op.tag
        pos = op.position
        b := op.backup
        s.replaceInCurDataNotify( op.position, op.tag, op.delLen,
                                  s.newData[b:b+op.insLen] )
    case multiPosOp:
        tag = op.tag
        pos = op.positions[0]
        b := op.backup
        s.replaceInCurDataAtMultipleLocations( op.positions, op.tag, op.delLen,
                                               s.newData[b:b+op.insLen], true )
    }
    return
}

func (s *Storage) AreUndoRedoPossible( ) (u, r bool) {
    if s.top < len( s.stack ) {
        r = true
    }
    if s.top > 0 {
        u = true
    }
    return
}

// push operation position(s), caller tag, deleted and inserted lengths onto the
// undo/redo stack. Usually this means just appending an operation to the stack,
// but if operations were previously undone, the top of the stack is not at the
// end of the slice anymore. In that case, all operations that could be redone
// until now are just removed from the slice and the backup (newData) and cut
// (cutData) storage are cleaned: afterwards, previously saved data for redo are
// forgotten and it is impossible to redo those operations.
// Because it might change newData and/or cutData length, pushOp must be called
// before updating newData or cutData (see insertInCurData and cutInCurData).
func (s *Storage) pushSinglePosOp( p, t, dl, il int64 ) {
    s.cleanStack()
    s.stack = append( s.stack, singlePosOp{ t, dl, il,
                                            int64(len(s.newData)),
                                            int64(len(s.cutData)), p } )
    s.top ++
    if s.top < len( s.stack ) {
        log.Panicln( "pushSinglePosOp not at top of stack" )
    }
    s.notifyUndoRedo( true, false )
}

func (s *Storage) pushMultiPosOp( p []int64, t, dl, il int64 ) {
    s.cleanStack()
    s.stack = append( s.stack, multiPosOp{ t, dl, il,
                                            int64(len(s.newData)),
                                            int64(len(s.cutData)), p } )
    s.top ++
    s.notifyUndoRedo( true, /*s.top < len( s.stack)*/ false )
}

func (s *Storage) cleanStack( ) {
    if s.top < len( s.stack ) {
        for i := len( s.stack)-1; i >= s.top; i-- {
            op := s.stack[i]

            var b, c int64
            switch op := op.(type) {
            case singlePosOp:
                b, c = op.backup, op.cut
            case multiPosOp:
                b, c = op.backup, op.cut
            }
            // REPLACE, or INSERT (no cut data) or DELETE (no new data)
            if int64(len(s.newData)) < b {
                log.Panicln( "push: newData inconsistent" );
            }
            if int64(len(s.cutData)) < c {
                log.Panicln( "push: cutData inconsistent" );
            }
            s.newData = s.newData[:b]
            s.cutData = s.cutData[:c]
            s.lost = true
        }
        s.stack = s.stack[:s.top]
    }
}

// insert data bytes at position pos in main storage.
// if save is true, push the INSERT command onto the undo/redo stack and
//                  save the inserted data bytes into the backup storage.
func (s *Storage) insertInCurData( pos, tag int64, data []byte, save bool ) error {
    cl := int64(len(s.curData))
    if cl < pos || pos < 0 {
        return fmt.Errorf("INSERT outside data boundaries\n")
    }
    il := int64(len(data))
    if save {
        s.pushSinglePosOp( pos, tag, 0, il )           // no byte deleted, il inserted
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
        log.Panicln("INSERT: Wrong length after insertion")
    }
    if s.notifyLenChange != nil {
        s.notifyLenChange( s.Length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

func (s *Storage) InsertClipboardAt( pos, tag int64 ) error {
    cl := int64(len(s.curData))
    if cl < pos || pos < 0 {
        return fmt.Errorf("INSERT outside data boundaries\n")
    }
    il := s.clip.Size()
    s.pushSinglePosOp( pos, tag, 0, il )           // no byte deleted, il inserted
    nPos := int64(len(s.newData))
    for i:= int64(0); i < il; i++ {
        s.newData = append( s.newData, s.clip.Get( ) )
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
        s.notifyLenChange( s.Length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

// insert a single byte at position pos
func (s *Storage) InsertByteAt( pos, tag int64, v byte ) error {
    return s.insertInCurData( pos, tag, []byte{ v }, true )
}

// insert multiple bytes at position pos
func (s *Storage) InsertBytesAt( pos, tag int64, v []byte ) error {
    return s.insertInCurData( pos, tag, v, true )
}

// delete or cut data starting at pos, for the given length.
// if save is true, push the DELETE command onto the undo/redo stack and
//                  save the deleted data into the cut storage.
func (s *Storage) cutInCurData( pos, tag, length int64, save bool ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || length < 0 || (pos + length) > curLen {
        return fmt.Errorf("DELETE outside data boundaries\n")
    }

    if save {
        s.pushSinglePosOp( pos, tag, length, 0 )   // length bytes removed, 0 inserted
        s.cutData = append( s.cutData, s.curData[pos:pos+length]... )
    }
    if pos == curLen-length {
        s.curData = s.curData[:pos]
    } else {
        copy( s.curData[pos:], s.curData[pos+length:] )
        s.curData = s.curData[:curLen-length]
    }
    if s.notifyLenChange != nil {
        s.notifyLenChange( s.Length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}

// delete one byte at position pos.
func (s *Storage) DeleteByteAt( pos, tag int64 ) error {
    return s.cutInCurData( pos, tag, 1, true )
}

// delete n bytes at position pos.
func (s *Storage) DeleteBytesAt( pos, tag, n int64 ) error {
    return s.cutInCurData( pos, tag, n, true )
}

// save n bytes starting at position pos into the paste buffer
// this is used for cut and copy
func (s *Storage) saveClipboardData( pos, n int64 ) error {
    if pos < 0 || pos + n > int64(len(s.curData)) {
        return fmt.Errorf("CUT/COPY outside data boundaries\n")
    }
//    s.pasteData = s.pasteData[0:0]
//    s.pasteData = append( s.pasteData, s.curData[pos:pos+n]... )
    s.clip.Set( s.curData[pos:pos+n] )
    return nil
}

// cut n bytes starting at position pos
func (s *Storage) CutBytesAt( pos, tag, n int64 ) error {
    if err := s.saveClipboardData( pos, n ); err != nil {
        return err
    }
    return s.cutInCurData( pos, tag, n, true )
}

// copy n bytes starting at position pos
func (s *Storage) CopyBytesAt( pos, n int64 ) error {
    return s.saveClipboardData( pos, n )
}

// replace existing data[pos:pos+dl] in the main buffer with new data.
func (s *Storage) replaceInCurData( pos, tag, dl int64, data []byte ) {
    curLen := int64(len(s.curData))
    il := int64(len(data))

    if il <= dl {
        // replace first il bytes in [pos:pos+il]
        copy( s.curData[pos:], data )
        // move remaining bytes beyond pos+dl to pos+il        
        copy( s.curData[pos+il:], s.curData[pos+dl:] )
        // shrink the slice to pos+(il-dl)
        s.curData = s.curData[:curLen-dl+il]
    } else if pos + il > curLen {
        // ==----xxxx  curData
        //   ^   ^   ^
        //  pos +dl curLen
        //   ++++++++>>>>>>>  data
        //   ^             ^
        //   0             il

        // ==----xxxx>>>>>>>
        s.curData = append( s.curData, data[curLen-pos:]... )

        // ==----xxxx>>>>>>>xxxx
        s.curData = append( s.curData, s.curData[pos+dl:curLen]... )

        // ==++++++++>>>>>>>xxxx
        copy( s.curData[pos:], data[:curLen-pos] )
    } else {
        // ==--------xxxx=====::::  s.curData
        //   ^       ^            ^
        //  pos     +dl        curLen
        //   ++++++++++++           data
        //   ^           ^
        //   0          il
        //                        v
        // ==--------xxxx=====::::::::  s.curData
        s.curData = append( s.curData, s.curData[curLen-(il-dl):]... )

        // ==--------xxxx=====::::::::  s.curData
        // ==--------xxxxxxxx=====::::  s.curData
        copy( s.curData[pos+il:curLen], s.curData[pos+dl:] )

        // ==--------xxxxxxxx=====::::  s.curData
        //   ++++++++++++
        // ==++++++++++++xxxx=====::::
        copy( s.curData[pos:], data )
    }
}

func (s *Storage) replaceInCurDataNotify( pos, tag, dl int64, data []byte ) {

    curLen := int64(len(s.curData))
    s.replaceInCurData( pos, tag, dl, data )
    if curLen != int64(len(s.curData)) && s.notifyLenChange != nil {
        s.notifyLenChange( s.Length())
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
}

func (s *Storage) replaceInCurDataNotifySave( pos, tag, dl int64,
                                              data []byte ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || dl < 0 || (pos + dl) > curLen {
        return fmt.Errorf("REPLACE outside data boundaries\n")
    }
    il := int64(len(data))
    s.pushSinglePosOp( pos, tag, dl, il )  // dl removed, il inserted

    s.cutData = append( s.cutData, s.curData[pos:pos+dl]... )
    s.newData = append( s.newData, data... )
    s.replaceInCurDataNotify( pos, tag, dl, data )
    return nil
}

func (s *Storage) ReplaceWithClipboardAt( pos, tag, dl int64 ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || dl < 0 || (pos + dl) > curLen {
        return fmt.Errorf("REPLACE outside data boundaries\n")
    }
    il := s.clip.Size()
    s.pushSinglePosOp( pos, tag, dl, il )      // dl removed, il inserted

    s.cutData = append( s.cutData, s.curData[pos:pos+dl]... )
    nPos := int64(len(s.newData))
    for i := int64(0); i < il; i++ {
        s.newData = append( s.newData, s.clip.Get( ) )
    }
    s.replaceInCurDataNotify( pos, tag, dl, s.newData[nPos:] )
    return nil
}

// replace one byte at position pos
func (s *Storage) ReplaceByteAt( pos, tag int64, v byte ) error {
    return s.replaceInCurDataNotifySave( pos, tag, 1, []byte{v} )
}

// replace multiple bytes at position pos
func (s *Storage) ReplaceBytesAt( pos, tag, l int64, v []byte ) error {
    return s.replaceInCurDataNotifySave( pos, tag, l, v )
}

func (s *Storage) replaceInCurDataAtMultipleLocations( pos []int64,
                                                       tag, l int64,
                                                       v []byte,
                                                       correct bool ) {
    nl := int64(len(v))
    nPos := len(pos)
    for i := 0; i < nPos; i++ {
        var p int64
        if correct {            // correct for length difference
            p = pos[i] + int64(i) * (nl - l)
        } else {
            p = pos[i]
        }
        s.replaceInCurData( p, 0, l, v )
    }
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
}

// replace multiple bytes at multiple locations given by the pos slice.
func (s *Storage) ReplaceBytesAtMultipleLocations( pos []int64, tag, l int64,
                                                   v []byte ) error {
    curLen := int64(len(s.curData))
    for _, p := range pos {
        if p < 0 || l < 0 || p + l > curLen {
            return fmt.Errorf("REPLACE outside data boundaries\n")
        }
    }
    nl := int64(len(v))
    s.pushMultiPosOp( pos, tag, l, nl )
    p := pos[0]
    s.cutData = append( s.cutData, s.curData[p:p+l]... )
    s.newData = append( s.newData, v... )
    s.replaceInCurDataAtMultipleLocations( pos, tag, l, v, true )
    return nil
}

// special operations to avoid creating extra slices when deleting bytes

// replaceByteAtAndEraseFollowingBytes replaces the first byte at pos with the
// given byte v, and erases the following n bytes . This is used in case of
// replace editing mode when a selection is replaced by a single byte. This is
// treated as a replaceBytesAt in order to allow undo/redo as usual, but the
// caller does not have to create an extra, potentially large, slice full of
// zeros.
func (s *Storage) ReplaceByteAtAndEraseFollowingBytes( pos, tag int64, v byte,
                                                       n int64 ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || n < 0 || (pos + n + 1) > curLen {
        return fmt.Errorf("REPLACE outside data boundaries\n")
    }
    il := int64(n+1)

//    s.push( REPLACE, pos, tag, il, il ) // il removed, il inserted
    s.pushSinglePosOp( pos, tag, il, il )          // il removed, il inserted
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
