package mq

import (
	"log"

	"github.com/streadway/amqp"
)

type Mq struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

//如果异常关闭,会接收到通知
var notifyClose chan *amqp.Error

func NewMq(RabbitURL string) (*Mq) {
	//是否开启异步转移功能，开启的时候才初始化RabbitMQ连接
	conn, err := amqp.Dial(RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	channel.NotifyClose(notifyClose)
	//断线自动重连
	go func() {
		for {
			select {
			case msg := <-notifyClose:
				conn = nil
				channel = nil
				log.Panicf("onNotifyChannelClosed: %+v\n", msg)
				conn, err := amqp.Dial(RabbitURL)
				if err != nil {
					log.Println(err.Error())
					return
				}

				channel, err = conn.Channel()
				if err != nil {
					log.Println(err.Error())
					return
				}
			}
		}
	}()

	return &Mq{
		conn:    conn,
		channel: channel,
	}
}

//Publish:发布消息
func (q *Mq) Publish(exchange, routeKey string, msg []byte) bool {

	if nil == q.channel.Publish(
		exchange,
		routeKey,
		false, //如果没有对应的queue，就会丢弃这条消息
		false,
		amqp.Publishing{
			ContentType: "text/palin",
			Body:        msg}) {
		return true
	}
	return false
}
