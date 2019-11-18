gcloud config set project grayll-app-f3f3f3

`gcloud builds submit --tag gcr.io/grayll-app-f3f3f3/grayll-app`

`gcloud beta run deploy --image gcr.io/grayll-app-f3f3f3/grayll-app --platform managed`

gcloud config set project grayll-federation
gcloud functions deploy Query --runtime go111 --trigger-http
