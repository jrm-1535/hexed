A binary file editor, showing binary data in hexadecimal characters and ascii
when possible. It allows selection in either the hexadecimal area ot the ascii
area, complete or temporary write protection, search for patterns and full undo
redo stack.

It relies on gotk3 (and therefore on gtk3 and Cairo) for all windowing and
graphics operations. It has only be somehow tested on linux but should be
portable on any platform that supports gtk3 and go.

It is currently far from complete (pre alpha).

