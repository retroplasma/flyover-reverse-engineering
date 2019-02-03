
This is an attempt to reverse-engineer the Flyover feature in Apple Maps. Main goal is to document the results and to provide code that emerges.

#### Authenticate URLs
```
go run auth.go [url] [session_id] [token_1] [token_2]
```
- `session_id`: 40 digits
- `token_1`: see GeoServices binary (function: GEOURLAuthenticationGenerateURL)
- `token_2`: see Geo Resource Manifest config (protobuf field 30)

#### Files on macOS
- `~/Library/Preferences/com.apple.GEO.plist`
- `/System/Library/PrivateFrameworks/GeoServices.framework/GeoServices`
- `/System/Library/PrivateFrameworks/VectorKit.framework/VectorKit`
- `/System/Library/PrivateFrameworks/GeoServices.framework/XPCServices/com.apple.geod.xpc`
