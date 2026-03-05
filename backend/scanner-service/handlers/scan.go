package handlers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/sin-engine/scanner-service/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var scanCollection *mongo.Collection

func init() {
    scanCollection = mongoClient.Database("scanner").Collection("scans")
}

type CreateScanRequest struct {
    Name     string   `json:"name" binding:"required"`
    Target   string   `json:"target" binding:"required"`
    ScanType []string `json:"scanType" binding:"required"`
    Config   map[string]interface{} `json:"config"`
}

func CreateScan(c *gin.Context) {
    var req CreateScanRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    userID := c.GetString("user_id")
    
    scan := models.NewScan()
    scan.UserID = userID
    scan.Name = req.Name
    scan.Target = req.Target
    scan.ScanType = req.ScanType
    scan.Config = req.Config
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    _, err := scanCollection.InsertOne(ctx, scan)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scan"})
        return
    }
    
    c.JSON(http.StatusCreated, scan)
}

func ListScans(c *gin.Context) {
    userID := c.GetString("user_id")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    filter := bson.M{"user_id": userID}
    
    // Add status filter if provided
    if status := c.Query("status"); status != "" {
        filter["status"] = status
    }
    
    opts := options.Find().
        SetSort(bson.D{{Key: "created_at", Value: -1}}).
        SetLimit(100)
    
    cursor, err := scanCollection.Find(ctx, filter, opts)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list scans"})
        return
    }
    defer cursor.Close(ctx)
    
    var scans []models.Scan
    if err := cursor.All(ctx, &scans); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse scans"})
        return
    }
    
    c.JSON(http.StatusOK, scans)
}

func GetScan(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetString("user_id")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    var scan models.Scan
    err := scanCollection.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&scan)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get scan"})
        }
        return
    }
    
    c.JSON(http.StatusOK, scan)
}

func StartScan(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetString("user_id")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    now := time.Now()
    update := bson.M{
        "$set": bson.M{
            "status":     models.StatusRunning,
            "started_at": now,
            "updated_at": now,
        },
    }
    
    result, err := scanCollection.UpdateOne(
        ctx,
        bson.M{"_id": id, "user_id": userID},
        update,
    )
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start scan"})
        return
    }
    
    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
        return
    }
    
    // Start scan in background
    go runScan(id)
    
    c.JSON(http.StatusOK, gin.H{"message": "Scan started"})
}

func StopScan(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetString("user_id")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    update := bson.M{
        "$set": bson.M{
            "status":     models.StatusCancelled,
            "updated_at": time.Now(),
        },
    }
    
    result, err := scanCollection.UpdateOne(
        ctx,
        bson.M{"_id": id, "user_id": userID, "status": models.StatusRunning},
        update,
    )
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop scan"})
        return
    }
    
    if result.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found or not running"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "Scan stopped"})
}

func DeleteScan(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetString("user_id")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    result, err := scanCollection.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scan"})
        return
    }
    
    if result.DeletedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
        return
    }
    
    // Delete associated vulnerabilities
    vulnCollection := mongoClient.Database("scanner").Collection("vulnerabilities")
    vulnCollection.DeleteMany(ctx, bson.M{"scan_id": id})
    
    c.JSON(http.StatusOK, gin.H{"message": "Scan deleted"})
}

func GetResults(c *gin.Context) {
    id := c.Param("id")
    userID := c.GetString("user_id")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // Verify scan exists
    var scan models.Scan
    err := scanCollection.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&scan)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Scan not found"})
        return
    }
    
    // Get vulnerabilities
    vulnCollection := mongoClient.Database("scanner").Collection("vulnerabilities")
    cursor, err := vulnCollection.Find(ctx, bson.M{"scan_id": id})
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get results"})
        return
    }
    defer cursor.Close(ctx)
    
    var vulnerabilities []models.Vulnerability
    if err := cursor.All(ctx, &vulnerabilities); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse results"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "scan":           scan,
        "vulnerabilities": vulnerabilities,
        "count":          len(vulnerabilities),
    })
}