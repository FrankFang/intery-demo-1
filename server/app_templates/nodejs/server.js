const http = require('http');
const fs = require('fs')

const requestListener = function (req, res) {
  res.writeHead(200);
  res.end('Hello, World!');
}

const server = http.createServer(requestListener);

server.on('error', console.log)
process.on('uncaughtException', console.log)
process.on('unhandledRejection', console.log)

// function appendLog(error) {
//   fs.appendFile('/tmp/log', error.message + '\n' + error.stack + '\n', (err) => { })
// }

const port = process.env.PORT || 8080
server.listen(port, () => {
  console.log('I am listening ' + port)
});
