<template>
<div>
  <h1>Recent Builds for {{ $route.params.name }}</h1>
  <v-data-table :headers="headers" :items="buildData" class="elevation-1" :rows-per-page-items="rowsPerPageItems"
    :pagination.sync="pagination">
    <template slot="items" slot-scope="props">
      <td>
        <router-link :to="props.item.buildLink">#{{props.item.id}}</router-link>
      </td>
      <td>
        <v-icon :color="statusColor(props.item.status)">{{ statusIcon(props.item.status) }}</v-icon>
      </td>
      <td>
        <a :href="props.item.ownerLink">{{ props.item.owner }}</a><br/>
        <a :href="props.item.repoLink">{{ props.item.repo }}</a>
      </td>
      <td>
        <a :href="props.item.shaLink">{{ shortSHA(props.item.sha) }}</a>
      </td>
      <td>
        <v-icon small>mdi-calendar-clock</v-icon>
        <timeago :datetime="props.item.time" :auto-update="10"></timeago>
      </td>
      <td>
        {{ props.item.duration.humanize() }}
      </td>
    </template>
  </v-data-table>
  <v-btn ripple color="info" @click="fetchBuilds">Fetch</v-btn>
</div>
</template>

<script>
import Moment from 'moment'

export default {
  name: 'builds',
  created() {
    if (!this.refresher) {
      const msg = { Kind: 'organization-builds', Data: { orgName: this.$route.params.name } }
      this.$socket.sendObj(msg)
      this.refresher = setInterval(() => {
        this.$socket.sendObj(msg)
      }, 5000)
    }
  },
  destroyed() {
    if (this.refresher) { clearInterval(this.refresher) }
  },
  methods: {
    tableRowClassName({ row }) {
      return `${row.Status}-row`
    },
    fetchBuilds() {
      this.$socket.sendObj({ Kind: 'last-builds' })
    },
    statusColor(status) {
      const statusMap = {
        failure: 'error',
        success: 'success',
      }
      return statusMap[status] || 'info'
    },
    statusIcon(status) {
      const statusMap = {
        failure: 'mdi-emoticon-sad',
        success: 'mdi-emoticon-happy',
      }
      return statusMap[status] || 'mdi-run'
    },
    shortSHA(sha) {
      return sha.substr(0, 7)
    },
    timeDiff(from, to) {
      // TODO: check when `to` is not there.
      if (to === undefined) { return 'pending' }
      return Moment.duration(Moment(from.Time).diff(Moment(to.Time)))
    },
    linkToBuild(build) {
      const { repo } = build.Hook.pull_request.head
      return `/builds/${repo.owner.login}/${repo.name}/${build.ID}`
    },
  },
  computed: {
    headers() {
      return [
        { text: 'Build', value: 'id' },
        { text: 'Status', value: 'status' },
        { text: 'Project', value: 'project' },
        { text: 'SHA', value: 'sha' },
        { text: 'Time', value: 'time' },
        { text: 'Duration', value: 'duration' },
      ]
    },
    buildData() {
      return this.$store.state.socket.organizationBuilds.map((build) => {
        const pr = build.Hook.pull_request

        return {
          value: false,
          time: build.CreatedAt.Time,
          project: build.ProjectName,
          owner: pr.head.repo.owner.login,
          ownerLink: pr.head.repo.owner.html_url,
          repo: pr.head.repo.name,
          repoLink: pr.head.repo.html_url,
          status: build.Status,
          sha: this.shortSHA(pr.head.sha),
          shaLink: `${pr.base.repo.html_url}/commit/${pr.head.sha}`,
          duration: this.timeDiff(build.FinishedAt, build.CreatedAt),
          prLink: pr.html_url,
          buildLink: this.linkToBuild(build),
          id: build.ID,
        }
      })
    },
  },
  data() {
    return {
      pagination: {
        sortBy: 'id',
        descending: true,
      },
      rowsPerPageItems: [25, 50, 100, { text: '$vuetify.dataIterator.rowsPerPageAll', value: -1 }],
    }
  },
}
</script>

<style scoped>
.v-table {
  width: 100%;
  max-width: 100%;
}
</style>
