package app

import (
	"encoding/json"
	"testing"

	"github.com/openbyte-os/sdk-go/translation"
)

func TestDefinitionHelpJsonRoundTrip(t *testing.T) {
	def := Definition{
		ID:   NewID("acme", "billing"),
		Name: translation.String("Billing"),
		Help: []HelpArticle{{
			ID:              "refunds",
			Type:            HelpArticleTypeHowTo,
			Title:           translation.String("Issue a refund"),
			Summary:         translation.String("Refund an invoice payment."),
			Content:         translation.String("Use this when a customer needs money returned against a paid invoice."),
			Icon:            "payments",
			Category:        "Payments",
			Tags:            []string{"billing", "refunds"},
			Keywords:        []string{"credit", "return"},
			DestinationPath: "help/refunds",
			Priority:        20,
			Pins: []HelpPin{{
				PathRegex: "^/invoices/",
				Priority:  100,
			}},
		}},
	}

	raw, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("marshal definition: %v", err)
	}

	decoded := Definition{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal definition: %v", err)
	}

	if len(decoded.Help) != 1 {
		t.Fatalf("help articles = %d, want 1", len(decoded.Help))
	}

	article := decoded.Help[0]
	if article.Type != HelpArticleTypeHowTo {
		t.Fatalf("article type = %q, want %q", article.Type, HelpArticleTypeHowTo)
	}
	if article.Title.Fallback != "Issue a refund" {
		t.Fatalf("title = %q", article.Title.Fallback)
	}
	if article.Content.Fallback == "" {
		t.Fatal("expected content to round trip")
	}
	if article.Pins[0].PathRegex != "^/invoices/" || article.Pins[0].Priority != 100 {
		t.Fatalf("pin = %#v", article.Pins[0])
	}
}

func TestDefinitionProvideHelpJsonRoundTrip(t *testing.T) {
	def := Definition{
		ID:          NewID("acme", "billing"),
		Name:        translation.String("Billing"),
		ProvideHelp: true,
	}

	raw, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("marshal definition: %v", err)
	}

	decoded := Definition{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal definition: %v", err)
	}

	if !decoded.ProvideHelp {
		t.Fatal("expected provideHelp to round trip")
	}
	if len(decoded.Help) != 0 {
		t.Fatalf("help articles = %d, want 0", len(decoded.Help))
	}
}

func TestHelpPath(t *testing.T) {
	if PathHelp != "/_kubex/help" {
		t.Fatalf("PathHelp = %q, want %q", PathHelp, "/_kubex/help")
	}
}
