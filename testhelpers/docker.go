package testhelpers

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type PostgresDockerContainer struct {
	ContainerId      string
	ConnectionString string
}

func CreatePostgresDockerContainer(ctx context.Context, t *testing.T) (*PostgresDockerContainer, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Error creating Docker client: %v", err)
	}

	imageName := "postgres:15.3-alpine"
	out, err := cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		log.Fatalf("Error pulling image: %v", err)
	}
	defer out.Close()

	dbName := "testdb"
	dbUser := "postgres"
	dbPassword := "postgres"
	initFile := filepath.Join("..", "testdata", "init-db.sql")

	port, err := nat.NewPort("tcp", "5432")
	if err != nil {
		log.Fatalf("Error configuring port: %v", err)
	}

	containerConfig := &container.Config{
		Image: imageName,
		Env: []string{
			fmt.Sprintf("POSTGRES_DB=%s", dbName),
			fmt.Sprintf("POSTGRES_USER=%s", dbUser),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", dbPassword),
		},
		ExposedPorts: nat.PortSet{
			port: struct{}{},
		},
	}

	hostPort := rand.Intn(65536)

	for !checkPortAvailable(strconv.Itoa(hostPort)) {
		hostPort = rand.Intn(65536)
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			port: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(hostPort),
				},
			},
		},
		Binds: []string{
			fmt.Sprintf("%s:/docker-entrypoint-initdb.d/init.sql", getAbsolutePath(initFile)),
		},
	}

	networkingConfig := &network.NetworkingConfig{}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, networkingConfig, nil, "")
	if err != nil {
		log.Fatalf("Error creating container: %v", err)
	}

	t.Cleanup(func() {
		log.Printf("removing container with id %s", resp.ID)
		if err := teardownDockerContainer(ctx, cli, resp.ID); err != nil {
			log.Fatalf("error terminating postgres container: %s", err)
		}
	})

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Fatalf("Error starting container: %v", err)
	}

	timeout := 60 * time.Second
	if err := waitForLogMessage(ctx, cli, resp.ID, "ready to accept connections", 2, timeout); err != nil {
		log.Fatalf("Error waiting for initialization: %v", err)
	}

	connStr := fmt.Sprintf("postgresql://postgres:postgres@localhost:%s/testdb?sslmode=disable", strconv.Itoa(hostPort))

	return &PostgresDockerContainer{
		ContainerId:      resp.ID,
		ConnectionString: connStr,
	}, nil
}

func teardownDockerContainer(ctx context.Context, cli *client.Client, containerId string) error {
	if err := cli.ContainerStop(ctx, containerId, container.StopOptions{}); err != nil {
		return err
	}

	if err := cli.ContainerRemove(ctx, containerId, container.RemoveOptions{}); err != nil {
		return err
	}

	return nil
}

func getAbsolutePath(fileName string) string {
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}
	return absPath
}

func checkPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)

	if err != nil {
		return false
	}

	_ = ln.Close()
	return true
}

func waitForLogMessage(ctx context.Context, cli *client.Client, containerID, message string, count int, timeout time.Duration) error {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "all",
	}

	logs, err := cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return fmt.Errorf("error getting container logs: %v", err)
	}
	defer logs.Close()

	scanner := bufio.NewScanner(logs)
	matches := 0

	timedOut := time.After(timeout)
	done := make(chan bool)

	go func() {
		for scanner.Scan() {
			logLine := scanner.Text()
			//fmt.Println(logLine)
			if strings.Contains(logLine, message) {
				matches++
				if matches >= count {
					done <- true
					return
				}
			}
		}
		done <- false
	}()

	select {
	case <-done:
		if matches >= count {
			return nil
		}
		return fmt.Errorf("expected log message '%s' not found %d times", message, count)
	case <-timedOut:
		return fmt.Errorf("timeout waiting for log message '%s' to appear %d times", message, count)
	}
}
