
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
gcloud compute scp ./streaming redis-server:~
gcloud compute scp ./trade redis-server:~

gcloud compute scp ./horizon  instance-group-horizon-nzvk:~

gcloud compute scp ./grayll-grz-arkady-firebase-adminsdk-9q3s2-3fb5715c06.json redis-server:~

sudo systemctl daemon-reload
sudo systemctl start trade
sudo systemctl stop trade
sudo systemctl stop streaming
sudo systemctl start streaming

sudo cp /home/bc/trade .
sudo cp /home/bc/streaming .

== ulimit

The ulimit command by default changes the HARD limits, which you (a user) can lower, but cannot raise.

Use the -S option to change the SOFT limit, which can range from 0-{HARD}.

I have actually aliased ulimit to ulimit -S, so it defaults to the soft limits all the time.

alias ulimit='ulimit -S'
As for your issue, you're missing a column in your entries in /etc/security/limits.conf.

There should be FOUR columns, but the first is missing in your example.

* soft nofile 4096
* hard nofile 4096
The first column describes WHO the limit is to apply for. '*' is a wildcard, meaning all users. To raise the limits for root, you have to explicitly enter 'root' instead of '*'.

You also need to edit /etc/pam.d/common-session* and add the following line to the end:

session required pam_limits.so

=======

check max open file
cat /run/redis/redis-server.pid
933
cat /proc/933/limits

lsof -p 933
lsof | awk '{print $1}' | sort | uniq -c | sort -r | head