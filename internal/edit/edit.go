// Package edit provides primitives for editing (inserting, deleting, replacing,
// copying. cutting, pasting, undoing, redoing) any sequence of bytes.
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

// clipboard interface required for cutting to and pasting from.
type Clipboard interface{
    Size( ) int64       // return clipboard current content size
    Set( data []byte )  // overwrite clipboard content with data slice
    Get( ) byte         // read one byte at a time from clipboard [0:size]
}

// Storage object required for any operation.
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

// internal debug functions
func (s *Storage) print(  ) {
    fmt.Printf("Storage %v\n", s.curData )
}

func (s *Storage) printInternal( ) {
    fmt.Printf( "newData storage %v\n", s.newData )
    fmt.Printf( "cutData storage %v\n", s.cutData )
}

func (s *Storage) printStack( ) {
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

// return current storage length.
func (s *Storage) Length() int64 {
    return int64(len(s.curData))
}

// return whether storage is presumed modified since creation/reload.
func (s *Storage)IsDirty( ) bool {
    if s.lost {
        return true
    }
    return s.top > 0
}

// attach a notification triggered each time the storage length changes.
func (s *Storage) SetNotifyLenChange( f func( int64 ) ) {
    s.notifyLenChange = f
}

// attach a notification triggered each time the undo/redo capability changes.
func (s *Storage) SetNotifyUndoRedoAble( f func( u, r bool ) ) {
    s.notifyUndoRedo = f
}

// attach a notification triggered each time data changes in storage.
func (s *Storage) SetNotifyDataChange( f func() ) {
    s.notifyDataChange = f
}

// GetData returns the current storage content.
// Beware, it is only a shallow copy: do NOT MODIFY the returned data
func (s *Storage) GetData( start, beyond int64 ) []byte {
    if start < 0 || start > beyond || beyond > int64(len(s.curData)) {
        return nil
    }
    return s.curData[start:beyond]
}

// NewStorage creates and initializes a storage. The argument path is a
// path to a file containing data bytes to edit. The argument clip is the
// clipboard interface to use when cutting, copying or pasting. In case path
// is empty, an initially empty storage is returned. An error is returned if
// the file corresponding to the path cannot be read, otherwise the whole
// file is read in memory, and the new storage is returned.
func NewStorage( path string, clip Clipboard ) ( *Storage, error ) {

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

// Reload performs storage re-initialization. It is used for example when
// reverting to the original data file. The original path must be provided.
// An error is returned if the file corresponding to the path cannot be read.
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

// Undo undo the last operation done on the storage. The position and tag
// provided when the operation was requested are returned to the caller.
// An error is returned if no operation can be undone.
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

// Redo does again the last undone operation. The position and tag provided
// when the operation was initially requested are returned to the caller.
// An error is returned if no operation can be redone.
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

// Checks whether Undo and/or Redo are possible and return a tuple (undo, redo)
// indicating the state of each.
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

// InsertClipboardAt implements a paste operation, inserting clipboard content
// at the given position in storage data. The arguments pos and tag are saved
// and returned when undoing or redoing the operation. An error is returned if
// the position is out of storage.
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

// InsertByteAt inserts a single byte in the storage at position given by the
// argument pos. The arguments pos and tag are saved and returned when undoing
// or redoing the operation. The argument b is the byte to insert. An error is
// returned if the position is out of storage.
func (s *Storage) InsertByteAt( pos, tag int64, b byte ) error {
    return s.insertInCurData( pos, tag, []byte{ b }, true )
}

// InsertBytesAt inserts multiple bytes in the storage at position given by the
// argument pos. The arguments pos and tag are saved and returned when undoing
// or redoing the operation. The argument b is the byte slice to insert. An
// error is returned if the position is out of storage.
func (s *Storage) InsertBytesAt( pos, tag int64, b []byte ) error {
    return s.insertInCurData( pos, tag, b, true )
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

// DeleteByteAt deletes one byte in the storage at position given by the
// argument pos. The arguments pos and tag are saved and returned when undoing
// or redoing the operation. An error is returned if the position is out of
// storage.
func (s *Storage) DeleteByteAt( pos, tag int64 ) error {
    return s.cutInCurData( pos, tag, 1, true )
}

// DeleteBytesAt deletes multiple bytes in the storage at position given by the
// argument pos. The segment to delete is defined by the starting pposition and
// the number of bytes n. The arguments pos and tag are saved and returned when
// undoing or redoing the operation. An error is returned if the segment to
// delete does not fit in storage.
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

// CutBytesAt implements a cut operation, from the storage to the clipboard.
// It moves the number of bytes given by the argument n, starting at position
// given by the argument pos from the storage, to the clipboard. The arguments
// pos and tag are saved and returned when undoing or redoing the operation.
// An error is returned if the segment to delete does not fit in storage.
func (s *Storage) CutBytesAt( pos, tag, n int64 ) error {
    if err := s.saveClipboardData( pos, n ); err != nil {
        return err
    }
    return s.cutInCurData( pos, tag, n, true )
}

// CopyBytesAt implements a copy operation, from storage to the clipboard. It
// copies the number of bytes given by the argument n, starting at position
// given by the argument pos from the storage, to the clipboard. Since this
// operation has no effect on the storage (nothing is added, removed or
// replaced), it is not possible to undo or redo it, and no tag is needed.
// An error is returned if the segment to copy does not fit in storage.
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

// ReplaceWithClipboardAt replaces a segment of data bytes in storage with
// the content of the clipboard. It has the same effect as deleting the segment
// and inserting the content of the clipboard at the same location. The
// arguments pos and dl indicate the starting point and the length of the data
// segment in storage respectively. The arguments pos and tag are saved and
// returned when undoing or redoing the operation. An error is returned if the
// segment to replace does not fit in storage.
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

// ReplaceByteAt replaces one byte in the storage at position given by the
// argument pos. It has the same effect as deleting one byte and inserting
// another one at the same location. The argument pos gives the location of
// the byte to replace. The arguments pos and tag are saved and returned when
// undoing or redoing the operation. The argument b is the byte that should
// replace the one deleted. An error is returned if the position is outside
// storage.
func (s *Storage) ReplaceByteAt( pos, tag int64, b byte ) error {
    return s.replaceInCurDataNotifySave( pos, tag, 1, []byte{b} )
}

// ReplaceBytesAt replaces multiple bytes in the storage at position given by
// the argument pos. It has the same effect as deleting a segment of bytes in
// storage and inserting a new segment of bytes at the same location. The
// arguments pos and dl indicate the starting point and the length of the data
// segment in storage respectively. The arguments pos and tag are saved and
// returned when undoing or redoing the operation. The argument b is the slice
// of bytes that should be inserted after the segment is deleted. An error is
// returned if the segment to replace does not fit in storage.
func (s *Storage) ReplaceBytesAt( pos, tag, dl int64, b []byte ) error {
    return s.replaceInCurDataNotifySave( pos, tag, dl, b )
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

// ReplaceBytesAtMultipleLocations replaces multiple bytes at multiple locations
// given by a slice of positions. It has the same effect as repeating deleting
// a segment of bytes in storage then inserting a new segment of bytes at the
// same location, for each position in the slice. The argument pos is the slice
// of position where replacement should happen. The argument dl indicates the
// length of bytes to delete ar eacj position, and the argument b is the new
// slice of data byte that should replace the deleted ones. When undoing or
// redoing the operation the first position (that is pos[0]) and tag are
// returned. An error is returned if any segment to replace does not fit in
// storage and no replacement occurs.
func (s *Storage) ReplaceBytesAtMultipleLocations( pos []int64, tag, dl int64,
                                                   b []byte ) error {
    curLen := int64(len(s.curData))
    for _, p := range pos {
        if p < 0 || dl < 0 || p + dl > curLen {
            return fmt.Errorf("REPLACE outside data boundaries\n")
        }
    }
    nl := int64(len(b))
    s.pushMultiPosOp( pos, tag, dl, nl )
    p := pos[0]
    s.cutData = append( s.cutData, s.curData[p:p+dl]... )
    s.newData = append( s.newData, b... )
    s.replaceInCurDataAtMultipleLocations( pos, tag, dl, b, true )
    return nil
}

// special operations to avoid creating extra slices when deleting bytes

// ReplaceByteAtAndEraseFollowingBytes replaces the first byte in storage at
// the location given by the argument pos with the given byte b, and erases the
// number of following bytes given by the argument n. This is used in case of
// replace editing mode when a selection is replaced by a single byte. This is
// treated as a replaceBytesAt in order to allow undo/redo as usual, but the
// caller does not have to create an extra, potentially large, slice full of
// zeros. An error is returned if the segment to replace does not fit in
// storage.
func (s *Storage) ReplaceByteAtAndEraseFollowingBytes( pos, tag int64, b byte,
                                                       n int64 ) error {
    curLen := int64(len(s.curData))
    if pos < 0 || n < 0 || (pos + n + 1) > curLen {
        return fmt.Errorf("REPLACE outside data boundaries\n")
    }
    il := int64(n+1)

    s.pushSinglePosOp( pos, tag, il, il )          // il removed, il inserted
    s.cutData = append( s.cutData, s.curData[pos:pos+il]... )
    nPos := int64(len(s.newData))
    s.newData = append( s.newData, b )

    for i := int64(0); i < n; i++ {
        s.newData = append( s.newData, 0 )
    }
    copy( s.curData[pos:pos+il], s.newData[nPos:nPos+il] )
    if s.notifyDataChange != nil {
        s.notifyDataChange()
    }
    return nil
}
