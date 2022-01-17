package log

import (
	"bufio"
	"fmt"
	"intery/server/config/dir"
	"intery/server/database"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
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
		"Access-Control-Allow-Origin": "http://127.0.0.1:3000",
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

	logChan := make(chan string)
	close := make(chan bool, 1)
	quit := make(chan bool, 1)

	if deployment.Status != "running" {
		go func() {
			userDir := dir.EnsureUserDir(deployment.UserId)
			projectDir := dir.EnsureProjectDir(userDir, deployment.ProjectId)
			content, _ := ioutil.ReadFile(filepath.Join(projectDir, fmt.Sprintf("%v_log", deployment.ContainerId)))
			logChan <- string(content)
			close <- true
		}()
	} else {
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
	}

	sse, err := sseApi.broker.Connect(fmt.Sprintf("%v", rand.Int63()), c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	go func() {
		count := 0
		for {
			select {
			case <-quit:
				return
			case <-close:
				sse.Send(net.StringEvent{
					Id:    fmt.Sprintf("%d", count),
					Event: "close",
					Data:  "",
				})
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
	quit <- true
}
