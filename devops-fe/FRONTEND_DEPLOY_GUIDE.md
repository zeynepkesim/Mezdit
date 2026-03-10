# Vue.js Frontend - Kubernetes Deployment Rehberi

Bu doküman Vue.js frontend uygulamanızı Kubernetes'e deploy etmek için gereken tüm adımları içerir.

---

## 📋 Genel Bakış

**Mevcut Durum:**
- Frontend şu anda VM'de çalışıyor: `http://devops-mezdit-vm13.cloud.ut.ee`

**Hedef:**
- Frontend ve Backend tek domain'de çalışacak: `https://devops-mezdit.kubernetes.devops.cs.ut.ee`
  - Frontend: `https://devops-mezdit.kubernetes.devops.cs.ut.ee/` (root path)
  - Backend API: `https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/` (API path)
- Otomatik deployment (GitLab CI/CD)
- HTTPS ile güvenli erişim
- CORS problemi yok (same-origin)

---

## 🚀 Adım Adım Kurulum

### 1. Dockerfile Oluşturma

Frontend repository'nizde **Dockerfile** oluşturun:

```dockerfile
# Build stage - Vue.js uygulamasını build et
FROM node:18 AS builder

WORKDIR /app

# package.json ve package-lock.json'ı kopyala
COPY package*.json ./

# Dependencies'leri yükle
RUN npm ci

# Tüm kaynak kodları kopyala
COPY . .

# Production build
RUN npm run build

# Production stage - NGINX ile serve et
FROM nginx:alpine

# Build edilen dosyaları NGINX'e kopyala
COPY --from=builder /app/dist /usr/share/nginx/html

# Custom NGINX config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

### 2. NGINX Configuration Oluşturma

**nginx.conf** dosyası oluşturun:

```nginx
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    # Vue.js SPA için - tüm route'ları index.html'e yönlendir
    # API ve WS path'leri hariç, diğer tüm istekler index.html'e gider
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/x-javascript application/xml+rss application/json;
}
```

**ÖNEMLİ NOT:** Bu basit NGINX config yeterlidir çünkü Kubernetes Ingress üzerinden routing yapılacak. API istekleri Ingress seviyesinde backend'e yönlendirilecek.

### 3. Kubernetes Manifest Klasörü Oluşturma

Frontend repository'nizde `k8s/` klasörü oluşturun:

```bash
mkdir k8s
cd k8s
```

### 4. ConfigMap Oluşturma

**k8s/configmap.yaml:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: frontend-config
  namespace: emzops-staging
data:
  # Backend API URL - Same origin olduğu için relative path kullanılabilir
  VUE_APP_API_URL: "/api"
  # WebSocket URL (eğer kullanıyorsanız) - Same origin
  VUE_APP_WS_URL: "/ws"
  # Diğer environment variables buraya eklenebilir
```

**Not:**
- Vue.js'te environment variable'lar `VUE_APP_` prefix'i ile başlamalı
- Tek domain kullandığımız için relative path'ler yeterli (CORS problemi yok)

### 5. Deployment Oluşturma

**k8s/deployment.yaml:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: devops-frontend
  namespace: emzops-staging
  labels:
    app: devops-frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: devops-frontend
  template:
    metadata:
      labels:
        app: devops-frontend
    spec:
      # GitLab Registry'den image çekmek için secret
      imagePullSecrets:
        - name: gitlab-registry-secret
      containers:
      - name: frontend
        # Image adını kendi frontend repository'nize göre değiştirin
        image: registry.gitlab.cs.ut.ee/devops/teams/emzops/devops-fe:latest
        ports:
        - containerPort: 80
          name: http
        # ConfigMap'ten environment variables al
        envFrom:
        - configMapRef:
            name: frontend-config
        # Resource limits (optional ama önerilen)
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
        # Health check (optional)
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 6. Service Oluşturma

**k8s/service.yaml:**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: devops-frontend-service
  namespace: emzops-staging
  labels:
    app: devops-frontend
spec:
  type: ClusterIP
  selector:
    app: devops-frontend
  ports:
  - protocol: TCP
    port: 80
    targetPort: 80
    name: http
```

### 7. Ingress Güncelleme (Mevcut Backend Ingress'i Güncelleyin)

**UYARI:** Yeni bir Ingress oluşturmayın! Mevcut backend Ingress'inizi güncelleyin.

Backend repository'nizdeki `k8s/week10/ingress.yaml` dosyasını aşağıdaki gibi güncelleyin:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: devops-ingress
  namespace: emzops-staging
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - devops-mezdit.kubernetes.devops.cs.ut.ee
    secretName: devops-tls-cert
  rules:
  - host: devops-mezdit.kubernetes.devops.cs.ut.ee
    http:
      paths:
      # Frontend - Root path (en son olmalı)
      - path: /
        pathType: Prefix
        backend:
          service:
            name: devops-frontend-service
            port:
              number: 80
      # Backend API - /api path'i
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: devops-backend-service
            port:
              number: 8080
      # WebSocket (eğer kullanıyorsanız)
      - path: /ws
        pathType: Prefix
        backend:
          service:
            name: devops-backend-service
            port:
              number: 8080
```

**ÖNEMLİ NOTLAR:**
1. Frontend için **AYRI bir Ingress oluşturmayın**
2. Path sıralaması önemli: `/api` ve `/ws` önce, `/` en son olmalı
3. Aynı TLS certificate her iki servis için kullanılacak
4. Bu yapılandırma same-origin sağlar (CORS problemi olmaz)

### 8. .gitignore Güncelleme

Frontend repository'nizde `.gitignore` dosyasına ekleyin:

```
# Kubernetes secrets (never commit)
k8s/secret.yaml
k8s/*-secret.yaml
```

### 9. GitLab CI/CD Pipeline Oluşturma/Güncelleme

Frontend repository'nizde `.gitlab-ci.yml` oluşturun veya güncelleyin:

```yaml
stages:
  - test
  - build-image
  - deploy

variables:
  IMAGE_NAME: $CI_REGISTRY_IMAGE
  IMAGE_TAG: $CI_COMMIT_SHORT_SHA

# Test stage (eğer testleriniz varsa)
test:
  stage: test
  image: node:18
  script:
    - npm ci
    - npm run test:unit
  only:
    - main
    - merge_requests

# Lint check (optional)
lint:
  stage: test
  image: node:18
  script:
    - npm ci
    - npm run lint
  only:
    - main
    - merge_requests

# Docker image build ve push
build-docker-image:
  stage: build-image
  image:
    name: gcr.io/kaniko-project/executor:v1.23.0-debug
    entrypoint: [""]
  script:
    - echo "Building frontend container image with Kaniko..."
    - mkdir -p /kaniko/.docker
    - echo "{\"auths\":{\"$CI_REGISTRY\":{\"auth\":\"$(printf "%s:%s" "${CI_REGISTRY_USER}" "${CI_REGISTRY_PASSWORD}" | base64 | tr -d '\n')\"}}}" > /kaniko/.docker/config.json
    - >-
      /kaniko/executor
      --context "${CI_PROJECT_DIR}"
      --dockerfile "${CI_PROJECT_DIR}/Dockerfile"
      --destination "${IMAGE_NAME}:${IMAGE_TAG}"
      --destination "${IMAGE_NAME}:latest"
      --cache=true
    - echo "✅ Frontend image pushed - $IMAGE_NAME:latest"
  only:
    - main

# Kubernetes deployment
deploy-to-k8s:
  stage: deploy
  image: bitnami/kubectl:latest
  before_script:
    - mkdir -p ~/.kube
    - echo "$KUBECONFIG_CONTENT" | base64 -d > ~/.kube/config
    - chmod 600 ~/.kube/config
  script:
    - echo "Deploying frontend to Kubernetes namespace emzops-staging..."
    - kubectl apply -f k8s/configmap.yaml
    - kubectl apply -f k8s/deployment.yaml
    - kubectl apply -f k8s/service.yaml
    - echo "Waiting for deployment to be ready..."
    - kubectl rollout status deployment/devops-frontend -n emzops-staging --timeout=5m
    - echo "✅ Deployment successful!"
    - echo "Frontend URL: https://devops-mezdit.kubernetes.devops.cs.ut.ee"
    - kubectl get pods -n emzops-staging -l app=devops-frontend
  environment:
    name: staging
    url: https://devops-mezdit.kubernetes.devops.cs.ut.ee
  only:
    - main
```

### 10. GitLab CI/CD Variable Ekleme

Frontend GitLab repository'nizde:

1. **Settings → CI/CD → Variables**
2. **Add variable** tıklayın
3. Aşağıdaki değerleri girin:

**Variable 1:**
- **Key**: `KUBECONFIG_CONTENT`
- **Value**: Backend'deki `KUBECONFIG_BASE64.txt` dosyasının içeriği
- **Type**: Variable
- **Protected**: ✅ Yes
- **Masked**: ❌ No
- **Expand variable reference**: ❌ No

### 11. Vue.js Environment Variables Ayarlama

Vue.js uygulamanızda API URL'i environment variable'dan alın:

**Örnek kullanım (src/api/axios.js veya benzeri):**

```javascript
import axios from 'axios';

// Same origin olduğu için relative path kullanın
const baseURL = process.env.VUE_APP_API_URL || '/api';

const apiClient = axios.create({
  baseURL: baseURL,
  withCredentials: true,  // Cookie'ler için gerekli
  headers: {
    'Content-Type': 'application/json',
  }
});

export default apiClient;
```

**WebSocket için (eğer kullanıyorsanız):**

```javascript
// Same origin olduğu için relative path kullanın
const wsURL = process.env.VUE_APP_WS_URL || '/ws';

// Relative path'i absolute URL'e çevir
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const fullWsURL = `${protocol}//${window.location.host}${wsURL}`;

const socket = new WebSocket(fullWsURL);
```

### 12. Backend CORS Ayarları

**ÖNEMLİ:** Same-origin olduğu için CORS ayarlarına gerek YOK! Ancak local development için ayar yapmanız gerekebilir:

```go
// Go backend örneği
router.Use(cors.New(cors.Config{
    AllowOrigins: []string{
        "https://devops-mezdit.kubernetes.devops.cs.ut.ee",  // Production (same-origin, CORS sorunu yok)
        "http://localhost:8080",      // Local development için
        "http://localhost:3000",      // Local Vue.js dev server
    },
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

**Not:** Production'da frontend ve backend aynı domain'de olduğu için CORS gereksizdir, sadece local development için gerekli.

### 13. Deployment

Artık her şey hazır! Commit ve push yapın:

```bash
git add .
git commit -m "Add Kubernetes deployment for frontend"
git push origin main
```

GitLab pipeline otomatik olarak:
1. ✅ Testleri çalıştıracak
2. ✅ Docker image build edecek
3. ✅ GitLab Registry'ye push edecek
4. ✅ Kubernetes'e deploy edecek

### 14. Deployment Doğrulama

```bash
# Pod'ları kontrol et
kubectl get pods -n emzops-staging -l app=devops-frontend

# Logs kontrol et
kubectl logs -f deployment/devops-frontend -n emzops-staging

# Service kontrol et
kubectl get service devops-frontend-service -n emzops-staging

# Ingress kontrol et
kubectl get ingress devops-frontend-ingress -n emzops-staging

# TLS certificate kontrol et
kubectl get certificate frontend-tls-cert -n emzops-staging

# Port forward ile local test
kubectl port-forward service/devops-frontend-service 3000:80 -n emzops-staging
# Tarayıcıda: http://localhost:3000
```

### 15. Browser'da Test

1. https://devops-mezdit.kubernetes.devops.cs.ut.ee adresine gidin (artık tek URL!)
2. TLS certificate'in çalıştığını kontrol edin (yeşil kilit ikonu)
3. Backend API çağrılarının çalıştığını kontrol edin (örn: https://devops-mezdit.kubernetes.devops.cs.ut.ee/api/health)
4. Browser console'da CORS hatası olmadığını kontrol edin (same-origin olduğu için olmayacak)

---

## 🔧 Sorun Giderme

### Problem 1: Image Pull Hatası

**Semptom:** `ImagePullBackOff` veya `ErrImagePull`

**Çözüm:**
```bash
# Registry secret kontrol et
kubectl get secret gitlab-registry-secret -n emzops-staging

# Eğer yoksa backend'deki gibi oluştur (zaten var olmalı)
```

### Problem 2: Pod Başlamıyor

```bash
# Pod detaylarını kontrol et
kubectl describe pod <pod-name> -n emzops-staging

# Logs kontrol et
kubectl logs <pod-name> -n emzops-staging
```

### Problem 3: CORS Hatası

**Same-origin kullandığımız için CORS hatası OLMAMALI!** Eğer görüyorsanız:

1. Frontend'in doğru URL'de çalıştığını kontrol edin (https://devops-mezdit.kubernetes.devops.cs.ut.ee)
2. API isteklerinin relative path kullandığını kontrol edin (`/api/...`)
3. Local development'ta farklı port kullanıyorsanız backend CORS ayarlarını kontrol edin

### Problem 4: TLS Certificate Oluşmuyor

```bash
# Certificate status kontrol et (backend ile aynı certificate kullanılıyor)
kubectl describe certificate devops-tls-cert -n emzops-staging

# 1-2 dakika bekleyin, otomatik oluşacak
```

### Problem 5: API İstekleri 404 Veriyor

1. Backend'in çalıştığını kontrol edin:
   ```bash
   kubectl get pods -n emzops-staging -l app=devops-backend
   ```

2. API URL'in doğru olduğunu kontrol edin:
   ```bash
   kubectl get configmap frontend-config -n emzops-staging -o yaml
   ```

3. Frontend kodunda API URL'in doğru kullanıldığını kontrol edin

---

## 📝 Önemli Notlar

1. **Environment Variables:**
   - Vue.js'te environment variables `VUE_APP_` prefix'i ile başlamalı
   - Build time'da embed edilir, runtime'da değiştirilemez
   - Değiştirmek için yeniden build gerekir

2. **NGINX Configuration:**
   - Vue.js SPA routing için `try_files` önemli
   - API proxy yapılandırması backend ile iletişim için gerekli

3. **CORS:**
   - Same-origin kullandığımız için CORS problemi yok
   - Local development için backend CORS ayarları gerekebilir
   - Credentials kullanıyorsanız `withCredentials: true` olmalı

4. **Secrets:**
   - Hassas bilgiler (API keys vs) varsa Kubernetes Secret kullanın
   - Git'e commit etmeyin

5. **Testing:**
   - Production'a push etmeden önce local'de test edin:
     ```bash
     docker build -t frontend-test .
     docker run -p 3000:80 frontend-test
     ```

---

## 🎯 Checklist

Deployment öncesi kontrol listesi:

- [ ] Dockerfile oluşturuldu
- [ ] nginx.conf oluşturuldu
- [ ] k8s/ klasörü ve manifest dosyaları oluşturuldu
- [ ] .gitlab-ci.yml güncellendi
- [ ] GitLab CI/CD variable (`KUBECONFIG_CONTENT`) eklendi
- [ ] Vue.js kodunda environment variables kullanılıyor
- [ ] Backend CORS ayarları yapıldı
- [ ] .gitignore güncellendi
- [ ] Local test yapıldı (docker build)
- [ ] Git'e commit ve push yapıldı
- [ ] Pipeline başarılı oldu
- [ ] Pod'lar çalışıyor
- [ ] TLS certificate oluştu
- [ ] Browser'da erişim sağlandı
- [ ] API çağrıları çalışıyor
- [ ] CORS hatası yok

---

**Sorularınız için:** Backend KUBERNETES_SETUP.md dosyasına bakın veya takım arkadaşlarınıza sorun.

**Başarılar!** 🚀
