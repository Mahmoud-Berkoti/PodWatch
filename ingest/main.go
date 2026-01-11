package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/podwatch/podwatch/pkg/models"
)

var (
	natsConn *nats.Conn
)

func main() {
	// 1. Config
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	// 2. NATS Connection
	var err error
	// Retry logic for NATS
	for i := 0; i < 5; i++ {
		natsConn, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		log.Printf("Waiting for NATS... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer natsConn.Close()

	// 3. Setup Gin
	r := gin.Default()

	r.POST("/v1/events", handleEvent)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 4. Run with mTLS
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	certFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	caFile := os.Getenv("TLS_CA_FILE")

	if certFile != "" && keyFile != "" && caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err != nil {
			log.Fatalf("Error reading CA file: %v", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}

		server := &http.Server{
			Addr:      ":" + port,
			Handler:   r,
			TLSConfig: tlsConfig,
		}

		log.Printf("Starting Ingest API on port %s with mTLS", port)
		if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	} else {
		log.Printf("Starting Ingest API on port %s WITHOUT mTLS (dev mode)", port)
		if err := r.Run(":" + port); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}
}

func handleEvent(c *gin.Context) {
	// Read raw body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "failed to read body"})
		return
	}
	// Restore body for any subsequent middlewares (if any, though we are leaf)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Validate JSON schema
	var event models.RuntimeEvent
	if err := json.Unmarshal(bodyBytes, &event); err != nil {
		c.JSON(400, gin.H{"error": "invalid json"})
		return
	}

	// Basic validation
	if event.ClusterID == "" || event.NodeID == "" {
		c.JSON(400, gin.H{"error": "missing cluster_id or node_id"})
		return
	}

	// 1. Publish to NATS
	// Subject: events.raw.<cluster>.<node>
	subject := fmt.Sprintf("events.raw.%s.%s", event.ClusterID, event.NodeID)
	if err := natsConn.Publish(subject, bodyBytes); err != nil {
		log.Printf("Error publishing to NATS: %v", err)
		c.JSON(500, gin.H{"error": "internal error"})
		return
	}

	// 2. Write to S3 (Async/Buffered) - Disabled for demo
	// In production, this would batch events to S3
	log.Printf("Event received: %s from %s/%s", event.EventType, event.ClusterID, event.NodeID)

	c.JSON(200, gin.H{"status": "accepted"})
}
