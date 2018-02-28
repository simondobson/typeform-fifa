package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
    "log"
    "encoding/json"
    "bytes"
)

const (
  nameFieldId = "nhyYf7gcVwRq"
  winnerFieldId = "57238811"
  loserFieldId = "57239047"
  jumpToFieldId = "57239858"
  authToken = "Bearer A9JeLy7TDc44VsA7kifsXAgfRwXoLRKMuqTCmKrTJnjR"
  formURL = "https://api.typeform.com/forms/DOfp3b"
)

type Webhook struct {
  Form_Response struct {
    Answers []struct {
      Text string
      Field struct {
        Id string
      }
    }
  }
}

type Form struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Variables     struct {
		Score int `json:"score"`
	} `json:"variables"`
	Theme struct {
		Href string `json:"href"`
	} `json:"theme"`
	Settings struct {
		Language             string `json:"language"`
		IsPublic             bool   `json:"is_public"`
		ProgressBar          string `json:"progress_bar"`
		ShowProgressBar      bool   `json:"show_progress_bar"`
		ShowTypeformBranding bool   `json:"show_typeform_branding"`
		Meta                 struct {
			AllowIndexing bool `json:"allow_indexing"`
		} `json:"meta"`
	} `json:"settings"`
	WelcomeScreens []struct {
		Ref        string `json:"ref"`
		Title      string `json:"title"`
		Properties struct {
			ShowButton bool   `json:"show_button"`
			ButtonText string `json:"button_text"`
		} `json:"properties"`
		Attachment struct {
			Type string `json:"type"`
			Href string `json:"href"`
		} `json:"attachment"`
	} `json:"welcome_screens"`
	ThankyouScreens []struct {
		Ref        string `json:"ref"`
		Title      string `json:"title"`
		Properties struct {
			ShowButton bool   `json:"show_button"`
			ShareIcons bool   `json:"share_icons"`
			ButtonMode string `json:"button_mode"`
			ButtonText string `json:"button_text"`
		} `json:"properties"`
		Attachment struct {
			Type string `json:"type"`
			Href string `json:"href"`
		} `json:"attachment"`
	} `json:"thankyou_screens"`
	Fields []Field `json:"fields"`
	Logic []Logic `json:"logic"`
}

type Field struct {
		ID         string `json:"id"`
		Title      string `json:"title"`
		Ref        string `json:"ref"`
		Properties struct {
			Description string `json:"description,omitempty"`
			HideMarks   bool   `json:"hide_marks,omitempty"`
			ButtonText  string `json:"button_text,omitempty"`
      Choices     []Choice `json:"choices,omitempty"`
		} `json:"properties,omitempty"`
		Attachment *Attachment `json:"attachment,omitempty"`
		Type        string `json:"type"`
		Validations struct {
			Required bool `json:"required,omitempty"`
		} `json:"validations,omitempty"`
}

type Attachment struct {
  Type string `json:"type,omitempty"`
  Href string `json:"href,omitempty"`
}

type Choice struct {
    Label string `json:"label"`
}

type Logic struct {
  Actions []Action `json:"actions"`
  Type string `json:"type"`
  Ref  string `json:"ref"`
}

type Action struct {
  Action  string `json:"action"`
  Details Details `json:"details"`
  Condition Condition `json:"condition"`
}

type Details struct {
  To TypeVal `json:"to"`
}

type Condition struct {
  Op string `json:"op"`
  Vars []LogicCondition `json:"vars"`
}

type LogicCondition struct {
  Op   string `json:"op"`
  Vars []TypeVal `json:"vars"`
}

type TypeVal struct {
  Type  string `json:"type"`
  Value string `json:"value"`
}

func updateForm(NewChallenger string) {
    resp, err := http.Get(formURL)
    if err != nil {
      log.Println(err)
      return
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
      log.Println(err)
      return
    }

    var f Form
    err = json.Unmarshal(body, &f)
    if err != nil {
      log.Println(err)
      return
    }

    newChoice := Choice{
      Label: NewChallenger,
    }

    var winnerFieldRef string
    var loserFieldRef string
    var jumpToFieldRef string

    for i, field := range f.Fields {
      if field.ID == winnerFieldId {
        winnerFieldRef = f.Fields[i].Ref
        f.Fields[i].Properties.Choices = append(field.Properties.Choices, newChoice)
      }
      if field.ID == loserFieldId {
        loserFieldRef = f.Fields[i].Ref
        f.Fields[i].Properties.Choices = append(field.Properties.Choices, newChoice)
      }
      if field.ID == jumpToFieldId {
        jumpToFieldRef = f.Fields[i].Ref
      }
    }

    newLogicJump := Action{
      Action: "jump",
      Details: Details{To: TypeVal{Type: "field", Value: jumpToFieldRef}},
      Condition: Condition{
        Op: "and",
        Vars: []LogicCondition{
          {
            Op: "equal",
            Vars: []TypeVal{TypeVal{Type: "field", Value: winnerFieldRef},TypeVal{Type: "constant", Value: NewChallenger}},
          },
          {
            Op: "equal",
            Vars: []TypeVal{TypeVal{Type: "field", Value: loserFieldRef},TypeVal{Type: "constant", Value: NewChallenger}},
          },
        },
      },
    }

    for i, logic := range f.Logic {
      if logic.Ref == loserFieldRef {
        f.Logic[i].Actions = append(logic.Actions, newLogicJump)
      }
    }

    for i, tys := range f.ThankyouScreens {
      if tys.Ref == "default_tys" {
        log.Println("helloooooooo")
        f.ThankyouScreens = append(f.ThankyouScreens[:i], f.ThankyouScreens[i+1:]...)
      }
    }
    
    updatedForm, err :=json.Marshal(f)
    if err != nil {
      log.Println(err)
      return
    }

    req, err := http.NewRequest("PUT", formURL, bytes.NewBuffer(updatedForm))
    req.Header.Set("Authorization", authToken)
    req.Header.Set("Content-Type", "application/json")


    client := &http.Client{}
    post_resp, err := client.Do(req)
    if err != nil {
      log.Println(err)
      return
    }

    if post_resp.StatusCode != 200 {
        log.Println("Error. Updating the form was unsuccessfull")
        log.Println(post_resp)
        return
    }

}

func handler(w http.ResponseWriter, req *http.Request) {
    switch req.Method {
    case "GET":
      fmt.Fprintf(w, htmlStr)
    case "POST":
      if req.URL.Path != "/add_new_challenger" {
  		    http.Error(w, "404 not found.", http.StatusNotFound)
  		    return
  	  }
      body, err := ioutil.ReadAll(req.Body)
      if err != nil {
        log.Println(err)
        return
      }

      var wh Webhook

      err = json.Unmarshal(body, &wh)
      if err != nil {
        log.Println(err)
        return
      }

      for _, answer := range wh.Form_Response.Answers {
        if answer.Field.Id == nameFieldId {
            updateForm(answer.Text)
        }
      }
    }
}

func removeElement(s []int, i int) []int {
    s[len(s)-1], s[i] = s[i], s[len(s)-1]
    return s[:len(s)-1]
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

var htmlStr = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
</head>
<body>
  <div>
      <h1>Typeform FIFA Challenge</h1>
  </div>
</body>
</html>
`
