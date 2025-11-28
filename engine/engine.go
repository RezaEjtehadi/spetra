package engine

import (
	"fmt"
	"log"
	"net/http"
	"time"

	googletrans "github.com/Conight/go-googletrans"
	"github.com/RezaEjtehadi/bintra/bingtranslator"
	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	"github.com/spf13/viper"
)

func TranslateText(text string) (string, error) {
    
    if text == "" {
        return "", fmt.Errorf("error")
    }
    
    var result string
    engine := viper.GetString("engineTranslate")
    
    switch engine {
case "google":
        result = googleTranslate(text)
    case "bing":
        result = bingTranslate(text)
    default:
        return "", fmt.Errorf("error")
    }
    
   /* if result == "" {
        return "", fmt.Errorf("error")
    }*/


    if result != "google not conected!" {
        return result, nil
    }
	
    return result, nil
}

func Speak(text string) {
	if viper.GetString("engineSpeech") == "google" {
		googleSpeech(text)
	} else if viper.GetString("engineSpeech")=="bing"{
		bingSpeech(text)
		time.Sleep(5 * time.Second)

	}

}

func bingTranslate(text string) string {
	client := &http.Client{}
	var session *bingtranslator.BingSession
	gg := viper.GetString("translatorlanguage")

	translation, err := bingtranslator.Translate(client, session, bingtranslator.AutoDetect, gg, text)
	if err != nil {
		log.Printf("Translation error: %v", err)
		
		return "You're not connected" 
	}
	return translation
}

func googleTranslate(text string) string {
	config := googletrans.Config{UserAgent: []string{"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}}
	t := googletrans.New(config)
	gg := viper.GetString("translatorlanguage")
	translation, err := t.Translate(text, "auto", gg)
	if err != nil {
		return "You're not connected"
	}
	
	return translation.Text
}

func googleSpeech(text string)  {
	speech := htgotts.Speech{
		Folder:   "audio",
		Language: viper.GetString("speechlanguage"),
		Handler:  &handlers.Native{},
	}
	speech.Speak(text)

}

func bingSpeech(text string)  {
	client := &http.Client{}
	var session *bingtranslator.BingSession
	err := bingtranslator.Pronounce(client, session, text, bingtranslator.English, "Female", "en-US-AriaNeural")
	if err != nil {
		log.Fatal(err)
	}

	//time.Sleep(5 * time.Second)

}
