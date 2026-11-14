<script setup lang="ts">
import { onMounted, ref } from "vue";

type Note = { id: number; text: string };
const status = ref("...");
const notes = ref<Note[]>([]);
const text = ref("");

async function load() {
  notes.value = await (await fetch("/api/notes")).json();
}
async function add() {
  await fetch("/api/notes", { method: "POST", body: JSON.stringify({ text: text.value }) });
  text.value = "";
  await load();
}
onMounted(async () => {
  status.value = (await (await fetch("/api/ping")).json()).status;
  await load();
});
</script>

<template>
  <main>
    <h1>SQLite + Vue spike</h1>
    <p>Backend ping: <strong>{{ status }}</strong></p>
    <form @submit.prevent="add">
      <input v-model="text" placeholder="note" />
      <button>Add</button>
    </form>
    <ul><li v-for="n in notes" :key="n.id">{{ n.text }}</li></ul>
  </main>
</template>
