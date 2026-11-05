package order

import (
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func formatOrderMessage(o order, frontendURL string) string {
	var sb strings.Builder

	if o.ToyID == nil {
		sb.WriteString("<b>🧸 New custom order</b>\n\n")
	} else {
		sb.WriteString("<b>🛍️ New order</b>\n\n")

		toyURL, err := url.JoinPath(frontendURL, "toys", o.ToyID.String())
		if err == nil {
			fmt.Fprintf(&sb, "<b>Toy:</b> <a href=\"%s\">open on web site</a>\n", html.EscapeString(toyURL))
		}
	}

	fmt.Fprintf(&sb, "<b>Phone number:</b> <a href=\"https://t.me/%s\">%s</a>\n",
		html.EscapeString(o.Phone),
		html.EscapeString(o.Phone),
	)

	if o.Message == nil {
		sb.WriteString("<b>Message:</b> not specified\n")
	} else {
		fmt.Fprintf(&sb, "<b>Message:</b>\n%s\n", html.EscapeString(*o.Message))
	}

	if o.Details == nil {
		sb.WriteString("<b>Details:</b> not specified\n")
	} else {
		fmt.Fprintf(&sb, "<b>Details:</b>\n%s\n", html.EscapeString(*o.Details))
	}

	fmt.Fprintf(&sb, "\n<b>ID:</b> <code>%s</code>", o.ID)
	adminOrderURL, err := url.JoinPath(frontendURL, "admin", "orders", o.ID.String())
	if err == nil {
		fmt.Fprintf(&sb, "\n<b>Admin panel:</b> <a href=\"%s\">view order</a>\n",
			html.EscapeString(adminOrderURL),
		)
	}

	return sb.String()
}

func formatMissedOrdersMessage(ids []uuid.UUID, frontendURL string) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "⚠️ <b>Missed orders: %d</b>\n\n", len(ids))

	sb.WriteString("Telegram did not deliver the messages about them right away\n")

	adminOrderURL, err := url.JoinPath(frontendURL, "admin", "orders")
	if err == nil {
		fmt.Fprintf(&sb, "\n<b>Admin panel:</b> <a href=\"%s\">view orders</a>\n",
			html.EscapeString(adminOrderURL),
		)
	}

	return sb.String()
}
