package router

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gateway/internal/config"
	"gateway/internal/handlers"
	"gateway/internal/middleware"
	"gateway/internal/proxy"
	pb "gateway/proto/compiled"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Router struct {
	engine          *gin.Engine
	config          *config.Config
	serviceRegistry *proxy.ServiceRegistry
	healthHandler   *handlers.HealthHandler
	toursClient     pb.TourServiceClient
}

func NewRouter(cfg *config.Config, toursClient pb.TourServiceClient) (*Router, error) {
	if cfg.Server.Port == "8080" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	serviceRegistry := proxy.NewServiceRegistry()

	if err := registerServices(serviceRegistry, cfg.Services); err != nil {
		return nil, err
	}

	router := &Router{
		engine:          engine,
		config:          cfg,
		serviceRegistry: serviceRegistry,
		healthHandler:   handlers.NewHealthHandler(),
		toursClient:     toursClient,
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router, nil
}

func (r *Router) setupMiddleware() {
	r.engine.Use(gin.Recovery())
	r.engine.Use(middleware.RequestID())
	r.engine.Use(middleware.Logger())
	r.engine.Use(middleware.CORS())
	r.engine.Use(middleware.RateLimit(100))

	r.engine.Use(middleware.JWTAuth(middleware.JWTConfig{
		Secret: r.config.Auth.JWTSecret,
	}))
}

func (r *Router) setupRoutes() {
	r.engine.GET("/health", r.healthHandler.HealthCheck)

	api := r.engine.Group("/api")
	{
		api.POST("/register", r.handleAuth("register"))
		api.POST("/login", r.handleAuth("login"))

		// Koristi novu funkciju koja ne menja putanju za blog servis
		api.Any("/blogs/*path", r.handleBlogProxyRequest())
		api.Any("/blog/*path", r.handleBlogProxyRequest())

		api.Any("/images/*path", r.handleImageProxyRequest())
		api.Any("/image/*path", r.handleImageProxyRequest())

		api.Any("/stakeholders/*path", r.handleStakeholdersProxyRequest())
		api.Any("/stakeholder/*path", r.handleStakeholdersProxyRequest())

		api.Any("/follow/*path", r.handleFollowerProxyRequest())
		toursGroup := api.Group("/tours")
		{
			toursGroup.POST("/create", r.handleCreateTour())  // Adapted to use gRPC client
			toursGroup.GET("/my-tours", r.handleGetMyTours()) // Adapted to use gRPC client
			toursGroup.GET("/:tourId", r.handleGetTourByID()) // Adapted to use gRPC client
			toursGroup.PUT("/:tourId/set-price", r.handleSetTourPrice()) // NOVA gRPC ruta
			toursGroup.GET("/:tourId/get-published", r.handleServiceRequest("tours"))
			toursGroup.PUT("/:tourId", r.handleServiceRequest("tours"))
			toursGroup.DELETE("/:tourId", r.handleServiceRequest("tours"))
			toursGroup.POST("/:tourId/publish", r.handleServiceRequest("tours"))
			toursGroup.POST("/:tourId/archive", r.handleServiceRequest("tours"))
			toursGroup.POST("/:tourId/set-price", r.handleServiceRequest("tours"))

			toursGroup.POST("/:tourId/create-keypoint", r.handleServiceRequest("tours"))
			toursGroup.GET("/:tourId/keypoints", r.handleServiceRequest("tours"))
			toursGroup.GET("/keypoints/:keypointId", r.handleServiceRequest("tours"))
			toursGroup.PUT("/keypoints/:keypointId", r.handleServiceRequest("tours"))
			toursGroup.DELETE("/keypoints/:keypointId", r.handleServiceRequest("tours"))
		}
	}

	r.engine.NoRoute(func(c *gin.Context) {
		log.Warn().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Msg("Route not found")

		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"path":  c.Request.URL.Path,
		})
	})
}

func (r *Router) handleCreateTour() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, userIDExists := c.Get("user_id")
		userRole, userRoleExists := c.Get("role")

		if !userIDExists || !userRoleExists {
			log.Error().Msg("User data missing in context after JWT authentication")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication data not found"})
			return
		}

		if userRole != "Guide" {
			log.Warn().Str("user_id", fmt.Sprintf("%v", userID)).Msg("Unauthorized access: user is not a GUIDE")
			c.JSON(http.StatusForbidden, gin.H{"error": "Only guides can create tours"})
			return
		}

		var reqBody pb.CreateTourRequest
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			log.Error().Err(err).Msg("Invalid request body for CreateTour")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		uidStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id type"})
			return
		}

		uidInt, err := strconv.Atoi(uidStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id format"})
			return
		}

		reqBody.UserId = int32(uidInt)

		resp, err := r.toursClient.CreateTour(c, &reqBody)
		if err != nil {
			log.Error().Err(err).Msg("Failed to call CreateTour via gRPC")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (r *Router) handleGetMyTours() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, userIDExists := c.Get("user_id")
		userRole, userRoleExists := c.Get("role")

		if !userIDExists || !userRoleExists {
			log.Error().Msg("User data missing in context after JWT authentication")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication data not found"})
			return
		}

		if userRole != "Guide" {
			log.Warn().Str("user_id", fmt.Sprintf("%v", userID)).Msg("Unauthorized access: user is not a GUIDE")
			c.JSON(http.StatusForbidden, gin.H{"error": "Only guides can view their tours"})
			return
		}

		uidStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id type"})
			return
		}

		uidInt, err := strconv.Atoi(uidStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user_id format"})
			return
		}

		req := &pb.GetToursByAuthorIDRequest{
			UserId: int32(uidInt),
		}

		resp, err := r.toursClient.GetToursByAuthorID(c, req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to call GetToursByAuthorID via gRPC")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tours"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (r *Router) handleGetTourByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		tourIDStr := c.Param("tourId")
		tourID, err := strconv.Atoi(tourIDStr)
		if err != nil {
			log.Error().Err(err).Msg("Invalid tour ID format")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID format"})
			return
		}

		req := &pb.GetTourByIDRequest{
			TourId: int32(tourID),
		}

		resp, err := r.toursClient.GetTourByID(c, req)
		if err != nil {
			log.Error().Err(err).Msg("Failed to call GetTourByID via gRPC")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tour"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (r *Router) handleSetTourPrice() gin.HandlerFunc {
	return func(c *gin.Context) {
		tourIDStr := c.Param("tourId")
		tourID, err := strconv.Atoi(tourIDStr)
		if err != nil {
			log.Error().Err(err).Msg("Invalid tour ID format for SetTourPrice")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID format"})
			return
		}

		var reqBody pb.SetTourPriceRequest
		if err := c.ShouldBindJSON(&reqBody); err != nil {
			log.Error().Err(err).Msg("Invalid request body for SetTourPrice")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		reqBody.TourId = int32(tourID)

		resp, err := r.toursClient.SetTourPrice(c, &reqBody)
		if err != nil {
			log.Error().Err(err).Msg("Failed to call SetTourPrice via gRPC")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set tour price"})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func (r *Router) handleServiceRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var serviceProxy *proxy.ServiceProxy
		var exists bool

		if serviceName == "tours" {
			serviceProxy, _ = proxy.NewServiceProxy(r.config.Services.ToursAPI)
		} else {
			serviceProxy, exists = r.serviceRegistry.GetService(serviceName)
			if !exists {
				log.Error().Str("service", serviceName).Msg("Service not found")
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "Service not available",
					"service": serviceName,
				})
				return
			}
		}

		originalPath := c.Request.URL.Path
		servicePrefix := "/api/" + serviceName
		newPath := strings.TrimPrefix(originalPath, servicePrefix)
		if newPath == "" {
			newPath = "/"
		}
		finalPath := "/api" + newPath
		c.Request.URL.Path = finalPath

		log.Debug().
			Str("service", serviceName).
			Str("original_path", originalPath).
			Str("new_path", finalPath).
			Msg("Routing request to service")

		serviceProxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (r *Router) handleAuth(action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceProxy, exists := r.serviceRegistry.GetService("stakeholders")
		if !exists {
			log.Error().Str("service", "stakeholders").Msg("Service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service not available",
				"service": "stakeholders",
			})
			return
		}

		c.Request.URL.Path = "/api/" + action
		c.Request.RequestURI = ""

		serviceProxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (r *Router) handleBlogProxyRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceProxy, exists := r.serviceRegistry.GetService("blog")
		if !exists {
			log.Error().Str("service", "blog").Msg("Blog service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Blog service not available"})
			c.Abort()
			return
		}

		// NE MODIFIKUJEMO PUTANJU.

		serviceProxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (r *Router) handleFollowerProxyRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceProxy, exists := r.serviceRegistry.GetService("follower")
		if !exists {
			log.Error().Str("service", "follower").Msg("Follower service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Follower service not available"})
			c.Abort()
			return
		}

		serviceProxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (r *Router) handleStakeholdersProxyRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceProxy, exists := r.serviceRegistry.GetService("stakeholders")
		if !exists {
			log.Error().Str("service", "stakeholders").Msg("Stakeholders service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Stakeholders service not available"})
			c.Abort()
			return
		}

		originalPath := c.Request.URL.Path
		newPath := strings.TrimPrefix(originalPath, "/api/stakeholders")
		finalPath := "/api" + newPath
		c.Request.URL.Path = finalPath

		serviceProxy.ServeHTTP(c.Writer, c.Request)
	}
}

func (r *Router) handleImageProxyRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceProxy, exists := r.serviceRegistry.GetService("image")
		if !exists {
			log.Error().Str("service", "image").Msg("Image service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Image service not available"})
			c.Abort()
			return
		}

		originalPath := c.Request.URL.Path
		// Putanja u gateway-u je /api/images/*path, a servis ocekuje /api/*path
		newPath := strings.TrimPrefix(originalPath, "/api/images")
		finalPath := "/api" + newPath
		c.Request.URL.Path = finalPath

		serviceProxy.ServeHTTP(c.Writer, c.Request)
	}
}

func registerServices(registry *proxy.ServiceRegistry, services config.ServicesConfig) error {
	serviceMappings := map[string]string{
		"blog":         services.Blog,
		"image":        services.Image,
		"stakeholders": services.Stakeholders,
		"tours":        services.Tours,
		"follower":     services.Follower,
	}

	for name, url := range serviceMappings {
		if err := registry.RegisterService(name, url); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
