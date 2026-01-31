package bot

import (
	"context"
	"fmt"
	"strings"

	"passportier-bot/internal/models"
	"passportier-bot/internal/security"
	"passportier-bot/internal/vault"

	"gopkg.in/telebot.v3"
	"gorm.io/gorm"
)

// HandleInlineQuery handles inline search reuqests (@BotName query).
func HandleInlineQuery(b *telebot.Bot, db *gorm.DB, sm *security.SessionManager) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		query := strings.ToLower(c.Query().Text)
		userID := c.Sender().ID

		userKey, err := sm.GetSession(context.Background(), userID)
		if err != nil {
			article := &telebot.ArticleResult{
				ResultBase: telebot.ResultBase{
					ID: "unlock",
				},
				Title:       "🔒 Session Locked",
				Description: "Tap here to unlock via private chat",
			}
			article.SetContent(&telebot.InputTextMessageContent{
				Text: "Please switch to private chat and allow /unlock to access your passwords.",
			})

			return c.Answer(&telebot.QueryResponse{
				Results:    []telebot.Result{article},
				CacheTime:  1,
				IsPersonal: true,
			})
		}

		var entries []models.PasswordEntry
		dbSearch := db.Where("user_id = ?", userID)
		if query != "" {
			dbSearch = dbSearch.Where("LOWER(service) LIKE ?", "%"+query+"%")
		} else {
			dbSearch = dbSearch.Limit(50)
		}
		
		if err := dbSearch.Find(&entries).Error; err != nil {
			return c.Answer(&telebot.QueryResponse{Results: []telebot.Result{}})
		}

		results := buildInlineResults(db, userID, userKey, entries)
		return c.Answer(&telebot.QueryResponse{Results: results, CacheTime: 5, IsPersonal: true})
	}
}

func buildInlineResults(db *gorm.DB, userID int64, userKey string, entries []models.PasswordEntry) []telebot.Result {
	results := make([]telebot.Result, 0, len(entries))
	for _, entry := range entries {
		decrypted, err := vault.RetrieveCredential(db, userID, entry.Service, userKey)
		if err != nil {
			continue
		}

		text := fmt.Sprintf("🔑 *%s*\n||%s||", entry.Service, decrypted)
		article := &telebot.ArticleResult{
			ResultBase: telebot.ResultBase{
				ID: fmt.Sprintf("%d", entry.ID),
			},
			Title:       entry.Service,
			Description: "Tap to send password",
		}
		article.SetContent(&telebot.InputTextMessageContent{
			Text:      text,
			ParseMode: telebot.ModeMarkdown,
		})
		results = append(results, article)
	}
	return results
}
