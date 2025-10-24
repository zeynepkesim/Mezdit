# Mezdit - Reddit-like Social Platform Frontend

A modern, full-featured Reddit-like social media platform built with Vue.js 3, featuring real-time updates via WebSocket, user authentication, post management, and voting system.

## ✨ Features

### 🔐 Authentication System
- **User Registration**: Create new accounts with name, username, and password
- **User Login**: Secure authentication with JWT-like session management
- **Change Password**: Secure password change with current password verification
- **Auto Profile Photos**: Automatic DiceBear avatar generation for all users
- **Persistent Sessions**: LocalStorage-based session management
- **Route Guards**: Protected routes requiring authentication

### 📝 Post Management
- **Create Posts**:
  - Posts with header (title) and content text
  - Optional image uploads (JPG, JPEG, PNG)
  - Custom display names or post anonymously
  - Privacy controls (public/private posts)
- **Edit Posts**:
  - Only post owners can edit their posts
  - Update text, images, display name, and privacy
  - Real-time validation
- **Anonymous Posting**: Toggle to post without revealing identity
- **Private Posts**: Posts visible only to the owner
- **Public Posts**: Posts visible to all users (including guests)
- **Image Support**: Upload and display images with posts

### 👥 User Experience
- **Guest Access**: Browse and view posts without authentication
- **Login Prompts**: Friendly prompts when guests try to interact (voting, posting)
- **Welcome Card**: Personalized greeting with member-since date
- **Profile Menu**: Quick access to change password and logout

### 👍👎 Voting System
- **Upvote/Downvote**: Like or dislike posts
- **Toggle Voting**: Click same vote again to remove it
- **Vote States**: Visual indicators (green for upvote, red for downvote, grey for no vote)
- **Real-time Updates**: Vote counts update instantly via WebSocket
- **User Vote Context**: See your current vote status on each post

### 🔄 Real-Time Features (WebSocket)
- **Live Post Creation**: New posts appear instantly without refresh
- **Live Post Updates**: Edits appear in real-time for all users
- **Live Vote Updates**: Vote counts update instantly across all clients
- **Privacy Changes**: Real-time visibility updates when privacy changes
- **Auto-Reconnect**: Exponential backoff reconnection strategy
- **Connection Status**: Visual indicator showing WebSocket connection state

### 🎨 Modern UI/UX
- **Vuetify 3**: Material Design components
- **Responsive Design**: Works on desktop, tablet, and mobile
- **Toast Notifications**: Success/error messages with auto-dismiss
- **Loading States**: Skeleton screens and progress indicators
- **Form Validation**: Real-time input validation with error messages
- **Password Strength Indicator**: Visual feedback for password security
- **Smooth Animations**: Transitions and hover effects

### 🌐 Public Access
- **Guest Viewing**: Anyone can browse posts without login
- **Login Button**: Prominent login button in header for guests
- **Conditional Features**: Create/Edit/Vote features require authentication
- **Seamless Flow**: Easy transition from guest to authenticated user

## 🏗️ Tech Stack

### Frontend
- **Vue.js 3** - Progressive JavaScript framework with Composition API
- **Vuetify 3** - Material Design component framework
- **Vite** - Next-generation frontend build tool
- **Vue Router 4** - Official routing library
- **Axios** - Promise-based HTTP client
- **Vue Toast Notification** - Toast notification system
- **WebSocket API** - Native browser WebSocket for real-time updates

### Backend Integration
- **Go Backend** - RESTful API with WebSocket support
- **PostgreSQL** - Database for persistent storage
- **Real-time Broadcasts** - WebSocket Hub for multi-client updates

## 📁 Project Structure

```
devops-fe/
├── src/
│   ├── api/                    # API layer
│   │   ├── axios.js           # Axios instance configuration
│   │   ├── auth.js            # Authentication API calls
│   │   └── posts.js           # Post management API calls
│   ├── composables/           # Vue composition functions
│   │   └── useWebSocket.js    # WebSocket connection manager
│   ├── pages/                 # Page components
│   │   ├── Login.vue          # Login page with form
│   │   ├── Register.vue       # Registration page
│   │   ├── Home.vue           # Main feed with posts list
│   │   ├── CreatePost.vue     # Post creation form
│   │   ├── EditPost.vue       # Post editing form
│   │   └── ChangePassword.vue # Password change form
│   ├── router/                # Vue Router configuration
│   │   └── index.js           # Routes and navigation guards
│   ├── store/                 # State management
│   │   ├── auth.js            # Authentication state
│   │   └── posts.js           # Posts state + WebSocket handlers
│   ├── App.vue                # Root component with header
│   └── main.js                # Application entry point
├── public/                    # Static assets
├── .gitignore                 # Git ignore rules
├── index.html                 # HTML entry point
├── package.json               # Dependencies and scripts
├── vite.config.js             # Vite configuration
└── README.md                  # This file
```

## 🚀 Getting Started

### Prerequisites

- **Node.js** (v18 or higher recommended)
- **npm** or **yarn** package manager
- **Backend API** running on `http://localhost:8080`

### Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd devops-fe
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Start the development server:**
   ```bash
   npm run dev
   ```

4. **Open your browser:**
   Navigate to `http://localhost:5173`

### Backend Setup

The frontend requires the backend API to be running. See backend documentation for setup instructions.

**Backend must be running on:**
- API: `http://localhost:8080`
- WebSocket: `ws://localhost:8080/ws`

## 📜 Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server with hot reload |
| `npm run build` | Build optimized production bundle |
| `npm run preview` | Preview production build locally |

## 🔌 API Integration

### HTTP Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/login` | User authentication | No |
| `POST` | `/register` | User registration | No |
| `POST` | `/change-password` | Change user password | Yes |
| `GET` | `/posts` | Fetch all posts (with optional userId) | No |
| `POST` | `/posts` | Create new post | Yes |
| `POST` | `/posts/upload` | Create post with image | Yes |
| `POST` | `/posts/edit` | Edit existing post | Yes |
| `POST` | `/posts/edit-with-image` | Edit post with image | Yes |
| `POST` | `/vote` | Vote on post (upvote/downvote) | Yes |
| `GET` | `/users/{username}` | Get user by username | No |
| `GET` | `/images/{filename}` | Get uploaded image | No |

### WebSocket Connection

**Endpoint:** `ws://localhost:8080/ws`

**Message Types:**
- `POST_CREATED` - New post created
- `POST_UPDATED` - Post edited
- `POST_VOTED` - Vote added/changed/removed
- `POST_PRIVACY_CHANGED` - Post privacy toggled

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

## 🎯 Key Features Explained

### Authentication Flow

1. **Registration:**
   - User enters name, username, password
   - Backend creates user with auto-generated avatar
   - Returns user data with profile photo URL
   - User is automatically logged in
   - Redirected to posts page

2. **Login:**
   - User enters username and password
   - Backend validates credentials
   - Returns user data
   - Stored in localStorage for persistence
   - Redirected to posts page

3. **Session Management:**
   - User data stored in localStorage
   - Auto-loaded on page refresh
   - Router guards protect authenticated routes
   - Logout clears localStorage and reloads posts

### Post Creation Workflow

1. **Navigate to Create Post** (authenticated users only)
2. **Fill Form:**
   - Header (required)
   - Content text (required)
   - Display name (optional, defaults to "Anonymous" if empty)
   - Image upload (optional)
   - Anonymous toggle (disables display name)
   - Private toggle (makes post private)
3. **Submit:**
   - Validated before submission
   - API call creates post
   - WebSocket broadcasts to all users
   - Redirect to home page
   - New post appears at top of feed

### Voting System

**States:**
- **No Vote**: Grey buttons, no background
- **Upvoted**: Green upvote button, flat variant
- **Downvoted**: Red downvote button, flat variant

**Toggle Behavior:**
- Click upvote when not voted → Add upvote
- Click upvote when already upvoted → Remove vote
- Click downvote when upvoted → Change to downvote
- Same logic for downvote button

**Real-time Updates:**
1. User clicks vote button
2. API call to `/vote`
3. Backend updates database
4. Backend broadcasts via WebSocket
5. All connected clients receive update
6. UI updates instantly (~100ms)

### WebSocket Real-Time System

**Connection Lifecycle:**
```
Page Load
    ↓
Connect to ws://localhost:8080/ws
    ↓
Connected (isConnected = true)
    ↓
Listen for messages
    ↓
[Connection Lost]
    ↓
Auto-reconnect (exponential backoff)
    ↓
1s → 2s → 4s → 8s → 16s → 30s (max)
    ↓
Reconnected
    ↓
Continue listening
```

**Message Handling:**
```javascript
// Incoming WebSocket message
{
  type: 'POST_VOTED',
  payload: { id: 1, voteCount: 42, isLiked: true, ... }
}
    ↓
handleWebSocketMessage(message)
    ↓
Switch on message.type
    ↓
handlePostVoted(payload)
    ↓
Find post in local state
    ↓
Update post with new data
    ↓
Vue reactivity updates UI
```

### Privacy System

**Public Posts:**
- Visible to all users (guests and authenticated)
- Included in `/posts` endpoint without userId
- Badge: No badge shown

**Private Posts:**
- Visible only to post owner
- Filtered by backend based on userId parameter
- Badge: Yellow "Private" chip with lock icon
- Guests cannot see private posts

### Anonymous Posting

When anonymous toggle is enabled:
- Display name field is disabled
- Display name sent as empty string to backend
- Backend stores as "Anonymous"
- Post shows "Anonymous" with incognito icon
- Owner still maintains ownership for editing

## 🎨 UI Components

### Header (App.vue)
- **Logo**: Mezdit with flash icon
- **For Guests**: Login button
- **For Users**:
  - Create Post button
  - User avatar menu (Profile, Change Password, Logout)

### Home Page (Main Feed)
- **Connection Status**: Real-time WebSocket indicator
- **Welcome Card**: Personalized greeting (authenticated users)
- **Posts List**:
  - Post header with private badge
  - Author avatar and name
  - Timestamp
  - Content text
  - Image (if present)
  - Vote buttons and count
  - Edit button (owner only)

### Forms
- **Login/Register**: Side-by-side branding and form layout
- **Create/Edit Post**:
  - Text fields with validation
  - Image upload with preview
  - Toggle switches (Anonymous, Private)
  - Password strength indicator (Change Password)

### Toast Notifications
- **Position**: Bottom-right
- **Types**: Success (green), Error (red)
- **Duration**: 3-5 seconds
- **Examples**:
  - "Successfully logged in!"
  - "Post created successfully!"
  - "You must login to vote"
  - "Invalid username or password"

## ⚙️ Configuration

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

This proxies all `/api/*` requests to the backend, avoiding CORS issues.

### Axios Configuration (src/api/axios.js)

```javascript
const api = axios.create({
  baseURL: "/api",
  headers: { "Content-Type": "application/json" }
})

// Error interceptor for logging
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error("API Error:", error.response?.data || error.message);
    return Promise.reject(error)
  }
)
```

### Router Guards (src/router/index.js)

```javascript
router.beforeEach((to, from, next) => {
  const publicRoutes = ['login', 'register', 'posts'];
  const requiresAuth = !publicRoutes.includes(to.name);

  if (requiresAuth && !isAuthenticated()) {
    next({ name: 'login' })  // Redirect to login
  } else if ((to.name === 'login' || to.name === 'register') && isAuthenticated()) {
    next({ name: 'posts' })  // Redirect to posts if already logged in
  } else {
    next()  // Allow navigation
  }
})
```

## 🔒 Security Features

- **Authentication Required**: Protected routes for create, edit, vote
- **Ownership Validation**: Only post owners can edit posts
- **Password Requirements**: Minimum 6 characters
- **Current Password Verification**: Required for password change
- **Private Posts**: Backend filters based on ownership
- **Input Validation**: Client-side and server-side validation
- **Error Handling**: Proper error messages without exposing internals

## 🐛 Error Handling

### API Errors
- Network errors caught and displayed as toast
- Backend error messages parsed and shown to user
- Validation errors shown inline in forms
- Loading states prevent multiple submissions

### WebSocket Errors
- Connection failures logged to console
- Auto-reconnect with exponential backoff
- Visual indicator for connection state
- Graceful degradation (app works without WebSocket)

### Form Validation
- Required field validation
- Password length validation
- Password match validation (confirm password)
- Display name required when not anonymous
- Real-time validation feedback

## 🧪 Testing Scenarios

### Test Real-Time Updates

1. **Open two browsers** (or normal + incognito)
2. **Login as different users** in each
3. **In Browser 1**: Create a new post
4. **In Browser 2**: Post appears instantly! ✨
5. **In Browser 1**: Vote on a post
6. **In Browser 2**: Vote count updates instantly!
7. **In Browser 1**: Edit a post
8. **In Browser 2**: Changes appear instantly!

### Test Authentication Flow

1. **Open app as guest** - See posts, login button
2. **Try to vote** - "You must login to vote" message
3. **Click login** - Navigate to login page
4. **Login** - Redirected to posts with welcome card
5. **Create post** - New post appears at top
6. **Logout** - Redirected to posts as guest

### Test Privacy

1. **Create private post** (toggle private on)
2. **Logout** - Private post disappears
3. **Login** - Private post reappears
4. **Login as different user** - Private post not visible

## 📊 Performance Considerations

### Optimizations
- **Vue 3 Composition API**: Better tree-shaking and performance
- **Vite**: Fast HMR and optimized builds
- **Lazy Loading**: Router-based code splitting (ready for implementation)
- **WebSocket**: More efficient than polling for real-time updates
- **LocalStorage**: Reduces auth API calls
- **Computed Properties**: Efficient reactive updates

### WebSocket Efficiency
- Single persistent connection vs. polling every X seconds
- ~100ms latency for updates
- Bidirectional communication
- Auto cleanup on disconnect

## 🚧 Future Enhancements

### Planned Features
- [ ] Comment system
- [ ] User profiles with post history
- [ ] Post search and filtering
- [ ] Categories/Tags
- [ ] Notifications system
- [ ] Dark mode
- [ ] Post bookmarking
- [ ] Share functionality
- [ ] Infinite scroll / Pagination
- [ ] Post analytics (views, engagement)

### Technical Improvements
- [ ] Unit tests (Vitest)
- [ ] E2E tests (Playwright/Cypress)
- [ ] TypeScript migration
- [ ] PWA support
- [ ] Image optimization
- [ ] CDN integration
- [ ] Service Worker for offline support
- [ ] Performance monitoring

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Branch Strategy
- `main` - Production-ready code
- `realtime-data` - Real-time features with WebSocket
- `feature/*` - New features
- `bugfix/*` - Bug fixes

## 📝 Git Workflow

```bash
# Switch to feature branch
git checkout -b feature/new-feature

# Make changes and commit
git add .
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/new-feature

# Create pull request on GitHub/GitLab
```

### Commit Message Convention

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `style:` Code style changes (formatting)
- `refactor:` Code refactoring
- `test:` Adding tests
- `chore:` Maintenance tasks

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- **Vue.js Team** - Amazing framework
- **Vuetify Team** - Beautiful Material Design components
- **Vite Team** - Lightning-fast build tool
- **DiceBear** - Avatar generation API
- **Backend Team** - Solid Go API with WebSocket support

## 📞 Support

For issues, questions, or contributions:
- Open an issue on GitHub
- Check existing issues first
- Provide clear reproduction steps for bugs
- Include screenshots for UI issues

## 🔗 Related Projects

- **Backend**: DevOps BE - Go backend with WebSocket support
- **Database**: PostgreSQL for data persistence

---

**Built with ❤️ using Vue.js 3 and modern web technologies**
