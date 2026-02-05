package handlers

import (
	"net/http"
	tele "social-network/shared/go/telemetry"
)

func (h *Handlers) testHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "test handler called")

		// err := utils.WriteJSON(ctx, w, http.StatusOK, map[string]string{
		// 	"message": "this request id is: " + r.Context().Value(ct.ReqID).(string),
		// })
		// if err != nil {
		// 	tele.Warn(ctx, "failed to send test ACK. @1", "error", err.Error())
		// 	utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send logout ACK")
		// 	return
		// }

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlPage))
	}
}

var htmlPage = `<!DOCTYPE html>
<html>
<head>
	<title>Example</title>
</head>
<body>
	<h1>Hello</h1>
</body>
<script>
	const socket = new WebSocket("ws://localhost:8082/live");
	let intervalId;

	socket.onopen = () => {
		console.log("WebSocket connected");

		intervalId = setInterval(() => {
			if (socket.readyState === WebSocket.OPEN) {
				socket.send('ch:{"category":"private", "conversation_id":"2VolejRejNmG", "body": "This is the body"}');
			}
		}, 800);
	};

	socket.onmessage = (event) => {
		// console.log(event.data);
	};

	socket.onerror = (error) => {
		console.error("WebSocket error:", error);
	};

	socket.onclose = (event) => {
		clearInterval(intervalId);
		console.log("WebSocket closed:", event.code, event.reason);
	};
</script>
</html>
`
