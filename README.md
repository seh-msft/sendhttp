# sendhttp

Sendhttp reads in raw HTTP requests and replays the request, emitting the HTTP response. 

Written in [Go](https://golang.org/). 

## Build

	go build

## Limitations

- Does not support HTTP 2

## Usage

```
$ ./sendhttp -h
Usage of ./sendhttp:
  -B    omit the body in the response (JSON)
  -D    verbose debug output
  -T    populate response TLS information (JSON)
  -b string
        substitute request body, if any
  -d    is the substitute body base64-encoded?
  -i string
        file to read request from (if not stdin)
  -j    emit response as JSON
  -p string
        protocol to use for request (default "https")
$
```

## Examples

Given a request in a file `req.http` such as:

```
GET /foo/bar/img.svg HTTP/1.1
Host: something.somewhere
Connection: close
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36
Accept: image/avif,image/webp,image/apng,image/*,*/*;q=0.8
Sec-Fetch-Site: cross-site
Sec-Fetch-Mode: no-cors
Sec-Fetch-Dest: image
Referer: https://something.somewhere/
Accept-Encoding: gzip, deflate
Accept-Language: en-US,en;q=0.9


```

Replay the request:

```
$ sendhttp < req.http
HTTP/2.0 200 OK
Content-Length: 394
Access-Control-Allow-Origin: *
Content-Encoding: gzip
Content-Type: image/svg+xml
Date: Fri, 16 Oct 2020 22:40:17 GMT

<some kind of xml>
$
```

Substitute the body, a HTTP header, and a URL query paramter, emitting the response as JSON:

```
$ sendhttp -i req.http -b 'A Body Here' -j 'some: where' 'foo? bar'
{…}
$
```
