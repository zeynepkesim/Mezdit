import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '../store/auth.js'
import Login from '../pages/Login.vue'
import Home from '../pages/Home.vue'
import CreatePost from '../pages/CreatePost.vue'
import EditPost from '../pages/EditPost.vue'
import Register from '../pages/Register.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "login",
      component: Login,
    },
    {
      path: "/register",
      name: "register",
      component: Register,
    },
    { 
      path: '/posts', 
      name: 'posts', 
      component: Home
    },
    { 
      path: '/create-post', 
      name: 'create-post', 
      component: CreatePost
    },
    { 
      path: '/edit-post/:id', 
      name: 'edit-post', 
      component: EditPost
    }
  ],
});

router.beforeEach((to, from, next) => {
  if ((to.name === 'posts' || to.name === 'create-post' || to.name === 'edit-post') && !isAuthenticated()) {
    next({ name: 'login' })
  } else if ((to.name === 'login' || to.name === 'register') && isAuthenticated()) {
    next({ name: 'posts' })
  } else {
    next();
  }
})

export default router
