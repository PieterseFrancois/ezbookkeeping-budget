<template>
    <v-dialog :model-value="modelValue" max-width="720" scrollable
              @update:model-value="emit('update:modelValue', $event)">
        <v-card :title="tt('How Budgeting Works')">
            <v-card-text class="budget-help-body">
                <section v-for="section in sections" :key="section.title" class="budget-help-section">
                    <h3 class="budget-help-heading">{{ section.title }}</h3>
                    <template v-for="(block, i) in section.blocks" :key="i">
                        <p v-if="block.type === 'p'" class="text-body-2 mb-3">{{ block.text }}</p>
                        <div v-else-if="block.type === 'note'" class="budget-help-note text-body-2 mb-3">{{ block.text }}</div>
                        <ul v-else-if="block.type === 'list'" class="text-body-2 mb-3 ps-5">
                            <li v-for="(item, j) in block.items" :key="j">{{ item }}</li>
                        </ul>
                        <details v-else-if="block.type === 'example'" class="budget-help-example mb-3">
                            <summary>{{ block.title }}</summary>
                            <pre>{{ block.lines.join('\n') }}</pre>
                        </details>
                        <div v-else-if="block.type === 'table'" class="budget-help-table-wrap mb-3">
                            <table class="budget-help-table">
                                <thead>
                                    <tr><th v-for="(h, j) in block.head" :key="j">{{ h }}</th></tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(row, j) in block.rows" :key="j">
                                        <td v-for="(cell, k) in row" :key="k">{{ cell }}</td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </template>
                </section>
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn color="primary" @click="emit('update:modelValue', false)">{{ tt('Close') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import { BUDGET_HELP_SECTIONS } from '@/views/base/BudgetHelpContent.ts';

defineProps<{ modelValue: boolean }>();
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>();

const { tt } = useI18n();
const sections = BUDGET_HELP_SECTIONS;
</script>

<style scoped>
.budget-help-body {
    max-height: 70vh;
}

/* Each section is a visually distinct card so it is obvious where one ends */
.budget-help-section {
    padding: 16px 18px;
    margin-bottom: 16px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 8px;
    background: rgba(var(--v-theme-on-surface), 0.02);
}

.budget-help-section:last-child {
    margin-bottom: 0;
}

.budget-help-heading {
    font-size: 1rem;
    font-weight: 600;
    line-height: 1.3;
    margin: 0 0 12px;
    padding-bottom: 8px;
    border-bottom: 2px solid rgb(var(--v-theme-primary));
    display: inline-block;
}

.budget-help-section > *:last-child {
    margin-bottom: 0 !important;
}

.budget-help-note {
    padding: 8px 12px;
    border-inline-start: 3px solid rgb(var(--v-theme-primary));
    background: rgba(var(--v-theme-on-surface), 0.04);
    border-radius: 4px;
}

.budget-help-example {
    border: 1px solid rgba(var(--v-theme-on-surface), 0.15);
    border-radius: 6px;
    overflow: hidden;
}

.budget-help-example > summary {
    cursor: pointer;
    padding: 8px 12px;
    font-size: 0.8125rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.7);
    user-select: none;
}

.budget-help-example > pre {
    margin: 0;
    padding: 12px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    font-size: 0.75rem;
    line-height: 1.55;
    overflow-x: auto;
    white-space: pre;
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.budget-help-table-wrap {
    overflow-x: auto;
}

.budget-help-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
}

.budget-help-table th,
.budget-help-table td {
    text-align: start;
    padding: 6px 10px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.1);
    vertical-align: top;
}

.budget-help-table th {
    font-weight: 600;
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgba(var(--v-theme-on-surface), 0.55);
}
</style>
