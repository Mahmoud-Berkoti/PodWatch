package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// BatchWriter buffers events and writes to S3
type BatchWriter struct {
	bucket        string
	region        string
	clusterID     string
	nodeID        string
	uploader      *s3manager.Uploader
	buffer        *bytes.Buffer
	gzipWriter    *gzip.Writer
	mu            sync.Mutex
	lastWrite     time.Time
	flushInterval time.Duration
}

func NewBatchWriter(bucket, region, clusterID string) *BatchWriter {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	
	uploader := s3manager.NewUploader(sess)

	bw := &BatchWriter{
		bucket:        bucket,
		region:        region,
		clusterID:     clusterID,
		uploader:      uploader,
		buffer:        new(bytes.Buffer),
		lastWrite:     time.Now(),
		flushInterval: 60 * time.Second,
	}
	bw.gzipWriter = gzip.NewWriter(bw.buffer)
	
	go bw.loop()
	return bw
}

func NewNodeBatchWriter(bucket, region, clusterID, nodeID string) *BatchWriter {
	bw := NewBatchWriter(bucket, region, clusterID)
	bw.nodeID = nodeID
	return bw
}


func (bw *BatchWriter) Write(data []byte) error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	if _, err := bw.gzipWriter.Write(data); err != nil {
		return err
	}
	if _, err := bw.gzipWriter.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}

func (bw *BatchWriter) loop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		bw.flushIfNeeded()
	}
}

func (bw *BatchWriter) flushIfNeeded() {
	bw.mu.Lock()
	if bw.buffer.Len() == 0 || time.Since(bw.lastWrite) < bw.flushInterval {
		bw.mu.Unlock()
		return
	}
	
	bw.gzipWriter.Close()
	payload := bw.buffer.Bytes()
	
	bw.buffer.Reset()
	bw.gzipWriter.Reset(bw.buffer)
	bw.lastWrite = time.Now()
	bw.mu.Unlock()

	go bw.upload(payload)
}

func (bw *BatchWriter) upload(data []byte) {
	now := time.Now().UTC()
	// path: raw/cluster/date/node/hour/*.jsonl.gz
	
	nodePath := bw.nodeID
	if nodePath == "" {
		nodePath = "unknown-node"
	}

	key := fmt.Sprintf("raw/%s/%s/%s/%s/%s.jsonl.gz", 
		bw.clusterID, 
		now.Format("2006-01-02"),
		nodePath,
		now.Format("15"), 
		fmt.Sprintf("%d", now.UnixNano()))

	_, err := bw.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bw.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		log.Printf("Failed to upload to S3: %v", err)
	} else {
		log.Printf("Uploaded %s to S3", key)
	}
}
