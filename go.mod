module bitbucket.org/grayll/grayll.io-user-app-back-end

go 1.13

//replace github.com/huyntsgs/stellar-service v0.0.0-20200511152020-7a130845cf0d => /home/bc/go/src/github.com/huyntsgs/stellar-service

//replace github.com/huyntsgs/cors v1.3.2-0.20200524025249-9865eda97561 => /home/bc/go/src/github.com/huyntsgs/cors

replace bitbucket.org/ww/goautoneg v0.0.0-20120707110453-75cd24fc2f2c => /home/bc/go/src/bitbucket.org/huykbc/goautoneg

require (
	//cloud.google.com/go v0.66.0
	cloud.google.com/go v0.51.0

	cloud.google.com/go/firestore v1.3.0
	cloud.google.com/go/pubsub v1.6.2
	firebase.google.com/go v3.12.0+incompatible
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/NeverBounce/NeverBounceApi-Go v0.0.0-20200129202202-ff62cbbd83aa
	github.com/SherClockHolmes/webpush-go v1.1.0
	github.com/algolia/algoliasearch-client-go/v3 v3.6.0
	github.com/antigloss/go v0.0.0-20200109080012-05d5d0918164
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/avct/uasurfer v0.0.0-20191028135549-26b5daa857f1
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/daemonfire300/azure-sdk-for-go v0.0.0-20160604125511-3b8a7044d504
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/dgoogauth v0.0.0-20190221195224-5a805980a5f3
	github.com/fatih/structs v1.0.0
	//github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.6.2
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/go-redis/redis/v7 v7.2.0
	github.com/go-sql-driver/mysql v1.4.0
	github.com/golang/protobuf v1.4.2
	github.com/google/uuid v1.1.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/huyntsgs/cors v1.3.2-0.20200526010755-d0b894e7ee83
	github.com/huyntsgs/hermes v0.0.0-20191119075450-4a9f1c2a64a0
	github.com/huyntsgs/stellar-service v0.0.0-20200811102404-492829a6952c

	//github.com/huyntsgs/stellar-service v0.0.0-20200511152020-7a130845cf0d

	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jaytaylor/html2text v0.0.0-20190408195923-01ec452cbe43 // indirect
	github.com/jinzhu/now v1.1.1
	github.com/kr/pty v1.1.8 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.3.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.2 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/stellar/go v0.0.0-20190920224012-d72ea298f1e9
	github.com/ulule/limiter/v3 v3.5.0
	go.uber.org/goleak v0.10.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	google.golang.org/api v0.32.0
	google.golang.org/genproto v0.0.0-20200921151605-7abf4a1a14d5
	google.golang.org/grpc v1.32.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
