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
                <v-dialog v-model="dialog">
                  <template #activator="{ props }">
                    <v-btn color="primary" dark class="mb-2" v-bind="props"> Nouveau Thing</v-btn>
                  </template>
                  <v-card>
                    <v-card-title>
                      <span class="text-h5">{{ formTitle }}</span>
                      <span class="ml-3 text-sm-caption">id: {{ editedItem.id }}</span>
                    </v-card-title>
                    <!--

    more_data: undefined,
  pos_x: 0,
  pos_y: 0,
-->
                    <v-card-text>
                      <v-container class="v-container--fluid ma-0">
                        <v-row>
                          <v-col class="d-none d-sm-flex" cols="0" sm="2" md="2" lg="1" xl="1">
                            <v-text-field v-model="editedItem.type_id" density="compact" :disabled="true" label="id-type"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="10" md="4" lg="3" xl="2">
                            <v-select v-model="editedItem.type_id" item-title="name" item-value="id" :items="arrListTypeThing" density="compact" label="TypeObjet*"></v-select>
                          </v-col>
                          <v-col cols="12" sm="12" md="6" lg="8">
                            <v-text-field v-model="editedItem.name" density="compact" label="Nom de l'objet*"></v-text-field>
                          </v-col>
                          <v-col cols="12">
                            <v-text-field v-model="editedItem.description" label="Description"></v-text-field>
                          </v-col>
                          <v-col cols="12">
                            <v-textarea
                              v-model="editedItem.comment"
                              rows="2"
                              row-height="15"
                              row
                              density="compact"
                              auto-grow
                              bg-color="amber-lighten-4"
                              color="orange orange-darken-4"
                              label="Commentaire"
                            ></v-textarea>
                          </v-col>
                          <v-col cols="12" sm="4" md="3" lg="3">
                            <v-text-field type="number" v-model="editedItem.external_id" density="compact" label="identifiant externe" />
                          </v-col>
                          <v-col cols="12" sm="4" md="3" lg="3">
                            <v-text-field v-model="editedItem.external_ref" density="compact" label="référence externe" />
                          </v-col>
                          <v-col cols="12" sm="4" md="3" lg="3">
                            <v-text-field v-model="editedItem.build_at" density="compact" label="Date construction" />
                          </v-col>
                          <v-col cols="12" sm="4" md="3" lg="3">
                            <v-text-field v-model="editedItem.status" density="compact" label="Etat de l'objet" />
                          </v-col>
                        </v-row>
                        <v-row>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.contained_by" density="compact" label="Contenu dans"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.contained_by_old" density="compact" label="Contenu dans(ancien)"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.managed_by" density="compact" label="Managé par"></v-text-field>
                          </v-col>
                        </v-row>
                        <v-row>
                          <v-col cols="6" sm="4" md="2" >
                            <v-checkbox v-model="editedItem.validated" density="compact" label="Validé?" />
                          </v-col>
                          <v-col cols="12" sm="6" md="5">
                            <v-text-field v-model="editedItem.validated_time" label="Date de validation" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="5">
                            <v-text-field v-model="editedItem.validated_by" label="Validé par" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="6" sm="4" md="2" lg="2" xl="1">
                            <v-checkbox v-model="editedItem.inactivated" density="compact" label="Inactif?" />
                          </v-col>
                          <v-col cols="12" sm="6" md="3">
                            <v-text-field v-model="editedItem.inactivated_time" label="Date d'inactivation" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="3">
                            <v-text-field v-model="editedItem.inactivated_by" label="Inactivé par" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.inactivated_reason" label="Raison de l'inactivation" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                        </v-row>
                        <v-row>
                          <v-col cols="6" sm="4" md="4" lg="2" xl="1">
                            <v-checkbox v-model="editedItem.deleted" density="compact" label="Effacé?" />
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.deleted_at" label="Date d'effacement" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="4">
                            <v-text-field v-model="editedItem.deleted_by" label="Effacé par" density="compact" :disabled="true"></v-text-field>
                          </v-col>

                          <v-col cols="12" sm="6" md="3">
                            <v-text-field v-model="editedItem.created_at" label="Date de Création" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="3">
                            <v-text-field v-model="editedItem.created_by" label="Création par" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="3">
                            <v-text-field v-model="editedItem.last_modified_at" label="Date de modification" density="compact" :disabled="true"></v-text-field>
                          </v-col>
                          <v-col cols="12" sm="6" md="3">
                            <v-text-field v-model="editedItem.last_modified_by" label="Modification par" density="compact" :disabled="true"></v-text-field>
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
//import { Configuration } from "@/typescript-axios-client-generated/configuration"
//import { Thing } from "@/typescript-axios-client-generated/models/thing2"
//import { ThingList } from "@/typescript-axios-client-generated/models/thing-list"
import { Configuration } from "../openapi-generator-cli_thing_typescript-axios/configuration"
import { DefaultApi, Thing, ThingList } from "../openapi-generator-cli_thing_typescript-axios/api"
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

interface typeThingSelect {
  id: number
  name: string
}

const arrListTypeThing: typeThingSelect[] = reactive([])
const records: ThingList[] = reactive([])
const defaultListItem: Ref<ThingList> = ref({
  id: crypto.randomUUID(),
  type_id: 0,
  name: "",
  description: undefined,
  external_id: undefined,
  inactivated: false,
  validated: undefined,
  status: undefined,
  created_by: 0,
  created_at: undefined,
  pos_x: 0,
  pos_y: 0,
})
const editedIndex = ref(-1)
const defaultItem: Ref<Thing> = ref({
  id: crypto.randomUUID(),
  type_id: 0,
  name: "",
  description: undefined,
  comment: undefined,
  external_id: undefined,
  external_ref: undefined,
  build_at: undefined,
  status: undefined,
  contained_by: undefined,
  contained_by_old: undefined,
  inactivated: false,
  inactivated_timeTime: undefined,
  inactivated_by: undefined,
  inactivated_reason: undefined,
  validated: undefined,
  validated_time: undefined,
  validated_by: undefined,
  managed_by: undefined,
  created_at: undefined,
  created_by: 0,
  last_modified_at: undefined,
  last_modified_by: undefined,
  deleted: false,
  deleted_at: undefined,
  deleted_by: undefined,
  more_data: undefined,
  pos_x: 0,
  pos_y: 0,
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
)
watch(
  () => myProps.validated,
  (val, oldValue) => {
    log.t(` watch myProps.validated old: ${oldValue}, new val: ${val}`)
    if (areWeReady.value == true) {
      if (val !== oldValue) {
        retrieveList(myProps.typeThing, myProps.createdBy)
      }
    }
  }
)

watch(
  () => myProps.inactivated,
  (val, oldValue) => {
    log.t(` watch myProps.inactivated old: ${oldValue}, new val: ${val}`)
    if (areWeReady.value == true) {
      if (val !== oldValue) {
        retrieveList(myProps.typeThing, myProps.createdBy)
      }
    }
  }
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
const editItem = async (item: ThingList) => {
  log.t(" #> entering EDIT ...", item)
  editedIndex.value = records.indexOf(item)
  const id = item.id
  const res = await getThing(id)
  if (res.data !== null) {
    log.l("ok, preparing editedItem with Thing data ", res.data)
    editedItem.value = Object.assign({}, res.data)
    dialog.value = true
  } else {
    log.w(`problem retrieving getThing(${id})`, res.err)
  }
}

const deleteItem = (item: ThingList) => {
  log.t(" #> entering DELETE ...", item)
  editedIndex.value = records.indexOf(item)
  //editedItem.value = Object.assign({}, item)
  dialogDelete.value = true
}

const deleteItemConfirm = () => {
  records.splice(editedIndex.value, 1)
  closeDelete()
}

const close = () => {
  log.t(" #> entering CLOSE ...")
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
  log.t(" #> entering SAVE ...")
  if (editedIndex.value > -1) {
    Object.assign(records[editedIndex.value], editedItem.value)
  } else {
    const newItem = Object.assign({}, defaultListItem.value)
    newItem.id = editedItem.value.id
    newItem.type_id = editedItem.value.type_id
    newItem.name = editedItem.value.name
    newItem.description = editedItem.value.description
    newItem.external_id = editedItem.value.external_id
    newItem.inactivated = editedItem.value.inactivated
    newItem.validated = editedItem.value.validated
    newItem.status = editedItem.value.status
    newItem.created_by = editedItem.value.created_by
    newItem.created_at = editedItem.value.created_at
    newItem.pos_x = editedItem.value.pos_x
    newItem.pos_y = editedItem.value.pos_y
    records.push(newItem)
  }
  close()
}

type netThing = { data: Thing | null; err: Error | null }

const getThing = async (id: string): Promise<netThing> => {
  log.t(`> Entering.. getThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myApi.get(id)
    log.l("myAPi.get : ", resp)
    if (resp.status == 200) {
      areWeReady.value = true
      return { data: resp.data, err: null }
    } else {
      areWeReady.value = true
      log.w("getThing got problem", resp)
      return { data: null, err: Error(`problem in getThing status : ${resp.status}, ${resp.statusText}`) }
    }
  } catch (error) {
    areWeReady.value = true
    if (axios.isAxiosError(error)) {
      log.w(`Try Catch Axios ERROR message:${error.message}, error:`, error)
      log.l("Axios error.response:", error.response)
      return { data: null, err: Error(`Axios error in getThing Try catch : ${error.message}`) }
    } else {
      log.e("unexpected error: ", error)
      return { data: null, err: Error(`unexpected error: in getThing Try catch : ${error}`) }
    }
  }
}

const clearRecords = (): void => {
  if (records.length > 0) {
    records.splice(0)
  }
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
          clearRecords()
          resp.data.forEach((r) => {
            records.push(r)
          })
          areWeReady.value = true
        } else {
          areWeReady.value = true
          log.w("retrieveList got problem", resp)
        }
      })
      .catch((err) => {
        log.w("# retrieveList in catch ERROR err: ", err)
        clearRecords()
        areWeReady.value = true
      })
  } catch (error) {
    clearRecords()
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
  myApi.typeThingList(undefined, undefined, undefined, undefined, 300, 0).then((resp) => {
    log.l("myAPi.typeThingList : ", resp)
    if (resp.status == 200) {
      resp.data.forEach((r) => {
        const temp = { id: r.id, name: r.name }
        arrListTypeThing.push(temp)
      })
    } else {
      //display alert with status code > 200
    }
  })
}

onMounted(() => {
  log.t("mounted()")
  initialize()
})
</script>
