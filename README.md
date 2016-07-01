# fatty
Tool for testing servers's max header and body size

# Usage

```
  -b	Test body size
  -dest string
    	Destination of a testing server
  -h	Test header size
  -i int
    	Testing value initial size (default 1)
  -m int
    	Max number of requests to perform [0 = limitless]
  -method string
    	Request method
  -proxy string
    	Destination of a proxy server
  -proxy-pass string
    	Proxy password
  -proxy-user string
    	Proxy username
  -r int
    	Testing value increase ratio (default 2)
  -strategy string
    	Testing value groving strategy [linear,expo]
  -v	Verbosive output
```
