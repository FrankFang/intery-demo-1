const http = require('http');
const fs = require('fs')

const requestListener = function (req, res) {
  res.writeHead(200);
  res.end('Hello, World!');
}

const server = http.createServer(requestListener);

server.on('error', appendLog)
process.on('uncaughtException', error => appendLog(error));
process.on('unhandledRejection', error => appendLog(error));

function appendLog(error) {
  fs.appendFile('/tmp/log', error.message + '\n' + error.stack + '\n', (err) => { })
}

server.listen(process.env.PORT || 8080);
