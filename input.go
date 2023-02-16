package main

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

const (
    NIBBLE = 1 + iota       // order is important: do not change
    LINE
    PAGE
    END
    ABSOLUTE 
)

func (pc *pageContext) setCaretPosition( offset int64,  unit int ) {

    adj := pc.barAdjust
    origin := int64(adj.GetValue())
    end := int64(adj.GetUpper())
    pageSize := int64(adj.GetPageSize())
    pageNibbles := (pageSize / int64(getCharHeight())) *
                                    int64(pc.nBytesLine << 1)
    dataNibbles := pc.store.Length() << 1

    if pc.sel.start != -1 {
        if unit < END && offset == 1  {     // next char, line or page starts at
            pc.caretPos = pc.sel.beyond * 2 // the bottom end of the selection
        }
        pc.resetSelection()     // moving caret always removes selection
    }

    switch unit {
    case NIBBLE: // always within current window, or +/-1 nibble outside
        nPos := pc.caretPos + int64(offset)
        if nPos < 0 {
            nPos = 0
        } else if nPos > dataNibbles {
            nPos = dataNibbles
        }
        pc.scrollPositionFollowCaret( nPos )

    case LINE:  // should always be +/- 1 line, within window or outside
        if offset != -1 && offset != 1 {
            panic("Bad offset for LINE\n")
        }
        nPos := pc.caretPos + (offset * int64(pc.nBytesLine << 1))
        if nPos >= 0 && nPos <= dataNibbles {
            pc.scrollPositionFollowCaret( nPos )
        }

    case PAGE:  // should always be +/- 1 page
        if offset != -1 && offset != 1 {
            panic("Bad offset for PAGE\n")
        }
        nPos := pc.caretPos + int64(offset) * pageNibbles
        if nPos >= 0 && nPos <= dataNibbles {
            pc.scrollPositionFollowPage( nPos, float64( int64(offset) * pageSize) )
        }

    case ABSOLUTE:  // can be anywhere, inside or outside current window
        pc.scrollPositionFollowCaret( offset )

    case END:   // should always be +/- 1 end
        if offset == -1 {
            pc.caretPos = 0
            origin = 0
        } else if offset == +1 {
            pc.caretPos = dataNibbles
            origin = end - pageSize
        } else {
            panic("Bad offset for END\n")
        }
        adj.SetValue( float64( origin ) )
        pc.showBytePosition()
    }
    if pc.caretPos & 1 == 0 {
        pc.setEvenCaretNoPending()
    } else {
        pc.setOddCaretNoPending()
    }
    explorePossible( pc.caretPos < dataNibbles )
    updateSearchPosition( pc.caretPos >> 1 )
}

/*  Nibble state machine: 8 states and 3 events (insert, delete, backspace)
        => 24 functions

    states:
        odd caret, no pending op
        odd caret, pending insert
        odd caret, pending delete
        odd caret, pending backspace
        even caret, no pending op
        even caret, pending insert
        even caret, pending delete
        even caret, pending backspace */

func (pc *pageContext) setEvenCaretNoPending( ) {
    pc.ins = (*pageContext).insertEvenCaretNoPending
    pc.del = (*pageContext).deleteEvenCaretNoPending
    pc.bck = (*pageContext).backspaceEvenCaretNoPending
}

func (pc *pageContext) setEvenCaretPendingInsert( ) {
    pc.ins = (*pageContext).insertEvenCaretPendingInsert
    pc.del = (*pageContext).deleteEvenCaretPendingInsert
    pc.bck = (*pageContext).backspaceEvenCaretPendingInsert
}

func (pc *pageContext) setEvenCaretPendingDelete( ) {
    pc.ins = (*pageContext).insertEvenCaretPendingDelete
    pc.del = (*pageContext).deleteEvenCaretPendingDelete
    pc.bck = (*pageContext).backspaceEvenCaretPendingDelete
}

func (pc *pageContext) setEvenCaretPendingBackspace( ) {
    pc.ins = (*pageContext).insertEvenCaretPendingBackspace
    pc.del = (*pageContext).deleteEvenCaretPendingBackspace
    pc.bck = (*pageContext).backspaceEvenCaretPendingBackspace
}

func (pc *pageContext) setOddCaretNoPending( ) {
    pc.ins = (*pageContext).insertOddCaretNoPending
    pc.del = (*pageContext).deleteOddCaretNoPending
    pc.bck = (*pageContext).backspaceOddCaretNoPending
}

func (pc *pageContext) setOddCaretPendingInsert( ) {
    pc.ins = (*pageContext).insertOddCaretPendingInsert
    pc.del = (*pageContext).deleteOddCaretPendingInsert
    pc.bck = (*pageContext).backspaceOddCaretPendingInsert
}

func (pc *pageContext) setOddCaretPendingDelete( ) {
    pc.ins = (*pageContext).insertOddCaretPendingDelete
    pc.del = (*pageContext).deleteOddCaretPendingDelete
    pc.bck = (*pageContext).backspaceOddCaretPendingDelete
}

func (pc *pageContext) setOddCaretPendingBackspace( ) {
    pc.ins = (*pageContext).insertOddCaretPendingBackspace
    pc.del = (*pageContext).deleteOddCaretPendingBackspace
    pc.bck = (*pageContext).backspaceOddCaretPendingBackspace
}

func (pc *pageContext) InitCaretPosition( ) {
//  pc.caretPos = 0
    pc.setEvenCaretNoPending()
}

func (pc *pageContext) insertEvenCaretNoPending( nibble byte ) {
//    ab cd ef gh         no pending op
//          ^             even caret
//   event insert nibble u
//    ab cd u0 ef gh      insert byte u0 @caret,
//          ++            pending op = insert
//           ^            move caret + 1 (odd)
    bPos := pc.caretPos / 2
//   when undoing caret goes back to same even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag = 0 | (1 << 2)
    pc.store.InsertByteAt( bPos, 1 << 2, nibble << 4 )
    pc.caretPos ++
    pc.setOddCaretPendingInsert()
}

func (pc *pageContext) insertEvenCaretPendingInsert( nibble byte ) {
//    ab cu 0d ef gh      pending op = insert
//          ^             even caret
//   event insert nibble v
//    ab cu vd ef gh      replace 0d <- vd @caret,
//          ==            no pending op
//           ^            move caret + 1 (odd)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    nb := data[0] | nibble << 4
//   when undoing caret goes back to same even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag = 0 | (1 << 2)
    pc.store.ReplaceByteAt( bPos, 1 << 2, nb )
    pc.caretPos ++
    pc.setOddCaretNoPending()
}

// same as NoPending
func (pc *pageContext) insertEvenCaretPendingDelete( nibble byte ) {
//    ab c0 ef gh         pending op = delete
//          ^             even caret
//   event insert nibble u
//    ab c0 u0 ef gh      insert byte u0 @caret,
//          ++            pending op = insert odd
//           ^            move caret + 1 (odd)
    bPos := pc.caretPos / 2
//   when undoing caret goes back to same even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag = 0 | (1 << 2)
    pc.store.InsertByteAt( bPos, 1 << 2, nibble << 4 )
    pc.caretPos ++
    pc.setOddCaretPendingInsert()
}

// same as PendingInsert
func (pc *pageContext) insertEvenCaretPendingBackspace( nibble byte ) {
//    ab 0d ef gh         pending op = backspace even
//       ^                even caret
//   event insert nibble u
//    ab ud ef gh         replace 0d <- ud @caret,
//       ==               no pending op
//        ^               move caret + 1 (odd)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    nb := data[0] | nibble << 4     // to allow undo
//   when undoing caret goes back to same even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag = 0 | (1 << 2)
    pc.store.ReplaceByteAt( bPos, 1 << 2,  nb )
    pc.caretPos ++
    pc.setOddCaretNoPending()
}

func (pc *pageContext) insertOddCaretNoPending( nibble byte ) {
//    ab cd ef gh         no pending op
//        ^               odd caret
//   event insert nibble u
//    ab cu 0d ef gh      replace byte cd <-cu @caret-1, insert byte 0d @caret+1,
//       == ++            pending op = insert
//          ^             move caret + 1 (even)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := (data[0] & 0xf0) | nibble
    nb := data[0] & 0x0f
    res := make( []byte, 2 )
    res[0] = ub
    res[1] = nb
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag = 1 | (2 << 2)
    pc.store.ReplaceBytesAt( bPos, 1 | (2 << 2), 1, res )
    pc.caretPos ++
    pc.setEvenCaretPendingInsert( )
}

func (pc *pageContext) insertOddCaretPendingInsert( nibble byte ) {
//    ab cd u0 ef gh      pending op = insert
//           ^            odd caret
//   event insert nibble v
//    ab cd uv ef gh      replace byte u0 <- uv @caret-1,
//          ==            no pending op
//             ^          move caret + 1 (even)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0] | nibble
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag = 1 | (2 << 2)
    pc.store.ReplaceByteAt( bPos, 1 | (2 << 2), ub )
    pc.caretPos ++
    pc.setEvenCaretNoPending( )
}

// alternatively same as noPending (insert byte 0f @caret+1, replace byte 0f <-0u @caret-1)
func (pc *pageContext) insertOddCaretPendingDelete( nibble byte ) {
//    ab cd 0f gh         pending op = delete
//           ^            odd caret
//   event insert nibble u
//    ab cd 0u 0f gh      insert byte 0u @caret-1,
//          ++            pending op = insert
//             ^          move caret + 1 (even)
    bPos := pc.caretPos / 2
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag = 1 | (2 <<2)
    pc.store.InsertByteAt( bPos, 1 | (2 <<2), nibble )
    pc.caretPos ++
    pc.setEvenCaretPendingInsert( )
}

// same as pendingInsert
func (pc *pageContext) insertOddCaretPendingBackspace( nibble byte ) { 
//    ab c0 ef gh         pending op = backspace odd
//        ^               odd caret
//   event insert nibble u
//    ab cu ef gh         replace c0 <- cu @caret-1,
//       ==               no pending op
//          ^             move caret + 1 (even)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0] | nibble
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag = 1 | (2 << 2)
    pc.store.ReplaceByteAt( bPos, 1 | (2 << 2), ub )
    pc.caretPos ++
    pc.setEvenCaretNoPending( )
}

func (pc *pageContext) deleteEvenCaretNoPending( ) {
//    ab cd ef gh         no pending op
//          ^             even caret
//   event delete
//    ab cd 0f gh         replace byte ef <- 0f @caret,
//          ==            pending op = delete
//           ^            move caret + 1 (odd)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0] & 0x0f
//   when undoing caret goes back to same even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag = 0 | (1 << 2)
    pc.store.ReplaceByteAt( bPos, 1 << 2, ub )
    pc.caretPos ++
    pc.setOddCaretPendingDelete( )
}

func (pc *pageContext) deleteEvenCaretPendingInsert( ) {
//    ab cu 0d ef gh      pending op = insert
//          ^             even caret
//   event delete
//    ab cu 0d ef gh      no byte update,
//          ==            pending op = delete
//           ^            move care + 1 (odd)
//   Even though there is no change, we replace the current byte
//   with itself to allow undoing/redoing and caret move.
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0]
//   when undoing caret goes back to same even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag = 0 | (1 << 2)
    pc.store.ReplaceByteAt( bPos, 1 << 2, ub )
    pc.caretPos ++
    pc.setOddCaretPendingDelete( )
}

func (pc *pageContext) deleteEvenCaretPendingDelete( ) {
//    ab c0 ef gh         pending op = delete
//          ^             even caret
//   event delete
//    ab cf    gh         replace byte c0 <- cf @caret-2,
//       == --            no pending op
//        ^               move caret - 1 (odd)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    nibble := data[0] & 0x0f
    data = pc.store.GetData( bPos - 1, bPos )
    ub := data[0] | nibble
    rep := make( []byte, 1 )
    rep[0] = ub
//   when undoing caret goes back to next even position: tag = 2
//   when redoing caret moves 1 nibble ahead tag: = 2 | (1 << 2)
    pc.store.ReplaceBytesAt( bPos - 1, 2 | (1 << 2), 2, rep )
    pc.caretPos --
    pc.setOddCaretNoPending( )
}

// same as PendingInsert
func (pc *pageContext) deleteEvenCaretPendingBackspace( ) {
//    ab 0d ef gh         pending op = backspace even
//       ^                even caret
//   event delete
//    ab 0d ef gh         no byte update,
//       ==               pending op = delete
//        ^               move caret + 1 (odd)
//   Even though there is no change, we replace the current byte
//   with itself to allow undoing/redoing and caret move.
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0]
//   when undoing caret goes back to same even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag = 0 | (1 << 2)
    pc.store.ReplaceByteAt( bPos, 1 << 2, ub )
    pc.caretPos ++
    pc.setOddCaretPendingDelete( )
}

func (pc *pageContext) deleteOddCaretNoPending( ) {
//    ab cd ef gh         no pending op
//        ^               odd caret
//   event delete
//    ab c0 ef gh         replace byte cd <- c0 @caret-1,
//       ==               pending op = delete
//          ^             move caret + 1 (even)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0] & 0xf0
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag = 1 | (2 << 2)
    pc.store.ReplaceByteAt( bPos, 1 | (2 << 2), ub )
    pc.caretPos ++
    pc.setEvenCaretPendingDelete( )
}

func (pc *pageContext) deleteOddCaretPendingInsert( ) {
//    ab cd u0 ef gh      pending op = insert
//           ^            odd caret
//   event delete
//    ab cd u0 ef gh      no byte update,
//          ==            pending op = delete
//             ^          move caret + 1 (even)
//   Even though there is no change, we replace the current byte
//   with itself to allow undoing/redoing and caret move.
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0]
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag = 1 | (2 << 2)
    pc.store.ReplaceByteAt( bPos, 1 | (2 << 2), ub )
    pc.caretPos ++
    pc.setEvenCaretPendingDelete( )
}

func (pc *pageContext) deleteOddCaretPendingDelete( ) {
//    ab cd 0f gh         pending op = delete
//           ^            odd caret
//   event delete
//    ab cd    gh         remove byte @caret-1,
//          --            no pending op
//          ^             move caret - 1 (even)
    bPos := pc.caretPos / 2
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret goes to the same even position: tag = 1 | (0 << 2)
    pc.store.DeleteByteAt( bPos, 1 )
    pc.caretPos --
    pc.setEvenCaretNoPending( )
}

// same as deleteOddCaretPendingInsert
func (pc *pageContext) deleteOddCaretPendingBackspace( ) {
//    ab c0 ef gh         pending op = backspace odd
//        ^               odd caret
//   event delete
//    ab c0 ef gh         no byte update,
//       ==               pending op = delete
//          ^             move caret + 1 (even)
//   Even though there is no change, we replace the current byte
//   with itself to allow undoing/redoing and caret move.
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0]
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag = 1 | (2 << 2)
    pc.store.ReplaceByteAt( bPos, 1 | (2 << 2), ub )
    pc.caretPos ++
    pc.setEvenCaretPendingDelete( )
}

func (pc *pageContext) backspaceEvenCaretNoPending( ) {
//    ab cd ef gh         no pending op
//          ^             even caret
//   event backspace
//    ab c0 ef gh         replace byte cd <- c0 @caret-2,
//       ==               pending op = backspace
//        ^               move caret - 1 (odd)
    bPos := (pc.caretPos / 2) - 1
    if bPos >= 0 {
        data := pc.store.GetData( bPos, bPos + 1 )
        ub := data[0] & 0xf0
//   when undoing caret goes back to next even position: tag = 2
//   when redoing caret moves 1 nibble ahead: tag = 2 | (1 << 2)
        pc.store.ReplaceByteAt( bPos, 2 | (1 << 2), ub )
        pc.caretPos --
        pc.setOddCaretPendingBackspace( )
    }
}

func (pc *pageContext) backspaceEvenCaretPendingInsert( ) {
//    ab cu 0d ef gh      pending op = insert
//          ^             even caret
//   event backspace
//    ab cd    ef gh      replace cu 0d <- cd @caret-2,
//       == --            no pending op
//        ^               move caret -1 (odd)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    nibble := data[0] & 0x0f
    data = pc.store.GetData( bPos - 1, bPos )
    ub := (data[0] & 0xf0) | nibble
    rep := make( []byte, 1 )
    rep[0] = ub
//   when undoing caret goes back to next even position: tag = 2
//   when redoing caret moves 1 nibble ahead: tag = 2 | (1 << 2)
    pc.store.ReplaceBytesAt( bPos - 1, 2 | (1 << 2), 2, rep )
    pc.caretPos --
    pc.setOddCaretNoPending( )
}

func (pc *pageContext) backspaceEvenCaretPendingDelete( ) {
//    ab c0 ef gh         pending op = delete
//          ^             even caret
//   event backspace
//    ab c0 ef gh         no byte update,
//       ==               pending op = backspace
//        ^               move caret - 1 (odd)
//   Even though there is no change, we replace the current byte
//   with itself to allow undoing/redoing and caret move.
    if pc.caretPos > 0 {
        bPos := (pc.caretPos / 2) - 1
        data := pc.store.GetData( bPos, bPos + 1 )
        ub := data[0]
//   when undoing caret goes back to next even position: tag = 2
//   when redoing caret moves 1 nibble ahead: tag = 2 | (1 << 2)
        pc.store.ReplaceByteAt( bPos, 2 | (1 << 2), ub )
        pc.caretPos --
        pc.setOddCaretPendingBackspace( )
    }
}

// same as backspaceEvenCaretPendingInsert
func (pc *pageContext) backspaceEvenCaretPendingBackspace( ) {
//    ab 0d ef gh         pending op = backspace even
//       ^                even caret
//   event backspace
//    ad    ef gh         replace ab <- ad @caret-2,
//    == --               no pending op
//     ^                  move caret -1 (odd)
    if pc.caretPos > 0 {
        bPos := pc.caretPos / 2
        data := pc.store.GetData( bPos, bPos + 1 )
        nibble := data[0] & 0x0f
        data = pc.store.GetData( bPos - 1, bPos )
        ub := (data[0] & 0xf0) | nibble
        rep := make( []byte, 1 )
        rep[0] = ub
//   when undoing caret goes back to next even position: tag = 2
//   when redoing caret moves 1 nibble ahead: tag = 2 | (1 << 2)
        pc.store.ReplaceBytesAt( bPos - 1, 2 | (1 << 2), 2, rep )
        pc.caretPos --
        pc.setOddCaretNoPending( )
    }
}

func (pc *pageContext) backspaceOddCaretNoPending( ) {
//    ab cd ef gh         no pending op
//        ^               odd caret
//   event backspace
//    ab 0d ef gh         replace byte cd <- 0d @caret-1,
//       ==               pending op = backspace
//       ^                move caret - 1 (even)
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0] & 0x0f
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret goes back to same even position: tag = 1 | (0 << 2)
    pc.store.ReplaceByteAt( bPos, 1, ub )
    pc.caretPos --
    pc.setOddCaretPendingBackspace( )
}

func (pc *pageContext) backspaceOddCaretPendingInsert( ) {
//    ab cd u0 ef gh      pending op = insert
//           ^            odd caret
//   event backspace
//    ab cd    ef gh      remove byte u0 @caret-1,
//          --            no pending op
//          ^             move caret -1 (even)
    bPos := pc.caretPos / 2
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret goes back to same even position: tag = 1 | (0 << 2)
    pc.store.DeleteByteAt( bPos, 1 )
    pc.caretPos --
    pc.setEvenCaretNoPending( )
}

func (pc *pageContext) backspaceOddCaretPendingDelete( ) {
//    ab cd 0f gh         pending op = delete
//           ^            odd caret
//   event backspace
//    ab cd 0f gh         no byte update,
//          ==            pending op = backspace
//          ^             move caret - 1 (even)
//   Even though there is no change, we replace the current byte
//   with itself to allow undoing/redoing and caret move.
    bPos := pc.caretPos / 2
    data := pc.store.GetData( bPos, bPos + 1 )
    ub := data[0]
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret goes back to same even position: tag = 1 | (0 << 2)
    pc.store.ReplaceByteAt( bPos, 1, ub )
    pc.caretPos --
    pc.setEvenCaretPendingBackspace( )
}

// same as backspaceOddCaretPendingInsert
func (pc *pageContext) backspaceOddCaretPendingBackspace( ) {
//    ab c0 ef gh         pending op = backspace odd
//        ^               odd caret
//   event backspace
//    ab    ef gh         remove byte c0 @caret-1,
//       --               no pending op
//       ^                move caret - 1 (even)
    bPos := pc.caretPos / 2
//   when undoing caret goes back to same odd position: tag = 1
//   when redoing caret goes back to same even position: tag = 1 | (0 << 2)
    pc.store.DeleteByteAt( bPos, 1 )
    pc.caretPos --
    pc.setEvenCaretNoPending( )
}

func (pc *pageContext)insCommand( nibble byte ) {
    if pc.tempReadOnly  {
        return
    }
    if pc.sel.start != -1 {   // valid selection
        pc.killOrCleanSelection( true, nibble )
    } else if pc.replaceMode {
        bPos := pc.caretPos / 2
        data := pc.store.GetData( bPos, bPos + 1 )
        var ( ub byte; tag int64 )
        if data != nil {
            ub = data[0]
        } else {                    // deal with no existing data!
            ub = 0
        }
        if pc.caretPos & 1 == 0 {   // even position
            ub = ( ub & 0x0f ) | ( nibble << 4 )
//   when undoing caret goes back to even position: tag = 0
//   when redoing caret moves 1 nibble ahead: tag |=  (1 << 2)
            tag = 1 << 2
        } else {                    // odd position
            ub = ( ub & 0xf0 ) | nibble
//   when undoing caret goes back to odd position: tag = 1
//   when redoing caret moves 2 nibbles ahead: tag |=  (2 << 2)
            tag = 1 | (2 << 2)
        }
        if data != nil {
            pc.store.ReplaceByteAt( bPos, tag, ub )
        } else {
            pc.store.InsertByteAt( bPos, tag, ub )
        }
        pc.caretPos ++
    } else {
        pc.ins( pc, nibble )
    }
    pc.scrollPositionFollowCaret( pc.caretPos )
    pc.virgin = false
}

// add means insert nibble after cleaning selection
// cleaning selection means replacing it with 0's if replaceMode, or
//                          deleting it if not replaceMode
func (pc *pageContext) killOrCleanSelection( add bool, nibble byte ) bool {
    if pc.sel.start != -1 {     // valid selection
        if pc.tempReadOnly {    // but no modification allowed
            return true
        }
        s := pc.sel.start     // in bytes
        l := pc.sel.beyond - s
        cp := 2 * s           // caret position in nibbles
        // selection is always even aligned:
        // undoing moves caret back to the even position (0)
        // redoing moves caret 1 nibble ahead to the following odd postion (1<<2)
        tag := int64(1<<2)
        if pc.replaceMode {
            if add {
                pc.store.ReplaceByteAtAndEraseFollowingBytes( s, tag, nibble << 4, l-1 )
                cp ++
            } else {
                pc.store.ReplaceByteAtAndEraseFollowingBytes( s, 0, 0, l-1 )
            }
        } else {
            if add {
                rep := make( []byte, 1 )
                rep[0] = nibble << 4
                pc.store.ReplaceBytesAt( s, tag, l, rep )
                pc.setOddCaretPendingInsert()
                cp ++
            } else {
                pc.store.DeleteBytesAt( s, 0, l )
                pc.setEvenCaretNoPending()
            }
        }
        pc.scrollPositionFollowCaret( cp )
        pc.resetSelection()
        return true
    }
    return false
}

func (pc *pageContext)delCommand( ) {
    if pc.tempReadOnly  {
        return
    }
    if ! pc.killOrCleanSelection( false, 0 ) {
        bPos := pc.caretPos / 2
        if bPos < pc.store.Length() {
            if pc.replaceMode {
                data := pc.store.GetData( bPos, bPos + 1 )
                if data == nil {
                    return
                }
                var ( ub byte; tag int64 )
                if pc.caretPos & 1 == 0 {   // even position
                    ub = data [0] & 0x0f
                    tag = 1 << 2
                } else {                    // odd podition
                    ub = data[0] & 0xf0
                    tag = 1 | (2 << 2)
                }

                pc.store.ReplaceByteAt( bPos, tag, ub )
                pc.caretPos ++
            } else {
                pc.del( pc )
            }
            pc.scrollPositionFollowCaret( pc.caretPos )
        }
    }
    pc.virgin = false
}

func (pc *pageContext)backCommand( ) {
    if pc.tempReadOnly  {
        return
    }
    if ! pc.killOrCleanSelection( false, 0 ) {
        if pc.caretPos > 0 {
            if pc.replaceMode {
                pc.caretPos --              // move caret 1 nibble backward
                bPos := pc.caretPos / 2
                data := pc.store.GetData( bPos, bPos + 1)
                if data == nil {
                    return
                }
                var ( ub byte; tag int64 )
                if pc.caretPos & 1 == 0 {   // even position after back
                    ub = data [0] & 0x0f
                    tag = 1 // undo: back to following odd, redo no move
                } else {                    // odd position after back
                    ub = data[0] & 0xf0
                    tag = 2 | (1 << 2) // undo next byte, redo next nibble
                }
                pc.store.ReplaceByteAt( bPos, tag, ub )
            } else {
                pc.bck( pc )
            }
            pc.scrollPositionFollowCaret( pc.caretPos )
        }
    }
    pc.virgin = false
}

const (
    ENTER_KEY = gdk.KEY_Return
    KEYPAD_ENTER_KEY = gdk.KEY_KP_Enter

    DELETE_KEY = gdk.KEY_Delete
    BACKSPACE_KEY = gdk.KEY_BackSpace

    HOME_KEY = gdk.KEY_Home
    END_KEY = gdk.KEY_End

    LEFT_KEY = gdk.KEY_Left
    UP_KEY = gdk.KEY_Up
    RIGHT_KEY = gdk.KEY_Right
    DOWN_KEY = gdk.KEY_Down

    PAGE_UP_KEY = gdk.KEY_Page_Up
    PAGE_DOWN_KEY = gdk.KEY_Page_Down

    INSERT_KEY = gdk.KEY_Insert
)

func editAtCaret( da *gtk.DrawingArea, event *gdk.Event ) bool {
// caret position in hexArea is expressed in nibbles (0.5 byte)
// it takes 2 * nBytesLine to reach the end of a line
    keyEvent := gdk.EventKeyNewFromEvent(event)
    modifiers := keyEvent.State()
//    printDebug( "Key modifiers=%#04x\n", modifiers )
    if modifiers & 0x0f != 0 {
        return false
    }
    keyVal := keyEvent.KeyVal()
    pc := getCurrentPageContext()
    switch keyVal {
    case HOME_KEY:
        pc.setCaretPosition( -1, END )
    case END_KEY:
        pc.setCaretPosition( +1, END )

    case LEFT_KEY:
        pc.setCaretPosition( -1, NIBBLE )
    case UP_KEY:
        pc.setCaretPosition( -1, LINE )
    case RIGHT_KEY:
        pc.setCaretPosition( +1, NIBBLE )
    case DOWN_KEY:
        pc.setCaretPosition( +1, LINE )

    case PAGE_UP_KEY:
        pc.setCaretPosition( -1, PAGE )

    case PAGE_DOWN_KEY:
        pc.setCaretPosition( +1, PAGE )

    case INSERT_KEY:
        if ! pc.tempReadOnly {
            pc.replaceMode = ! pc.replaceMode
            if pc.caretPos & 1 == 0 { // ensure proper state setting for insert
                pc.setEvenCaretNoPending()
            } else {
                pc.setOddCaretNoPending()
            }
        }
        showInputMode( pc.tempReadOnly, pc.replaceMode )

    case BACKSPACE_KEY:
        pc.backCommand( )

    case DELETE_KEY:
        pc.delCommand( )

    default:
        if hex, nibble := getNibbleFromKey( keyVal ); hex {
            pc.insCommand( nibble )
        } else {
            return false
        }
    }
    pc.canvas.QueueDraw( )    // force redraw
    return true
}

func getNibbleFromKey( keyVal uint ) (hex bool, nibble byte) {
    if keyVal & 0xff00 == 0 {                       // regular keys
        b := byte(keyVal & 0xff)
        if b < '0' || (b < 'A' && b > '9') || (b < 'a' && b > 'F') || b > 'f' {
            return false, 0
        }
        if b >= 'a' {
            nibble = b - ('a'-10)
        } else if b >= 'A' {
            nibble = b - ('A'-10)
        } else {
            nibble = b - ('0')
        }
    } else if keyVal & 0xfff0 == 0xffb0 {           // num from keypad
        nibble = byte(keyVal & 0x0f)
    } else {
        return
    }
    hex = true
    return
}
