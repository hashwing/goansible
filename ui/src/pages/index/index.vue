<template>
  <div fluid class="px-0" >
    <v-breadcrumbs :items="navs">
      <template v-slot:item="{ item }">
        <v-breadcrumbs-item :to="item.href" :disabled="item.disabled">
          {{ item.text.toUpperCase() }}
        </v-breadcrumbs-item>
      </template>
    </v-breadcrumbs>
    <v-col align="right">
      <v-dialog v-model="dialog" persistent max-width="600px">
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            color="primary"
            align="right"
            dark
            v-bind="attrs"
            text
            v-on="on"
          >
            <v-icon left>mdi-plus</v-icon> 创建
          </v-btn>
        </template>
        <v-card>
          <v-card-title>
            <span>添加</span>
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col cols="12">
                  <v-text-field
                    label="Name*"
                    v-model="actionParam.args.name"
                    hint="仅支持字母、数字和‘-’"
                    required
                  ></v-text-field>
                </v-col>
              </v-row>
            </v-container>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn
              color="blue darken-1"
              text
              @click="
                dialog = false;
                initActionParam();
              "
              >取消</v-btn
            >
            <v-btn
              color="blue darken-1"
              text
              @click="
                actionParam.key = 'create';
                doAction();
              "
              >创建</v-btn
            >
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-col>
    <v-simple-table class="float-none">
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">名称</th>
            <th class="text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in playbooks" :key="item.name">
            <td>{{ item.name }}</td>
            <td class="text-right">
              <v-menu bottom :offset-x="true" :offset-y="true">
                <template v-slot:activator="{ on, attrs }">
                  <v-btn icon v-bind="attrs" v-on="on">
                    <v-icon>mdi-dots-vertical</v-icon>
                  </v-btn>
                </template>
                <v-list dense>
                  <v-list-item
                    v-for="(action, i) in actions"
                    :key="i"
                    @click="
                      actionParam.key = action.key;
                      actionParam.args.name = item.name;
                      doAction();
                    "
                  >
                    <v-list-item-title>
                      <v-icon dense :color="action.color">{{
                        action.icon
                      }}</v-icon>
                      {{ action.title }}
                    </v-list-item-title>
                  </v-list-item>
                </v-list>
              </v-menu>
            </td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>
    <v-dialog v-model="delDialog" persistent max-width="290">
      <v-card>
        <v-card-title></v-card-title>
        <v-card-text>确定删除‘{{ actionParam.args.name }}’?</v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            text
            @click="
              delDialog = false;
              initActionParam();
            "
            >取消</v-btn
          >
          <v-btn
            color="red"
            text
            @click="
              actionParam.key = 'delete';
              doAction();
            "
            >确定</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-overlay :value="overlay">
      <v-progress-circular indeterminate size="64"></v-progress-circular>
    </v-overlay>
  </div>
</template>

<script>
export default {
  data() {
    return {
      dialog: false,
      delDialog: false,
      overlay: false,
      navs: [
        {
          text: "主页",
          href: "/",
        },
        {
          text: "Playbook",
        },
      ],
      playbooks: [],
      actions: [
        {
          title: "运行",
          icon: "mdi-play",
          key: "run",
           color: "green",
        },
        {
          title: "编辑",
          icon: "mdi-square-edit-outline",
          key: "edit",
        },
        {
          title: "删除",
          icon: "mdi-trash-can-outline",
          color: "red",
          key: "confirmDelete",
        },
      ],
      actionParam: {},
    };
  },
  created() {
    this.initActionParam();
    this.overlay = true;
    this.list();
  },
  methods: {
    list: async function () {
      let _that = this;

      let playbooks = await listPlaybook();
      this.playbooks = playbooks;
      _that.overlay = false;
    },
    initActionParam: function () {
      this.actionParam = {
        args: {},
      };
    },
    doAction: async function () {
      let _that = this;
      switch (this.actionParam.key) {
        case "confirmDelete":
          _that.delDialog = true;
          break;
        case "delete":
          _that.delDialog = false;
          this.overlay = true;
          let err = await deletePlaybook(_that.actionParam.args.name);
          if (err) {
            alert(err);
          }
          this.overlay = false;
          _that.initActionParam();
          this.list()
          break;
        case "create":
          this.overlay = true;
          await createPlaybook(_that.actionParam.args.name); 
          this.dialog=false
          this.list()
          break;
        case "edit":
          editPlaybook(_that.actionParam.args.name); 
          break;
        case "run":
          err = await runPlaybook(_that.actionParam.args.name,"");
          if (err){
            alert(err)
            return
          }
         
          this.$router.push('/log');
          break;
      }
    },
  },
};
</script>

<style>
</style>