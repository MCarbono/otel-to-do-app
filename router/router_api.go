package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"otel-to-do-app/mongodb"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

func NewRouterAPI() *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("api-service"))
	r.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, output{
			TraceID: trace.SpanFromContext(ctx.Request.Context()).SpanContext().TraceID().String(),
			Data:    "Hello",
			Error:   "",
		})
	})
	r.GET("/", func(ctx *gin.Context) {
		client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
		req, err := http.NewRequestWithContext(ctx.Request.Context(), "GET", "http://localhost:8080/todo", nil)
		if err != nil {
			err = fmt.Errorf("failed creating req: %w", err)
			ctx.JSON(http.StatusInternalServerError, output{
				TraceID: trace.SpanFromContext(ctx.Request.Context()).SpanContext().TraceID().String(),
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}
		res, err := client.Do(req)
		if err != nil {
			err = fmt.Errorf("failed doing req: %w", err)
			ctx.JSON(http.StatusInternalServerError, output{
				TraceID: trace.SpanFromContext(ctx.Request.Context()).SpanContext().TraceID().String(),
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			err = fmt.Errorf("failed reading body req: %w", err)
			ctx.JSON(http.StatusInternalServerError, output{
				TraceID: trace.SpanFromContext(ctx.Request.Context()).SpanContext().TraceID().String(),
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}
		var data []mongodb.TodoModel
		if err = json.Unmarshal(body, &data); err != nil {
			err = fmt.Errorf("failed parsing body req: %w", err)
			ctx.JSON(http.StatusInternalServerError, output{
				TraceID: trace.SpanFromContext(ctx.Request.Context()).SpanContext().TraceID().String(),
				Data:    nil,
				Error:   err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, output{
			TraceID: trace.SpanFromContext(ctx.Request.Context()).SpanContext().TraceID().String(),
			Data:    data,
		})
	})
	return r
}

type output struct {
	Error   string `json:"error"`
	TraceID string `json:"trace_id"`
	Data    any    `json:"data"`
}
