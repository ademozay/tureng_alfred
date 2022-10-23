package main

import (
	"fmt"
	"regexp"
	"strings"

	aw "github.com/deanishe/awgo"
	"github.com/hzenginx/tureng/tureng"
	"golang.org/x/text/unicode/norm"
)

var wf *aw.Workflow

func init() {
	wf = aw.New()
}

func splitInput(query string) (string, []string) {
	splited := strings.Split(query, ":")
	return splited[0], splited[1:]
}

func getIcon(category string) *aw.Icon {
	var icon string
	matched, _ := regexp.MatchString(`en->tr`, category)
	if matched {
		icon = "en_tr.png"
	} else {
		icon = "tr_en.png"
	}
	return &aw.Icon{
		Value: icon,
	}
}

func getAccentIcon(country string) *aw.Icon {
	var icon string
	if country == "us" {
		icon = "flag-us.png"
	} else if country == "uk" {
		icon = "flag-uk.png"
	} else if country == "au" {
		icon = "flag-au.png"
	}
	return &aw.Icon{
		Value: icon,
	}
}

func handleVoiceResponse(voices tureng.VoiceResponse, wf *aw.Workflow) {
	for _, voice := range voices {
		wf.
			NewItem(voice.AccentName).
			Icon(getAccentIcon(voice.Country)).
			Arg(voice.VoiceUrl).
			Valid(true)
	}
	wf.SendFeedback()
}

func handleSearchResponse(response *tureng.SearchResponse, word string, wf *aw.Workflow) {
	if response.IsSuccessful {
		if response.Result.IsFound == 1 {
			for _, result := range response.Result.Results {
				arg := result.Term
				if response.Result.IsEnglishToTurkish == 1 {
					arg = word
				}
				icon := getIcon(result.Category)
				wf.
					NewItem(result.Term).
					Subtitle(result.Category).
					Icon(icon).
					Arg(arg).
					Largetype(fmt.Sprintf("\n\n\n%v\n%v", word, result.Term)).
					Valid(true)
			}
		} else {
			for _, suggestion := range response.Result.Suggestions {
				wf.
					NewWarningItem(suggestion, "Did you mean this?").
					Autocomplete(fmt.Sprintf("translate:%s", suggestion))
			}
		}
	} else {
		wf.Fatal(response.Exception)
	}
	wf.SendFeedback()
}

func handleAutocompleteResponse(response *tureng.AutoCompleteResponse, wf *aw.Workflow) {
	for _, word := range response.Words {
		wf.
			NewItem(word).
			Autocomplete(fmt.Sprintf("translate:%s", word)).
			Icon(&aw.Icon{Value: "tureng.png"})
	}
	wf.SendFeedback()
}

func run() {
	var query = wf.Args()[0]
	query = norm.NFC.String(query)
	command, args := splitInput(query)

	if command == "translate" {
		if len(args) > 0 {
			word := args[0]
			response, err := tureng.Search(word)
			if err != nil {
				wf.FatalError(err)
			} else {
				handleSearchResponse(response, word, wf)
			}
		}
	} else if command == "voice" {
		word := args[0]
		voices, err := tureng.Voice(word)
		if err != nil {
			wf.FatalError(err)
		} else {
			handleVoiceResponse(voices, wf)
		}
	} else {
		response, err := tureng.AutoComplete(command)
		if err != nil {
			wf.FatalError(err)
		} else {
			handleAutocompleteResponse(response, wf)
		}
	}
}

func main() {
	wf.Run(run)
}
