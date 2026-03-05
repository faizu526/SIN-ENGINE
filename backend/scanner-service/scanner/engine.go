package scanner

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "time"
    
    "github.com/sin-engine/scanner-service/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.uber.org/zap"
)

type ScanEngine struct {
    scanID       string
    config       map[string]interface{}
    target       string
    scanTypes    []string
    client       *http.Client
    logger       *zap.Logger
    vulnChan     chan *models.Vulnerability
    done         chan bool
    wg           sync.WaitGroup
    scanCollection *mongo.Collection
    vulnCollection *mongo.Collection
}

func NewScanEngine(scan *models.Scan, logger *zap.Logger, mongoClient *mongo.Client) *ScanEngine {
    return &ScanEngine{
        scanID:    scan.ID,
        config:    scan.Config,
        target:    scan.Target,
        scanTypes: scan.ScanType,
        client: &http.Client{
            Timeout: 10 * time.Second,
            CheckRedirect: func(req *http.Request, via []*http.Request) error {
                return http.ErrUseLastResponse
            },
        },
        logger:        logger,
        vulnChan:      make(chan *models.Vulnerability, 100),
        done:          make(chan bool),
        scanCollection: mongoClient.Database("scanner").Collection("scans"),
        vulnCollection: mongoClient.Database("scanner").Collection("vulnerabilities"),
    }
}

func (e *ScanEngine) Run() {
    e.logger.Info("Starting scan", zap.String("scan_id", e.scanID), zap.String("target", e.target))
    
    // Update scan status
    e.updateStatus(models.StatusRunning, nil)
    
    // Start vulnerability collector
    go e.collectVulnerabilities()
    
    // Run modules based on scan types
    for _, scanType := range e.scanTypes {
        e.wg.Add(1)
        go e.runModule(scanType)
    }
    
    // Wait for all modules to complete
    e.wg.Wait()
    close(e.vulnChan)
    <-e.done
    
    // Update scan stats
    e.updateStats()
    
    e.logger.Info("Scan completed", zap.String("scan_id", e.scanID))
}

func (e *ScanEngine) runModule(moduleName string) {
    defer e.wg.Done()
    
    e.logger.Debug("Running module", zap.String("module", moduleName))
    
    switch moduleName {
    case "open_redirect":
        e.testOpenRedirect()
    case "xss":
        e.testXSS()
    case "sqli":
        e.testSQLi()
    case "ssrf":
        e.testSSRF()
    case "idor":
        e.testIDOR()
    case "rate_limit":
        e.testRateLimit()
    default:
        e.logger.Warn("Unknown module", zap.String("module", moduleName))
    }
}

func (e *ScanEngine) collectVulnerabilities() {
    for vuln := range e.vulnChan {
        vuln.ScanID = e.scanID
        vuln.CreatedAt = time.Now()
        
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        _, err := e.vulnCollection.InsertOne(ctx, vuln)
        cancel()
        
        if err != nil {
            e.logger.Error("Failed to save vulnerability", zap.Error(err))
        } else {
            e.logger.Info("Vulnerability found",
                zap.String("type", vuln.Type),
                zap.String("severity", string(vuln.Severity)),
                zap.String("url", vuln.URL))
        }
    }
    e.done <- true
}

func (e *ScanEngine) updateStatus(status models.ScanStatus, stats *models.ScanStats) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    update := bson.M{
        "$set": bson.M{
            "status":     status,
            "updated_at": time.Now(),
        },
    }
    
    if stats != nil {
        update["$set"].(bson.M)["stats"] = stats
    }
    
    if status == models.StatusRunning {
        update["$set"].(bson.M)["started_at"] = time.Now()
    } else if status == models.StatusCompleted {
        update["$set"].(bson.M)["completed_at"] = time.Now()
        update["$set"].(bson.M)["progress"] = 100
    }
    
    e.scanCollection.UpdateOne(ctx, bson.M{"_id": e.scanID}, update)
}

func (e *ScanEngine) updateStats() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Count vulnerabilities by severity
    pipeline := bson.A{
        bson.M{"$match": bson.M{"scan_id": e.scanID}},
        bson.M{"$group": bson.M{
            "_id":   "$severity",
            "count": bson.M{"$sum": 1},
        }},
    }
    
    cursor, err := e.vulnCollection.Aggregate(ctx, pipeline)
    if err != nil {
        e.logger.Error("Failed to aggregate vulnerabilities", zap.Error(err))
        return
    }
    defer cursor.Close(ctx)
    
    vulnStats := make(map[string]int)
    for cursor.Next(ctx) {
        var result struct {
            ID    string `bson:"_id"`
            Count int    `bson:"count"`
        }
        if err := cursor.Decode(&result); err == nil {
            vulnStats[result.ID] = result.Count
        }
    }
    
    stats := &models.ScanStats{
        Vulnerabilities: vulnStats,
    }
    
    e.updateStatus(models.StatusCompleted, stats)
}