package router

import (
	"net/http"
	"otel-to-do-app/mongodb"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func NewRouterTodoAPP(client *mongo.Client) *gin.Engine {
	r := gin.Default()
	r.Use(otelgin.Middleware("todo-service"))
	r.GET("/todo", func(ctx *gin.Context) {
		cur, err := client.Database("todo").Collection("todos").Find(ctx.Request.Context(), bson.M{})
		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}
		var result []mongodb.TodoModel
		err = cur.All(ctx, &result)
		if err != nil {
			ctx.AbortWithError(500, err)
			return
		}
		defer cur.Close(ctx.Request.Context())
		ctx.JSON(http.StatusOK, result)
	})
	return r
}
