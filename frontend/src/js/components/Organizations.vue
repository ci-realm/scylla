<template>
<div>
  <h1>Organizations</h1>
  <v-data-table :headers="headers" :items="orgData" class="elevation-1" :rows-per-page-items="rowsPerPageItems"
    :pagination.sync="pagination">
    <template slot="items" slot-scope="props">
      <td>
        <router-link :to="props.item.orgLink">{{ props.item.owner }}</router-link>
      </td>
      <td>
        <a :href="props.item.url">{{ props.item.url }}</a><br/>
      </td>
      <td>
        {{ props.item.buildCount }}
      </td>
    </template>
  </v-data-table>
</div>
</template>

<script>
export default {
  name: 'organizations',
  created() {
    if (!this.refresher) {
      this.$socket.sendObj({ Kind: 'organizations' })
      this.refresher = setInterval(() => {
        this.$socket.sendObj({ Kind: 'organizations' })
      }, 5000)
    }
  },
  destroyed() {
    if (this.refresher) { clearInterval(this.refresher) }
  },
  methods: {
    linkToBuild(build) {
      const { repo } = build.Hook.pull_request.head
      return `/builds/${repo.owner.login}/${repo.name}/${build.ID}`
    },
  },
  computed: {
    headers() {
      return [
        { text: 'Organization', value: 'owner' },
        { text: 'URL', value: 'url' },
        { text: 'Builds', value: 'buildCount' },
      ]
    },
    orgData() {
      return this.$store.state.socket.organizations.map(org => ({
        value: false,
        owner: org.Owner,
        url: org.URL,
        buildCount: org.BuildCount,
        orgLink: `/organizations/${org.Owner}`,
      }))
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
