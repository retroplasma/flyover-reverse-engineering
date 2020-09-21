<img width="100%" alt="Santa Monica Pier" title="Santa Monica Pier" src="https://user-images.githubusercontent.com/46618410/53480398-0a4d5100-3a73-11e9-983e-99f24ebdf674.png">

This is an attempt to reverse-engineer *Flyover* (= 3D satellite mode) from Apple Maps. The goal is to document the results and to provide code that emerges. Similar work is being done for Google Earth [here](https://github.com/retroplasma/earth-reverse-engineering).

#### Status
Roughly speaking, these parts have been figured out:
- bootstrap of manifests
- URL structure
- authentication algorithm
- map tiling and conversion from geo coordinates
- mesh decompression (huffman tables, edgebreaker variant etc.)
- tile lookup using octree

We can authenticate URLs and retrieve textured 3D models from given coordinates (latitude, longitude).

#### General
Data is stored in map tiles. These five tile styles are used for Flyover:

|Type  | Purpose                                     | URL structure                                        |
|------|---------------------------------------------|------------------------------------------------------|
|C3M   | Texture, Mesh, Transformation(, Animation)  | 🅐(?\|&)style=15&v=⓿&region=❶&x=❷&y=❸&z=❹&h=❺    |
|C3MM 1| Metadata                                    | 🅐(?\|&)style=14&v=⓿&part=❻&region=❶                |   
|C3MM 2| Metadata                                    | 🅐(?\|&)style=52&v=⓿&region=❶&x=❷&y=❸&z=❹&h=❺    |   
|DTM 1 | Terrain/Surface/Elevation                   | 🅐(?\|&)style=16&v=⓿&region=❶&x=❷&y=❸&z=❹         |
|DTM 2 | Terrain/Surface/Elevation                   | 🅐(?\|&)style=17&v=⓿&size=❼&scale=❽&x=❷&y=❸&z=❹  |

- 🅐: URL prefix from resource manifest
- ⓿: Version from resource manifest or altitude manifest using region
- ❶: Region ID from altitude manifest
- ❷❸❹: Map tile numbers ([tiled web map](https://en.wikipedia.org/wiki/Tiled_web_map) scheme)
- ❺: Height/altitude index. Probably from C3MM
- ❻: Incremental part number
- ❼❽: Size/scale. Not sure where its values come from

#### Resource hierarchy
```
ResourceManifest
└─ AltitudeManifest
   ├─ C3MM
   │  └─ C3M
   └─ DTM?
```
Focusing on C3M(M) for now. DTMs are images with a footer and are probably used for the [grid](https://user-images.githubusercontent.com/46618410/53483243-fdcbf700-3a78-11e9-8fc0-ad6cfa8c57cd.png) that is displayed when Maps is loading.

#### Code
This repository is structured as follows:

|Directory           | Description                  |
|--------------------|------------------------------|
|[cmd](./cmd)        | command line programs        |
|[pkg](./pkg)        | most of the actual code      |
|[proto](./proto)    | protobuf files               |
|[scripts](./scripts)| additional scripts           |
|[vendor](./vendor)  | dependencies                 |

##### Install
Install [Go](https://golang.org/) and run `go get github.com/retroplasma/flyover-reverse-engineering`. Then edit [config.json](config.json) in `$GOPATH/src/github.com/retroplasma/flyover-reverse-engineering`:
- automatically on macOS:
  - `./scripts/get_config_macos.sh > config.json`
- or manually:
  - `resourceManifestURL`: from [com.apple.GEO.plist](#files-on-macos) or [GeoServices](#files-on-macos) binary
  - `tokenP1`: from [GeoServices](#files-on-macos) binary (function: `GEOURLAuthenticationGenerateURL`)

##### Command line programs
Here are some independent command line programs that use code from [pkg](./pkg):

###### Export OBJ (proof of concept)

Usage:
```
go run cmd/poc-export-obj/main.go [lat] [lon] [zoom] [tryXY] [tryH]

Parameter   Description       Example
--------------------------------------
lat         Latitude          34.007603
lon         Longitude         -118.499741
zoom        Zoom (~ 13-20)    20
tryXY       Area scan         3
tryH        Altitude scan     40
```

This exports Santa Monica Pier to `./downloaded_files/obj/...`:
```
go run cmd/poc-export-obj/main.go 34.007603 -118.499741 20 3 40
```

Optional: Center-scale OBJ using node.js script:
```
node scripts/center_scale_obj.js
```

In Blender (compatible tutorial [here](https://github.com/retroplasma/earth-reverse-engineering/blob/1dd24a723513d7e96f249e2c635416d4596992c4/BLENDER.md)):

<img src="https://user-images.githubusercontent.com/46618410/65068957-fa06b000-d989-11e9-9091-1e71874b0b0c.png" width="300px">


###### Authenticate URLs
This authenticates a URL using parameters from `config.json`:
```
go run cmd/auth/main.go [url]
```

###### Parse C3M file
This parses a C3M v3 file, decompresses meshes, reads JPEG textures and produces a struct that contains a textured 3d model:
```
go run cmd/parse-c3m/main.go [file]
```

###### Parse C3MM file
This parses a C3MM v1 file. The C3MM files in a region span octrees whose roots are indexed in the first file.
```
go run cmd/parse-c3mm/main.go [file] [[file_number]]
```

#### Files on macOS
- `~/Library/Preferences/com.apple.GEO.plist`
  - last resource manifest url
- `~/Library/Caches/GeoServices/Resources/altitude-*.xml`
  - defines regions for c3m urls
  - `altitude-*.xml` url in resource manifest
- `~/Library/Containers/com.apple.geod/Data/Library/Caches/com.apple.geod/MapTiles/MapTiles.sqlitedb`
  - local map tile cache
- `/System/Library/PrivateFrameworks/GeoServices.framework/GeoServices`
  - resource manifest base url, networking, caching, authentication
- `/System/Library/PrivateFrameworks/VectorKit.framework/VectorKit`
  - parsers, decoders
- `/System/Library/PrivateFrameworks/GeoServices.framework/XPCServices/com.apple.geod.xpc`
  - loads `GeoServices`
- `/Applications/Maps.app/Contents/MacOS/Maps`
  - loads `VectorKit`

#### Important
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
