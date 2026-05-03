package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"bytes"
)

func GetTranscript(file string, apiKey string) string {
	audioBase64 := EncodeAudio(file)

 	payload := map[string]any {
        "model": "openai/whisper-1",
        "input_audio": map[string]string {
	        "data": audioBase64,
	        "format": "mp3",
        },
    }

    body, err := json.Marshal(payload)
    if err != nil {
   		panic(err)
    }

    req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/audio/transcriptions", bytes.NewReader(body))
    if err != nil {
    	panic(err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer " + apiKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
    	panic(err)
    }
    defer resp.Body.Close()

    respBody, _ := io.ReadAll(resp.Body)

    var respText map[string]any
    if err := json.Unmarshal(respBody, &respText);  err != nil {
    	panic(err)
    }
    return respText["text"].(string)
}

func PostProcess(transcript string, apiKey string) string {
	payload := map[string]any {
        "model": "google/gemini-3.1-flash-lite-preview",
        "messages": []map[string]string {
        	{
		        "role": "user",
		        "content": `You are a dictation post-processor. You receive raw speech-to-text output and return clean text ready to be typed into an application.

Your job:
- Remove filler words (um, uh, you know, like) unless they carry meaning.
- Fix spelling, grammar, and punctuation errors.
- When the transcript already contains a word that is a close misspelling of a name or term from the context or custom vocabulary,
correct the spelling. Never insert names or terms from context that the speaker did not say.
- Preserve the speakers intent, tone, and meaning exactly.

Output rules:
- Return ONLY the cleaned transcript text, nothing else. So NEVER output words like "Here is the cleaned transcript text:"
- If the transcription is empty, return exactly: EMPTY
- Do not add words, names, or content that are not in the transcription. The context is only for correcting spelling of words already
spoken.
- Do not change the meaning of what was said.

Example:
RAW_TRANSCRIPTION: "hey um so i just wanted to like follow up on the meating from yesterday i think we should definately move the
dedline to next friday becuz the desine team still needs more time to finish the mock ups and um yeah let me know if that works for you
ok thanks"

Then your response would be ONLY the cleaned up text, so here your response is ONLY:
"Hey, I just wanted to follow up on the meeting from yesterday. I think we should definitely move the deadline to next Friday because
the design team still needs more time to finish the mockups. Let me know if that works for you. Thanks."

Raw output:
` + transcript,
         	},
        },
        "reasoning": map[string]any {
        	"enabled": false,
        },
    }

    body, err := json.Marshal(payload)
    if err != nil {
   		panic(err)
    }

    req, err := http.NewRequest(http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(body))
    if err != nil {
    	panic(err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer " + apiKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
    	panic(err)
    }
    defer resp.Body.Close()

    respBody, _ := io.ReadAll(resp.Body)

    var respText map[string]any
    if err := json.Unmarshal(respBody, &respText);  err != nil { panic(err) }
    choices := respText["choices"].([]any)
    message := choices[0].(map[string]any)["message"].(map[string]any)
    content := message["content"].(string)
    return content
}
