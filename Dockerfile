# Dockerfile pour OnlyFeed Backend (Go)

# Étape 1: Builder - Image Go complète pour compiler
FROM golang:1.24-alpine AS builder

# Installer git et autres dépendances nécessaires
RUN apk add --no-cache git ca-certificates

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances Go
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download

# Copier tout le code source
COPY . .

# Compiler l'application
# Ajuste le chemin selon ta structure (tu as cmd/server/main.go)
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Étape 2: Runtime - Image finale légère
FROM alpine:latest

# Installer certificats SSL (nécessaire pour les appels HTTPS)
RUN apk --no-cache add ca-certificates tzdata

# Créer un utilisateur non-root pour la sécurité
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Définir le répertoire de travail
WORKDIR /app

# Copier l'exécutable depuis l'étape builder
COPY --from=builder /app/main .

# Changer le propriétaire des fichiers
RUN chown -R appuser:appgroup /app

# Utiliser l'utilisateur non-root
USER appuser

# Variables d'environnement par défaut
ENV GIN_MODE=release
ENV PORT=8080

# Exposer le port
EXPOSE 8080

# Commande de santé (optionnel mais utile)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Commande pour lancer l'application
CMD ["./main"]