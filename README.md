<img width="100%" alt="img" src="https://user-images.githubusercontent.com/46618410/53480398-0a4d5100-3a73-11e9-983e-99f24ebdf674.png">

This is an attempt to reverse-engineer *Flyover* (= 3D satellite mode) from Apple Maps. Main goal is to document the results and to provide code that emerges.

#### Motivation
Noticed differences between Google Earth and Apple Flyover during [previous project](https://github.com/retroplasma/earth-reverse-engineering). Extreme example:

|Google Earth|Apple Flyover|
|------------|-------------|
|<img src="https://user-images.githubusercontent.com/46618410/52183147-db89e500-27fc-11e9-9c75-fc78ff6cda58.jpg" alt="Google" title="Google"  width=100%>|<img src="https://user-images.githubusercontent.com/46618410/52183145-d62c9a80-27fc-11e9-9396-2d0acb34ec03.jpg" alt="Apple" title="Apple" width=100%>|

#### General
Data in stored in map tiles. These four tile styles are used for flyover:

|Type  | Purpose                                     | URL structure                                        |
|------|---------------------------------------------|------------------------------------------------------|
|DTM 1 | Terrain/Surface/Elevation                   | 🅐(?\|&)style=16&v=⓿&region=❶&x=❹&y=❺&z=❻         |
|DTM 2 | Terrain/Surface/Elevation                   | 🅐(?\|&)style=17&v=⓿&size=❷&scale=❸&x=❹&y=❺&z=❻  |
|C3M   | Texture, Mesh, ECEF, Rotation, Scene Graph? | 🅐(?\|&)style=15&v=⓿&region=❶&x=❹&y=❺&z=❻&h=❼    |
|C3MM  | Metadata                                    | 🅐(?\|&)style=52&v=⓿&region=❶&x=❹&y=❺&z=❻&h=❼    |   

- 🅐: URL prefix from resource manifest
- ⓿: Version from resource manifest or altitude manifest using region
- ❶: From altitude manifest
- ❷❸: todo
- ❹❺❻: Map tile. Regular scheme?
- ❼: Height/altitude? Only for C3M and C3MM

Focusing on C3M(M) for now. DTMs are just images with a footer; they're probably used for the [grid](https://user-images.githubusercontent.com/46618410/53483243-fdcbf700-3a78-11e9-8fc0-ad6cfa8c57cd.png) that is displayed when Maps is loading.

#### Authenticate URLs
```
go run cmd/auth/main.go [url] [session_id] [token_1] [token_2]
```
- `session_id`: 40 digits
- `token_1`: see GeoServices binary (function: GEOURLAuthenticationGenerateURL)
- `token_2`: see Geo Resource Manifest config (protobuf field 30)

#### Parse C3M file
This parses a C3M v3 file. It decompresses meshes, reads JPEG textures and produces a struct that contains a textured 3d model:
```
go run cmd/parse-c3m/main.go [file]
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
