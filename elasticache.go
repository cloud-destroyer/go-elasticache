package elasticache

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// Client is an Elasticache-aware memcache client
type Client struct {
	*memcache.Client
	// Custom ServerList so we can replace endpoints automatically at runtime
	ServerList *memcache.ServerList
	// Elasticache config endpoint
	Endpoint string
}

// Watch polls the Elasticache endpoint every 60 seconds to update cluster endpoints.
// It can be stopped by calling ctx.Done()
func (c *Client) Watch(ctx context.Context) {
	t := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-t.C:
			urls, err := clusterNodes(c.Endpoint)
			if err != nil {
				continue
			}

			c.ServerList.SetServers(urls...)
		case <-ctx.Done():
			t.Stop()
			return
		}
	}
}

// New takes an Elasticache configuration endpoint and returns an instance of the memcache client.
func New(endpoint string) (*Client, error) {
	servers, err := clusterNodes(endpoint)
	if err != nil {
		return nil, err
	}

	ss := new(memcache.ServerList)
	if err := ss.SetServers(servers...); err != nil {
		return nil, err
	}

	client := &Client{
		Client:     memcache.NewFromSelector(ss),
		ServerList: ss,
		Endpoint:   endpoint,
	}

	return client, nil
}

func clusterNodes(endpoint string) ([]string, error) {
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	command := "config get cluster\r\n"
	fmt.Fprintf(conn, command)

	response, err := parseNodes(conn)
	if err != nil {
		return nil, err
	}

	urls, err := parseURLs(response)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func parseNodes(conn io.Reader) (string, error) {
	var response string

	count := 0
	location := 3 // AWS docs suggest that nodes will always be listed on line 3

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		count++
		if count == location {
			response = scanner.Text()
		}
		if scanner.Text() == "END" {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return response, nil
}

func parseURLs(response string) ([]string, error) {
	var urls []string

	for _, v := range strings.Split(response, " ") {
		fields := strings.Split(v, "|") // ["host", "ip", "port"]

		port, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		urls = append(urls, fmt.Sprintf("%s:%d", fields[1], port))
	}

	return urls, nil
}
