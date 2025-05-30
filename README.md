# OnlyFeed Backend 🚀

Backend Go pour l'application OnlyFeed avec Docker.

## 🏃‍♂️ Démarrage rapide

### Prérequis
- Docker Desktop installé et lancé
- Git

### Lancement en 30 secondes
```bash
# 1. Cloner le repo
git clone https://github.com/ArthurDelaporte/OnlyFeed-Back.git

cd OnlyFeed-Back

# 2. Configurer les variables d'environnement
cp .env.example .env
# Puis édite .env avec tes vraies valeurs

# 3. Lancer l'application
docker-compose up --build
```

L'API sera disponible sur **http://localhost:8080**

## 📡 Routes API

- `GET /` - Health check
- `POST /api/auth/login` - Connexion
- `GET /api/posts` - Liste des posts
- `GET /api/posts/:id/comments` - Commentaires d'un post
- `POST /api/comments` - Créer un commentaire
- Et bien d'autres...

## 🔧 Commandes utiles

```bash
# Lancer en arrière-plan
docker-compose up -d

# Voir les logs
docker-compose logs -f

# Arrêter
docker-compose down

# Rebuild après modification du code
docker-compose up --build

# Nettoyer tout
docker-compose down --volumes --rmi all
```

## 🛠️ Développement

### Modification du code
1. Modifie ton code Go
2. `docker-compose up --build` pour rebuilder
3. Tes changements sont pris en compte

### Variables d'environnement
Modifie le fichier `.env` :
```
SUPABASE_DB_URL=ta_connection_string
GIN_MODE=debug
```

## 🐛 Dépannage

### Port déjà utilisé
```bash
# Trouver quel processus utilise le port 8080
netstat -ano | findstr :8080
# Arrêter le processus ou changer le port dans docker-compose.yml
```

### Problème de build
```bash
# Clean et rebuild
docker-compose down
docker system prune -f
docker-compose up --build
```

### Base de données
Vérifier que ta `SUPABASE_DB_URL` est correcte dans `.env`

## 📦 Structure Docker

- `Dockerfile` : Image de production légère
- `docker-compose.yml` : Configuration de développement
- `.env` : Variables d'environnement (ne pas commit !)

## 🔄 Workflow équipe

1. `git pull` - Récupérer les changements
2. `docker-compose up --build` - Lancer avec les dernières modifs
3. Développer normalement
4. Commit & push

**Fini les "ça marche chez moi" !** 🎯