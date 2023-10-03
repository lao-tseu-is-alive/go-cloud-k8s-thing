<style></style>
<template>
  <v-container class="fill-height">
    <v-responsive class="d-flex fill-height">
      <v-row>
        <!-- BEGIN FORM EDIT  -->
        <v-dialog v-model="dialog" :persistent="true" transition="dialog-top-transition" width="960">
          <v-card>
            <v-toolbar color="primary">
              <v-toolbar-title>
                {{ formTitle }} <span class="ml-3 text-body-1">id: {{ editedItem.id }}</span>
              </v-toolbar-title>
              <template #extension>
                <v-tabs v-model="dialogTab" color="primary-accent" bg-color="secondary" centered density="compact">
                  <v-tab value="tab-info-base">
                    <v-icon>mdi-information-outline</v-icon>
                    Infos principales
                  </v-tab>
                  <v-tab value="tab-info-more">
                    <v-icon>mdi-more</v-icon>
                    Infos détaillées
                  </v-tab>
                  <v-tab value="tab-info-record">
                    <v-icon>mdi-account-details</v-icon>
                    historique enregistrement
                  </v-tab>
                </v-tabs>
              </template>
            </v-toolbar>
            <v-card-text>
              <v-container class="v-container--fluid ma-0">
                <v-window v-model="dialogTab">
                  <v-window-item value="tab-info-base">
                    <v-row>
                      <v-col class="d-none d-sm-flex" cols="0" sm="2" md="2" lg="1" xl="1">
                        <v-text-field v-model="editedItem.type_id" density="compact" label="id-type"></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="10" md="4" lg="3" xl="2">
                        <v-select
                          v-model="editedItem.type_id"
                          item-title="name"
                          item-value="id"
                          :items="arrListTypeThing"
                          density="compact"
                          label="TypeObjet*"
                        ></v-select>
                      </v-col>
                      <v-col cols="12" sm="12" md="6" lg="8">
                        <v-text-field
                          v-model="editedItem.name"
                          density="compact"
                          required
                          label="Nom de l'objet*"
                        ></v-text-field>
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
                        <v-text-field
                          type="number"
                          v-model="editedItem.external_id"
                          density="compact"
                          label="identifiant externe"
                        />
                      </v-col>
                      <v-col cols="12" sm="4" md="3" lg="3">
                        <v-text-field v-model="editedItem.external_ref" density="compact" label="référence externe" />
                      </v-col>
                      <v-col cols="12" sm="4" md="3" lg="3">
                        <v-menu
                          v-model="menuDateConstruction"
                          :close-on-content-click="false"
                          :nudge-right="40"
                          transition="scale-transition"
                          offset-y
                          min-width="auto"
                          location="end"
                        >
                          <template #activator="{ props }">
                            <v-text-field
                              v-model="editedItem.build_at"
                              prepend-icon="mdi-calendar"
                              density="compact"
                              label="Date construction"
                              v-bind="props"
                              @click="menuDateConstruction = true"
                            />
                          </template>
                          <v-locale-provider locale="fr">
                            <v-date-picker
                              v-model="editedItem.build_at"
                              cancel-text="ANNULER"
                              header="Choisissez une date SVP"
                              title="Date de construction"
                              show-adjacent-months
                              show-week
                              color="primary"
                              elevation="24"
                              position="relative"
                              input-mode="calendar"
                              @click:save="editDateBuild(false)"
                              @click:cancel="editDateBuild(true)"
                            >
                            </v-date-picker>
                          </v-locale-provider>
                        </v-menu>
                      </v-col>
                      <v-col cols="12" sm="4" md="3" lg="3">
                        <v-text-field v-model="editedItem.status" density="compact" label="Etat de l'objet" />
                      </v-col>
                    </v-row>
                  </v-window-item>
                  <v-window-item value="tab-info-more">
                    <v-row>
                      <v-col cols="12" sm="6" md="4">
                        <v-text-field
                          v-model="editedItem.contained_by"
                          density="compact"
                          label="Contenu dans"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="4">
                        <v-text-field
                          v-model="editedItem.contained_by_old"
                          density="compact"
                          label="Contenu dans(ancien)"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="4">
                        <v-text-field
                          v-model="editedItem.managed_by"
                          density="compact"
                          label="Managé par"
                        ></v-text-field>
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col cols="6" sm="4" md="2">
                        <v-checkbox v-model="editedItem.validated" density="compact" label="Validé?" />
                      </v-col>
                      <v-col cols="12" sm="6" md="5">
                        <v-text-field
                          v-model="editedItem.validated_time"
                          label="Date de validation"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="5">
                        <v-text-field
                          v-model="editedItem.validated_by"
                          label="Validé par"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="6" sm="4" md="2" lg="2" xl="1">
                        <v-checkbox v-model="editedItem.inactivated" density="compact" label="Inactif?" />
                      </v-col>
                      <v-col cols="12" sm="6" md="3">
                        <v-text-field
                          v-model="editedItem.inactivated_time"
                          label="Date d'inactivation"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="3">
                        <v-text-field
                          v-model="editedItem.inactivated_by"
                          label="Inactivé par"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="4">
                        <v-text-field
                          v-model="editedItem.inactivated_reason"
                          label="Raison de l'inactivation"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                    </v-row>
                  </v-window-item>
                  <v-window-item value="tab-info-record">
                    <v-row>
                      <v-col cols="6" sm="4" md="4" lg="2" xl="1">
                        <v-checkbox v-model="editedItem.deleted" density="compact" label="Effacé?" />
                      </v-col>
                      <v-col cols="12" sm="6" md="4">
                        <v-text-field
                          v-model="editedItem.deleted_at"
                          label="Date d'effacement"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="4">
                        <v-text-field
                          v-model="editedItem.deleted_by"
                          label="Effacé par"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col cols="12" sm="6" md="3">
                        <v-text-field
                          v-model="editedItem.created_at"
                          label="Date de Création"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="3">
                        <v-text-field
                          v-model="editedItem.created_by"
                          label="Création par"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="3">
                        <v-text-field
                          v-model="editedItem.last_modified_at"
                          label="Date de modification"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="3">
                        <v-text-field
                          v-model="editedItem.last_modified_by"
                          label="Modification par"
                          density="compact"
                          :disabled="true"
                        ></v-text-field>
                      </v-col>
                    </v-row>
                  </v-window-item>
                </v-window>
              </v-container>
            </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn dark color="primary" variant="flat" @click="close">Annuler</v-btn>
              <v-btn dark color="primary" variant="flat" @click="save">Sauver</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <!-- END FORM EDIT  -->
        <!-- BEGIN FORM DELETE  -->
        <v-dialog v-model="dialogDelete" :persistent="true" transition="dialog-top-transition" width="560">
          <v-card>
            <v-card-title class="text-h5">Voulez-vous vraiment effacer ?</v-card-title>
            <v-card-text> {{ deletedItem.external_id }} : {{ deletedItem.name }} </v-card-text>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn dark color="primary" variant="flat" @click="closeDelete">Annuler</v-btn>
              <v-btn dark color="primary" variant="flat" @click="deleteItemConfirm">Sauver</v-btn>
              <v-spacer></v-spacer>
            </v-card-actions>
          </v-card>
        </v-dialog>
        <!-- END FORM DELETE  -->
      </v-row>
      <v-row>
        <v-col cols="10"
          >Trouvé {{ numThingsFound }} Thing(s) avec ces filtres:{{ propsValues }} ready:{{ areWeReady }}
        </v-col>
        <v-col cols="2">
          <template v-if="!areWeReady">
            <v-btn :loading="!areWeReady" class="flex-grow-1" height="48" variant="tonal"> chargement... </v-btn>
          </template>
        </v-col>
      </v-row>
      <v-row class="d-flex align-center justify-center">
        <v-col cols="12">
          <v-data-table
            :headers="getHeaderVTable as any"
            :items="records"
            :sort-by="[{ key: 'external_id', order: 'asc' }]"
            class="elevation-1"
          >
            <template #top>
              <v-toolbar density="compact">
                <v-toolbar-title style="text-align: left">{{ numThingsFound }} Thing trouvés...</v-toolbar-title>
                <v-spacer></v-spacer>
                <v-btn
                  dark
                  color="primary"
                  variant="flat"
                  prepend-icon="mdi-creation"
                  density="default"
                  class="m-2"
                  @click="newThing"
                >
                  Nouveau Thing</v-btn
                >
              </v-toolbar>
            </template>
            <template #item.type_id="{ item }">
              <v-label> {{ getTypeThingName(item.columns.type_id) }}</v-label>
            </template>
            <template #item.inactivated="{ item }">
              <v-checkbox-btn
                v-model="item.columns.inactivated"
                :disabled="true"
                class="d-none d-lg-block"
              ></v-checkbox-btn>
            </template>
            <template #item.validated="{ item }">
              <v-checkbox-btn
                v-model="item.columns.validated"
                :disabled="true"
                class="d-none d-lg-block"
              ></v-checkbox-btn>
            </template>
            <template #item.created_at="{ item }">
              <v-label> {{ getDateFromTimeStamp(item.columns.created_at) }}</v-label>
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
import { useDisplay } from "vuetify"
import { getLog, BACKEND_URL, defaultAxiosTimeout } from "@/config"
import { getDateFromTimeStamp, getDateIsoFromTimeStamp, isNullOrUndefined } from "@/tools/utils"
import { getLocalJwtTokenAuth, getSessionId, getUserId } from "@/components/Login"
import { Configuration } from "@/openapi-generator-cli_thing_typescript-axios"
import { DefaultApi, Thing, ThingList } from "@/openapi-generator-cli_thing_typescript-axios"
import axios, { AxiosInstance, CreateAxiosDefaults } from "axios"
import { VDataTable } from "vuetify/labs/VDataTable"
// import { ThingStatus } from "@/typescript-axios-client-generated/models/thing-status"

const log = getLog("ThingListVue", 4, 2)
const displaySize = reactive(useDisplay())
const areWeReady = ref(false)
let myApi: DefaultApi
let myAxios: AxiosInstance
const dialog = ref(false)
const dialogDelete = ref(false)
const menuDateConstruction = ref(false)

interface typeThingSelect {
  id: number
  name: string
}
interface IDictionary {
  [key: number]: string
}
const dialogTab = ref(null)
let dicoTypeThing: IDictionary = {}
const arrListTypeThing: typeThingSelect[] = []
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
const editedItem: Ref<Thing> = ref(defaultItem)
const deletedItem: Ref<ThingList> = ref(Object.assign({}, defaultListItem))
const numThingsFound = ref(0)
const myProps = defineProps<{
  typeThing?: number | undefined
  createdBy?: number | undefined
  inactivated?: boolean | undefined
  validated?: boolean | undefined
  limit?: number | undefined
  offset?: number | undefined
}>()

//// EVENT SECTION

const emit = defineEmits(["thing-ok", "thing-error"])

//// WATCH SECTION
watch(
  () => myProps.typeThing,
  (val, oldValue) => {
    log.t(` watch myProps.typeThing old: ${oldValue}, new val: ${val}`)
    if (val !== undefined && areWeReady.value == true) {
      if (val !== oldValue) {
        listThing(val, myProps.createdBy)
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
        listThing(myProps.typeThing, val)
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
        listThing(myProps.typeThing, myProps.createdBy)
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
        listThing(myProps.typeThing, myProps.createdBy)
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
        listThing(myProps.typeThing, myProps.createdBy)
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
  return editedIndex.value === -1 ? "Nouvel objet" : "Edition de l'objet"
})

const propsValues = computed(() => {
  return JSON.stringify(myProps, undefined, 3)
})

// responsive header
const getHeaderVTable = computed(() => {
  log.t(`#> Entering... ${displaySize.name}`)
  if (displaySize.name === "lg") {
    return [
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
  } else {
    return [
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
      { title: "Created", key: "created_at" },
      { title: "Actions", key: "actions", sortable: false },
    ]
  }
})

//// FUNCTIONS SECTION
const editItem = async (item: ThingList) => {
  log.t(" #> entering EDIT ...", item)
  editedIndex.value = records.indexOf(item)
  const id = item.id
  const res = await getThing(id)
  if (res.data !== null) {
    log.l(`ok, filling editedItem with Thing id : ${res.data.id}`)
    editedItem.value = Object.assign({}, res.data)
    if (editedItem.value.build_at !== undefined) {
      if (editedItem.value.build_at.indexOf("T") > 0) {
        editedItem.value.build_at = editedItem.value.build_at.split("T")[0]
      }
    }
    dialog.value = true
    log.l(`Now inside editing id : ${res.data.id}`)
  } else {
    log.w(`problem retrieving getThing(${id})`, res.err)
  }
}

const editDateBuild = (cancel: boolean) => {
  log.t(`#> entering ... cancel=${cancel}`)
  if (cancel !== true) {
    if (editedItem.value.build_at !== undefined) {
      if (editedItem.value.build_at.indexOf("T") > 0) {
        editedItem.value.build_at = editedItem.value.build_at.split("T")[0]
      } else {
        editedItem.value.build_at = new Date(editedItem.value.build_at).toISOString().split("T")[0]
      }
    }
  }
  log.t(`<# exit ... build_at=${editedItem.value.build_at}`)
  menuDateConstruction.value = false
}

const newThing = () => {
  log.t("#> entering ...")
  editedItem.value = Object.assign({}, defaultItem.value)
  editedItem.value.id = crypto.randomUUID()
  editedItem.value.created_by = getUserId()
  const justNow = new Date()
  editedItem.value.created_at = justNow.toISOString()
  dialog.value = true
}

const deleteItem = (item: ThingList) => {
  log.t(" #> entering DELETE ...", item)
  editedIndex.value = records.indexOf(item)
  deletedItem.value = Object.assign({}, item)
  dialogDelete.value = true
}

const deleteItemConfirm = async () => {
  const id = deletedItem.value.id
  const res = await deleteThing(id)
  if (res.err === null) {
    log.l(`ok, doing deletedItem(${id}) `)
    records.splice(editedIndex.value, 1)
  } else {
    const msg = `problem doing deletedItem(${id}) ${res.err.message}`
    log.w(msg)
    emit("thing-error", msg)
  }
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
    deletedItem.value = Object.assign({}, defaultListItem.value)
    editedIndex.value = -1
  })
}

const save = async () => {
  log.t(" #> entering SAVE ...")
  if (editedItem.value.external_id !== undefined) {
    editedItem.value.external_id = +editedItem.value.external_id
  }
  if (editedIndex.value > -1) {
    //// HANDLING UPDATE OF EXISTING ITEM
    Object.assign(records[editedIndex.value], editedItem.value)
    log.l(`build_at : ${editedItem.value.build_at}`)
    if (editedItem.value.build_at != undefined) {
      const tmpDate = new Date(editedItem.value.build_at).toISOString()
      log.l(`tmpDate : ${tmpDate}`)
      editedItem.value.build_at = tmpDate
    }
    log.l(`build_at : ${editedItem.value.build_at}`)
    const res = await updateThing(editedItem.value.id, editedItem.value)
    if (res.data === null) {
      const msg = `Save update failed. Problem:  ${res.err?.message}`
      log.w(msg)
      emit("thing-error", msg)
    } else {
      Object.assign(records[editedIndex.value], res.data)
      const msg = `Vos modifications ont été enregistrées dans la Base avec succès.`
      log.w(msg)
      emit("thing-ok", msg)
    }
  } else {
    //// HANDLING CREATE OF NEW ITEM
    const res = await createThing(editedItem.value.id, editedItem.value)
    if (res.data === null) {
      const msg = `Save createThing failed. Problem:  ${res.err?.message}`
      log.w(msg)
      emit("thing-error", msg)
    } else {
      Object.assign(editedItem.value, res.data)
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
      //reset of editedItem is done in close()
      const msg = `Nouvel enregistrement sauvé dans la Base  id: ${res.data?.external_id}`
      log.w(msg)
      emit("thing-ok", msg)
    }
  }
  close()
}
const clearRecords = (): void => {
  if (records.length > 0) {
    records.splice(0)
  }
}

type netThing = { data: Thing | null; err: Error | null }

const getThing = async (id: string): Promise<netThing> => {
  log.t(`> Entering.. getThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myApi.get(id)
    log.l(`SUCCESS myAPi.get(id:${resp.data.id}`)
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
      if (error.response !== undefined && error.response.data !== undefined) {
        const srvMessage = isNullOrUndefined(error.response.data.message) ? "" : error.response.data.message
        const msg = `getThing error : ${error.message}. Server says : ${srvMessage}`
        log.w(msg)
        emit("thing-error", msg)
        return { data: null, err: Error(msg) }
      } else {
        return { data: null, err: Error(`getThing error : ${error.message}.`) }
      }
    } else {
      log.e("unexpected error: ", error)
      return { data: null, err: Error(`unexpected error: in getThing Try catch : ${error}`) }
    }
  }
}

/**
 * listThing a list of things from server based on current criteria
 * @param typeThing filter the list with only this type of thing
 * @param createdBy filter the list with only records created by given user id
 */
const listThing = async (typeThing?: number, createdBy?: number) => {
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
    const resp = await myApi.list(
      typeThing,
      createdBy,
      myProps.inactivated,
      myProps.validated,
      myProps.limit,
      myProps.offset
    )
    clearRecords()
    resp.data.forEach((r) => {
      records.push(r)
    })
    numThingsFound.value = await countThing(undefined, typeThing, createdBy)
    areWeReady.value = true
  } catch (error) {
    clearRecords()
    numThingsFound.value = await countThing(undefined, typeThing, createdBy)
    areWeReady.value = true
    if (axios.isAxiosError(error)) {
      if (error.response !== undefined) {
        log.w(`getThing error : ${error.message}. Server says : ${error.response.data}`)
      } else {
        log.w(`getThing error : ${error.message}.`)
      }
      numThingsFound.value = await countThing(undefined, typeThing, createdBy)
    } else {
      log.e("unexpected error: ", error)
    }
  }
}

const createThing = async (id: string, t: Thing): Promise<netThing> => {
  log.t(`> Entering.. createThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myAxios.post("thing", t)
    log.l("myAxios.post(thing) : ", resp)
    areWeReady.value = true
    return { data: resp.data, err: null }
  } catch (err) {
    areWeReady.value = true
    if (axios.isAxiosError(err)) {
      log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
      if (err.response !== undefined && err.response.data !== undefined) {
        const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
        return { data: null, err: Error(`createThing error : ${err.message}. Server says : ${srvMessage}`) }
      } else {
        return { data: null, err: Error(`createThing error : ${err.message}.`) }
      }
    } else {
      log.e("💥💥 unexpected error: ", err)
      return { data: null, err: Error(`💥💥 createThing error: in deleteThing Try catch : ${err}`) }
    }
  }
}
const updateThing = async (id: string, t: Thing): Promise<netThing> => {
  log.t(`> Entering.. updateThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myAxios.put("thing/" + id, t)
    log.l("myAxios.put(thing/id) : ", resp)
    areWeReady.value = true
    return { data: resp.data, err: null }
  } catch (err) {
    areWeReady.value = true
    if (axios.isAxiosError(err)) {
      log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
      if (err.response !== undefined && err.response.data !== undefined) {
        const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
        return { data: null, err: Error(`updateThing error : ${err.message}. Server says : ${srvMessage}`) }
      } else {
        return { data: null, err: Error(`updateThing error : ${err.message}.`) }
      }
    } else {
      log.e("💥💥 unexpected error: ", err)
      return { data: null, err: Error(`💥💥 updateThing error: in deleteThing Try catch : ${err}`) }
    }
  }
}

const countThing = async (keywords?: string, typeThing?: number, createdBy?: number): Promise<number> => {
  log.t(`> Entering.. typeThing: ${typeThing}, createdBy: ${createdBy} `)
  if (typeThing != undefined) {
    typeThing = typeThing == 0 ? undefined : typeThing
  }
  if (createdBy != undefined) {
    createdBy = createdBy == 0 ? undefined : createdBy
  }
  log.t(`After adjusting typeThing: ${typeThing}, createdBy: ${createdBy} `)
  try {
    const resp = await myApi.count(keywords, typeThing, createdBy, myProps.inactivated, myProps.validated)
    return resp.data
  } catch (error) {
    if (axios.isAxiosError(error)) {
      if (error.response !== undefined) {
        log.w(`countThing error : ${error.message}. Server says : ${error.response.data}`)
      } else {
        log.w(`countThing error : ${error.message}.`)
      }
    } else {
      log.e("unexpected error: ", error)
    }
    return 0
  }
}

const deleteThing = async (id: string): Promise<netThing> => {
  log.t(`> Entering.. deleteThing: ${id}`)
  areWeReady.value = false
  try {
    const resp = await myApi._delete(id)
    log.l("myAPi._delete : ", resp)
    areWeReady.value = true
    return { data: null, err: null }
  } catch (err) {
    areWeReady.value = true
    if (axios.isAxiosError(err)) {
      log.w(`Try Catch Axios ERROR message:${err.message}, error:`, err)
      if (err.response !== undefined && err.response.data !== undefined) {
        const srvMessage = isNullOrUndefined(err.response.data.message) ? "" : err.response.data.message
        return { data: null, err: Error(`deleteThing error : ${err.message}. Server says : ${srvMessage}`) }
      } else {
        return { data: null, err: Error(`deleteThing error : ${err.message}.`) }
      }
    } else {
      log.e("💥💥 unexpected error: ", err)
      return { data: null, err: Error(`💥💥 unexpected error: in deleteThing Try catch : ${err}`) }
    }
  }
}

const getTypeThingName = (id: number): string => {
  if (id in dicoTypeThing) {
    return dicoTypeThing[id]
  }
  return "# type inconnu #"
}

const loadTypeThingData = async (): Promise<void> => {
  const resp = await myApi.typeThingList(undefined, undefined, undefined, undefined, 300, 0)
  log.l("myAPi.typeThingList : ", resp)
  resp.data.forEach((r) => {
    const temp = { id: r.id, name: r.name }
    arrListTypeThing.push(temp)
  })
  dicoTypeThing = Object.fromEntries(arrListTypeThing.map((x) => [x.id, x.name]))
}

const initialize = async () => {
  const token = getLocalJwtTokenAuth()
  const myConf = new Configuration({
    accessToken: token,
    baseOptions: { timeout: defaultAxiosTimeout, headers: { "X-Goeland-Token": getSessionId() } },
    basePath: BACKEND_URL + "/goapi/v1",
  })
  myApi = new DefaultApi(myConf)
  myAxios = axios.create({
    baseURL: BACKEND_URL + "/goapi/v1",
    timeout: defaultAxiosTimeout,
    headers: { "X-Goeland-Token": getSessionId(), Authorization: `Bearer ${getLocalJwtTokenAuth()}` },
  } as CreateAxiosDefaults)
  areWeReady.value = true

  await loadTypeThingData()
  await listThing(myProps.typeThing, myProps.createdBy)
}

onMounted(() => {
  log.t("mounted()")
  initialize()
})
</script>
