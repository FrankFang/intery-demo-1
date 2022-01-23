user=intery
host=$user@i.xiedaimala.com
task_name=$1
echo "请先部署最新代码到服务器"
ssh $host "/home/intery/backend/current/intery task $task_name"
echo "OK"