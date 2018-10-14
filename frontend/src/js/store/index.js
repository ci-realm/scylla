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
    LAST_BUILDS(state, message) {
      state.socket.lastBuilds = message.Data.builds
    },
    BUILD(state, message) {
      state.socket.build = message.Data.build
      state.socket.build_lines = message.Data.build.Log.map((log) => {
        const time = Moment(log.created_at).format('HH:mm:ss:SS')
        return { time, line: log.line }
      })
    },
    BUILD_LOG(state, message) {
      const time = Moment(message.Data.time).format('HH:mm:ss:SS')
      state.socket.build_lines.push({ time, line: message.Data.line })
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
