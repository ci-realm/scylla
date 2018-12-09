import Vue from 'vue'
import Vuex from 'vuex'
import Moment from 'moment'

Vue.use(Vuex)

export default new Vuex.Store({
  strict: true,
  state: {
    socket: {
      isConnected: false,
      message: '',
      error: '',
      reconnects: 0,
      lastBuilds: [],
      organizations: [],
      organizationBuilds: [],
      build: {},
      build_lines: [],
      reconnectError: false,
    },
  },
  mutations: {
    SOCKET_ONOPEN(state, event) {
      Vue.prototype.$socket = event.currentTarget
      state.socket.isConnected = true
    },
    // SOCKET_ONCLOSE(state, event) {
    SOCKET_ONCLOSE(state) {
      state.socket.isConnected = false
    },
    SOCKET_ONERROR(state, event) {
      state.socket.error = event
    },
    // default handler called for all methods
    SOCKET_ONMESSAGE(state, message) {
      state.socket.message = message
    },
    // mutations for reconnect methods
    SOCKET_RECONNECT(state, count) {
      state.socket.reconnects = count
    },
    SOCKET_RECONNECT_ERROR(state) {
      state.socket.reconnectError = true
    },

    // Custom mutation messages

    lastBuilds(state, message) {
      state.socket.lastBuilds = message.data.builds
    },
    organizations(state, message) {
      state.socket.organizations = message.data.organizations
    },
    organizationBuilds(state, message) {
      state.socket.organizationBuilds = message.data.organizationBuilds
    },
    build(state, message) {
      state.socket.build = message.data.build
      state.socket.build_lines = message.data.build.log.map((log) => {
        const time = Moment(log.created_at).format('HH:mm:ss:SS')
        return { time, line: log.line }
      })
    },
    buildLog(state, message) {
      const time = Moment(message.data.time).format('HH:mm:ss:SS')
      state.socket.build_lines.push({ time, line: message.data.line })
    },
  },
  actions: {
    sendMessage(context, message) {
      // .....
      Vue.prototype.$socket.send(message)
      // .....
    },
  },
})
