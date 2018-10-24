<template>
<div v-if="build.Status">
  <h1>{{ $route.params.owner }} / {{ $route.params.repo }}</h1>
  <h2>Build {{ $route.params.id }}</h2>
  <div>
    <span>Pull Request {{ pr.number }}</span>
    <span>{{ pr.title }}</span>
  </div>
  <div>Commit {{ pr.head.sha }}</div>
  <div>#{{ pr.number }}: {{ pr.title }}</div>
  <div>Branch {{ pr.base.ref }}</div>
  <div>{{ pr.user.login }}</div>
  <div>{{ duration() }}</div>
  <timeago :datetime="createdAt()" :auto-update="10"></timeago>
  <v-btn @click="toggleLive">
    <v-icon v-if="live">mdi-pause</v-icon>
    <v-icon v-else>mdi-play</v-icon>
  </v-btn>
  <v-data-iterator
    id="console-log"
    :rows-per-page-items="rowsPerPageItems"
    :pagination.sync="pagination"
    :items="filteredLines"
    >
    <v-flex slot="item" slot-scope="props" xs12 sm6 md4 lg3>
      <div class="line"><time>{{props.item.time}}</time><pre>{{props.item.line}}</pre></div>
    </v-flex>
  </v-data-iterator>
</div>
</template>

<script>
import Moment from 'moment'

export default {
  name: 'build',
  created() {
    const { owner, repo, id } = this.$route.params
    const projectName = `${owner}/${repo}`
    this.$socket.sendObj({ Kind: 'build', Data: { id, projectName } })
    this.$socket.sendObj({ Kind: 'build-log-watch', Data: { id, projectName } })
  },
  updated() {
    this.scrollToEnd()
  },
  destroyed() {
    const { owner, repo, id } = this.$route.params
    const projectName = `${owner}/${repo}`
    this.$socket.sendObj({ Kind: 'build-log-unwatch', Data: { id, projectName } })
  },
  methods: {
    toggleLive() {
      this.live = !this.live
    },
    scrollToEnd() {
      if (this.live && this.$el.querySelector) {
        const container = this.$el.querySelector('#console-log')
        container.scrollTop = container.scrollHeight
      }
    },
    duration() {
      if (this.finishedAt() === undefined) { return 'pending' }
      const { createdAt, finishedAt } = this
      const d = Moment.duration(Moment(createdAt()).diff(Moment(finishedAt())))
      return `Ran for ${d?.humanize()}`
    },
    finishedAt() {
      const time = this.build.FinishedAt?.Time
      if (time === undefined) { return undefined }
      return Moment(time)
    },
    createdAt() {
      const time = this.build.CreatedAt.Time
      if (time === undefined) { return undefined }
      return Moment(time)
    },
    age() {
      return this.createdAt()
    },
  },
  computed: {
    build() {
      return this.$store.state.socket.build
    },
    oldLines() {
      return this.$store.state.socket.build.Log.Elements.map(line => (
        line
      )).join('\n')
    },
    lines() {
      return this.$store.state.socket.build_lines
    },
    pr() {
      return this.build.Hook.pull_request
    },
    filteredLines() {
      return this.$store.state.socket.build_lines
    },
  },
  data() {
    return {
      live: true,
      rowsPerPageItems: [25, 50, 100],
      pagination: {
        rowsPerPage: 4,
      },
    }
  },
}
</script>

<style scoped>
  code:before {
    content: "";
  }
  #console-log {
    max-height: 30em;
    min-height: 30em;
    overflow-y: scroll;
    overflow-x: scroll;
  }
  #console-log pre {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-size: 1em;
    background: #212121;
  }
  #console-log time {
    background: #212121;
    float: left;
    padding-right: 1em;
  }
</style>
