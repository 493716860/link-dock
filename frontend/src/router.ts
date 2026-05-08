import { createRouter, createWebHistory } from 'vue-router';

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: { template: '<div />' },
    },
    {
      path: '/admin/seed',
      name: 'admin-seed',
      component: { template: '<div />' },
      meta: {
        requiresSuperAdmin: true,
      },
    },
  ],
});

