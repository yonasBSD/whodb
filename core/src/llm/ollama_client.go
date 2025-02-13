package llm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func prepareOllamaRequest(prompt string, model LLMModel) (string, []byte, map[string]string, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  string(model),
		"prompt": prompt,
	})
	if err != nil {
		return "", nil, nil, err
	}
	url := fmt.Sprintf("%v/generate", getOllamaEndpoint())
	return url, requestBody, nil, nil
}

func prepareOllamaModelsRequest() (string, map[string]string) {
	url := fmt.Sprintf("%v/tags", getOllamaEndpoint())
	return url, nil
}

func parseOllamaResponse(body io.ReadCloser, receiverChan *chan string, responseBuilder *strings.Builder) (*string, error) {
	defer body.Close()
	reader := bufio.NewReader(body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		var completionResponse struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := json.Unmarshal([]byte(line), &completionResponse); err != nil {
			return nil, err
		}

		if receiverChan != nil {
			*receiverChan <- completionResponse.Response
		}

		if _, err := responseBuilder.WriteString(completionResponse.Response); err != nil {
			return nil, err
		}

		if completionResponse.Done {
			response := responseBuilder.String()
			return &response, nil
		}
	}

	return nil, nil
}

func parseOllamaModelsResponse(body io.ReadCloser) ([]string, error) {
	defer body.Close()

	var modelsResp struct {
		Models []struct {
			Name string `json:"model"`
		} `json:"models"`
	}
	if err := json.NewDecoder(body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	models := []string{}
	for _, model := range modelsResp.Models {
		models = append(models, model.Name)
	}
	return models, nil
}
