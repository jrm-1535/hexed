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

func showClipboard() {
    data, err := clipBoard.cb.WaitForText( )
    if err == nil {
        fmt.Printf( "Clipboard: %s\n", data )
    } else {
        fmt.Printf( "unable to get clipboard data\n" )
    }
}

func ownerChanged( cb *gtk.Clipboard, event *gdk.Event ) {
    fmt.Printf( "Clipboard changed\n")
//    if clipBoard != cb {
//        fmt.Printf("Not the hexed clipBoard!\n")
//    }
    showClipboard()
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
    if l < 2 || (l & 1) != 0 {
        fmt.Printf("isClipboardAvailable returns false\n")
        return false, nil
    }
    clipBoard.index = 0
    clipBoard.extraValid = false
    return true, &clipBoard
}

// set an extra byte for prepending and appending nibbles in case of pasting
// at an odd nibble location. This must be done by the entity managing the
// paste operation, and is only valid for that paste operation
func (hcp *hexClipboard) setExtraNibbles( extra byte ) {
    hcp.extraByte = extra
    hcp.extraValid = true
//fmt.Printf("clipboard setExtraNibbles: first nibble=%#2x, last nibble=%#2x\n",
//           extra >> 4, extra & 0x0f)
}

func (hcp *hexClipboard)size() ( n int64 ) {
    n = int64(len(hcp.data)) / 2    // in bytes
    if hcp.extraValid { // including extra 2 nibbles if extraByte is valid
        n ++
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
//fmt.Printf( "clipboard size=%d index=%d extra=%#02x valid=%v data=%v\n",
//            l, hcp.index, hcp.extraByte, hcp.extraValid, hcp.data )
    if hcp.index >= 2 * l {
        panic("out of clipboard data\n")
    }
    var b byte
    if hcp.extraValid && hcp.index == 0 {
        b = hcp.extraByte >> 4
    } else {
        b = getNibbleFromHexDigit( hcp.data[hcp.index] )
        hcp.index++
    }
    b <<= 4
    if hcp.extraValid && hcp.index == 2 * (l-1) {
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
    b := make( []byte, l * 2 )  // 2 char per data byte
    writeHexDigitsFromSlice( b, data )
    clipBoard.cb.SetText( string(b) )
    pasteDataExists( true )
}
