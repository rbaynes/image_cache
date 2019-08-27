#!/usr/bin/env python3.6

from collections import deque
from typing import Dict, Any


#------------------------------------------------------------------------------
class Cache:
    ''' Helper class to cache HTTP headers and file contents from an URL.
        A Least Recently Used (LRU) list of URLs is maintained, so that when 
        the cache is about to exceed its maximum capacity, it can evict the
        LRU item and recover its space.
    '''
    # Content cache in a hashtable.  Assumes a single host.
    # Format is {'URL', {'header': 'value', 
    #                    'file_bytes': <file contents>}}
    __cache: Dict[str, Dict[str, Any]] = {}


    def __init__(self, max_bytes: int = 200*1024) -> None:
        self.__LRU = deque() # Least Recently Used list of keys
        self.max_bytes = max_bytes
        self.current_bytes = 0


    def __repr__(self) -> str:
        ret = ''
        for k in self.__cache.keys():
            ret += k + '\n'
            for sk in self.__cache[k].items():
                if sk[1] is not None and len(sk[1]) <= 70: 
                    ret += f'  {sk[0]}: {sk[1]}\n'
                else:
                    ret += f'  {sk[0]}: ...\n' # don't print large values
        ret += f'LRU list (last key is LRU):\n'
        for k in self.__LRU:
            ret += f'  {k}\n'
        ret += f'Max cache size: {self.max_bytes} bytes\n'
        ret += f'  Current size: {self.current_bytes} bytes\n'
        ret += f'        Unused: {self.max_bytes - self.current_bytes} bytes\n'
        multiple = 's'
        if 1 == len(self.__LRU):
            multiple = ''
        return f'{len(self.__LRU)} Cache Item{multiple}:\n{ret}'


    def clear(self) -> None:
        self.__cache.clear()
        self.__LRU.clear()
        self.current_bytes = 0


    def set(self, key: str, subkey: str, value: Any) -> None:
        if value is None:
            return
        # Make room for this item if necessary.
        self.__check_size_and_evict(value)
        # Add the new item to the cache
        if key not in self.__cache: # add the first entry
            self.__cache[key] = {subkey: value}
        else: 
            if subkey in self.__cache[key]:
                # we are over-writing a subkey, so recover its bytes
                self.current_bytes -= len(self.__cache[key][subkey])
            self.__cache[key][subkey] = value
        # Account for the memory this value takes up (ignore keys)
        self.current_bytes += len(value)
        # Put this key in the front (most recently used) spot in the LRU list.
        self.__add_head_to_LRU(key)


    def get(self, key: str, subkey: str) -> Any:
        if key not in self.__cache: 
            return None
        elif subkey not in self.__cache[key]:
            return None
        # Move this key to the front of the LRU list.
        self.__add_head_to_LRU(key)
        return self.__cache[key][subkey]


    def __remove_from_LRU(self, key: str) -> None:
        # If the key is in the list, remove it.
        # (The linear search below will be slow if many items in the list.)
        if key in self.__LRU:      
            self.__LRU.remove(key)  


    def __add_head_to_LRU(self, key: str) -> None:
        # Remove the key if it's already in the list.
        self.__remove_from_LRU(key)
        # Put the key at the front of the list.
        self.__LRU.appendleft(key)


    def __get_LRU(self) -> str:
        # Returns the key of the LRU item, or None if the LRU list is empty.
        if 0 == len(self.__LRU):
            return None
        return self.__LRU.pop() # Note: removes item from the list.


    # Will adding this value exceed our our capacity?
    # If so, evict the LRU item to make room.
    def __check_size_and_evict(self, value_to_add: Any) -> None:
        if self.current_bytes + len(value_to_add) < self.max_bytes:
            return # Nothing to do, there is enough space.
        # Capacity will be exceeded, so get and evict the LRU
        lru = self.__get_LRU()
        if lru is None:
            return
        recovered_bytes = 0
        # Recover the bytes from the subkeys of this item
        for sk in self.__cache[lru].items():
            recovered_bytes += len(sk[1])
        self.current_bytes -= recovered_bytes
        # Remove the LRU from the cache dict
        self.__cache.pop(lru)
        # Remove the LRU from the list
        self.__remove_from_LRU(lru)
        print(f'Evicted {lru} from the cache and recovered '
                f'{recovered_bytes} bytes.') 

