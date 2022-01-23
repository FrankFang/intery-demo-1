user=intery
host=$user@i.xiedaimala.com
task_name=$1
ssh $host "/home/intery/backend/current/intery task $task_name"
echo "OK"