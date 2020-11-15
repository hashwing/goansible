import 'core-js/stable';
import '@mdi/font/css/materialdesignicons.css';
import Vue from 'vue';
import Vuetify from 'vuetify';
import 'vuetify/dist/vuetify.min.css';

Vue.use(Vuetify);

import App from './App.vue';

Vue.config.productionTip = false;
Vue.config.devtools = true;


import router from './router/router'
router.beforeEach((to, from, next) => {
	if (to.meta.title) {
		document.title = to.meta.title
	}
	next()
})


new Vue({
	router,
	vuetify: new Vuetify({
		icons: {
			iconfont: 'mdi'
		},
		theme: {
			dark: false
		}
	}),
	render: h => h(App),
	mounted() {
		this.$router.replace('/')
	},
}).$mount('#app');