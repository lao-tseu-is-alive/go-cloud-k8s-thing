<style></style>
<template>
  <v-container class="fill-height">
    <v-responsive class="d-flex align-center text-center fill-height">
      <v-row>
        <v-col cols="10">filtres:{{ propsValues }} ready:{{ areWeReady }} </v-col>
        <v-col cols="2">
          <template v-if="!areWeReady">
            <v-btn :loading="!areWeReady" class="flex-grow-1" height="48" variant="tonal"> chargement... </v-btn>
          </template>
        </v-col>
      </v-row>
      <v-row class="d-flex align-center justify-center">
        <v-col cols="12">
          <v-data-table :headers="headers as any" :items="records" :sort-by="[{ key: 'external_id', order: 'asc' }]" class="elevation-1">
            <template #top>
              <v-toolbar density="compact">
                <v-toolbar-title style="text-align: left">Liste de Thing...</v-toolbar-title>
                <v-spacer></v-spacer>
                <!-- BEGIN FORM EDIT  -->
                <v-dialog v-model="dialog" max-width="500px">
                  <template #activator="{ props }">
                    <v-btn color="primary" dark class="mb-2" v-bind="props"> Nouveau Thing</v-btn>
                  </template>
                  <v-card>
                    <v-card-title>
                      <span class="text-h5">{{ formTitle }}</span>
                    </v-card-title>

                    <v-card-text>
                      <v-container>
                        <v-row>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.name" label="Nom de l'objet"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.typeId" label="Type"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.validated" label="Valid"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.inactivated" label="Inactif"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.createdAt" label="Creation"></v-text-field>
                          </v-col>
                        </v-row>
                      </v-container>
                    </v-card-text>

                    <v-card-actions>
                      <v-spacer></v-spacer>
                      <v-btn color="blue-darken-1" variant="text" @click="close"> Cancel</v-btn>
                      <v-btn color="blue-darken-1" variant="text" @click="save"> Save</v-btn>
                    </v-card-actions>
                  </v-card>
                </v-dialog>
                <!-- END FORM EDIT  -->
                <v-dialog v-model="dialogDelete" max-width="500px">
                  <v-card>
                    <v-card-title class="text-h5">Are you sure you want to delete this item?</v-card-title>
                    <v-card-actions>
                      <v-spacer></v-spacer>
                      <v-btn color="blue-darken-1" variant="text" @click="closeDelete">Cancel</v-btn>
                      <v-btn color="blue-darken-1" variant="text" @click="deleteItemConfirm">OK</v-btn>
                      <v-spacer></v-spacer>
                    </v-card-actions>
                  </v-card>
                </v-dialog>
              </v-toolbar>
            </template>
            <template #item.inactivated="{ item }">
              <v-checkbox-btn v-model="item.columns.inactivated" :disabled="true"></v-checkbox-btn>
            </template>
            <template #item.validated="{ item }">
              <v-checkbox-btn v-model="item.columns.validated" :disabled="true"></v-checkbox-btn>
            </template>
            <template #item.actions="{ item }">
              <v-icon size="small" class="me-2" @click="editItem(item.raw)"> mdi-pencil</v-icon>
              <v-icon size="small" @click="deleteItem(item.raw)"> mdi-delete</v-icon>
            </template>
            <template #no-data>
              <v-alert type="warning">No data available</v-alert>
              <v-btn color="primary" @click="initialize"> Reset</v-btn>
            </template>
          </v-data-table>
        </v-col>
      </v-row>
    </v-responsive>
  </v-container>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed, nextTick, watch } from "vue"
import type { Ref } from "vue"
import { getLog, BACKEND_URL } from "@/config"
import { getLocalJwtTokenAuth } from "@/components/Login"
import { Configuration } from "@/typescript-axios-client-generated/configuration"
import { Thing } from "@/typescript-axios-client-generated/models/thing"
import { ThingList } from "@/typescript-axios-client-generated/models/thing-list"
import { DefaultApi } from "@/typescript-axios-client-generated/apis/default-api"
import axios from "axios"
import { VDataTable } from "vuetify/labs/VDataTable"
// import { ThingStatus } from "@/typescript-axios-client-generated/models/thing-status"

const log = getLog("ThingListVue", 4, 2)
const areWeReady = ref(false)
let myApi: DefaultApi
const dialog = ref(false)
const dialogDelete = ref(false)
const headers = [
  {
    title: "id (external)",
    sortable: true,
    key: "external_id",
  },
  {
    title: "Nom",
    align: "start",
    sortable: true,
    key: "name",
  },
  { title: "Type", key: "type_id" },
  { title: "Inactif", key: "inactivated" },
  { title: "Valide", key: "validated" },
  { title: "Created", key: "created_at" },
  { title: "Actions", key: "actions", sortable: false },
]

const records: ThingList[] = reactive([])
/*const defaultListItem: Ref<ThingList> = ref({
  id: crypto.randomUUID(),
  typeId: 0,
  name: "",
  description: undefined,
  externalId: undefined,
  inactivated: false,
  validated: undefined,
  status: undefined,
  createdBy: 0,
  createdAt: undefined,
  posX: 0,
  posY: 0,
})*/
const editedIndex = ref(-1)
const defaultItem: Ref<Thing> = ref({
  id: crypto.randomUUID(),
  typeId: 0,
  name: "",
  description: undefined,
  comment: undefined,
  externalId: undefined,
  externalRef: undefined,
  buildAt: undefined,
  status: undefined,
  containedBy: undefined,
  containedByOld: undefined,
  inactivated: false,
  inactivatedTime: undefined,
  inactivatedBy: undefined,
  inactivatedReason: undefined,
  validated: undefined,
  validatedTime: undefined,
  validatedBy: undefined,
  managedBy: undefined,
  createdAt: undefined,
  createdBy: 0,
  lastModifiedAt: undefined,
  lastModifiedBy: undefined,
  deleted: false,
  deletedAt: undefined,
  deletedBy: undefined,
  moreData: undefined,
  posX: 0,
  posY: 0,
})
const editedItem: Ref<Thing> = ref(Object.assign({}, defaultItem))

const myProps = defineProps<{
  typeThing?: number | undefined
  createdBy?: number | undefined
  inactivated?: boolean | undefined
  validated?: boolean | undefined
  limit?: number | undefined
  offset?: number | undefined
}>()

//// WATCH SECTION
watch(
  () => myProps.typeThing,
  (val, oldValue) => {
    log.t(` watch myProps.typeThing old: ${oldValue}, new val: ${val}`)
    if (val !== undefined && areWeReady.value == true) {
      if (val !== oldValue) {
        retrieveList(val, myProps.createdBy)
      }
    }
  }
  //  { immediate: true }
)
watch(
  () => myProps.createdBy,
  (val, oldValue) => {
    log.t(` watch myProps.createdBy old: ${oldValue}, new val: ${val}`)
    if (val !== undefined && areWeReady.value == true) {
      if (val !== oldValue) {
        retrieveList(myProps.typeThing, val)
      }
    }
  }
  //  { immediate: true }
)

watch(
  () => myProps.limit,
  (val, oldValue) => {
    log.t(` watch myProps.limit old: ${oldValue}, new val: ${val}`)
    if (val !== undefined && areWeReady.value == true) {
      if (val !== oldValue) {
        retrieveList(myProps.typeThing, myProps.createdBy)
      }
    }
  }
  //  { immediate: true }
)

watch(dialog, (val, oldValue) => {
  log.t(` watch dialog old: ${oldValue}, new val: ${val}`)
  val || close()
})
watch(dialogDelete, (val, oldValue) => {
  log.t(` watch dialogDelete old: ${oldValue}, new val: ${val}`)
  val || closeDelete()
})
//// COMPUTED SECTION
const formTitle = computed(() => {
  return editedIndex.value === -1 ? "New Item" : "Edit Item"
})

const propsValues = computed(() => {
  return JSON.stringify(myProps, undefined, 3)
})

//// FUNCTIONS SECTION
const editItem = (item: Thing) => {
  editedIndex.value = records.indexOf(item)
  editedItem.value = Object.assign({}, item)
  dialog.value = true
}

const deleteItem = (item: Thing) => {
  editedIndex.value = records.indexOf(item)
  editedItem.value = Object.assign({}, item)
  dialogDelete.value = true
}

const deleteItemConfirm = () => {
  records.splice(editedIndex.value, 1)
  closeDelete()
}

const close = () => {
  dialog.value = false
  nextTick(() => {
    editedItem.value = Object.assign({}, defaultItem.value)
    editedIndex.value = -1
  })
}

const closeDelete = () => {
  dialogDelete.value = false
  nextTick(() => {
    editedItem.value = Object.assign({}, defaultItem.value)
    editedIndex.value = -1
  })
}

const save = () => {
  if (editedIndex.value > -1) {
    Object.assign(records[editedIndex.value], editedItem.value)
  } else {
    records.push(editedItem.value)
  }
  close()
}

const retrieveList = (typeThing?: number, createdBy?: number) => {
  log.t(`> Entering.. typeThing: ${typeThing}, createdBy: ${createdBy} `)
  areWeReady.value = false
  if (typeThing != undefined) {
    typeThing = typeThing == 0 ? undefined : typeThing
  }
  if (createdBy != undefined) {
    createdBy = createdBy == 0 ? undefined : createdBy
  }
  log.t(`After adjusting typeThing: ${typeThing}, createdBy: ${createdBy} `)
  try {
    myApi
      .list(typeThing, createdBy, myProps.inactivated, myProps.validated, myProps.limit, myProps.offset)
      .then((resp) => {
        log.l("myAPi.list : ", resp)
        if (resp.status == 200) {
          if (records.length > 0) {
            records.splice(0)
          }
          resp.data.forEach((r) => {
            records.push(r)
          })
          areWeReady.value = true
        } else {
          log.w("retrieveList got problem", resp)
        }
      })
      .catch((err) => {
        log.w("# retrieveList in catch ERROR err: ", err)
        if (records.length > 0) {
          records.splice(0)
        }
        areWeReady.value = true
      })
  } catch (error) {
    if (records.length > 0) {
      records.splice(0)
    }
    areWeReady.value = true
    if (axios.isAxiosError(error)) {
      log.w(`Try Catch Axios ERROR message:${error.message}, error:`, error)
      log.l("Axios error.response:", error.response)
    } else {
      log.e("unexpected error: ", error)
    }
  }
}

const initialize = () => {
  const token = getLocalJwtTokenAuth()
  const myConf = new Configuration({ accessToken: token, basePath: BACKEND_URL + "/goapi/v1" })
  myApi = new DefaultApi(myConf)
  areWeReady.value = true
  retrieveList(myProps.typeThing, myProps.createdBy)
}

onMounted(() => {
  log.t("mounted()")
  initialize()
})
</script>
