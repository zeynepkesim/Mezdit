# Reddit Clone Frontend

A simple Reddit-like frontend application built with Vue.js 3 and Vite.

## Features

- 📝 **Create Posts**: Users can create posts with author name and content
- 👍👎 **Voting System**: Upvote and downvote posts with real-time updates
- 🔄 **Real-time Updates**: Posts are fetched from backend API
- 🎨 **Modern UI**: Clean, responsive design with toast notifications
- 🚀 **Fast Development**: Built with Vite for lightning-fast development

## Tech Stack

- **Vue.js 3** - Progressive JavaScript framework
- **Vite** - Next-generation frontend build tool
- **Vue Router** - Official router for Vue.js
- **Axios** - HTTP client for API requests
- **CSS3** - Modern styling with flexbox and grid

## Project Structure

```
src/
├── api/           # API layer (HTTP requests)
├── components/    # Reusable Vue components
├── composables/   # Vue composition functions
├── pages/         # Page components
├── router/        # Vue Router configuration
├── store/         # State management
└── assets/        # Static assets
```

## Getting Started

### Prerequisites

- Node.js (v16 or higher)
- npm or yarn
- Backend API running on `http://localhost:8080`

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd devops-fe
   ```

2. Install dependencies:
   ```bash
   npm install
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open your browser and navigate to `http://localhost:5173`

### Backend Integration

The frontend communicates with a Go backend API. Make sure the backend is running on `http://localhost:8080` before starting the frontend.

**API Endpoints:**
- `GET /posts` - Fetch all posts
- `POST /posts` - Create a new post
- `POST /vote` - Vote on a post (upvote/downvote)

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build locally

## Features Overview

### Home Page (`/`)
- Displays list of posts sorted by creation date (newest first)
- Each post shows:
  - Author avatar (generated from username)
  - Author name and creation timestamp
  - Post content
  - Upvote/Downvote buttons with counts
- Loading states and error handling
- Responsive design

### Create Post Page (`/create`)
- Form with author name and content fields
- Real-time validation
- Loading states during submission
- Success/error toast notifications
- Automatic redirect to home after successful creation

### Voting System
- Users can upvote or downvote posts
- Visual feedback with active states (green/red backgrounds)
- Real-time vote count updates
- Backend integration for persistent voting

### Toast Notifications
- Success notifications for successful actions
- Error notifications with backend error messages
- Auto-dismiss after 3 seconds
- Click to dismiss manually
- Smooth animations

## Configuration

### Vite Proxy Setup

The project uses Vite's proxy feature to avoid CORS issues during development:

```javascript
// vite.config.js
export default defineConfig({
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

### Environment Variables

Create a `.env` file in the root directory if you need to customize the API URL:

```env
VITE_API_BASE_URL=/api
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Commit your changes: `git commit -am 'Add feature'`
4. Push to the branch: `git push origin feature-name`
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Architecture Notes

### State Management
- Uses Vue 3 Composition API for reactive state
- Centralized store pattern without Vuex/Pinia for simplicity
- Local state for UI interactions (voting states)

### API Layer
- Axios-based HTTP client with error handling
- JSON API communication with proper Content-Type headers
- Response/error transformation for consistent UI feedback

### Routing
- Vue Router 4 with history mode
- Simple two-page application (Home, Create Post)
- Programmatic navigation for form submissions

### Styling
- Scoped CSS for component isolation
- Utility-first approach with consistent design tokens
- Responsive design with mobile-first approach
- CSS Grid and Flexbox for layouts
