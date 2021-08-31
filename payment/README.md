
## Setup redis
sudo apt-get update
sudo apt-get install redis-server
ps -f -u redis
sudo tail -f /var/log/redis/redis-server.log

## Set root pwd for compute engine:
sudo passwd => prompt enter passwd
su - => switch to root

sudo chown redis:redis /var/lib/redis
sudo chmod 770 /var/lib/redis
echo 'vm.overcommit_memory = 1' >> /etc/sysctl.conf
sysctl vm.overcommit_memory=1

https://cloud.google.com/community/tutorials/setting-up-redis

# The filename where to dump the DB
dbfilename dump.rdb
# The working directory.
#
# The DB will be written inside this directory, with the filename specified
# above using the 'dbfilename' configuration directive.
#
# The Append Only File will also be created inside this directory.
#
# Note that you must specify a directory here, not a file name.
dir /var/lib/redis

https://stackoverflow.com/questions/22160753/redis-failed-opening-rdb-for-saving-permission-denied/28686802#28686802

## Redis known issue
/etc/redis/redis.conf

- err:  MISCONF Redis is configured to save RDB snapshots, but is currently not able to persist on disk. Commands that may modify the data set are disabled. Please check Redis logs for details about the error.
https://github.com/antirez/redis/issues/584

### Fix 1:
sudo chown redis:redis /var/lib/redis
sudo chmod 770 /var/lib/redis
echo 'vm.overcommit_memory = 1' >> /etc/sysctl.conf
sysctl vm.overcommit_memory=1

### Temporarily fix:
- ssh to computer engine
grayll@redis-server:~$ redis-cli
127.0.0.1:6379> config set stop-writes-on-bgsave-error no

## Copy streaming server to computer Engine
gcloud config set project grayll-app-f3f3f3

gcloud compute scp payment redis-server:/home/bc
gcloud compute scp trade redis-server:/home/bc
gcloud compute scp streaming1.service redis-server:/home/bc
gcloud compute scp trade1.service redis-server:/home/bc

gcloud compute scp config1.json redis-server:/home/bc
gcloud compute scp grayll-gry-1-balthazar-firebase-adminsdk.json redis-server:/home/bc


gcloud compute scp config1.json redis-server:/home/bc

sudo systemctl stop streaming
sudo cp /home/bc/streaming .

curl http://localhost:8080/debug/pprof/heap > heap.pprof

gcloud compute scp redis-server:/home/bc/heap.pprof .
