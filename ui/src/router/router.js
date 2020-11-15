import Vue from 'vue'
import Router from 'vue-router'
import Index from '@/pages/index/index'
import Log from '@/pages/index/log'
import Setting from '@/pages/setting/index'

Vue.use(Router)

let router = new Router({
  mode: '',
  routes: [
    {
      path: '/',
      name: 'Index',
      component: Index,
      meta:{
        title: 'GoAnsible'
      }
    },
    {
      path: '/log',
      name: 'Log',
      component: Log,
      meta:{
        title: '日志' 
      }
    },
    {
      path: '/setting',
      name: 'Setting',
      component: Setting,
      meta:{
        title: '设置'
      }
    },
    ]
})


export default router

