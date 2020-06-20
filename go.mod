module bitbucket.org/grayll/grayll.io-user-app-back-end

go 1.12

//replace github.com/huyntsgs/stellar-service v0.0.0-20200511152020-7a130845cf0d => /home/bc/go/src/github.com/huyntsgs/stellar-service

//replace github.com/huyntsgs/cors v1.3.2-0.20200524025249-9865eda97561 => /home/bc/go/src/github.com/huyntsgs/cors

require (
	cloud.google.com/go v0.51.0
	cloud.google.com/go/firestore v1.1.0
	cloud.google.com/go/pubsub v1.0.1
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
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dgryski/dgoogauth v0.0.0-20190221195224-5a805980a5f3
	github.com/fatih/structs v1.0.0
	//github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.6.2
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/golang/protobuf v1.3.3
	github.com/google/uuid v1.1.1 // indirect
	github.com/huyntsgs/cors v1.3.2-0.20200526010755-d0b894e7ee83
	github.com/huyntsgs/hermes v0.0.0-20191119075450-4a9f1c2a64a0
	github.com/huyntsgs/stellar-service v0.0.0-20200511152020-7a130845cf0d

	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jaytaylor/html2text v0.0.0-20190408195923-01ec452cbe43 // indirect
	github.com/jinzhu/now v1.1.1
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.3.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.2 // indirect
	github.com/sendgrid/rest v2.4.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	//github.com/stellar/go v0.0.0-20200428193902-20797e3e2f1a

	github.com/stellar/go v0.0.0-20190920224012-d72ea298f1e9
	go.uber.org/goleak v0.10.0 // indirect
	golang.org/x/crypto v0.0.0-20191112222119-e1110fd1c708
	google.golang.org/api v0.15.0
	google.golang.org/genproto v0.0.0-20200117163144-32f20d992d24
	google.golang.org/grpc v1.26.0
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
