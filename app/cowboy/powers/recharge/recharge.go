package recharge

import (
  "bytes"
  "encoding/json"
  "errors"
  "github.com/vishaltelangre/cowboy/Godeps/_workspace/src/github.com/gin-gonic/gin"
  "github.com/vishaltelangre/cowboy/app/cowboy/utils"
  "log"
  "net/http"
  "net/url"
  "text/template"
  "strings"
)

const (
  // slackRespTmpl is a customised template to display rich formatted movie details on Slack
  slackRespTmpl = `
  {{range $element := .}}
    *Value:* {{.Value}}
    *Talktime:* {{.Talktime}}
    *Validity:* {{.Validity}}
    *ShortDescription:* {{.ShortDescription}}
    *Description:* {{.Description}}
    *DescriptionMore:* {{.DescriptionMore}}
    *ProductType:* {{.ProductType}}
    *Circle:* {{.Circle}}
    *Operator:* {{.Operator}}
    *Category:* {{.Category}}
    *IsPrepaid:* {{.IsPrepaid}}
  {{end}}
`
)

// Plan defines structure of a plan with a lot details
type Plan struct {
  Value               string `json:"recharge_value, omitempty"`
  Talktime            string `json:"recharge_talktime, omitempty"`
  Validity            string `json:"recharge_validity, omitempty"`
  ShortDescription    string `json:"recharge_short_description, omitempty"`
  Description         string `json:"recharge_description, omitempty"`
  DescriptionMore     string `json:"recharge_description_more, omitempty"`
  ProductType         string `json:"product_type, omitempty"`
  Circle              string `json:"circle_master, omitempty"`
  Operator            string `json:"operator_master, omitempty"`
  Category            string `json:"recharge_master, omitempty"`
  IsPrepaid           string `json:"is_prepaid, omitempty"`
}

type APIRresponse struct {
  Status        int `json:"status_code, omitempty"`
  StatusText    string `json:"status_text, omitempty"`
  List          [] Plan `json:"data, omitempty"`
}

// findPlans fetches plans details from a third-party API
func findPlans(circle string, operator string) ([]Plan, error) {
  var apiURL *url.URL
  apiURL, err := url.Parse("http://api.dataweave.in/v1/telecom_data/listByCircle/?")
  if err != nil {
    return nil, err
  }

  parameters := url.Values{}
  parameters.Add("api_key", "4efd9ff6fefe4968624aa22362272d427dec05e0")
  parameters.Add("operator", operator)
  parameters.Add("circle", circle)
  apiURL.RawQuery = parameters.Encode()

  content, err := utils.GetContent(apiURL.String())

  var apiResp APIRresponse
  err = json.Unmarshal(content, &apiResp)

  if err != nil {
    return nil, err
  }

  if apiResp.Status != 200 {
    return nil, errors.New(apiResp.StatusText)
  }

  return apiResp.List, err
}

// formatSlackResp creates Slack-compatible plan details string
func formatSlackResp(list [] Plan) (string, error) {
  buf := new(bytes.Buffer)
  t := template.New("SlackRechargePlans")
  t, err := t.Parse(slackRespTmpl)
  if err != nil {
  log.Printf("Error: %s", err)
    return "", err
  }

  err = t.Execute(buf, list)
  if err != nil {
    log.Printf("Error: %s", err)
    panic(err)
    return "", err
  }

  return buf.String(), nil
}

// Separate out circle and operator from input text
func parsedReqText(str string) (string, string) {
  var splitedStr [] string
  splitedStr = utils.DeleteEmpty(strings.Split(str, " "))

  return splitedStr[0], splitedStr[1]
}

// Handler is a route handler for '/movie.:format' route
func Handler(c *gin.Context) {
  requestType := c.Param("format")
  text := c.Request.PostFormValue("text")

  operator, circle := parsedReqText(text)
  log.Printf("Recharge Query:: Circle := %s, Operator := %s", circle, operator)

  var list []Plan
  list, err := findPlans(circle, operator)

  switch requestType {
  case "json":
    if err != nil {
      c.JSON(http.StatusNotFound, gin.H{"Response": err.Error()})
    } else {
      c.IndentedJSON(http.StatusOK, list)
    }
  case "slack":
    text, err := formatSlackResp(list)
    if err != nil {
      c.String(http.StatusNotFound, "Not Found")
    }

    c.String(http.StatusOK, text)
  default:
    c.JSON(http.StatusUnsupportedMediaType, nil)
  }
}
