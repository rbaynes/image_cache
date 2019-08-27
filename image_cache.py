#!/usr/bin/env python3.6

# rbaynes Aug. 22, 2019
# Class to fetch files over HTTP and keep a local cache of them.

import sys, traceback, argparse, http.client, hashlib
from typing import Dict, ClassVar, Any
from utils.hash import Hash
from utils.cache import Cache


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
        if etag is not None:
            self.cache.set(URL, FileFetchAndCache.__etag, etag)
        last_mod = resp.getheader(FileFetchAndCache.__last_modified)
        if last_mod is not None:
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
        URL1 = '/images/landing/Rollercoaster/whatsroblox_12072017.jpg'
        URL2 = '/images/landing/Rollercoaster/gameimage3_12072017.jpg' 
        URL3 = '/images/landing/Rollercoaster/devices_people_12072017.png'

        # Command line arg parsing
        parser = argparse.ArgumentParser(description='file fetch and cache')
        parser.add_argument('-v', '--verbose', dest='verbose', 
                action='store_true', 
                help='Enable verbose logging')
        parser.add_argument('-H', '--host', type=str, default=HOST, 
                help=f'Host in hostname:port format.  Default is {HOST}')
        parser.add_argument('-U1', '--URL1', type=str, default=URL1, 
                help=f'URL1 to fetch.  Default is {URL1}')
        parser.add_argument('-U2', '--URL2', type=str, default=URL2, 
                help=f'URL2 to fetch.  Default is {URL2}')
        parser.add_argument('-U3', '--URL3', type=str, default=URL3, 
                help=f'URL3 to fetch.  Default is {URL3}')
        args = parser.parse_args()
        
        # Our fetch and cache class
        ffac = FileFetchAndCache(verbose=args.verbose)

        # Fetch and cache the files.
        files = [args.URL1, args.URL2, args.URL3]
        fetched_file_hashes = [None] * len(files)
        cached_file_hashes = [None] * len(files)
        for i in range(len(files)): # Fetch each file.
            for _ in range(2): # Fetch each two times, to test cache.
                success, from_cache, file_bytes = ffac.get(args.host, files[i])
                if success and not from_cache:
                    fetched_file_hashes[i] = Hash.md5(file_bytes)
                    print(f'Fetched and cached {files[i]}')
                elif success and from_cache:
                    cached_file_hashes[i] = Hash.md5(file_bytes)
                    print(f'Cache hit for {files[i]}')
                else:
                    print(f'Error: Did not fetch file {files[i]}')
                    break
            print()

        # Validate
        for i in range(len(files)): 
            if fetched_file_hashes[i] != cached_file_hashes[i]:
                print('Error: hashes do not match.')

    except Exception as e:
        exc_type, exc_value, exc_traceback = sys.exc_info()
        print(f"Exception: {e}")
        traceback.print_tb(exc_traceback, file=sys.stdout)



