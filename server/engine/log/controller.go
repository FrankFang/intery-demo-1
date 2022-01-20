package log

import (
	"bufio"
	"fmt"
	"intery/server/config"
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
		"Access-Control-Allow-Origin": config.GetDomain(),
	})
	sseClientBroker.SetDisconnectCallback(func(clientId string, sessionId string) {
		log.Printf("session %v of client %v was disconnected.", sessionId, clientId)
	})
	sseApi = &SseApi{broker: sseClientBroker}
}

func (ctrl *Controller) Index(c *gin.Context) {
	log.Println("11111111111111")
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
	log.Println("4444444444444444444")

	logChan := make(chan string)
	close := make(chan bool, 1)
	quit := make(chan bool, 1)

	log.Println("33333333333333333333")
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
		log.Println("555555555555555555555555")
		go func(logChan chan string) {
			log.Println("888888888888888888")
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

	log.Println("666666666666666666666")
	sse, err := sseApi.broker.Connect(fmt.Sprintf("%v", rand.Int63()), c.Writer, c.Request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	go func() {
		log.Println("7777777777777777777777")
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
				log.Println("101010101010101010", logs)
				sse.Send(net.StringEvent{
					Id:    fmt.Sprintf("%d", count),
					Event: "log",
					Data:  logs,
				})
				sse.Send(net.StringEvent{
					Id:    fmt.Sprintf("%d", count),
					Event: "message",
					Data:  logs,
				})
				count++
			}
		}
	}()

	log.Println("9999999999999999")
	<-sse.Done()
	quit <- true
	log.Println("2222222222")
}
