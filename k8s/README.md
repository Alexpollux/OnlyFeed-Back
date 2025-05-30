# 🚀 OnlyFeed Backend - Déploiement Kubernetes

Déploiement du backend OnlyFeed sur Kubernetes.

## 📋 Prérequis

- Docker Desktop avec Kubernetes activé
- kubectl installé
- Image Docker du backend construite

## 🔧 Setup rapide

### Option 1 : Docker Desktop (recommandé pour le développement)

```bash
# 1. Builder l'image Docker
docker-compose build

# 2. Créer le namespace
kubectl apply -f k8s/namespace.yaml

# 3. Configurer les variables (IMPORTANT: modifier avec tes vraies valeurs)
kubectl apply -f k8s/configmap.yaml

# 4. Déployer l'application
kubectl apply -f k8s/deployment.yaml

# 5. Vérifier le déploiement
kubectl get pods -n onlyfeed-backend
```

### Option 2 : Minikube (si Docker Desktop pose problème)

```bash
# Démarrer Minikube
minikube start

# Utiliser l'environnement Docker de Minikube
eval $(minikube docker-env)

# Builder l'image dans Minikube
docker-compose build

# Déployer
kubectl apply -f k8s/
```

## 🔍 Vérifications

### Statut des pods
```bash
kubectl get pods -n onlyfeed-backend
```

### Logs de l'application
```bash
kubectl logs -f deployment/onlyfeed-backend -n onlyfeed-backend
```

### Services
```bash
kubectl get services -n onlyfeed-backend
```

## 🌐 Accès à l'application

### Via NodePort (recommandé)
```
http://localhost:30080
```

### Via Port Forward
```bash
kubectl port-forward svc/onlyfeed-backend-service 8080:8080 -n onlyfeed-backend
```
Puis : `http://localhost:8080`

## ⚙️ Configuration

### Variables d'environnement
Modifie `k8s/configmap.yaml` avec tes vraies valeurs :

```yaml
# ConfigMap (variables non sensibles)
AWS_REGION: "eu-west-1"
AWS_BUCKET_NAME: "ton-bucket"
NEXT_PUBLIC_SUPABASE_URL: "https://ton-projet.supabase.co"

# Secret (variables sensibles)
SUPABASE_DB_URL: "postgresql://user:pass@host:port/db"
JWT_SECRET: "ton-jwt-secret"
# ... autres secrets
```

⚠️ **IMPORTANT** : Ne commit jamais le fichier avec tes vraies valeurs !

## 🛠️ Commandes utiles

### Redémarrer le déploiement
```bash
kubectl rollout restart deployment/onlyfeed-backend -n onlyfeed-backend
```

### Supprimer tout
```bash
kubectl delete namespace onlyfeed-backend
```

### Debug des pods
```bash
kubectl describe pod -n onlyfeed-backend
kubectl get events -n onlyfeed-backend --sort-by=.metadata.creationTimestamp
```

## 🎯 Architecture

```
┌─────────────────┐    ┌─────────────────┐
│   LoadBalancer  │    │    NodePort     │
│  (futur, cloud) │    │  localhost:30080│
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┘
                     │
         ┌─────────────────┐
         │     Service     │
         │ onlyfeed-backend│
         └─────────────────┘
                     │
         ┌─────────────────┐
         │   Deployment    │
         │   (2 replicas)  │
         └─────────────────┘
                     │
      ┌──────────────┼──────────────┐
      │              │              │
 ┌─────────┐    ┌─────────┐    ┌─────────┐
 │  Pod 1  │    │  Pod 2  │    │ConfigMap│
 │Backend  │    │Backend  │    │Secret   │
 └─────────┘    └─────────┘    └─────────┘
```

## 🚨 Troubleshooting

### Problème : ErrImagePull / ErrImageNeverPull
```bash
# Vérifier que l'image existe
docker images | grep onlyfeed

# Si l'image n'existe pas, la builder
docker-compose build

# Pour Minikube, utiliser l'environnement Docker
eval $(minikube docker-env)
docker-compose build
```

### Problème : CrashLoopBackOff
```bash
# Vérifier les logs
kubectl logs -f deployment/onlyfeed-backend -n onlyfeed-backend

# Vérifier les variables d'environnement
kubectl get configmap backend-config -n onlyfeed-backend -o yaml
kubectl get secret backend-secrets -n onlyfeed-backend -o yaml
```

### Problème : Service inaccessible
```bash
# Vérifier que les pods sont Running
kubectl get pods -n onlyfeed-backend

# Vérifier les services
kubectl get svc -n onlyfeed-backend

# Test de connectivité interne
kubectl exec -it deployment/onlyfeed-backend -n onlyfeed-backend -- wget -qO- http://localhost:8080
```

## 📚 Ressources

- [Documentation Kubernetes](https://kubernetes.io/docs/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
- [Configuration des variables d'environnement](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/)

---

## 🎓 Pour aller plus loin

### Horizontal Pod Autoscaler
```bash
kubectl autoscale deployment onlyfeed-backend --cpu-percent=50 --min=1 --max=10 -n onlyfeed-backend
```

### Ingress (pour un domaine personnalisé)
Créer un fichier `k8s/ingress.yaml` pour exposer l'application via un nom de domaine.

### Monitoring
Intégrer Prometheus + Grafana pour surveiller l'application.

---

**🚀 Bon déploiement !**