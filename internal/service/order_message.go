package service

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/repository"
)

func formatOrderMessage(order repository.Order, siteURL string) string {
	var sb strings.Builder

	if order.ToyID == nil {
		sb.WriteString("<b>🧸 New custom order</b>\n\n")
	} else {
		sb.WriteString("<b>🧸 New order</b>\n\n")

		toyURL, err := url.JoinPath(siteURL, "toys", order.ToyID.String())
		if err == nil {
			fmt.Fprintf(&sb, "<b>Toy:</b> <a href=\"%s\">open on site</a>\n", html.EscapeString(toyURL))
		}
	}

	fmt.Fprintf(&sb, "<b>Phone number:</b> <a href=\"https://t.me/%s\">%s</a>\n",
		html.EscapeString(order.Phone),
		html.EscapeString(order.Phone),
	)

	if order.Message != "" {
		fmt.Fprintf(&sb, "<b>Message:</b>\n%s\n", html.EscapeString(order.Message))
	}

	fmt.Fprintf(&sb, "\n<b>ID:</b> <code>%s</code>", order.ID)

	return sb.String()
}
