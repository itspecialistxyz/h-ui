package frontend

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed dist/*
var staticFiles embed.FS

func InitFrontend(router *gin.Engine, huiWebContext *string) {

	relativePath := "/"
	if huiWebContext != nil && strings.HasPrefix(*huiWebContext, "/") {
		relativePath = *huiWebContext
	}
	// Serve index.html for the root context
	router.GET(relativePath, func(c *gin.Context) {
		indexHTML, err := staticFiles.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error: index.html not found in embed FS")
			return
		}
		c.Data(http.StatusOK, "text/html", indexHTML)
	})
	// Serve index.html for all subpaths (SPA deep link support)
	router.GET(relativePath+"/*any", func(c *gin.Context) {
		indexHTML, err := staticFiles.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error: index.html for /*any not found")
			return
		}
		c.Data(http.StatusOK, "text/html", indexHTML)
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		faviconBytes, err := staticFiles.ReadFile("dist/favicon.ico")
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal Server Error: favicon.ico not found")
			return
		}
		c.Data(http.StatusOK, "image/x-icon", faviconBytes)
	})

	// Serve static assets from the "assets" subdirectory within the embedded dist
	assetsFS, err := fs.Sub(staticFiles, "dist/assets")
	if err != nil {
		// This would mean the "dist/assets" directory wasn't embedded or is empty
		// Handle this error appropriately, perhaps by logging and not setting up the route
		// For now, let's assume it will exist if dist/* is embedded correctly.
		// If dist/assets doesn't exist, http.FS will serve a 404, which is acceptable.
	} else {
		router.StaticFS(relativePath+"assets", http.FS(assetsFS))
	}

	router.NoRoute(func(c *gin.Context) {
		filePath := strings.TrimPrefix(c.Request.URL.Path, relativePath) // Ensure path is relative to web context
		filePath = strings.TrimPrefix(filePath, "/")                     // Ensure no leading slash for ReadFile

		// Avoid serving index.html again if it's a direct request to it via NoRoute
		if filePath == "index.html" || filePath == "" {
			indexHTML, err := staticFiles.ReadFile("dist/index.html")
			if err != nil {
				c.String(http.StatusInternalServerError, "Internal Server Error")
				return
			}
			c.Data(http.StatusOK, "text/html", indexHTML)
			return
		}

		fileContent, err := staticFiles.ReadFile("dist/" + filePath)
		if err != nil {
			// If file not found in embed FS, try serving index.html for SPA routing
			indexHTML, err := staticFiles.ReadFile("dist/index.html")
			if err != nil {
				c.String(http.StatusInternalServerError, "Internal Server Error")
				return
			}
			c.Data(http.StatusOK, "text/html", indexHTML)
			return
		}
		c.Data(http.StatusOK, http.DetectContentType(fileContent), fileContent)
	})
}

// getStaticFS is no longer needed as fs.Sub is used directly.
// func getStaticFS() fs.FS {
// 	staticFs, _ := fs.Sub(staticFiles, "assets") // Changed: removed "dist/"
// 	return staticFs
// }

// getFileContent is effectively replaced by the NoRoute handler logic using staticFiles.ReadFile directly.
// func getFileContent(filePath string) ([]byte, error) {
// 	cleanedPath := strings.TrimPrefix(filePath, "/") // Ensure no leading slash
// 	fileContent, err := staticFiles.ReadFile(cleanedPath) // Changed: removed "dist" prefix
// 	if err != nil {
// 		return nil, err
// 	}
// 	return fileContent, nil
// }
