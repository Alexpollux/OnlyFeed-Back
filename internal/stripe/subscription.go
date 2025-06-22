package stripe

import (
	"fmt"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/price"
	"github.com/stripe/stripe-go/v78/product"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v78"

	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/database"
	"github.com/ArthurDelaporte/OnlyFeed-Back/internal/user"
)

func CreateSubscriptionSession(c *gin.Context) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	domain := os.Getenv("DOMAIN_URL")

	creatorID := c.Param("creator_id")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")

	// Récupérer les infos du créateur
	var creator user.User
	if err := database.DB.First(&creator, "id = ? AND is_creator = true", creatorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Créateur introuvable"})
		return
	}
	if creator.StripeAccountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Le créateur n'a pas de compte Stripe"})
		return
	}

	// 🔥 Injecter StripeAccount dans les paramètres
	baseParams := &stripe.Params{}
	baseParams.StripeAccount = &creator.StripeAccountID

	// Création du produit
	productParams := &stripe.ProductParams{
		Params: *baseParams,
		Name:   stripe.String(fmt.Sprintf("Abonnement OnlyFeed au créateur %s", creator.Username)),
	}
	createdProduct, err := product.New(productParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création produit Stripe"})
		return
	}

	// Création du prix
	priceParams := &stripe.PriceParams{
		Params:     *baseParams,
		Product:    stripe.String(createdProduct.ID),
		Currency:   stripe.String("eur"),
		UnitAmount: stripe.Int64(int64(creator.SubscriptionPrice * 100)),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		},
	}
	createdPrice, err := price.New(priceParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création tarif Stripe"})
		return
	}

	// Création de la session d’abonnement
	sessionParams := &stripe.CheckoutSessionParams{
		Params:     *baseParams,
		Mode:       stripe.String("subscription"),
		SuccessURL: stripe.String(fmt.Sprintf("%s/%s?subscribe=success", domain, creator.Username)),
		CancelURL:  stripe.String(fmt.Sprintf("%s/%s?subscribe=error", domain, creator.Username)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(createdPrice.ID),
				Quantity: stripe.Int64(1),
			},
		},
		CustomerEmail: stripe.String(userEmail),
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			ApplicationFeePercent: stripe.Float64(20.0),
		},
		Metadata: map[string]string{
			"creator_id":    creator.ID,
			"subscriber_id": userID,
		},
	}

	createdSession, err := session.New(sessionParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création session Stripe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": createdSession.URL})
}
