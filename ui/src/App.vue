<template>
  <v-app id="app" >
    <div>
    <v-navigation-drawer
      app
      permanent
      color="#E0E0E0"
      mini-variant
      mini-variant-width="56"
      absolute
    >
      <v-list-item class="px-2">
        <v-list-item-avatar>
          <v-img :src="require('@/assets/images/logo.jpg')"></v-img>
        </v-list-item-avatar>
      </v-list-item>
      <v-divider></v-divider>
      <v-list-item-group dense nav v-model="itemIndex">
        <v-list-item v-for="(item) in items" :key="item.title" :to="item.path" link>
          <v-list-item-action>
            <v-icon>{{ item.icon }}</v-icon>
          </v-list-item-action>
          <v-list-item-content>
            <v-list-item-title>{{ item.title }}</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-list-item-group>
    </v-navigation-drawer>
    <v-content>
      <v-container fluid>
        <router-view></router-view>
      </v-container>
    </v-content>
    </div>
  </v-app>
</template>

<script>
export default {
  name: "app",
  data() {
    return {
      items: [
        { title: "Home", icon: "mdi-view-dashboard",path:"/"},
      ],
      mini: true,
      subMemu: false,
      itemIndex: 0
    };
  },
  created(){
    window.oncontextmenu=function(e){
      e.preventDefault();
    }
  },
  methods: {
    clickItem(index) {
      if (index == 1) {
        this.subMemu = true;
        return;
      }
      this.subMemu = false;
    }
  }
};
</script>

<style>
html {
  overflow-y: auto !important;
}
</style>
