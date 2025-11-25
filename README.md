# Mezdit - Reddit-like Social Platform

[![pipeline status](https://gitlab.cs.ut.ee/devops/teams/emzops/devops-fe/badges/main/pipeline.svg)](https://gitlab.cs.ut.ee/devops/teams/emzops/devops-fe/-/commits/main)

**🚀 Production:** [https://devops-mezdit.kubernetes.devops.cs.ut.ee/posts](https://devops-mezdit.kubernetes.devops.cs.ut.ee/posts)
**📦 Namespace:** `emzops-staging`

A modern, full-featured social media platform built with Vue.js 3, featuring real-time updates via WebSocket, user authentication, post management, and voting system. Deployed on Kubernetes with automated CI/CD.

## Features

### Core Functionality
- **Authentication**: Registration, login, password management with JWT session handling
- **Post Management**: Create, edit, view posts with optional images (JPG, PNG, GIF up to 10MB)
- **Anonymous Posting**: Post without revealing identity while maintaining ownership
- **Privacy Controls**: Public posts (visible to all) or private posts (owner only)
- **Voting System**: Upvote/downvote with real-time updates and visual feedback
- **Guest Access**: Browse posts without authentication, with login prompts for interactions
- **Real-Time Updates**: WebSocket for instant post creation, edits, votes, and privacy changes

### UI/UX
- Material Design with Vuetify 3
- Responsive layout (mobile, tablet, desktop)
- Toast notifications for user feedback
- Live preview while creating/editing posts
- Auto-generated DiceBear avatars
- Loading states and form validation
- WebSocket connection status indicator

## Tech Stack

**Frontend:**
- Vue.js 3 (Composition API)
- Vuetify 3 (Material Design)
- Vue Router 4
- Vite
- Axios
- WebSocket API
- Vue Toast Notification

**Backend Integration:**
- Go backend with RESTful API
- PostgreSQL database
- WebSocket Hub for real-time broadcasts

## Project Structure

```
devops-fe/
├── src/
│   ├── api/                    # API integration
│   │   ├── axios.js           # Axios instance with JWT auth
│   │   ├── auth.js            # Auth endpoints
│   │   └── posts.js           # Post endpoints
│   ├── composables/
│   │   └── useWebSocket.js    # WebSocket manager
│   ├── pages/                 # Route components
│   │   ├── Login.vue
│   │   ├── Register.vue
│   │   ├── Home.vue           # Main feed
│   │   ├── CreatePost.vue
│   │   ├── EditPost.vue
│   │   └── ChangePassword.vue
│   ├── router/
│   │   └── index.js           # Routes & guards
│   ├── store/                 # State management
│   │   ├── auth.js
│   │   └── posts.js           # Posts + WebSocket handlers
│   ├── App.vue                # Root component
│   └── main.js                # Entry point
├── package.json
├── vite.config.js
└── README.md
```

## Getting Started

### Prerequisites
- Node.js v18 or higher
- Backend API running on `http://localhost:8080`

### Installation

```bash
# Clone repository
git clone <repository-url>
cd devops-fe

# Install dependencies
npm install

# Start development server
npm run dev
```

Open browser at `http://localhost:5173`

### Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build for production |
| `npm run preview` | Preview production build |

## API Integration

### HTTP Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| `POST` | `/login` | User authentication | No |
| `POST` | `/register` | User registration | No |
| `POST` | `/change-password` | Change password | Yes |
| `GET` | `/posts` | Fetch posts (optional userId param) | No |
| `POST` | `/posts` | Create post | Yes |
| `POST` | `/posts/upload` | Create post with image | Yes |
| `POST` | `/posts/edit` | Edit post | Yes |
| `POST` | `/posts/edit-with-image` | Edit post with image | Yes |
| `POST` | `/vote` | Vote on post | Yes |
| `GET` | `/images/{filename}` | Get uploaded image | No |

### WebSocket

**Endpoint:** `ws://localhost:8080/ws`

**Message Types:**
- `POST_CREATED` - New post created
- `POST_UPDATED` - Post edited
- `POST_VOTED` - Vote changed
- `POST_PRIVACY_CHANGED` - Privacy toggled

**Message Format:**
```json
{
  "type": "POST_VOTED",
  "payload": {
    "id": 1,
    "publisherId": 2,
    "publisherName": "John Doe",
    "header": "Post Title",
    "contentText": "Post content",
    "voteCount": 42,
    "isLiked": true,
    "isPrivate": false,
    "createdAt": "2023-12-01T10:00:00Z"
  }
}
```

## Key Features Explained

### Authentication Flow
1. **Registration/Login**: Credentials validated, user data stored in localStorage
2. **Session Management**: Auto-loaded on page refresh, persists across sessions
3. **Route Guards**: Protected routes redirect unauthenticated users to login

### Post Creation
1. Fill form: Title (required), content (required, max 500 chars), display name (optional), image (optional)
2. Toggle anonymous mode to hide identity
3. Toggle private mode to restrict visibility
4. Real-time preview shows how post will appear
5. Submit → API creates post → WebSocket broadcasts → appears in feed

### Voting System
- **No Vote**: Grey buttons
- **Upvoted**: Green button
- **Downvoted**: Red button
- Click same vote to remove it
- Real-time updates via WebSocket (~100ms latency)

### WebSocket Real-Time
- Auto-connect on page load
- Auto-reconnect with exponential backoff (1s → 2s → 4s → 8s → 16s → 30s max)
- Visual connection status indicator
- Message handlers update local state and UI reactively

### Privacy System
- **Public**: Visible to all users and guests
- **Private**: Only visible to post owner, filtered by backend

## Configuration

### Vite Proxy (vite.config.js)
```javascript
export default defineConfig({
  plugins: [vue(), vuetify()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
})
```

### Router Guards (src/router/index.js)
```javascript
router.beforeEach((to, from, next) => {
  const publicRoutes = ['login', 'register', 'posts'];
  const requiresAuth = !publicRoutes.includes(to.name);

  if (requiresAuth && !isAuthenticated()) {
    next({ name: 'login' })
  } else if ((to.name === 'login' || to.name === 'register') && isAuthenticated()) {
    next({ name: 'posts' })
  } else {
    next()
  }
})
```

## Security Features

- Authentication required for create/edit/vote operations
- Ownership validation (only post owners can edit)
- Password requirements (minimum 6 characters)
- Current password verification for password changes
- Private post filtering by backend
- JWT token stored in localStorage with Bearer auth
- Input validation (client-side and server-side)
- Proper error handling without exposing internals

## Testing

### Test Real-Time Updates
1. Open two browsers (or normal + incognito)
2. Login as different users in each
3. Create/edit/vote in one browser → see instant updates in the other

### Test Authentication
1. Browse as guest → see posts, login button
2. Try to vote → "You must login" message
3. Login → welcome card appears, can create posts
4. Logout → back to guest view

### Test Privacy
1. Create private post → logout → post disappears
2. Login as different user → private post not visible
3. Login as owner → private post visible

## Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit: `git commit -m 'feat: add amazing feature'`
4. Push: `git push origin feature/amazing-feature`
5. Open Pull Request

### Commit Convention
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `style:` Code formatting
- `refactor:` Code refactoring
- `test:` Tests
- `chore:` Maintenance

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Vue.js Team - Amazing framework
- Vuetify Team - Material Design components
- Vite Team - Fast build tool
- DiceBear - Avatar generation API
- Backend Team - Go API with WebSocket support

---

## Recent Updates

### Ingress Configuration Fix (2025-01)
- **Issue**: 404 errors on all API endpoints, 413 errors on image uploads
- **Root Cause**: Missing NGINX Ingress rewrite annotations and body size limits
- **Fix**: Updated `k8s/ingress.yaml` with proper path rewriting and size limits
- **Changes**:
  - Added `nginx.ingress.kubernetes.io/rewrite-target: /$2` annotation
  - Added `nginx.ingress.kubernetes.io/use-regex: "true"` annotation
  - Added `nginx.ingress.kubernetes.io/proxy-body-size: "10m"` for file uploads
  - Updated path patterns to use regex: `/api(/|$)(.*)`
- **Impact**: All API endpoints now work correctly, image uploads up to 10MB supported

### Kubernetes Deployment (2025-01)
- Migrated from VM deployment to Kubernetes cluster
- **Production URL**: https://devops-mezdit.kubernetes.devops.cs.ut.ee/posts
- **Namespace**: `emzops-staging`
- **Components**:
  - Frontend deployment with NGINX
  - Backend deployment with persistent storage
  - PostgreSQL StatefulSet with Longhorn storage
  - Ingress with TLS/SSL support
  - RBAC for GitLab CI/CD deployment
  - NetworkPolicy for Prometheus metrics
- **CI/CD**: Automated deployment via GitLab pipelines using Kaniko for container builds

### Week 11 Compliance (2025-01)
- ✅ Separated metrics endpoint to port 9090
- ✅ Added NetworkPolicy for Prometheus scraping
- ✅ Configured persistent storage with Longhorn
- ✅ Implemented liveness and readiness probes
- ✅ Set up automated monitoring with Prometheus/Grafana

---

## AI Usage Disclosure

This project was developed with AI assistance for documentation purposes:

**AI was used for:**
- Writing and formatting README documentation
- Generating code comments and inline documentation
- Creating technical explanations and setup guides
- Structuring documentation sections and examples

The core development work, including all code implementation, design decisions, and technical solutions, was performed entirely by human developers. AI tools served only as assistants for documentation and explanation tasks.

---

**Built with ❤️ using Vue.js 3 and modern web technologies**
