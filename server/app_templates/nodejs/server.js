const http = require('http');
const fs = require('fs')

const requestListener = function (req, res) {
  res.writeHead(200);
  res.end('你好，世界！');
}

const server = http.createServer(requestListener);

const socket = process.env.SOCKET
if (!socket) {
  console.log("未指定 socket")
  process.exit(1)
}
server.listen(socket, () => {
  console.log('I am listening on ' + socket)
});
