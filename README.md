# OnlyFeed 

OnlyFeed est une plateforme de partage de contenu avec un système d'abonnement payant pour les créateurs, développée en Go (backend) et Flutter (frontend).

##  Équipe de développement

| Développeur | GitHub | Contributions principales |
|-------------|--------|---------------------------|
| **Arthur DELAPORTE** | [@ArthurDelaporte](https://github.com/ArthurDelaporte) | Gestion des utilisateurs, Monétisation, Fonctionnalités sociales, Architecture technique |
| **Alexandre MEME** | [@Alexpollux](https://github.com/Alexpollux) | Gestion des posts, Interface utilisateur, Administration & Modération, Fonctionnalités sociales |
| **Thibaud LEFOUR** | [@ThibGit99](https://github.com/ThibGit99) | Système de messagerie, Interface utilisateur |

##  Fonctionnalités

###  Gestion des utilisateurs (Arthur DELAPORTE)
-  Inscription et connexion avec authentification Supabase
-  Profils utilisateur avec avatar et bio
-  Édition des profils
-  Système de rôles (utilisateur, créateur, admin)
-  Recherche d'utilisateurs
-  Gestion multilingue (FR/EN)
-  Thème clair/sombre/système

###  Gestion des posts (Alexandre MEME & Arthur DELAPORTE)
-  Création de posts avec images
-  Posts gratuits et premium (payants)
-  Grille de posts sur les profils
-  Page de détail des posts avec commentaires
-  Système de likes avec animations
-  Feed d'actualité avec pagination
-  Partage de posts via messagerie

###  Monétisation (Arthur DELAPORTE)
-  Intégration Stripe Connect pour les créateurs
-  Système d'abonnement mensuel
-  Commission de 20% sur les abonnements
-  Gestion des webhooks Stripe
-  Désabonnement et gestion des statuts

###  Système de messagerie (Thibaud LEFOUR)
-  Conversations privées entre utilisateurs
-  Envoi de messages texte et images
-  Partage de posts dans les conversations
-  Messages non lus et notifications
-  Suppression de conversations
-  Interface responsive mobile/desktop

###  Interface utilisateur (Alexandre MEME & Thibaud LEFOUR)
-  Design responsive (mobile/tablet/desktop)
-  Navigation avec sidebar adaptative
-  Thème personnalisé avec couleurs abricot/basilic
-  Interface multilingue
-  Animations et micro-interactions
-  Gestion des états de chargement

###  Fonctionnalités sociales (Arthur DELAPORTE & Alexandre MEME)
-  Système de follow/unfollow
-  Statistiques des profils (followers, posts, etc.)
-  Interactions entre utilisateurs
-  Feed personnalisé

###  Administration & Modération (Alexandre MEME)
-  Dashboard administrateur avec statistiques
-  Graphiques d'évolution et de distribution
-  Gestion des utilisateurs
-  Système de signalement (posts, utilisateurs, commentaires)
-  Modération des contenus

###  Architecture technique (Arthur DELAPORTE)
-  Backend Go avec Gin framework
-  Base de données PostgreSQL (Supabase)
-  Stockage S3 AWS pour les médias
-  Architecture RESTful
-  Middleware d'authentification JWT
-  Gestion des erreurs et logs structurés

##  Technologies utilisées

### Backend
- **Go** - Langage principal
- **Gin** - Framework web
- **GORM** - ORM pour PostgreSQL
- **Supabase** - Base de données et authentification
- **AWS S3** - Stockage des fichiers
- **Stripe** - Paiements et abonnements

### Frontend
- **Flutter** - Framework UI multiplateforme
- **Dart** - Langage de programmation
- **Provider** - Gestion d'état
- **Dio** - Client HTTP
- **Go Router** - Navigation
- **Easy Localization** - Internationalisation

### Services externes
- **Supabase** - BaaS (Backend as a Service)
- **AWS S3** - Stockage cloud
- **Stripe** - Plateforme de paiement

##  Structure du projet

```
onlyfeed/
├── README.md                          # Ce fichier
├── backend/                           # Backend Go
│   ├── cmd/server/main.go             # Point d'entrée
│   ├── internal/                      # Code métier
│   │   ├── auth/                      # Authentification
│   │   ├── user/                      # Gestion utilisateurs
│   │   ├── post/                      # Gestion posts
│   │   ├── message/                   # Système messagerie
│   │   ├── stripe/                    # Intégration Stripe
│   │   ├── admin/                     # Administration
│   │   └── middleware/                # Middlewares
│   ├── go.mod                         # Dépendances Go
│   └── .env.example                   # Variables d'environnement
└── frontend/                          # Frontend Flutter
    ├── lib/                          # Code Dart
    │   ├── features/                 # Fonctionnalités métier
    │   │   ├── auth/                 # Authentification
    │   │   ├── profile/              # Profils
    │   │   ├── post/                 # Posts
    │   │   ├── message/              # Messagerie
    │   │   ├── admin/                # Administration
    │   │   └── like/                 # Système de likes
    │   ├── shared/                   # Code partagé
    │   └── core/                     # Widgets centraux
    ├── assets/                       # Ressources statiques
    ├── pubspec.yaml                  # Dépendances Flutter
    └── web/                          # Configuration web
```

##  Procédure de lancement en local

### Prérequis
- Go 1.21+ installé
- Flutter 3.0+ installé
- PostgreSQL (ou compte Supabase)
- Compte AWS S3
- Compte Stripe

### 1. Configuration du Backend

```bash
# Cloner le repository
git clone https://github.com/ArthurDelaporte/OnlyFeed-Back.git
cd OnlyFeed-Back

# Installer les dépendances Go
go mod tidy

# Copier et configurer les variables d'environnement
cp .env.example .env
```

Éditer le fichier `.env` avec vos vraies valeurs :

```env
# Base de données
SUPABASE_DB_URL=postgresql://user:password@host:port/database

# Supabase
NEXT_PUBLIC_SUPABASE_URL=https://votre-projet.supabase.co
SUPABASE_ANON_KEY=votre_anon_key
SUPABASE_SERVICE_ROLE_KEY=votre_service_role_key
SUPABASE_JWKS_URL=https://votre-projet.supabase.co/rest/v1/rpc

# JWT
JWT_SECRET=votre_jwt_secret_tres_securise

# AWS S3
AWS_ACCESS_KEY_ID=votre_access_key
AWS_SECRET_ACCESS_KEY=votre_secret_key
AWS_BUCKET_NAME=votre-bucket-s3
AWS_REGION=eu-west-1

# Stripe
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Application
GIN_MODE=debug
DOMAIN_URL=http://localhost:5000
```

```bash
# Lancer le serveur backend
go run ./cmd/server
```

Le backend sera accessible sur **http://localhost:8080**

### 2. Configuration du Frontend

```bash
# Aller dans le dossier frontend
cd ../frontend

# Installer les dépendances Flutter
flutter pub get

# Lancer l'application web
flutter run -d chrome --web-hostname=localhost --web-port=5000 --dart-define=BASE_URL=http://localhost:8080
```

L'application sera accessible sur **http://localhost:5000**

### 3. Configuration des services externes

#### Supabase
1. Créer un projet sur [supabase.com](https://supabase.com)
2. Récupérer l'URL du projet et les clés API
3. Configurer l'authentification et les tables

#### AWS S3
1. Créer un bucket S3
2. Configurer les permissions CORS
3. Créer un utilisateur IAM avec accès S3

#### Stripe
1. Créer un compte sur [stripe.com](https://stripe.com)
2. Activer Stripe Connect
3. Configurer les webhooks pour les événements d'abonnement

##  Scripts utiles

### Backend
```bash
# Lancer en mode développement
go run ./cmd/server

# Construire l'exécutable
go build -o onlyfeed ./cmd/server

# Lancer les tests
go test ./...

# Vérifier le code
go vet ./...
go fmt ./...
```

### Frontend
```bash
# Lancer en développement web
flutter run -d chrome --web-hostname=localhost --web-port=5000 --dart-define=BASE_URL=http://localhost:8080

# Construire pour le web
flutter build web --dart-define=BASE_URL=https://votre-api.com

# Lancer les tests
flutter test

# Analyser le code
flutter analyze
```

##  API Endpoints principaux

### Authentification
- `POST /api/auth/signup` - Inscription
- `POST /api/auth/login` - Connexion
- `POST /api/auth/logout` - Déconnexion

### Utilisateurs
- `GET /api/me` - Profil utilisateur connecté
- `PUT /api/me` - Modifier son profil
- `GET /api/users/username/:username` - Profil par nom d'utilisateur
- `GET /api/users/search` - Recherche d'utilisateurs

### Posts
- `GET /api/posts` - Liste des posts
- `POST /api/posts` - Créer un post
- `GET /api/posts/:id` - Détail d'un post
- `DELETE /api/posts/:id` - Supprimer un post
- `POST /api/posts/:id/like` - Liker/Déliker un post

### Messagerie
- `GET /api/messages/conversations` - Liste des conversations
- `GET /api/messages/conversations/:id` - Messages d'une conversation
- `POST /api/messages/send` - Envoyer un message

### Administration
- `GET /api/admin/stats` - Statistiques générales
- `GET /api/admin/charts/:type` - Données pour graphiques
- `GET /api/admin/reports` - Signalements

##  Dépannage

### Problèmes courants

**Erreur de connexion à la base de données**
- Vérifier la variable `SUPABASE_DB_URL` dans le `.env`
- S'assurer que Supabase est accessible

**Erreur CORS côté frontend**
- Vérifier que l'URL du backend est correcte
- S'assurer que le backend accepte les requêtes depuis localhost:5000

**Problème d'upload S3**
- Vérifier les clés AWS dans le `.env`
- S'assurer que le bucket existe et est accessible

**Erreur Stripe**
- Vérifier les clés Stripe (test/live)
- Configurer correctement les webhooks

### Logs et debugging

Le backend utilise des logs structurés en JSON. Pour voir les logs en temps réel :

```bash
# Backend
go run ./cmd/server | jq .

# Frontend (console du navigateur)
# Ouvrir les outils de développement (F12)
```

##  Documentation supplémentaire

- [Documentation Go/Gin](https://gin-gonic.com/)
- [Documentation Flutter](https://flutter.dev/docs)
- [API Supabase](https://supabase.com/docs)
- [Documentation Stripe Connect](https://stripe.com/docs/connect)
- [AWS S3 Go SDK](https://docs.aws.amazon.com/sdk-for-go/)

**OnlyFeed** - Une plateforme moderne de partage de contenu avec monétisation intégrée 