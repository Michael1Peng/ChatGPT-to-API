package chatgpt

import (
	"fmt"
	"freechatgpt/internal/tokens"
	chatgpt_types "freechatgpt/typings/chatgpt"
	official_types "freechatgpt/typings/official"
	"regexp"
	"strings"

	arkose "github.com/xqdoo00o/funcaptcha"
)

var gpt4Regexp = regexp.MustCompile(`^(gpt-4|gpt-4o)(?:-gizmo-g-(\w+))?$`)

func ConvertAPIRequest(api_request official_types.APIRequest, account string, secret *tokens.Secret, deviceId string, requireArk bool, dx string, proxy string) chatgpt_types.ChatGPTRequest {
	chatgpt_request := chatgpt_types.NewChatGPTRequest()
	var api_version int
	if secret.PUID == "" {
		api_request.Model = "gpt-3.5"
	}
	if strings.HasPrefix(api_request.Model, "gpt-3.5") {
		api_version = 3
		chatgpt_request.Model = "text-davinci-002-render-sha"
	} else if strings.HasPrefix(api_request.Model, "gpt-4") {
		api_version = 4
		matches := gpt4Regexp.FindStringSubmatch(api_request.Model)
		if len(matches) > 0 {
			chatgpt_request.Model = matches[1]
		} else {
			chatgpt_request.Model = "gpt-4"
		}
		if len(matches) == 3 && matches[2] != "" {
			chatgpt_request.ConversationMode.Kind = "gizmo_interaction"
			chatgpt_request.ConversationMode.GizmoId = "g-" + matches[2]
		}
	}
	if requireArk {
		token, err := arkose.GetOpenAIToken(api_version, secret.PUID, dx, proxy)
		if err == nil {
			chatgpt_request.ArkoseToken = token
		} else {
			fmt.Println("Error getting Arkose token: ", err)
		}
	}
	ifMultimodel := api_version == 4
	for _, api_message := range api_request.Messages {
		if api_message.Role == "system" {
			api_message.Role = "critic"
		}
		chatgpt_request.AddMessage(api_message.Role, api_message.Content, ifMultimodel, account, secret, deviceId, proxy)
	}
	return chatgpt_request
}

func RenewTokenForRequest(request *chatgpt_types.ChatGPTRequest, puid string, dx string, proxy string) {
	var api_version int
	if strings.HasPrefix(request.Model, "gpt-4") {
		api_version = 4
	} else {
		api_version = 3
	}
	token, err := arkose.GetOpenAIToken(api_version, puid, dx, proxy)
	if err == nil {
		request.ArkoseToken = token
	} else {
		fmt.Println("Error getting Arkose token: ", err)
	}
}
