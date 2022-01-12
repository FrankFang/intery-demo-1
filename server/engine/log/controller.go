package log

import (
	"bufio"
	"fmt"
	"intery/server/database"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/ahmetb/dlog"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	net "github.com/subchord/go-sse"
)

type Controller struct {
}
type SseApi struct {
	broker *net.Broker
}

var sseApi *SseApi

func init() {
	rand.Seed(time.Now().Unix())
	sseClientBroker := net.NewBroker(map[string]string{
		"Access-Control-Allow-Origin": "http://localhost:3000",
	})
	sseClientBroker.SetDisconnectCallback(func(clientId string, sessionId string) {
		log.Printf("session %v of client %v was disconnected.", sessionId, clientId)
	})
	sseApi = &SseApi{broker: sseClientBroker}
}

func (ctrl *Controller) Index(c *gin.Context) {
	deploymentIdString := c.Query("deployment_id")
	deploymentId, err := strconv.Atoi(deploymentIdString)
	if err != nil {
		// c.JSON(http.StatusBadGateway, gin.H{"reason": "deployment_id 必须是数字"})
		panic(err)
	}
	d := database.GetQuery().Deployment
	deployment, err := d.WithContext(c).Where(d.ID.Eq(uint(deploymentId))).First()
	if err != nil {
		// c.JSON(http.StatusNotFound, gin.H{"reason": err.Error()})
		panic(err)
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}
	logReader, err := cli.ContainerLogs(c, deployment.ContainerId, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	logChan := make(chan string)
	go func(logChan chan string) {
		defer logReader.Close()
		r := dlog.NewReader(logReader)
		s := bufio.NewScanner(r)
		for s.Scan() {
			logChan <- s.Text()
		}
		if err := s.Err(); err != nil {
			log.Fatalf("read error: %v", err)
		}
	}(logChan)

	sse, err := sseApi.broker.Connect(fmt.Sprintf("%v", rand.Int63()), c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	stop := make(chan interface{}, 1)

	go func() {
		count := 0
		for {
			select {
			case <-stop:
				return
			case logs := <-logChan:
				sse.Send(net.StringEvent{
					Id:    fmt.Sprintf("%d", count),
					Event: "log",
					Data:  logs,
				})
				count++
			}
		}
	}()

	<-sse.Done()
	stop <- true
}
