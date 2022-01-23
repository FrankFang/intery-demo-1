user=intery
host=$user@i.xiedaimala.com
ssh $host 'kill -9 $(cat /home/intery/pid/intery.pid)'
ssh $host 'rm /home/intery/socket/intery.sock'
ssh $host '/home/intery/backend/current/intery server >> /home/intery/log/intery.log 2>&1 & echo $! > /home/intery/pid/intery.pid' 
ssh $host 'docker exec nginx1 sh -c "chmod 777 /tmp/socket/*.sock"'
echo "OK"