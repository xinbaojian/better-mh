<script setup lang="ts">

import {nextTick, onMounted, reactive, ref} from "vue";
import {CaptureMat, CheckConnected, Connect, Disconnect} from "../../wailsjs/go/adb/Adb";
import {StartGame, StopGame, TestButton} from "../../wailsjs/go/main/Game";
import {ElMessage} from "element-plus";
import {EventsOn} from "../../wailsjs/runtime";

const form = reactive({
  ip: '100.94.171.85',
  port: '16384',
  // port: '16416',
  baoTu: false,
  waBaoTu: false,
  shimen: false,
  zhuoGui: false,
  yunBiao: false,
})

const state = reactive({
  connected: false,
  startGame: false,
})

const logs = ref<string[]>([])
const logContainer = ref<HTMLElement | null>(null)

function connect() {
  Connect(form.ip, form.port).then((res) => {
    state.connected = res
    if (res) {
      ElMessage.success("连接成功")
    } else {
      ElMessage.error("连接失败")
    }
  })
}

function disconnect() {
  Disconnect(form.ip, form.port).then((res) => {
    state.connected = !res
    if (res) {
      ElMessage.success("断开成功")
    } else {
      ElMessage.error("断开失败")
    }
  })
}

function startGameClick() {
  console.log("开始游戏", form)
  logs.value = []
  state.startGame = true
  StartGame(form.baoTu, form.waBaoTu, form.shimen, form.zhuoGui,form.yunBiao)
}

function stopGameClick() {
  StopGame()
  state.startGame = false
}

function clearLog() {
  logs.value = []
}

onMounted(() => {
  if (form.ip) {
    CheckConnected(form.ip, form.port).then((res: boolean) => {
      state.connected = res
    })
  }
  EventsOn("backendLog", (message: string) => {
    logs.value.push(message)
    if (message.includes("请先连接ADB")){
      state.connected = false
      state.startGame = false
    }
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  })
})
</script>

<template>
  <div class="main">
    <el-form :inline="true" :model="form" class="demo-form-inline">
      <el-row>
        <el-col :span="8">
          <el-form-item label="IP">
            <el-input v-model="form.ip" placeholder="请输入IP" clearable/>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="Port">
            <el-input v-model="form.port" placeholder="请输入端口" clearable/>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row>
        <el-col :span="24">
          <el-form-item>
            <el-checkbox v-model="form.baoTu">打宝图任务</el-checkbox>
            <el-checkbox v-model="form.waBaoTu">挖宝图任务</el-checkbox>
            <el-checkbox v-model="form.shimen">师门任务</el-checkbox>
            <el-checkbox v-model="form.zhuoGui">捉鬼任务</el-checkbox>
            <el-checkbox v-model="form.yunBiao">运镖任务</el-checkbox>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row>

        <el-col :span="12">
          <el-form-item>
            <el-button v-if="!state.connected" type="primary" @click="connect">Adb链接</el-button>
            <el-button v-else type="danger" @click="disconnect">断开ADB</el-button>
            <el-button v-if="!state.startGame" type="primary" @click="startGameClick">开始</el-button>
            <el-button v-else type="danger" @click="stopGameClick">停止</el-button>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item>
            <el-button type="primary" @click="TestButton">测试</el-button>
            <el-button type="primary" @click="CaptureMat">截图</el-button>
            <el-button type="danger" @click="clearLog">清空日志</el-button>
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <div class="log-container" ref="logContainer">
      <div v-for="log in logs" class="log-entry">{{ log }}</div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.main {
  width: 100%;
  height: 100vh;
  text-align: center;
  display: flex;
  flex-direction: column;
}

.el-form {
  padding-top: 20px;
  width: 100%;
}

.demo-form-inline .el-input {
  --el-input-width: 220px;
}

.demo-form-inline .el-select {
  --el-select-width: 220px;
}

.log-container {
  margin-top: 20px;
  flex: 1;
  background-color: #f0f0f0;
  border: 1px solid #ccc;
  overflow-y: auto;
  padding: 10px;
  text-align: left;
}

.log-entry {
  font-family: monospace;
  white-space: pre-wrap;
  word-break: break-all;
  margin-bottom: 5px;
}
</style>
