// Go mod init
module github.com/curol/network

// Go Version
go 1.21.0

// Workspace modules for `go.work`
// Example:
//     use ./util

// Dependencies
require (
	github.com/go-redis/redis/v7 v7.4.1
	github.com/google/uuid v1.5.0
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/CloudyKit/fastprinter v0.0.0-20200109182630-33d98a066a53 // indirect
	github.com/CloudyKit/jet/v4 v4.1.0 // indirect
	github.com/CloudyKit/jet/v6 v6.2.0 // indirect
	github.com/Joker/jade v1.1.3 // indirect
	github.com/Shopify/goreferrer v0.0.0-20220729165902-8cddb4f5de06 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/aymerick/raymond v2.0.3-0.20180322193309-b565731e1464+incompatible // indirect
	github.com/caddyserver/certmagic v0.20.0 // indirect
	github.com/chris-ramon/douceur v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/duke-git/lancet v1.4.2 // indirect
	github.com/duke-git/lancet/v2 v2.2.8 // indirect
	github.com/eknkc/amber v0.0.0-20171010120322-cdade1c07385 // indirect
	github.com/fatih/color v1.15.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/flosch/pongo2/v4 v4.0.2 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gomarkdown/markdown v0.0.0-20231222211730-1d6d20845b47 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gorilla/css v1.0.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/imkira/go-interpol v1.1.0 // indirect
	github.com/iris-contrib/httpexpect/v2 v2.15.2 // indirect
	github.com/iris-contrib/jade v1.1.4 // indirect
	github.com/iris-contrib/pongo2 v0.0.1 // indirect
	github.com/iris-contrib/schema v0.0.6 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kataras/blocks v0.0.8 // indirect
	github.com/kataras/golog v0.1.11 // indirect
	github.com/kataras/iris v0.0.2 // indirect
	github.com/kataras/iris/v12 v12.2.10 // indirect
	github.com/kataras/pio v0.0.13 // indirect
	github.com/kataras/sitemap v0.0.6 // indirect
	github.com/kataras/tunnel v0.0.4 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/libdns/libdns v0.2.1 // indirect
	github.com/mailgun/raymond/v2 v2.0.48 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/mholt/acmez v1.2.0 // indirect
	github.com/microcosm-cc/bluemonday v1.0.26 // indirect
	github.com/miekg/dns v1.1.55 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ryanuber/columnize v2.1.0+incompatible // indirect
	github.com/sanity-io/litter v1.5.5 // indirect
	github.com/schollz/closestmatch v2.1.0+incompatible // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/tdewolff/minify/v2 v2.20.14 // indirect
	github.com/tdewolff/parse/v2 v2.7.8 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser v0.1.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yosssi/ace v0.0.5 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.26.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/exp v0.0.0-20240112132812-db7319d0e0e3 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.17.0 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	moul.io/http2curl/v2 v2.3.0 // indirect
)
