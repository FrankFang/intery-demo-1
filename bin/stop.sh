user=intery
host=$user@i.xiedaimala.com
ssh $host 'kill -9 $(cat /home/intery/pid/intery.pid)'
ssh $host 'rm /home/intery/socket/intery.sock'
echo "OK"