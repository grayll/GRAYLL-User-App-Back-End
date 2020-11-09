## Fix issue: build bitbucket.org/grayll/grayll.io-user-app-back-end: cannot load gopkg.in/russross/blackfriday.v2: cannot find module providing package gopkg.in/russross/blackfriday.v2

=> go mod edit -replace=gopkg.in/russross/blackfriday.v2@v2.0.1=github.com/russross/blackfriday/v2@v2.0.1

## Build and deploy
gcloud config set project grayll-app-f3f3f3 &&
gcloud builds submit --tag gcr.io/grayll-app-f3f3f3/grayll-app &&
gcloud beta run deploy --image gcr.io/grayll-app-f3f3f3/grayll-app --platform managed --region us-central1 --vpc-connector app-connector --set-env-vars  REDIS_HOST=10.128.0.6,SERVER=prod,SUPER_ADMIN_ADDRESS=,SUPER_ADMIN_SEED=

gcloud config set project grayll-app-f3f3f3 &&
gcloud builds submit --tag gcr.io/grayll-app-f3f3f3/grayll-app-dev &&
gcloud beta run deploy --image gcr.io/grayll-app-f3f3f3/grayll-app-dev --platform managed --region us-central1 --set-env-vars REDIS_HOST=10.128.0.6,SERVER=dev SELLING_PRICE=0.3,SERVER=prod,SUPER_ADMIN_ADDRESS=,SUPER_ADMIN_SEED=,SELLING_PERCENT=

## Federation project
gcloud config set project grayll-federation
gcloud functions deploy Query --runtime go111 --trigger-http

## Copy stellar horizon
gcloud compute scp ./stellar-core.cfg stellar-node:~
gcloud compute scp ./horizon stellar-node:~
gcloud compute scp ./horizon horizon-node:~
## Cloud task
gcloud tasks queues update xlm-loan-reminder --max-attempts=2

gcloud tasks queues create data-report --max-attempts=7200 --min-backoff=300s --max-backoff=3600s --max-doublings=16 --max-retry-duration=432000s --max-dispatches-per-second=500 --max-concurrent-dispatches=5000

## Set cloud run env
--set-env-vars
--update-env-vars
--remove-env-vars
--clear-env-vars
For example to add or update the environment variables:
superAdminAddress := os.Getenv("SUPER_ADMIN_ADDRESS")
	superAdminSeed := os.Getenv("SUPER_ADMIN_SEED")
	sellingPrice := os.Getenv("SELLING_PRICE")
	sellingPercent := os.Getenv("SELLING_PERCENT")

gcloud compute ssl-certificates describe horizon-lb-ssl \
    --global \
    --format="get(name,managed.status)"

gcloud run services update [SERVICE] --set-env-vars KEY1=VALUE1,KEY2=VALUE2
gcloud run deploy [SERVICE] --image gcr.io/[PROJECT-ID]/[IMAGE] --update-env-vars KEY1=VALUE1,KEY2=VALUE2