#!/usr/bin/env python3.6

import hashlib


#------------------------------------------------------------------------------
class Hash:
    ''' Helper class to let us know if a file changes. 
    '''
    @staticmethod
    def md5(contents: bytes) -> str:
        h = hashlib.md5()
        h.update(contents)
        return h.digest()


