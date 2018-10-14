<template>
  <v-app id="scylla" dark>
    <v-navigation-drawer
      v-model="drawer"
      clipped
      fixed
      app
    >
      <v-list dense>
        <v-list-tile to="/">
          <v-list-tile-action>
            <v-icon>mdi-home</v-icon>
          </v-list-tile-action>
          <v-list-tile-content>
            <v-list-tile-title>
              Dashboard
            </v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>

        <v-list-tile to="/builds">
          <v-list-tile-action>
            <v-icon>mdi-history</v-icon>
          </v-list-tile-action>
          <v-list-tile-content>
            <v-list-tile-title>
              Builds
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
        <v-layout align-center justify-center>
          <v-flex v-if="isConnected">
            <router-view/>
          </v-flex>
          <v-flex v-else style="text-align: center;">
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
    }
  },
}
</script>

<style>
.success-row { background-color: rgba(0,255,0,0.2); }
.queue-row   { background-color: rgba(255,255,0,0.2); }
.build-row   { background-color: rgba(255,255,255,0.2); }
.failure-row { background-color: rgba(255,0,0,0.2); }

pre.console-output {
  max-width: 96vw;
  overflow: auto;
  border-spacing: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: monospace;
}
</style>
