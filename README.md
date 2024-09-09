<div align="center">
<a href="https://pkg.go.dev/github.com/lostdusty/gobalt" title="Go API Reference" rel="nofollow"><img src="https://img.shields.io/badge/go-documentation-blue.svg?style=for-the-badge" alt="Go API Reference"></a>
<a href="https://github.com/lostdusty/gobalt/releases/lastest" title="Latest Release" rel="nofollow"><img src="https://img.shields.io/github/v/release/lostdusty/gobalt?include_prereleases&style=for-the-badge" alt="Latest Release"></a>

[![Static Badge](https://img.shields.io/badge/cobalt_discord-join-blue?style=for-the-badge&logo=discord)](https://discord.gg/pQPt8HBUPu)
[![Static Badge](https://img.shields.io/badge/supported-services-0077b6?style=for-the-badge&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA2NDAgNTEyIj48IS0tIUZvbnQgQXdlc29tZSBGcmVlIDYuNS4xIGJ5IEBmb250YXdlc29tZSAtIGh0dHBzOi8vZm9udGF3ZXNvbWUuY29tIExpY2Vuc2UgLSBodHRwczovL2ZvbnRhd2Vzb21lLmNvbS9saWNlbnNlL2ZyZWUgQ29weXJpZ2h0IDIwMjQgRm9udGljb25zLCBJbmMuLS0%2BPHBhdGggZD0iTTU3OS44IDI2Ny43YzU2LjUtNTYuNSA1Ni41LTE0OCAwLTIwNC41Yy01MC01MC0xMjguOC01Ni41LTE4Ni4zLTE1LjRsLTEuNiAxLjFjLTE0LjQgMTAuMy0xNy43IDMwLjMtNy40IDQ0LjZzMzAuMyAxNy43IDQ0LjYgNy40bDEuNi0xLjFjMzIuMS0yMi45IDc2LTE5LjMgMTAzLjggOC42YzMxLjUgMzEuNSAzMS41IDgyLjUgMCAxMTRMNDIyLjMgMzM0LjhjLTMxLjUgMzEuNS04Mi41IDMxLjUtMTE0IDBjLTI3LjktMjcuOS0zMS41LTcxLjgtOC42LTEwMy44bDEuMS0xLjZjMTAuMy0xNC40IDYuOS0zNC40LTcuNC00NC42cy0zNC40LTYuOS00NC42IDcuNGwtMS4xIDEuNkMyMDYuNSAyNTEuMiAyMTMgMzMwIDI2MyAzODBjNTYuNSA1Ni41IDE0OCA1Ni41IDIwNC41IDBMNTc5LjggMjY3Ljd6TTYwLjIgMjQ0LjNjLTU2LjUgNTYuNS01Ni41IDE0OCAwIDIwNC41YzUwIDUwIDEyOC44IDU2LjUgMTg2LjMgMTUuNGwxLjYtMS4xYzE0LjQtMTAuMyAxNy43LTMwLjMgNy40LTQ0LjZzLTMwLjMtMTcuNy00NC42LTcuNGwtMS42IDEuMWMtMzIuMSAyMi45LTc2IDE5LjMtMTAzLjgtOC42Qzc0IDM3MiA3NCAzMjEgMTA1LjUgMjg5LjVMMjE3LjcgMTc3LjJjMzEuNS0zMS41IDgyLjUtMzEuNSAxMTQgMGMyNy45IDI3LjkgMzEuNSA3MS44IDguNiAxMDMuOWwtMS4xIDEuNmMtMTAuMyAxNC40LTYuOSAzNC40IDcuNCA0NC42czM0LjQgNi45IDQ0LjYtNy40bDEuMS0xLjZDNDMzLjUgMjYwLjggNDI3IDE4MiAzNzcgMTMyYy01Ni41LTU2LjUtMTQ4LTU2LjUtMjA0LjUgMEw2MC4yIDI0NC4zeiIvPjwvc3ZnPg%3D%3D&logoColor=0077b6)](https://github.com/wukko/cobalt?tab=readme-ov-file#supported-services)
<h1>gobalt v1</h1>
<p> new update, see v2 branch </p>
</div>

> [!IMPORTANT]  
> Cobalt API recently had a breaking change update, if you want to keep using cobalt's api, we recommend [updating this package to version 2](https://github.com/lostdusty/gobalt/tree/v2). [(see changes)](https://github.com/imputnet/cobalt/commit/08bc5022a1170c2f739ae468a9b787a56e504887)

Gobalt provides a way to communicate with [cobalt.tools](https://cobalt.tools) using Go. To use it in your projects, simply run this command:
```sh
go get github.com/lostdusty/gobalt
```

## Usage
First, make sure you call `gobalt.CreateDefaultSettings()` to create the `Settings` struct with default values, then set an url.

```go
//Creates a Settings struct with default values, and save it to downloadMedia variable.
downloadMedia := gobalt.CreateDefaultSettings()

//Sets the URL, you MUST set one before downloading the media.
downloadMedia.Url = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

//After changing the url, Do() will make the necessary requests to cobalt to download your media
destination, err := gobalt.Do(downloadMedia)
if err != nil {
    //Handle errors here
}

//Prints out the url from cobalt to download the requested media.
fmt.Println(destination.URL)
//Output example: https://us4-co.wuk.sh/api/stream?t=wTn-71aaWAcV2RBejNFNV&e=1711777798864&h=fMtwwtW8AUmtTLB24DGjJJjrq1EJDBFaCDoDuZpX0pA&s=_YVZhz8fnzBBKKo7UZmGWOqfe4wWwH5P1azdgBqwf-I&i=6tGZqAXbW08_6KmAiLevZA
```

## Features
### Server info
You can query information about any cobalt server by using `CobaltServerInfo(apiurl)`. This will return a `ServerInfo` struct with [this info](https://github.com/wukko/cobalt/blob/current/docs/api.md#response-body-variables-1).

Example code:
```go
server, err := gobalt.CobaltServerInfo(gobalt.CobaltApi)
if err != nil {
	 //Handle the error here
}
fmt.Printf("Downloading from %v!\n", server.URL)
//Output: Downloading from https://us4-co.wuk.sh/!
```

### Use/Query other cobalt instances
Using `GetCobaltInstances()` fetches a community maintaned list of third-party cobalt instances ran by the community, except for ``*-co.wuk.sh``, none of them are "official" cobalt instances. Use them if you can't download from the main instance for whatever reason.

Example:
```go
cobalt, err := gobalt.GetCobaltInstances()
if err != nil { 
    //Handle errors here
}
fmt.Printf("Found %v online cobalt servers: ", len(cobalt))
for _, value := range cobalt {
	fmt.Printf("%v, ", value.URL)
}
/* Output: 
* Found 8 online cobalt servers: co.wuk.sh, cobalt-api.hyper.lol, cobalt.api.timelessnesses.me, api-dl.cgm.rs, cobalt.synzr.space, capi.oak.li, co.tskau.team, api.co.rooot.gay
*/
```
