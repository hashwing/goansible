<template>
  <v-container fluid class="px-0">
    <v-breadcrumbs :items="navs">
      <template v-slot:item="{ item }">
        <v-breadcrumbs-item :to="item.href" :disabled="item.disabled">
          {{ item.text.toUpperCase() }}
        </v-breadcrumbs-item>
      </template>
    </v-breadcrumbs>
    <div id="terminal1" class="xterm"></div>
  </v-container>
</template>

<script>
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";
export default {
  data() {
    return {
      log: "",
      navs: [
        {
          text: "主页",
          href: "/",
        },
        {
          text: "Playbook",
          href: "/",
        },
        {
          text: "日志",
        },
      ],
    };
  },
  mounted() {
    this.getLog();
  },

  methods: {
    getLog: function () {
      let _that = this;
      let i = 0;
      let term = new Terminal();
      let terminalContainer = document.getElementById("terminal1");
      const fitAddon = new FitAddon();
      term.loadAddon(fitAddon);
      term.open(terminalContainer);
      fitAddon.fit();
      term.focus();
      var t2 = window.setInterval(async function () {
        let s = await getLog(i);
        s.forEach((element) => {
          term.write(element + "\r\n");
        });
        i += s.length;
      }, 300);
    },
  },
};
</script>

<style>
</style>