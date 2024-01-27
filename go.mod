module github.com/carbocation/traceoverlay

go 1.21.0

require (
	cloud.google.com/go/storage v1.37.0
	github.com/carbocation/genomisc v0.0.0-20221110225648-66a475457014
	github.com/davecgh/go-spew v1.1.1
	github.com/gorilla/mux v1.8.1
	github.com/interpose/middleware v0.0.0-20150216143757-05ed56ed52fa
	github.com/jmoiron/sqlx v1.3.5
	github.com/justinas/alice v1.2.0
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/minio/blake2b-simd v0.0.0-20160723061019-3f5f724cb5b1
	golang.org/x/image v0.15.0
)

require (
	cloud.google.com/go v0.112.0 // indirect
	cloud.google.com/go/compute v1.23.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v1.1.5 // indirect
	github.com/carbocation/go-quantize v0.0.0-20220308192728-857cc7c8fdfc // indirect
	github.com/carbocation/handlers v0.0.0-20140528190747-c939c6d9ef31 // indirect
	github.com/carbocation/interpose v0.0.0-20161206215253-723534742ba3 // indirect
	github.com/carbocation/pfx v0.0.0-20230108194214-fcea663adae5 // indirect
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/csimplestring/go-csv v0.0.0-20180328183906-5b8b3cd94f2c // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-martini/martini v0.0.0-20170121215854-22fa46961aab // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/goods/httpbuf v0.0.0-20120503183857-5709e9bb814c // indirect
	github.com/google/s2a-go v0.1.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.2 // indirect
	github.com/googleapis/gax-go/v2 v2.12.0 // indirect
	github.com/justinas/nosurf v1.1.1 // indirect
	github.com/krolaw/zipstream v0.0.0-20180621105154-0a2661891f94 // indirect
	github.com/meatballhat/negroni-logrus v0.0.0-20201129033903-bc51654b0848 // indirect
	github.com/phyber/negroni-gzip v1.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/suyashkumar/dicom v0.4.6-0.20200816032854-6ffe547e2a08 // indirect
	github.com/theodesp/unionfind v0.0.0-20200112172429-2bf90fd5b8c5 // indirect
	github.com/tj/go-rle v0.0.0-20180508204109-877ab66bb189 // indirect
	github.com/urfave/negroni v1.0.0 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.47.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.47.0 // indirect
	go.opentelemetry.io/otel v1.22.0 // indirect
	go.opentelemetry.io/otel/metric v1.22.0 // indirect
	go.opentelemetry.io/otel/trace v1.22.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/oauth2 v0.16.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	google.golang.org/api v0.159.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/grpc v1.61.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

// This is done since otherwise, because the dicom module was fully rewritten
// after v0.4.6, go mod tidy will crash if you don't specify the version. This
// is not due to any *direct* imports, but oddly it is due to *indirect* imports
// from modules that have already correctly specified the version in their own
// go module files.
replace github.com/suyashkumar/dicom => github.com/suyashkumar/dicom v0.4.6-0.20200816032854-6ffe547e2a08

// Can be updated on a periodic basis by simply using 'main' or 'master' in
// place of the version string
replace github.com/carbocation/genomisc => github.com/carbocation/genomisc-prerelease v0.0.0-20240124192534-69d3ebe47159
