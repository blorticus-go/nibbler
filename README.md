# nibbler

Nibble chunks from Reader streams and slice in a common way

## Overview

This is a golang module that provides an interface for treating a Reader and a byte slice in a similar way when reading one byte at-a-time, or a sequence of bytes based on a character set, where each character is in the UTF-8 range of codepoint, 1..255, inclusive.  It is possible to peek ahead in the byte stream, or return a character previously read to the stream.  When using a Reader as the stream source, the package will automatically trigger a Read() whenever necessary.

