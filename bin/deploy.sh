user=intery
host=$user@i.xiedaimala.com
CGO_ENABLED=0 go build .
ssh $host 'kill -9 $(cat /home/intery/pid/intery.pid)'
ssh $host 'rm /home/intery/socket/intery.sock'
scp intery $host:/home/intery/backend/current
ssh $host '/home/intery/backend/current/intery server > /home/intery/log/intery.log 2>&1 & echo $! > /home/intery/pid/intery.pid' 
ssh $host "docker exec nginx1 chmod 777 /tmp/socket/intery.sock"