import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      redirect: '/home',
    },
    {
      path: '/home',
      name: 'Home',
      component: () => import('../views/home/index.vue'),
    },
    {
      path: '/metronome',
      name: 'Metronome',
      component: () => import('../views/metronome/index.vue'),
    },
  ],
})

export default router
