package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoInstance contains the Mongo client and database objects


var dataBase *mongo.Database


func main()  {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	Connect()
	setupHttp()
}

func Connect()  {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URI")))
	if err != nil {
		log.Fatal(err)
	}

	dataBase = client.Database("imgs")
}


func setupHttp(){
	engine:= html.New("./views", ".html")

	engine.Reload(true)

	app:= fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/:id",func(c *fiber.Ctx) error {
		var file bson.M

		fileName, err := url.PathUnescape(c.Params("id"))

		if err != nil {
			return err
		}

		fmt.Println(c.Params(fileName))

		if err := dataBase.Collection("files").FindOne(context.TODO(), bson.M{"fileName": fileName}).Decode(&file); err != nil {
			return c.Status(404).Render("404", nil)
		}

		fmt.Println(file)
		mimeType:= strings.Split(file["mimeType"].(string), "/")[0]
		cdnUrl:= os.Getenv("CDN_URL") + file["cdnFileName"].(string)
		embed:= file["embed"].(primitive.M)

		return c.Render("index", fiber.Map{
			"Image": mimeType == "image",
			"FileURL": cdnUrl,
			"OEmbedURL": "https://" + c.Hostname() + "/json/" + fileName + ".json",
			"Desc": embed["description"],
			"Color": embed["color"],
			"Name": file["fileName"],
			"OGName": file["originalFileName"],
			"Video": mimeType == "video",
		})
	})

	app.Get("/json/:id",func(c *fiber.Ctx) error {
		var file bson.M

		if !strings.HasSuffix(c.Params("id"), ".json"){
			return c.Status(404).Render("404", nil)

		}
		fileName, err := url.PathUnescape(strings.Replace(c.Params("id"), ".json", "", 1))

		if err != nil {
			return c.Status(404).Render("404", nil)
		}

		fmt.Println(c.Params(fileName))

		if err := dataBase.Collection("files").FindOne(context.TODO(), bson.M{"fileName": fileName}).Decode(&file); err != nil {
			return c.Status(404).Render("404", nil)
		}

		embed:= file["embed"].(primitive.M)

		author:= embed["author"].(primitive.M)

		header:= embed["header"].(primitive.M)


		return c.JSON(fiber.Map{
			"version": "1.0",
			"type":    "link",
			"title": embed["title"],
			"author_name": author["text"],
			"author_url": author["url"],
			"provider_name": header["text"],
			"provider_url": header["url"],
		})
		
		
	})

	log.Fatal(app.Listen(":3000"))
}

