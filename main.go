package main

import (
        "context"
        "log"

        "github.com/go-redis/redis/v8"
        "github.com/gofiber/fiber/v2"
        "github.com/gofiber/websocket/v2"
)

func main() {
        app := fiber.New()
        webs := websocket.New(func(c *websocket.Conn) {
                ctx := context.Background()
                quitSubscribeGoRutine := make(chan bool)
                redisClient := redis.NewClient(&redis.Options{Addr: "redis:6379", Password: "", DB: 0})
                subscriber := redisClient.Subscribe(ctx, c.Params("id"))
		log.Println("subscribe : ", c.Params("id"))
                var (
                        msg []byte
                        err error
                )
                go func() {
                        for {
                                select {
                                case <-quitSubscribeGoRutine:
                                        return
                                default:
                                        msg, err := subscriber.ReceiveMessage(ctx)
                                        if err != nil {
                                                log.Println("err occur  : ", err)
                                                break
                                        }
                                        c.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
                                }
                        }
                }()
                defer func() {
                        redisClient.Close()
                        quitSubscribeGoRutine <- true
                }()
                for {
                        if _, msg, err = c.ReadMessage(); err != nil {
                                log.Println("read:", err)
                                break
                        }
                        if err := redisClient.Publish(ctx, c.Params("id"), msg); err == nil {
                                log.Println("public err : ", err)
                                break
                        }
                }
        })
        app.Use("/ws", func(c *fiber.Ctx) error {
                if websocket.IsWebSocketUpgrade(c) {
                        c.Locals("allowed", true)
                        return c.Next()
                }
                return fiber.ErrUpgradeRequired
        })
        app.Get("/ws/:id", webs)
        app.Listen(":3000")
}

