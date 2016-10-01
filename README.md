# Komblobulate

Komblobulate is an encrypting, error resistant file container format, and a Go library for wrapping a file in this format.  The library is available freely under the MIT license.

For a simple example of usage, see my [kblob_cmd](https://github.com/kaiekkrin/kblob_cmd) project, which implements a simple command line interface.  Some day I might write some documentation.

Komblobulate provides:

* Resistance against random bit flips in the file data
* An encryption layer which may or may not prevent a user without the password from acquiring the plaintext of the file :)

