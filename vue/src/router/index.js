import { createRouter, createWebHashHistory } from 'vue-router'
import BasicLayout from '../layouts/BasicLayout.vue'
import Main from '@/views/index/Main.vue'
import Authorize from '@/views/settings/Authorize.vue'
import ApiInfo from '@/views/api/index.vue'
import SwaggerModels from '@/views/settings/SwaggerModels.vue'
import GlobalParameters from '@/views/settings/GlobalParameters.vue'
import OfficelineDocument from '@/views/settings/OfficelineDocument.vue'
import Settings from '@/views/settings/Settings.vue'
import Othermarkdown from '@/views/othermarkdown/index.vue'

export const routes = [{
  path: '/',
  name: 'home',
  component: BasicLayout,
  redirect: '/home',
  children: [{
    path: '/home',
    component: Main
  }, {
    path: '/home/:i18n',
    component: Main
  }, {
    path: '/plus',
    component: Main
  }, {
    path: '/plus/:i18n',
    component: Main
  },
    {
      path: '/Authorize/:groupName',
      component: Authorize
    },
    {
      path: '/:groupName/:controller/:summary',
      component: ApiInfo
    }, {
      path: '/SwaggerModels/:groupName',
      component: SwaggerModels
    }, {
      path: '/documentManager/GlobalParameters-:groupName',
      component: GlobalParameters
    }, {
      path: '/documentManager/OfficelineDocument-:groupName',
      component: OfficelineDocument
    }, {
      path: '/documentManager/Settings',
      component: Settings
    }, {
      path: '/:groupName-:mdid-omd/:id',
      component: Othermarkdown
    }
  ]
},
  {
    path: '/oauth2',
    name: 'oauth2',
    component: () => import('@/views/settings/OAuth2.vue')
  }];

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes
})

export default router
