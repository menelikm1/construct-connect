package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"

	"qetero/internal/models"
	"qetero/internal/repository"
)

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	b.send(msg.Chat.ID, `*Welcome to Qetero* 🇪🇹

Ethiopia's equipment rental marketplace.

To get started, link your account with your phone number:
/link +251912345678

Don't have an account yet? Register at our API or ask an admin.

Type /help to see all commands.`)
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	b.send(msg.Chat.ID, `*Available commands:*

/browse - Browse all available equipment
/search [category] [location] - Filter listings
/listing [number] - View details of a listing
/book [number] [start] [end] - Book equipment (dates: YYYY-MM-DD)
/mybookings - View your active bookings
/link [phone] - Link your account by phone number

*Categories:* excavator, crane, scaffold, compactor, loader, forklift, generator, water\_truck, concrete\_mixer, dump\_truck, dozer, roller, other`)
}

func (b *Bot) handleLink(msg *tgbotapi.Message, args string) {
	phone := strings.TrimSpace(args)
	if phone == "" {
		b.send(msg.Chat.ID, "Please provide your phone number.\nExample: /link +251912345678")
		return
	}

	ctx := context.Background()
	user, err := b.users.GetByPhone(ctx, phone)
	if err != nil {
		b.send(msg.Chat.ID, "No account found with that phone number. Please register first.")
		return
	}

	if err := b.users.LinkTelegramChatID(ctx, user.ID, msg.Chat.ID); err != nil {
		b.send(msg.Chat.ID, "Failed to link account. Please try again.")
		return
	}

	b.send(msg.Chat.ID, fmt.Sprintf("Account linked successfully. Welcome, *%s*!\n\nType /browse to see available equipment.", user.Name))
}

func (b *Bot) handleBrowse(msg *tgbotapi.Message, args string) {
	ctx := context.Background()

	listings, err := b.listings.Browse(ctx, repository.ListingFilter{Page: 1, Limit: 10})
	if err != nil {
		b.send(msg.Chat.ID, "Failed to fetch listings. Please try again.")
		return
	}
	if len(listings) == 0 {
		b.send(msg.Chat.ID, "No equipment available right now. Check back soon.")
		return
	}

	b.sendListingResults(msg.Chat.ID, listings)
}

func (b *Bot) handleSearch(msg *tgbotapi.Message, args string) {
	parts := strings.Fields(args)
	f := repository.ListingFilter{Page: 1, Limit: 10}

	if len(parts) >= 1 {
		f.Category = parts[0]
	}
	if len(parts) >= 2 {
		f.Location = strings.Join(parts[1:], " ")
	}

	if f.Category == "" && f.Location == "" {
		b.send(msg.Chat.ID, "Usage: /search [category] [location]\nExample: /search excavator Addis")
		return
	}

	ctx := context.Background()
	listings, err := b.listings.Browse(ctx, f)
	if err != nil {
		b.send(msg.Chat.ID, "Failed to fetch listings. Please try again.")
		return
	}
	if len(listings) == 0 {
		b.send(msg.Chat.ID, "No listings found matching your search.")
		return
	}

	b.sendListingResults(msg.Chat.ID, listings)
}

func (b *Bot) sendListingResults(chatID int64, listings []models.Listing) {
	ids := make([]uuid.UUID, len(listings))
	var sb strings.Builder
	sb.WriteString("*Available equipment:*\n\n")

	for i, l := range listings {
		ids[i] = l.ID
		sb.WriteString(fmt.Sprintf(
			"%d. *%s*\n   %s — %.0f ETB/day (min %d days)\n\n",
			i+1, l.Title, l.Location, l.PricePerDay, l.MinimumDays,
		))
	}
	sb.WriteString("Type /listing [number] for details.")

	b.sessions.setListings(chatID, ids)
	b.send(chatID, sb.String())
}

func (b *Bot) handleListing(msg *tgbotapi.Message, args string) {
	n, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil || n < 1 {
		b.send(msg.Chat.ID, "Usage: /listing [number]\nBrowse first with /browse to get numbers.")
		return
	}

	sess := b.sessions.get(msg.Chat.ID)
	if n > len(sess.LastListings) {
		b.send(msg.Chat.ID, "Invalid number. Use /browse to refresh the list.")
		return
	}

	ctx := context.Background()
	l, err := b.listings.GetByID(ctx, sess.LastListings[n-1])
	if err != nil {
		b.send(msg.Chat.ID, "Listing not found.")
		return
	}

	today := time.Now().Format("2006-01-02")
	weekLater := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	b.send(msg.Chat.ID, fmt.Sprintf(
		"*%s*\nCategory: %s\nLocation: %s\nPrice: *%.0f ETB/day* (min %d days)\n\n%s\n\nTo book:\n`/book %d %s %s`",
		l.Title, l.Category, l.Location, l.PricePerDay, l.MinimumDays,
		l.Description,
		n, today, weekLater,
	))
}

func (b *Bot) handleBook(msg *tgbotapi.Message, args string) {
	ctx := context.Background()

	// Must be linked
	user, err := b.users.GetByChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.send(msg.Chat.ID, "You need to link your account first.\nUse: /link +251912345678")
		return
	}

	// Parse: /book [number] [start] [end]
	parts := strings.Fields(args)
	if len(parts) != 3 {
		b.send(msg.Chat.ID, "Usage: /book [number] [start] [end]\nExample: /book 1 2026-03-15 2026-03-18")
		return
	}

	n, err := strconv.Atoi(parts[0])
	if err != nil || n < 1 {
		b.send(msg.Chat.ID, "Invalid listing number. Use /browse to see listings.")
		return
	}

	sess := b.sessions.get(msg.Chat.ID)
	if n > len(sess.LastListings) {
		b.send(msg.Chat.ID, "Invalid number. Use /browse to refresh the list.")
		return
	}

	start, err := time.Parse("2006-01-02", parts[1])
	if err != nil {
		b.send(msg.Chat.ID, "Invalid start date. Use format: YYYY-MM-DD")
		return
	}
	end, err := time.Parse("2006-01-02", parts[2])
	if err != nil {
		b.send(msg.Chat.ID, "Invalid end date. Use format: YYYY-MM-DD")
		return
	}

	if !end.After(start) {
		b.send(msg.Chat.ID, "End date must be after start date.")
		return
	}
	if start.Before(time.Now().Truncate(24 * time.Hour)) {
		b.send(msg.Chat.ID, "Start date cannot be in the past.")
		return
	}

	listingID := sess.LastListings[n-1]
	listing, err := b.listings.GetByID(ctx, listingID)
	if err != nil {
		b.send(msg.Chat.ID, "Listing not found.")
		return
	}
	if !listing.IsAvailable {
		b.send(msg.Chat.ID, "Sorry, this listing is not currently available.")
		return
	}
	if listing.OwnerID == user.ID {
		b.send(msg.Chat.ID, "You cannot book your own listing.")
		return
	}

	days := int(end.Sub(start).Hours()/24) + 1
	if days < listing.MinimumDays {
		b.send(msg.Chat.ID, fmt.Sprintf("Minimum rental is %d days.", listing.MinimumDays))
		return
	}

	conflict, err := b.bookings.HasConflict(ctx, listingID, start, end)
	if err != nil {
		b.send(msg.Chat.ID, "Failed to check availability. Please try again.")
		return
	}
	if conflict {
		b.send(msg.Chat.ID, "Those dates are already booked. Use /listing to see another option.")
		return
	}

	booking := &models.Booking{
		ID:         uuid.New(),
		ListingID:  listingID,
		RenterID:   user.ID,
		OwnerID:    listing.OwnerID,
		StartDate:  start,
		EndDate:    end,
		TotalDays:  days,
		TotalPrice: float64(days) * listing.PricePerDay,
		Status:     models.StatusPending,
	}

	if err := b.bookings.Create(ctx, booking); err != nil {
		b.send(msg.Chat.ID, "Failed to create booking. Please try again.")
		return
	}

	b.send(msg.Chat.ID, fmt.Sprintf(
		"Booking request sent!\n\n*%s*\n%s to %s (%d days)\nTotal: *%.0f ETB*\n\nPayment: arrange directly with the owner via Telebirr, CBE, or cash.\n\nYou'll be notified when the owner confirms.",
		listing.Title,
		start.Format("Jan 2"),
		end.Format("Jan 2, 2006"),
		days,
		booking.TotalPrice,
	))
}

func (b *Bot) handleMyBookings(msg *tgbotapi.Message) {
	ctx := context.Background()

	user, err := b.users.GetByChatID(ctx, msg.Chat.ID)
	if err != nil {
		b.send(msg.Chat.ID, "You need to link your account first.\nUse: /link +251912345678")
		return
	}

	bookings, err := b.bookings.GetByRenter(ctx, user.ID)
	if err != nil {
		b.send(msg.Chat.ID, "Failed to fetch bookings.")
		return
	}
	if len(bookings) == 0 {
		b.send(msg.Chat.ID, "You have no bookings yet.\nUse /browse to find equipment.")
		return
	}

	var sb strings.Builder
	sb.WriteString("*Your bookings:*\n\n")
	for _, bk := range bookings {
		sb.WriteString(fmt.Sprintf(
			"• %s to %s — *%s* — %.0f ETB\n",
			bk.StartDate.Format("Jan 2"),
			bk.EndDate.Format("Jan 2"),
			strings.ToUpper(string(bk.Status)),
			bk.TotalPrice,
		))
	}

	b.send(msg.Chat.ID, sb.String())
}
