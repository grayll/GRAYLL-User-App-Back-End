## Fix issue: build bitbucket.org/grayll/grayll.io-user-app-back-end: cannot load gopkg.in/russross/blackfriday.v2: cannot find module providing package gopkg.in/russross/blackfriday.v2

=> go mod edit -replace=gopkg.in/russross/blackfriday.v2@v2.0.1=github.com/russross/blackfriday/v2@v2.0.1

## Build and deploy
gcloud config set project grayll-app-f3f3f3

gcloud builds submit --tag gcr.io/grayll-app-f3f3f3/grayll-app

gcloud beta run deploy --image gcr.io/grayll-app-f3f3f3/grayll-app --platform managed

gcloud config set project grayll-federation
gcloud functions deploy Query --runtime go111 --trigger-http
