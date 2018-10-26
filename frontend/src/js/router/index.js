import Vue from 'vue'
import Router from 'vue-router'
import Hello from '../components/Hello'
import Builds from '../components/Builds'
import Build from '../components/Build'
import Organizations from '../components/Organizations'
import Organization from '../components/Organization'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Hello',
      component: Hello,
    },
    {
      path: '/builds',
      name: 'Builds',
      component: Builds,
    },
    {
      path: '/builds/:owner/:repo/:id',
      name: 'Build',
      component: Build,
    },
    {
      path: '/organizations',
      name: 'Organizations',
      component: Organizations,
    },
    {
      path: '/organizations/:name',
      name: 'Organization',
      component: Organization,
    },
  ],
})
