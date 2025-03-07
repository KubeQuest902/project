package projectserver

import (
	"encoding/json"
	"net/http"

	models "github.com/KubeQuest902/project/models"
	"github.com/mediocregopher/radix/v3"
)

var (
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisPool     *radix.Pool
)

var Animals = []string{"cat", "dog"}

func writeStandardResponse(r *http.Request, w http.ResponseWriter, message string) {
	responseObj := &models.Response{
		Message: message,
	}

	response, _ := json.Marshal(responseObj)
	w.Write(response)
}
