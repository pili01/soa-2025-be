package router

import (
    "net/http"
    "strings"

    "gateway/internal/config"
    "gateway/internal/handlers"
    "gateway/internal/middleware"
    "gateway/internal/proxy"

    "github.com/gin-gonic/gin"
    "github.com/rs/zerolog/log"
)

type Router struct {
    engine          *gin.Engine
    config          *config.Config
    serviceRegistry *proxy.ServiceRegistry
    healthHandler   *handlers.HealthHandler
}

func NewRouter(cfg *config.Config) (*Router, error) {
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
        
        toursGroup := api.Group("/tours")
        {
            toursGroup.POST("/create", r.handleServiceRequest("tours"))
            toursGroup.GET("/my-tours", r.handleServiceRequest("tours"))
            toursGroup.GET("/:tourId", r.handleServiceRequest("tours"))
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
        
        // Purchase service routes
        api.Any("/cart", r.handlePurchaseProxyRequest())
        api.Any("/cart/*path", r.handlePurchaseProxyRequest())
        api.Any("/checkout", r.handlePurchaseProxyRequest())
        api.Any("/checkout/*path", r.handlePurchaseProxyRequest())
        api.Any("/purchases", r.handlePurchaseProxyRequest())
        api.Any("/purchases/*path", r.handlePurchaseProxyRequest())
        api.Any("/validate-token", r.handlePurchaseProxyRequest())
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


func (r *Router) handleServiceRequest(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceProxy, exists := r.serviceRegistry.GetService(serviceName)
		if !exists {
			log.Error().Str("service", serviceName).Msg("Service not found")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service not available",
				"service": serviceName,
			})
			return
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
			Str("new_path", c.Request.URL.Path).
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

func (r *Router) handlePurchaseProxyRequest() gin.HandlerFunc {
    return func(c *gin.Context) {
        serviceProxy, exists := r.serviceRegistry.GetService("purchase")
        if !exists {
            log.Error().Str("service", "purchase").Msg("Purchase service not found")
            c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Purchase service not available"})
            c.Abort()
            return
        }
        
        originalPath := c.Request.URL.Path
        // Putanja u gateway-u je /api/cart/*path, a servis ocekuje /cart/*path
        newPath := strings.TrimPrefix(originalPath, "/api")
        c.Request.URL.Path = newPath

        // Prosleđujemo sve header-e, uključujući Authorization
        log.Debug().
            Str("original_path", originalPath).
            Str("new_path", c.Request.URL.Path).
            Str("authorization", c.Request.Header.Get("Authorization")).
            Msg("Routing purchase request")

        serviceProxy.ServeHTTP(c.Writer, c.Request)
    }
}

func registerServices(registry *proxy.ServiceRegistry, services config.ServicesConfig) error {
    serviceMappings := map[string]string{
        "blog":         services.Blog,
        "image":        services.Image,
        "stakeholders": services.Stakeholders,
        "tours":        services.Tours,
        "purchase":     services.Purchase,
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


