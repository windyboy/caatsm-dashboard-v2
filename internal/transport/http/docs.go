// Package http provides HTTP transport handlers for the CAATSM Dashboard API
//
// @title CAATSM Dashboard API
// @version 1.0
// @description Civil Aviation Aerogram Traffic Stream Monitor Dashboard API provides realtime monitoring, search, and analytics for aviation telegram traffic.
// @description 
// @description Features:
// @description - Full-text search with Meilisearch
// @description - Real-time analytics and statistics
// @description - Streaming CSV export (handles 50k+ records)
// @description - Time range validation (max 90 days)
// @description - WebSocket real-time updates
//
// @contact.name API Support
// @contact.email support@caatsm.example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:3002
// @BasePath /api
//
// @schemes http https
//
// @securityDefinitions.basic BasicAuth
//
// @x-logo {"url": "https://placeholder.com/logo.png", "altText": "CAATSM Dashboard"}
package http

