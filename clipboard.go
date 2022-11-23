package main

import (
    "fmt"
	"github.com/gotk3/gotk3/gtk"
	"github.com/gotk3/gotk3/gdk"
)

type hexClipboard struct {
    cb          *gtk.Clipboard
    data        string
    index       int64   // in nibbles
    extraByte   byte    // 2 nibbles, 4 MS Bits to prepend, 4 LS Bits to append
    extraValid  bool    // true if extraByte is meaningful
}

var clipBoard hexClipboard

func ownerChanged( cb *gtk.Clipboard, event *gdk.Event ) {
    fmt.Printf( "Clipboard changed\n")
//    if clipBoard != cb {
//        fmt.Printf("Not the hexed clipBoard!\n")
//    }
}

func initClipboard( ) (err error) {  // which clipboard?
    clipBoard.cb, err = gtk.ClipboardGet( gdk.SELECTION_CLIPBOARD /*SELECTION_PRIMARY*/ )
    if err == nil {
        clipBoard.cb.Connect( "owner-change", ownerChanged )
    } else {
        fmt.Printf( "Unable to get clipBoard: error %v\n", err )
    }
    return
}

// return whether data is available in the clipboard and if true, an interface
// to access data. Must be called before any attempt at retreiving clipboard
// data
func isClipboardDataAvailable( ) ( bool, *hexClipboard) {
    if clipBoard.cb == nil {
        panic("Hexed clipboard not initialized\n")
    }
    // TODO: when setWithOwner is available in gotk3, use WaitForContent and
    // binary data (application/octet-stream) as content target
    data, err := clipBoard.cb.WaitForText( )
    if err == nil {
        clipBoard.data = data
    }

    l := len(clipBoard.data)    // in bytes
//    fmt.Printf("Ascii: %s len=%d\n", ascii, l )
// read binary data from hex chars, prefixed with "0x"
    if l < 4 || (l & 1) != 0 || clipBoard.data[0] != '0' ||
                (clipBoard.data[1] != 'x' && clipBoard.data[1] != 'X') {
fmt.Printf("isClipboardAvalaible returns false\n")
        return false, nil
    }
    clipBoard.index = 2         // in nibbles after removing leading 0x
    clipBoard.extraValid = false
    return true, &clipBoard
}

// set an extra byte for prepending and appending nibbles in case of pasting
// at an odd nibble location. This must be done by the entity managing the
// paste operation, and is only valid for that paste operation
func (hcp *hexClipboard) setExtraNibbles( extra byte ) {
    hcp.extraByte = extra
    hcp.extraValid = true
}

func (hcp *hexClipboard)size() ( n int64 ) {
    n = int64(len(hcp.data)) / 2    // in bytes, including extra 2 nibbles
    if ! hcp.extraValid {           // unless extraByte is not valid
        n --
    }
    return
}

func getNibbleFromHexDigit( hc byte ) (nibble byte) {
    if hc >= 'a' {
        nibble = hc - ('a'-10)
    } else if hc >= 'A' {
        nibble = hc - ('A'-10)
    } else {
        nibble = hc - ('0')
    }
    return
}

// TODO: when using WaitForContent with binary data (application/octet-stream) 
// as content target, return directly 1 binary byte
func (hcp *hexClipboard)get() byte {
    l := hcp.size()
//fmt.Printf( "clipboard size=%d index=%d extra=%#02x valid=%v\n",
//            l, hcp.index, hcp.extraByte, hcp.extraValid )
    if hcp.index - 2 >= 2 * l {
        panic("out of clipboard data\n")
    }
    var b byte
    if hcp.extraValid && hcp.index == 2 {
        b = hcp.extraByte >> 4
    } else {
        b = getNibbleFromHexDigit( hcp.data[hcp.index] )
        hcp.index++
    }
    b <<= 4
    if hcp.extraValid && hcp.index == (2 * l) {
        b += hcp.extraByte & 0x0f
    } else {
        b += getNibbleFromHexDigit( hcp.data[hcp.index] )
        hcp.index++
    }
    return b
}

func getHexDigitFromNibble( nibble byte ) (hd byte) {

    if nibble < 10 {
        hd = '0' + nibble
    } else {
        hd = ('a'-10) + nibble
    }
    return
}

// write hex digits in out slice, from the source data bytes
// out must have been properly allocated for the data source:
// slice of length & capacity == len(data) * 2
func writeHexDigitsFromSlice( out []byte, data []byte ) {
    for i := 0; i < len(data); i++ {
        db := data[i]
        j := i << 1
        out[j] = getHexDigitFromNibble( db >> 4 )
        j++
        out[j] = getHexDigitFromNibble( db & 0x0f )
    }
}

// save binary data into hex chars, prefixed with "0x"
func setClipboardData( data []byte ) {
// TODO: when available in gotk3, use setWithOwner instead of immediate copy here
    l := len( data )
//fmt.Printf("setClipboardData: data len=%d\n", l)
    b := make( []byte, (l+1) * 2 )  // 0x + 2 char per data byte
    b[0] = '0'
    b[1] = 'x'
    writeHexDigitsFromSlice( b[2:], data )
    clipBoard.cb.SetText( string(b) )
    pasteDataExists( true )
}
