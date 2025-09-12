import { createRouter, createWebHistory } from 'vue-router'
import Home from '../pages/Home.vue'
import CreatePost from '../pages/CreatePost.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: Home },
    { path: '/create', name: 'create', component: CreatePost },
  ],
})

export default router 