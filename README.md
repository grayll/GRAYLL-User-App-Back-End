## Fix issue: build bitbucket.org/grayll/grayll.io-user-app-back-end: cannot load gopkg.in/russross/blackfriday.v2: cannot find module providing package gopkg.in/russross/blackfriday.v2

=> go mod edit -replace=gopkg.in/russross/blackfriday.v2@v2.0.1=github.com/russross/blackfriday/v2@v2.0.1

## Build and deploy
gcloud config set project grayll-app-f3f3f3

gcloud builds submit --tag gcr.io/grayll-app-f3f3f3/grayll-app

gcloud beta run deploy --image gcr.io/grayll-app-f3f3f3/grayll-app --platform managed --set-env-vars SELLING_PRICE=0.3,SUPER_ADMIN_ADDRESS=,SUPER_ADMIN_SEED=,SELLING_PERCENT=

## Federation project
gcloud config set project grayll-federation
gcloud functions deploy Query --runtime go111 --trigger-http

## Copy stellar horizon
gcloud compute scp ./stellar-core.cfg stellar-node:~
gcloud compute scp ./horizon stellar-node:~

## Cloud task
gcloud tasks queues update xlm-loan-reminder --max-attempts=2

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


gcloud run services update [SERVICE] --set-env-vars KEY1=VALUE1,KEY2=VALUE2
gcloud run deploy [SERVICE] --image gcr.io/[PROJECT-ID]/[IMAGE] --update-env-vars KEY1=VALUE1,KEY2=VALUE2