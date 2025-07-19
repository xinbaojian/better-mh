<script setup lang="ts">

import {reactive} from "vue";
import {Connect, Disconnect,Screenshot} from "../../wailsjs/go/adb/Adb";
import {ShowMessageDialog} from "../../wailsjs/go/message/Message";

const form = reactive({
  ip: '100.94.171.85',
  port: '16384',
  date: '',
})

const state = reactive({
  connected: false,
})

function connect() {
  Connect(form.ip, form.port).then((res) => {
    state.connected = res
    ShowMessageDialog("提示", res ? "连接成功" : "连接失败")
  })
}

function disconnect() {
  Disconnect(form.ip, form.port).then((res) => {
    state.connected = !res
    ShowMessageDialog("提示", res ? "断开成功" : "断开失败")
  })
}
</script>

<template>
  <div class="main">
    <el-form :inline="true" :model="form" class="demo-form-inline">
      <el-form-item label="IP">
        <el-input v-model="form.ip" placeholder="请输入IP" clearable/>
      </el-form-item>
      <el-form-item label="Port">
        <el-input v-model="form.port" placeholder="请输入端口" clearable/>
      </el-form-item>
      <el-form-item>
        <el-button v-if="!state.connected" type="primary" @click="connect">Connect</el-button>
        <el-button v-else type="danger" @click="disconnect">Disconnect</el-button>
        <el-button type="primary" @click="Screenshot('/Users/xinbaojian/Downloads')">Screenshot</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<style scoped lang="scss">
.main {
  width: 100%;
  height: 100%;
  text-align: center;
}

.demo-form-inline .el-input {
  --el-input-width: 220px;
}

.demo-form-inline .el-select {
  --el-select-width: 220px;
}
</style>
