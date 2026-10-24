package order

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
)

func formatOrderMessage(order sqlc.Order, frontendURL string) string {
	var sb strings.Builder

	if order.ToyID == nil {
		sb.WriteString("<b>🧸 New custom order</b>\n\n")
	} else {
		sb.WriteString("<b>🛍️ New order</b>\n\n")

		toyURL, err := url.JoinPath(frontendURL, "toys", order.ToyID.String())
		if err == nil {
			fmt.Fprintf(&sb, "<b>Toy:</b> <a href=\"%s\">open on web site</a>\n", html.EscapeString(toyURL))
		}
	}

	fmt.Fprintf(&sb, "<b>Phone number:</b> <a href=\"https://t.me/%s\">%s</a>\n",
		html.EscapeString(order.Phone),
		html.EscapeString(order.Phone),
	)

	if order.Message == nil {
		sb.WriteString("<b>Message:</b> not specified\n")
	} else {
		fmt.Fprintf(&sb, "<b>Message:</b>\n%s\n", html.EscapeString(*order.Message))
	}

	if order.Details == nil {
		sb.WriteString("<b>Details:</b> not specified\n")
	} else {
		fmt.Fprintf(&sb, "<b>Details:</b>\n%s\n", html.EscapeString(*order.Details))
	}

	fmt.Fprintf(&sb, "\n<b>ID:</b> <code>%s</code>", order.ID)
	adminOrderURL, err := url.JoinPath(frontendURL, "admin", "orders", order.ID.String())
	if err == nil {
		fmt.Fprintf(&sb, "\n<b>Admin panel:</b> <a href=\"%s\">view order</a>\n",
			html.EscapeString(adminOrderURL),
		)
	}

	return sb.String()
}
