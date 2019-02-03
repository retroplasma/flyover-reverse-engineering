
This is an attempt to reverse-engineer the Flyover feature in Apple Maps. Main goal is to document the results and to provide code that emerges.

#### Motivation
Differences between Google Earth and Apple Flyover. Extreme example:

<img src="https://user-images.githubusercontent.com/46618410/52183147-db89e500-27fc-11e9-9c75-fc78ff6cda58.jpg" alt="Google" title="Google"  width=50%><img src="https://user-images.githubusercontent.com/46618410/52183145-d62c9a80-27fc-11e9-9396-2d0acb34ec03.jpg" alt="Apple" title="Apple" width=50%>

#### Authenticate URLs
```
go run auth.go [url] [session_id] [token_1] [token_2]
```
- `session_id`: 40 digits
- `token_1`: see GeoServices binary (function: GEOURLAuthenticationGenerateURL)
- `token_2`: see Geo Resource Manifest config (protobuf field 30)

#### Files on macOS
- `~/Library/Preferences/com.apple.GEO.plist`
- `~/Library/Containers/com.apple.geod/Data/Library/Caches/com.apple.geod/MapTiles/MapTiles.sqlitedb`
- `/System/Library/PrivateFrameworks/GeoServices.framework/GeoServices`
- `/System/Library/PrivateFrameworks/VectorKit.framework/VectorKit`
- `/System/Library/PrivateFrameworks/GeoServices.framework/XPCServices/com.apple.geod.xpc`
