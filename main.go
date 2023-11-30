package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type env struct {
	AppPort     string `mapstructure:"APP_PORT"`
	MongoUri    string `mapstructure:"MONGO_URI"`
	MongoDbName string `mapstructure:"MONGO_DB_NAME"`
}

func main() {
	// initialize env config
	var env env
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	if err := viper.Unmarshal(&env); err != nil {
		log.Fatal(err)
	}

	// initialize mongo client
	mongoOptions := options.Client()
	mongoOptions.ApplyURI(env.MongoUri)
	mongoClient, err := mongo.Connect(context.Background(), mongoOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err := mongoClient.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	defer mongoClient.Disconnect(context.Background())

	// initialize gofiber
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Get("/transaction/success", func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		session, err := mongoClient.StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		callback := func(sessionContext mongo.SessionContext) (any, error) {
			ctx := mongo.NewSessionContext(ctx, session)

			_, err := mongoClient.Database(env.MongoDbName).Collection("users").InsertOne(ctx, map[string]string{
				"name":        "Leo Messi",
				"age":         "34",
				"nationality": "Argentina",
			})
			if err != nil {
				return nil, err
			}

			_, err = mongoClient.Database(env.MongoDbName).Collection("users").InsertOne(ctx, map[string]string{
				"name":        "Neymar",
				"age":         "31",
				"nationality": "Brazil",
			})
			if err != nil {
				return nil, err
			}

			return nil, nil
		}

		result, err := session.WithTransaction(
			ctx,
			callback,
		)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]any{
				"success": false,
				"data":    err.Error(),
			})
		}

		return c.Status(http.StatusOK).JSON(map[string]any{
			"success": true,
			"data":    result,
		})
	})
	app.Get("/transaction/error", func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		session, err := mongoClient.StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		callback := func(sessionContext mongo.SessionContext) (any, error) {
			ctx := mongo.NewSessionContext(ctx, session)

			_, err := mongoClient.Database(env.MongoDbName).Collection("users").InsertOne(ctx, map[string]string{
				"name":        "Leo Messi",
				"age":         "34",
				"nationality": "Argentina",
			})
			if err != nil {
				return nil, err
			}

			_, err = mongoClient.Database(env.MongoDbName).Collection("users").InsertOne(ctx, map[string]string{
				"name":        "Neymar",
				"age":         "31",
				"nationality": "Brazil",
			})
			err = errors.New("simulasi error")
			if err != nil {
				return nil, err
			}

			return nil, nil
		}

		result, err := session.WithTransaction(
			ctx,
			callback,
		)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]any{
				"success": false,
				"data":    err.Error(),
			})
		}

		return c.Status(http.StatusOK).JSON(map[string]any{
			"success": true,
			"data":    result,
		})
	})
	app.Listen(":" + env.AppPort)
}
