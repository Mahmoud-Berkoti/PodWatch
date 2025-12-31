package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/podwatch/podwatch/pkg/models"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

var (
	db       *sql.DB
	natsConn *nats.Conn
)

func main() {
	var err error

	// 1. Postgres
	pgURL := os.Getenv("DATABASE_URL")
	if pgURL == "" {
		pgURL = "postgres://kubeguard:kubeguard@localhost:5432/kubeguard?sslmode=disable"
	}
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", pgURL)
		if err == nil {
			if err = db.Ping(); err == nil {
				break
			}
		}
		log.Printf("Waiting for Postgres... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Error connecting to Postgres: %v", err)
	}
	defer db.Close()

	// Init schema
	initSchema()

	// 2. NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	for i := 0; i < 5; i++ {
		natsConn, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Error connecting to NATS: %v", err)
	}
	defer natsConn.Close()

	// 3. Subscribe to alerts
	go subscribeAlerts()

	// 4. HTTP API
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	// Alerts API
	r.GET("/v1/alerts", listAlerts)
	r.GET("/v1/alerts/:id", getAlert)
	
	// Incidents API
	r.GET("/v1/incidents", listIncidents)
	r.GET("/v1/incidents/:id", getIncident)
	r.PATCH("/v1/incidents/:id", updateIncident)
	r.GET("/v1/incidents/:id/timeline", getIncidentTimeline)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Incident service started on port %s", port)
	r.Run(":" + port)
}

func initSchema() {
	schema := `
	CREATE TABLE IF NOT EXISTS alerts (
		id TEXT PRIMARY KEY,
		timestamp TIMESTAMPTZ NOT NULL,
		rule_name TEXT NOT NULL,
		severity TEXT NOT NULL,
		description TEXT,
		event JSONB,
		incident_id TEXT,
		response TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS incidents (
		id TEXT PRIMARY KEY,
		status TEXT NOT NULL DEFAULT 'open',
		severity TEXT NOT NULL,
		title TEXT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT NOW(),
		updated_at TIMESTAMPTZ DEFAULT NOW(),
		alert_ids TEXT[],
		triggering_event_id TEXT
	);

	CREATE TABLE IF NOT EXISTS action_logs (
		id TEXT PRIMARY KEY,
		incident_id TEXT,
		action_type TEXT NOT NULL,
		target TEXT NOT NULL,
		status TEXT NOT NULL,
		message TEXT,
		created_at TIMESTAMPTZ DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_incident ON alerts(incident_id);
	CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
	`
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatalf("Error creating schema: %v", err)
	}
}

func subscribeAlerts() {
	_, err := natsConn.QueueSubscribe("alerts", "incident-workers", func(msg *nats.Msg) {
		var alert models.Alert
		if err := json.Unmarshal(msg.Data, &alert); err != nil {
			log.Printf("Error decoding alert: %v", err)
			return
		}

		// Store alert
		eventJSON, _ := json.Marshal(alert.Event)
		_, err := db.Exec(`
			INSERT INTO alerts (id, timestamp, rule_name, severity, description, event, response)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, alert.ID, alert.Timestamp, alert.RuleName, alert.Severity, alert.Description, eventJSON, alert.Response)
		if err != nil {
			log.Printf("Error storing alert: %v", err)
			return
		}

		// Find or create incident
		incidentID, err := findOrCreateIncident(alert)
		if err != nil {
			log.Printf("Error handling incident: %v", err)
			return
		}

		// Update alert with incident ID
		db.Exec(`UPDATE alerts SET incident_id = $1 WHERE id = $2`, incidentID, alert.ID)

		// Publish for response orchestrator
		alert.IncidentID = incidentID
		data, _ := json.Marshal(alert)
		natsConn.Publish("alerts.processed", data)
	})
	if err != nil {
		log.Fatalf("Error subscribing to alerts: %v", err)
	}
}

func findOrCreateIncident(alert models.Alert) (string, error) {
	// Look for recent open incident with same severity and namespace
	namespace := ""
	if alert.Event != nil && alert.Event.Container != nil {
		namespace = alert.Event.Container.Namespace
	}

	var incidentID string
	err := db.QueryRow(`
		SELECT id FROM incidents 
		WHERE status = 'open' 
		AND severity = $1 
		AND created_at > NOW() - INTERVAL '1 hour'
		ORDER BY created_at DESC
		LIMIT 1
	`, alert.Severity).Scan(&incidentID)

	if err == sql.ErrNoRows {
		// Create new incident
		incidentID = uuid.New().String()
		title := alert.RuleName
		if namespace != "" {
			title = title + " in " + namespace
		}
		
		triggeringEventID := ""
		if alert.Event != nil {
			triggeringEventID = alert.Event.EventID
		}

		_, err = db.Exec(`
			INSERT INTO incidents (id, status, severity, title, alert_ids, triggering_event_id)
			VALUES ($1, 'open', $2, $3, $4, $5)
		`, incidentID, alert.Severity, title, []string{alert.ID}, triggeringEventID)
		if err != nil {
			return "", err
		}
		log.Printf("Created incident %s for alert %s", incidentID, alert.ID)
	} else if err != nil {
		return "", err
	} else {
		// Add alert to existing incident
		_, err = db.Exec(`
			UPDATE incidents 
			SET alert_ids = array_append(alert_ids, $1), updated_at = NOW()
			WHERE id = $2
		`, alert.ID, incidentID)
		if err != nil {
			return "", err
		}
		log.Printf("Added alert %s to incident %s", alert.ID, incidentID)
	}

	return incidentID, nil
}

func listAlerts(c *gin.Context) {
	rows, err := db.Query(`
		SELECT id, timestamp, rule_name, severity, description, incident_id, response
		FROM alerts
		ORDER BY timestamp DESC
		LIMIT 100
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var alerts []map[string]interface{}
	for rows.Next() {
		var id, ruleName, severity, description, incidentID, response sql.NullString
		var timestamp time.Time
		rows.Scan(&id, &timestamp, &ruleName, &severity, &description, &incidentID, &response)
		alerts = append(alerts, map[string]interface{}{
			"id":          id.String,
			"timestamp":   timestamp,
			"rule_name":   ruleName.String,
			"severity":    severity.String,
			"description": description.String,
			"incident_id": incidentID.String,
			"response":    response.String,
		})
	}
	c.JSON(200, alerts)
}

func getAlert(c *gin.Context) {
	id := c.Param("id")
	var alert struct {
		ID          string          `json:"id"`
		Timestamp   time.Time       `json:"timestamp"`
		RuleName    string          `json:"rule_name"`
		Severity    string          `json:"severity"`
		Description string          `json:"description"`
		Event       json.RawMessage `json:"event"`
		IncidentID  string          `json:"incident_id"`
		Response    string          `json:"response"`
	}
	err := db.QueryRow(`
		SELECT id, timestamp, rule_name, severity, description, event, incident_id, response
		FROM alerts WHERE id = $1
	`, id).Scan(&alert.ID, &alert.Timestamp, &alert.RuleName, &alert.Severity, &alert.Description, &alert.Event, &alert.IncidentID, &alert.Response)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, alert)
}

func listIncidents(c *gin.Context) {
	status := c.Query("status")
	query := `SELECT id, status, severity, title, created_at, updated_at FROM incidents`
	args := []interface{}{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := db.Query(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var incidents []map[string]interface{}
	for rows.Next() {
		var id, status, severity, title string
		var createdAt, updatedAt time.Time
		rows.Scan(&id, &status, &severity, &title, &createdAt, &updatedAt)
		incidents = append(incidents, map[string]interface{}{
			"id":         id,
			"status":     status,
			"severity":   severity,
			"title":      title,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}
	c.JSON(200, incidents)
}

func getIncident(c *gin.Context) {
	id := c.Param("id")
	var incident struct {
		ID               string    `json:"id"`
		Status           string    `json:"status"`
		Severity         string    `json:"severity"`
		Title            string    `json:"title"`
		CreatedAt        time.Time `json:"created_at"`
		UpdatedAt        time.Time `json:"updated_at"`
		TriggeringEvent  string    `json:"triggering_event_id"`
	}
	err := db.QueryRow(`
		SELECT id, status, severity, title, created_at, updated_at, triggering_event_id
		FROM incidents WHERE id = $1
	`, id).Scan(&incident.ID, &incident.Status, &incident.Severity, &incident.Title, &incident.CreatedAt, &incident.UpdatedAt, &incident.TriggeringEvent)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, incident)
}

func updateIncident(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validStatuses := map[string]bool{
		"open": true, "investigating": true, "contained": true, "resolved": true,
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	_, err := db.Exec(`UPDATE incidents SET status = $1, updated_at = NOW() WHERE id = $2`, req.Status, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "updated"})
}

func getIncidentTimeline(c *gin.Context) {
	id := c.Param("id")
	
	// Get alerts for this incident
	rows, err := db.Query(`
		SELECT id, timestamp, rule_name, severity, description, event
		FROM alerts
		WHERE incident_id = $1
		ORDER BY timestamp ASC
	`, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var timeline []map[string]interface{}
	for rows.Next() {
		var alertID, ruleName, severity, description string
		var timestamp time.Time
		var event json.RawMessage
		rows.Scan(&alertID, &timestamp, &ruleName, &severity, &description, &event)
		timeline = append(timeline, map[string]interface{}{
			"type":        "alert",
			"id":          alertID,
			"timestamp":   timestamp,
			"rule_name":   ruleName,
			"severity":    severity,
			"description": description,
			"event":       event,
		})
	}

	// Get actions for this incident
	actionRows, err := db.Query(`
		SELECT id, action_type, target, status, message, created_at
		FROM action_logs
		WHERE incident_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer actionRows.Close()

	for actionRows.Next() {
		var actionID, actionType, target, status, message string
		var createdAt time.Time
		actionRows.Scan(&actionID, &actionType, &target, &status, &message, &createdAt)
		timeline = append(timeline, map[string]interface{}{
			"type":        "action",
			"id":          actionID,
			"timestamp":   createdAt,
			"action_type": actionType,
			"target":      target,
			"status":      status,
			"message":     message,
		})
	}

	c.JSON(200, timeline)
}
