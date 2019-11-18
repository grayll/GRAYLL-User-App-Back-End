# Build & publish cloud run

sudo gcloud projects list

gcloud config set project grayll-system-test

gcloud pubsub topics create gryprice
gcloud pubsub topics create grzprice

gcloud pubsub subscriptions create mySubscription --topic myTopic --push-endpoint="https://myapp.appspot.com/push"

   gcloud pubsub subscriptions list
   gcloud pubsub topics list-subscriptions myTopic
   gcloud pubsub subscriptions delete mySubscription

gcloud builds submit --tag gcr.io/grayll-system-test/gryz

- Build docker image
sudo gcloud beta run deploy pubsub-tutorial --image gcr.io/grayll-system-test/gryz

- Deploy and run
sudo gcloud beta run deploy gryz --image gcr.io/grayll-system-test/gryz