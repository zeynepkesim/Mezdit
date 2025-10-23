import { createRouter, createWebHistory } from 'vue-router'
import { isAuthenticated } from '../store/auth.js'
import Login from '../pages/Login.vue'
import Home from '../pages/Home.vue'
import CreatePost from '../pages/CreatePost.vue'
import EditPost from '../pages/EditPost.vue'
import Register from '../pages/Register.vue'
import ChangePassword from '../pages/ChangePassword.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      redirect: "/posts",
    },
    {
      path: "/login",
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
    },
    {
      path: '/change-password',
      name: 'change-password',
      component: ChangePassword
    }
  ],
});

router.beforeEach((to, from, next) => {
  // Posts sayfası herkese açık, diğer sayfalar authentication gerektirir
  const publicRoutes = ['login', 'register', 'posts'];
  const requiresAuth = !publicRoutes.includes(to.name);

  if (requiresAuth && !isAuthenticated()) {
    // Auth gerektiren sayfalara giriş yapmadan erişilmeye çalışılırsa login'e yönlendir
    next({ name: 'login' })
  } else if ((to.name === 'login' || to.name === 'register') && isAuthenticated()) {
    // Giriş yapmış kullanıcı login/register'a gitmeye çalışırsa posts'a yönlendir
    next({ name: 'posts' })
  } else {
    next();
  }
})

export default router
