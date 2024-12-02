package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	//"github.com/gospider007/ja3"
	"github.com/gospider007/requests"
	spiderws "github.com/gospider007/websocket"
)

func root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

var upgrader = websocket.Upgrader{}

var (
	j  = "772,4865-4867-4866-49195-49199-52393-52392-49196-49200-49171-49172-156-157-47-53,0-23-65281-10-11-16-5-34-18-51-43-13-45-28-65037,4588-29-23-24-25-256-257,0"
	ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:134.0) Gecko/20100101 Firefox/134.0"
)

func convertToSpider(in int) (spiderws.MessageType, error) {
	switch in {
	case websocket.TextMessage:
		return spiderws.TextMessage, nil
	case websocket.BinaryMessage:
		return spiderws.BinaryMessage, nil
	case websocket.CloseMessage:
		return spiderws.CloseMessage, nil
	case websocket.PingMessage:
		return spiderws.PingMessage, nil
	case websocket.PongMessage:
		return spiderws.PongMessage, nil
	default:
		return spiderws.BinaryMessage, errors.New("Unknown message type given")
	}
}

func convertToWS(in spiderws.MessageType) (int, error) {
	switch in {
	case spiderws.TextMessage:
		return websocket.TextMessage, nil
	case spiderws.BinaryMessage:
		return websocket.BinaryMessage, nil
	case spiderws.CloseMessage:
		return websocket.CloseMessage, nil
	case spiderws.PingMessage:
		return websocket.PingMessage, nil
	case spiderws.PongMessage:
		return websocket.PongMessage, nil
	default:
		return -1, errors.New("Unknown message type given")
	}
}

func socket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}
	defer c.Close()

	/*ja3Spec, err := ja3.CreateSpecWithStr(j)
	if err != nil {
		log.Println("JA3 failed:", err)
		c.Close()
		return
	}*/

	res, err := requests.Get(nil, "wss://ws.fishtank.live/socket.io/?EIO=4&transport=websocket", requests.RequestOption{
		// Ja3Spec: ja3Spec,
		Headers: map[string]string{
			"Origin":     "https://www.fishtank.live",
			"User-Agent": ua,
		},
	})
	if err != nil {
		log.Println("Websocket failed:", err)
		m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
		if e, ok := err.(*websocket.CloseError); ok {
			if e.Code != websocket.CloseNoStatusReceived {
				m = websocket.FormatCloseMessage(e.Code, e.Text)
			}
		}
		c.WriteMessage(websocket.CloseMessage, m)
		return
	}
	defer res.CloseBody()

	ws := res.WebSocket()
	defer ws.Close()

	errIn := make(chan error, 1)
	errOut := make(chan error, 1)

	incoming := func(ws *spiderws.Conn, c *websocket.Conn, errc chan error) {
		for {
			msgType, msg, err := ws.ReadMessage()
			if err != nil {
				m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						m = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}

				errc <- err

				c.WriteMessage(websocket.CloseMessage, m)

				break
			}

			// log.Printf("C<-S %s\r\n", msg)

			newType, err := convertToWS(msgType)
			if err != nil {
				m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						m = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}

				errc <- err

				c.WriteMessage(websocket.CloseMessage, m)

				break
			}

			err = c.WriteMessage(newType, msg)
			if err != nil {
				m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						m = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}

				errc <- err

				c.WriteMessage(websocket.CloseMessage, m)

				break
			}
		}
	}
	outgoing := func(ws *spiderws.Conn, c *websocket.Conn, errc chan error) {
		for {
			msgType, msg, err := c.ReadMessage()
			if err != nil {
				m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						m = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}

				errc <- err

				c.WriteMessage(websocket.CloseMessage, m)

				break
			}

			// log.Printf("C->S %s\r\n", msg)

			newType, err := convertToSpider(msgType)
			if err != nil {
				m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						m = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}

				errc <- err

				c.WriteMessage(websocket.CloseMessage, m)

				break
			}

			err = ws.WriteMessage(newType, msg)
			if err != nil {
				m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err))
				if e, ok := err.(*websocket.CloseError); ok {
					if e.Code != websocket.CloseNoStatusReceived {
						m = websocket.FormatCloseMessage(e.Code, e.Text)
					}
				}

				c.WriteMessage(websocket.CloseMessage, m)

				break
			}
		}
	}

	go incoming(ws, c, errIn)
	go outgoing(ws, c, errOut)

	var message string
	select {
	case err = <-errIn:
		message = "Error when copying from server to client: %v"
	case err = <-errOut:
		message = "Error when copying from client to server: %v"

	}
	if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
		log.Printf(message, err)
	}
}

func auth(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Cookie")
	if auth == "" {
		log.Println("Failed to get cookie from proxied auth request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res, err := requests.Get(nil, "https://api.fishtank.live/v1/auth", requests.RequestOption{
		// Ja3Spec: ja3Spec,
		Headers: map[string]string{
			"Origin":     "https://www.fishtank.live",
			"User-Agent": ua,
			"Cookie":     auth,
		},
	})
	if err != nil {
		log.Println("Failed to proxy auth request:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer res.CloseBody()

	w.Write(res.Content())
}

func main() {
	http.HandleFunc("/socket.io/", socket)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/", root)
	log.Fatalln(http.ListenAndServe("localhost:6958", nil))
}
