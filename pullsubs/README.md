# Build & publish cloud run

sudo gcloud projects list

gcloud config set project grayll-system-test

gcloud pubsub topics create gryprice
gcloud pubsub topics create grzprice

gcloud pubsub subscriptions create mySubscription --topic myTopic --push-endpoint="https://myapp.appspot.com/push"

   gcloud pubsub subscriptions list
   gcloud pubsub topics list-subscriptions myTopic
   gcloud pubsub subscriptions delete mySubscription

gcloud builds submit --tag gcr.io/grayll-app-f3f3f3/pullsubscription

- Build docker image
gcloud beta run deploy --image gcr.io/grayll-app-f3f3f3/pullsubscription --platform managed

- Deploy and run
sudo gcloud beta run deploy pubsub-tutorial --image gcr.io/grayll-app-f3f3f3/pullsubscription

gcloud compute scp ./pullsubscription.service pull-subscription:~

$ sudo systemctl daemon-reload
Start the service with:

$ sudo systemctl start pullsubscription
$ sudo systemctl enable pullsubscription
$ systemctl status pullsubscription