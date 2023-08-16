<template>
  <v-app>
    <v-app-bar color="primary"                prominent>
      <v-app-bar-nav-icon variant="text" @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
      <v-toolbar-title>{{ `${APP}:v${VERSION}`}}</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn variant="text" icon="mdi-magnify"></v-btn>
      <v-btn variant="text" icon="mdi-filter"></v-btn>
      <v-btn variant="text" icon="mdi-dots-vertical"></v-btn>
    </v-app-bar>
    <v-main>
      <template v-if="feedbackVisible">
        <v-alert closable
                 :type=feedbackType
                 :color=feedbackType
                 elevation="2"
                 :text="feedbackMsg"
                 >
        </v-alert>
      </template>
      <HelloWorld/>
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import {   onMounted, ref } from 'vue'
import type { Ref } from 'vue'
import {APP, getLog, DEV, BUILD_DATE, VERSION} from '@/config'
import HelloWorld from '@/components/HelloWorld.vue'

const log = getLog(APP, 4, 2);

type LevelAlert ='success' | 'info' | 'warning' | 'error' ;

const isNetworkOk = ref(true);
const drawer = ref(false);
const feedback = ref(null);
const feedbackMsg = ref(`${APP}, v.${VERSION}`);
const feedbackType: Ref<LevelAlert> = ref("info");
const feedbackVisible = ref(false);
const displayFeedBack = (text: string, type: LevelAlert= 'info') => {
  log.t(`displayFeedBack() text:'${text}' type:'${type}'`);
  feedbackType.value = type;
  feedbackMsg.value = text;
  feedbackVisible.value = true;
};
onMounted(() => {
  log.t('mounted()');
  log.w(`${APP} - ${VERSION}, du ${BUILD_DATE}`);

  window.addEventListener('online', () => {
    log.w('ONLINE AGAIN :)');
    isNetworkOk.value = true;
    displayFeedBack('⚡⚡🚀  CONNEXION RESEAU RETABLIE :  😊 vous êtes "ONLINE"  ', 'success');
  });
  window.addEventListener('offline', () => {
    log.w('OFFLINE :((');
    isNetworkOk.value = false;
    displayFeedBack('⚡⚡⚠ PAS DE RESEAU ! ☹ vous êtes "OFFLINE" ', 'error');
  });
});
</script>
