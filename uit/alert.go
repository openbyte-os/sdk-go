package uit

import (
	"encoding/json"
	"net/http"

	"github.com/openbyte-os/sdk-go/transport"
)

type AlertStyle string

const (
	AlertStyleSuccess AlertStyle = "success"
	AlertStyleDanger  AlertStyle = "error"
	AlertStyleWarning AlertStyle = "warning"
	AlertStyleInfo    AlertStyle = "info"
)

type AlertButton struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	Link    string `json:"link"`
	Uri     string `json:"uri"`
	Target  string `json:"target"`
}
type Alert struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Style       AlertStyle    `json:"style"`
	Icon        string        `json:"icon"`
	Closer      bool          `json:"closer"`
	Actions     []AlertButton `json:"actions"`
}

func NewAlert(title, description, icon string) *Alert {
	return &Alert{
		Title:       title,
		Description: description,
		Icon:        icon,
	}
}

func (a *Alert) WithAction(actions ...AlertButton) *Alert {
	a.Actions = append(a.Actions, actions...)
	return a
}

func (a *Alert) WithClose() *Alert {
	a.Closer = true
	return a
}

func (a *Alert) Info() *Alert    { a.Style = AlertStyleInfo; return a }
func (a *Alert) Warning() *Alert { a.Style = AlertStyleWarning; return a }
func (a *Alert) Success() *Alert { a.Style = AlertStyleSuccess; return a }
func (a *Alert) Danger() *Alert  { a.Style = AlertStyleDanger; return a }

func (a *Alert) WriteHeader(w http.ResponseWriter) {
	bytes, _ := json.Marshal(a)
	w.Header().Set(transport.ResponseAlert, string(bytes))
}

func AlertText(w http.ResponseWriter, title string) {
	w.Header().Set(transport.ResponseAlert, title)
}

func AlertInfo(w http.ResponseWriter, title string) {
	w.Header().Set(transport.ResponseAlertInfo, title)
}
func AlertSuccess(w http.ResponseWriter, title string) {
	w.Header().Set(transport.ResponseAlertSuccess, title)
}
func AlertWarning(w http.ResponseWriter, title string) {
	w.Header().Set(transport.ResponseAlertWarning, title)
}
func AlertDanger(w http.ResponseWriter, title string) {
	w.Header().Set(transport.ResponseAlertDanger, title)
}
