package report

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/logs"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/post"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

// CreateReport POST /api/reports
func CreateReport(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")

	var input CreateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		logs.LogJSON("WARN", "Invalid report data", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": userID,
		})
		return
	}

	// Validation des types et raisons
	if !input.TargetType.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type de signalement invalide"})
		return
	}

	if !input.Reason.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Raison de signalement invalide"})
		return
	}

	// Vérifier que la cible existe
	if err := validateTargetExists(input.TargetType, input.TargetID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Élément à signaler non trouvé"})
		logs.LogJSON("WARN", "Report target not found", map[string]interface{}{
			"targetType": input.TargetType,
			"targetID":   input.TargetID,
			"route":      route,
			"userID":     userID,
		})
		return
	}

	// Vérifier si l'utilisateur a déjà signalé cette cible
	var existingReport Report
	err := database.DB.Where("reporter_id = ? AND target_type = ? AND target_id = ?",
		userID, input.TargetType, input.TargetID).First(&existingReport).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Vous avez déjà signalé cet élément"})
		logs.LogJSON("WARN", "User already reported this target", map[string]interface{}{
			"targetType": input.TargetType,
			"targetID":   input.TargetID,
			"route":      route,
			"userID":     userID,
		})
		return
	}

	// Créer le signalement
	report := Report{
		ReporterID:  userID,
		TargetType:  input.TargetType,
		TargetID:    input.TargetID,
		Reason:      input.Reason,
		Description: input.Description,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := database.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du signalement"})
		logs.LogJSON("ERROR", "Error creating report", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": userID,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Signalement créé avec succès",
		"report":  report,
	})

	logs.LogJSON("INFO", "Report created successfully", map[string]interface{}{
		"reportID":   report.ID,
		"targetType": input.TargetType,
		"targetID":   input.TargetID,
		"route":      route,
		"userID":     userID,
	})
}

// GetReports GET /api/admin/reports (Admin seulement)
func GetReports(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")

	// Paramètres de pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Filtres
	status := c.Query("status")
	targetType := c.Query("target_type")
	reason := c.Query("reason")

	// Construction de la requête
	query := database.DB.Model(&Report{}).
		Preload("Reporter").
		Preload("Admin").
		Order("created_at DESC")

	// Application des filtres
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if targetType != "" {
		query = query.Where("target_type = ?", targetType)
	}
	if reason != "" {
		query = query.Where("reason = ?", reason)
	}

	// Compter le total
	var total int64
	query.Count(&total)

	// Récupérer les signalements avec pagination
	var reports []Report
	if err := query.Limit(limit).Offset(offset).Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des signalements"})
		logs.LogJSON("ERROR", "Error fetching reports", map[string]interface{}{
			"error":  err.Error(),
			"route":  route,
			"userID": userID,
		})
		return
	}

	// Enrichir avec les détails des cibles
	reportsWithTargets := make([]ReportWithTarget, len(reports))
	for i, report := range reports {
		reportWithTarget := ReportWithTarget{Report: report}

		// Charger les détails de la cible selon le type
		switch report.TargetType {
		case ReportTypePost:
			var targetPost post.Post
			if err := database.DB.First(&targetPost, "id = ?", report.TargetID).Error; err == nil {
				reportWithTarget.TargetPost = &targetPost
			}
		case ReportTypeUser:
			var targetUser user.User
			if err := database.DB.First(&targetUser, "id = ?", report.TargetID).Error; err == nil {
				reportWithTarget.TargetUser = &targetUser
			}
		case ReportTypeComment:
			var targetComment post.Comment
			if err := database.DB.First(&targetComment, "id = ?", report.TargetID).Error; err == nil {
				reportWithTarget.TargetComment = &targetComment
			}
		}

		reportsWithTargets[i] = reportWithTarget
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": reportsWithTargets,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})

	logs.LogJSON("INFO", "Reports fetched successfully", map[string]interface{}{
		"route":  route,
		"userID": userID,
		"total":  total,
	})
}

// UpdateReport PUT /api/admin/reports/:id (Admin seulement)
func UpdateReport(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")
	reportID := c.Param("id")

	var input UpdateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides", "details": err.Error()})
		return
	}

	// Récupérer le signalement
	var report Report
	if err := database.DB.First(&report, "id = ?", reportID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Signalement non trouvé"})
		logs.LogJSON("WARN", "Report not found", map[string]interface{}{
			"reportID": reportID,
			"route":    route,
			"userID":   userID,
		})
		return
	}

	// Mettre à jour le signalement
	updates := map[string]interface{}{
		"status":     input.Status,
		"admin_note": input.AdminNote,
		"admin_id":   userID,
		"updated_at": time.Now(),
	}

	// Si le statut est résolu ou rejeté, mettre la date de résolution
	if input.Status == StatusResolved || input.Status == StatusRejected {
		now := time.Now()
		updates["resolved_at"] = &now
	}

	if err := database.DB.Model(&report).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour"})
		logs.LogJSON("ERROR", "Error updating report", map[string]interface{}{
			"error":    err.Error(),
			"reportID": reportID,
			"route":    route,
			"userID":   userID,
		})
		return
	}

	// Recharger le signalement avec les relations
	if err := database.DB.Preload("Reporter").Preload("Admin").First(&report, "id = ?", reportID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors du rechargement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Signalement mis à jour avec succès",
		"report":  report,
	})

	logs.LogJSON("INFO", "Report updated successfully", map[string]interface{}{
		"reportID": reportID,
		"status":   input.Status,
		"route":    route,
		"userID":   userID,
	})
}

// DeleteReport DELETE /api/admin/reports/:id (Admin seulement)
func DeleteReport(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")
	reportID := c.Param("id")

	// Vérifier que le signalement existe
	var report Report
	if err := database.DB.First(&report, "id = ?", reportID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Signalement non trouvé"})
		logs.LogJSON("WARN", "Report not found", map[string]interface{}{
			"reportID": reportID,
			"route":    route,
			"userID":   userID,
		})
		return
	}

	// Supprimer le signalement
	if err := database.DB.Delete(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression"})
		logs.LogJSON("ERROR", "Error deleting report", map[string]interface{}{
			"error":    err.Error(),
			"reportID": reportID,
			"route":    route,
			"userID":   userID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signalement supprimé avec succès"})

	logs.LogJSON("INFO", "Report deleted successfully", map[string]interface{}{
		"reportID": reportID,
		"route":    route,
		"userID":   userID,
	})
}

// GetReportStats GET /api/admin/reports/stats (Admin seulement)
func GetReportStats(c *gin.Context) {
	route := c.FullPath()
	userID := c.GetString("user_id")

	// Statistiques par statut
	var statsByStatus []struct {
		Status ReportStatus `json:"status"`
		Count  int64        `json:"count"`
	}
	database.DB.Model(&Report{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statsByStatus)

	// Statistiques par type
	var statsByType []struct {
		TargetType ReportType `json:"target_type"`
		Count      int64      `json:"count"`
	}
	database.DB.Model(&Report{}).
		Select("target_type, COUNT(*) as count").
		Group("target_type").
		Scan(&statsByType)

	// Statistiques par raison
	var statsByReason []struct {
		Reason ReportReason `json:"reason"`
		Count  int64        `json:"count"`
	}
	database.DB.Model(&Report{}).
		Select("reason, COUNT(*) as count").
		Group("reason").
		Scan(&statsByReason)

	// Signalements récents (dernières 24h)
	var recentCount int64
	database.DB.Model(&Report{}).
		Where("created_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&recentCount)

	c.JSON(http.StatusOK, gin.H{
		"stats_by_status": statsByStatus,
		"stats_by_type":   statsByType,
		"stats_by_reason": statsByReason,
		"recent_count":    recentCount,
	})

	logs.LogJSON("INFO", "Report stats fetched successfully", map[string]interface{}{
		"route":  route,
		"userID": userID,
	})
}

// Fonction utilitaire pour valider l'existence de la cible
func validateTargetExists(targetType ReportType, targetID string) error {
	switch targetType {
	case ReportTypePost:
		var post post.Post
		return database.DB.First(&post, "id = ?", targetID).Error
	case ReportTypeUser:
		var user user.User
		return database.DB.First(&user, "id = ?", targetID).Error
	case ReportTypeComment:
		var comment post.Comment
		return database.DB.First(&comment, "id = ?", targetID).Error
	default:
		return fmt.Errorf("type de cible invalide")
	}
}
