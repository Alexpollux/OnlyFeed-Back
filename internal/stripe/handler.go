package stripe

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/account"
	"github.com/stripe/stripe-go/v78/accountlink"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

func CreateAccountLink(c *gin.Context) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	domain := os.Getenv("DOMAIN_URL")

	userId := c.GetString("user_id")

	// Création d’un compte connecté Stripe (standard)
	acctParams := &stripe.AccountParams{
		Type: stripe.String("standard"),
	}
	acct, err := account.New(acctParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du compte Stripe"})
		return
	}

	// Lien d'onboarding Stripe
	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(acct.ID),
		RefreshURL: stripe.String(fmt.Sprintf("%s/become-creator/error", domain)),
		ReturnURL:  stripe.String(fmt.Sprintf("%s/become-creator/success?account_id=%s", domain, acct.ID)),
		Type:       stripe.String("account_onboarding"),
	}
	link, err := accountlink.New(linkParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du lien d'onboarding Stripe"})
		return
	}

	// Enregistrer StripeAccountID dans la DB de l'utilisateur
	if err := database.DB.Model(&user.User{}).Where("id = ?", userId).Update("stripe_account_id", acct.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour StripeAccountID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": link.URL})
}

func CompleteConnect(c *gin.Context) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	userId := c.GetString("user_id")

	accountID := c.Query("account_id")
	if accountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Paramètre account_id manquant"})
		return
	}

	// Vérifie l'état du compte Stripe
	acct, err := account.GetByID(accountID, nil)
	if err != nil || !acct.ChargesEnabled {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Le compte n’est pas encore activé"})
		return
	}

	// Mise à jour de l’utilisateur : il devient créateur avec un prix par défaut de 5€
	updateData := map[string]interface{}{
		"is_creator":         true,
		"subscription_price": 5.0,
	}

	if err := database.DB.Model(&user.User{}).Where("id = ?", userId).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur mise à jour utilisateur"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

//// Exemple d’intention de paiement avec commission de 10%
//params := &stripe.PaymentIntentParams{
//Amount:   stripe.Int64(1000), // 10.00€
//Currency: stripe.String(string(stripe.CurrencyEUR)),
//ApplicationFeeAmount: stripe.Int64(100), // 1.00€ pour toi
//TransferData: &stripe.PaymentIntentTransferDataParams{
//Destination: stripe.String(creatorsStripeAccountID), // ID du créateur
//},
//}
//intent, err := paymentintent.New(params)
