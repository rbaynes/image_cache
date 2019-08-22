#!/usr/bin/env python3.6

# rbaynes Aug. 22, 2019
# Class to fetch files over HTTP and keep a local cache of them.

import sys, traceback, argparse, http.client, hashlib
from typing import Dict, ClassVar, Any


#------------------------------------------------------------------------------
class Hash:
    ''' Helper class to let us know if a file changes. 
    '''
    @staticmethod
    def md5(contents: bytes) -> str:
        h = hashlib.md5()
        h.update(contents)
        return h.digest()


#------------------------------------------------------------------------------
class Cache:
    ''' Helper class to cache HTTP headers and file contents from an URL.
    '''
    # Content cache in a hashtable.  Assumes a single host.
    # Format is {'URL', {'header': 'value', 
    #                    'file_bytes': <file contents>}}
    __cache: Dict[str, Dict[str, Any]] = {}

    def __init__(self) -> None:
        pass

    def __repr__(self) -> str:
        ret = ''
        for k in self.__cache.keys():
            ret += k + '\n'
            for sk in self.__cache[k].items():
                if sk[1] is not None and len(sk[1]) <= 70: 
                    ret += f'  {sk[0]}: {sk[1]}\n'
                else:
                    ret += f'  {sk[0]}: ...\n' # don't print large values
        return f'Cache:\n{ret}'

    def clear(self) -> None:
        self.__cache.clear()

    def set(self, key: str, subkey: str, value: Any) -> None:
        if key not in self.__cache: # add the first entry
            self.__cache[key] = {subkey: value}
        else: 
            self.__cache[key][subkey] = value

    def get(self, key: str, subkey: str) -> Any:
        if key not in self.__cache: 
            return None
        elif subkey not in self.__cache[key]:
            return None
        else:
            return self.__cache[key][subkey]


#------------------------------------------------------------------------------
class FileFetchAndCache:
    ''' Fetch files over HTTP and cache their headers and contents.
        Return files from the cache if they have not changed.
    '''
    # HTTP header name constants (and cache keys)
    __if_none_match: ClassVar[str]     = 'If-None-Match'
    __if_modified_since: ClassVar[str] = 'If-Modified-Since'
    __etag: ClassVar[str]              = 'ETag'
    __last_modified: ClassVar[str]     = 'Last-Modified'
    __file_bytes: ClassVar[str]        = 'file_bytes' # cache key
    __file_hash: ClassVar[str]         = 'file_hash' # cache key

    def __init__(self, verbose: bool = False) -> None:
        self.verbose = verbose
        self.cache = Cache()

    ''' Fetch an URL and cache the contents.
        Arguments:
            host: hostname:port
            URL: URL to fetch starting with /
        Returns a tuple of:
            success: bool
            from_cache: bool
            file_bytes: bytes
    '''
    def get(self, host: str = None, URL: str = None) -> tuple:
        success: bool = False 
        from_cache: bool = False 
        file_bytes: bytes = None
        file_hash: str = None
        request_headers: dict = {}

        # First check our cache for the headers from the URL, 
        # if we find them, add headers to our request
        etag = self.cache.get(URL, FileFetchAndCache.__etag)
        last_mod = self.cache.get(URL, FileFetchAndCache.__last_modified)
        if etag is not None and last_mod is not None:
            request_headers[FileFetchAndCache.__if_none_match] = etag
            request_headers[FileFetchAndCache.__if_modified_since] = last_mod

        conn = http.client.HTTPSConnection(host)
        if self.verbose:
            conn.set_debuglevel(1) 
        conn.request('GET', URL, headers=request_headers)
        resp = conn.getresponse()

        # Get and cache the headers we care about from the response
        etag = resp.getheader(FileFetchAndCache.__etag)
        self.cache.set(URL, FileFetchAndCache.__etag, etag)
        last_mod = resp.getheader(FileFetchAndCache.__last_modified)
        self.cache.set(URL, FileFetchAndCache.__last_modified, last_mod)

        # If we fetched the file the first time, cache it
        if resp.status == 200:
            file_bytes = resp.read()
            self.cache.set(URL, FileFetchAndCache.__file_bytes, file_bytes)
            file_hash = Hash.md5(file_bytes)
            self.cache.set(URL, FileFetchAndCache.__file_hash, file_hash)
            success = True
        elif resp.status == 304:
            file_bytes = self.cache.get(URL, FileFetchAndCache.__file_bytes)
            file_hash = self.cache.get(URL, FileFetchAndCache.__file_hash)
            success = True
            from_cache = True
        conn.close()

        if self.verbose:
            print(resp.status, resp.reason)
            print(self.cache) 

        return success, from_cache, file_bytes


#------------------------------------------------------------------------------
if __name__ == "__main__":
    try:
        # Application defaults
        HOST = 'static.rbxcdn.com'
        URL = '/images/landing/Rollercoaster/whatsroblox_12072017.jpg'

        # Command line arg parsing
        parser = argparse.ArgumentParser(description='file fetch and cache')
        parser.add_argument('-v', '--verbose', dest='verbose', 
                action='store_true', 
                help='Enable verbose logging')
        parser.add_argument('-H', '--host', type=str, default=HOST, 
                help=f'Host in hostname:port format.  Default is {HOST}')
        parser.add_argument('-U', '--URL', type=str, default=URL, 
                help=f'URL to fetch.  Default is {URL}')
        args = parser.parse_args()
        
        # Our fetch and cache class
        ffac = FileFetchAndCache(verbose=args.verbose)

        # Fetch the same file two times, and see if it cached
        fetched_file_hash = None
        cached_file_hash = None
        for _ in range(2): # Get the file two times
            success, from_cache, file_bytes = ffac.get(args.host, args.URL)
            if success and not from_cache:
                fetched_file_hash = Hash.md5(file_bytes)
                print(f'Fetched and cached the file!')
            elif success and from_cache:
                cached_file_hash = Hash.md5(file_bytes)
                print(f'Got the file from the cache!')
            else:
                print(f'Error: Did not fetch file {args.URL}')

        if fetched_file_hash != cached_file_hash:
            print('Error: the file we fetched does not match the cached file')

    except Exception as e:
        exc_type, exc_value, exc_traceback = sys.exc_info()
        print(f"Exception: {e}")
        traceback.print_tb(exc_traceback, file=sys.stdout)



