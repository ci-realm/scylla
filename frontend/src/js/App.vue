<template>
  <v-app id="scylla" dark>
    <v-navigation-drawer
      v-model="drawer"
      clipped
      fixed
      app
    >
      <v-list dense>
        <v-list-tile v-for="item in menuItems" :to="item.link">
          <v-list-tile-action>
            <v-icon>{{item.icon}}</v-icon>
          </v-list-tile-action>
          <v-list-tile-content>
            <v-list-tile-title>
              {{item.title}}
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-navigation-drawer>

    <v-toolbar app fixed clipped-left>
      <v-toolbar-side-icon @click.stop="drawer = !drawer"></v-toolbar-side-icon>
      <v-toolbar-title>Scylla</v-toolbar-title>
    </v-toolbar>

    <v-content>
      <v-container fluid fill-height>
        <v-layout justify-center v-if="isConnected">
          <v-flex>
            <router-view/>
          </v-flex>
        </v-layout>
        <v-layout v-else align-center justify-center>
          <v-flex style="text-align: center;">
            <v-progress-circular
              :size="70"
              :width="7"
              color="teal"
              indeterminate
            ></v-progress-circular>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>

    <v-footer app fixed height="auto">
      <v-layout row wrap justify-center style="text-align: center;">
        <v-flex xs4>
          <a href="https://builtwithnix.org">built with Nix</a>
        </v-flex>
        <v-flex xs4>
          <span class="copy">&copy; 2018 by manveru</span>
        </v-flex>
        <v-flex xs4>
          <a :href="scyllaVersionLink">{{ scyllaHostname }}</a>
        </v-flex>
      </v-layout>
    </v-footer>
  </v-app>
</template>

<script>
export default {
  name: 'app',
  props: {
    source: String,
  },
  computed: {
    isConnected() {
      return this.$store.state.socket.isConnected
    },
  },
  data() {
    return {
      drawer: true,
      scyllaVersionLink: 'localhost',
      scyllaHostname: 'localhost',
      menuItems: [
        { link: '/', icon: 'mdi-home', title: 'Home' },
        { link: '/builds', icon: 'mdi-history', title: 'Builds' },
        { link: '/organizations', icon: 'mdi-apps', title: 'Organizations' },
      ],
    }
  },
}
</script>

<style>
.success-row { background-color: rgba(0,255,0,0.2); }
.queue-row   { background-color: rgba(255,255,0,0.2); }
.build-row   { background-color: rgba(255,255,255,0.2); }
.failure-row { background-color: rgba(255,0,0,0.2); }
</style>
