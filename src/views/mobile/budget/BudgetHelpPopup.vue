<template>
    <f7-popup :opened="opened" tablet-fullscreen @popup:closed="emit('update:opened', false)">
        <f7-page>
            <f7-navbar :title="tt('How Budgeting Works')">
                <f7-nav-right>
                    <f7-link popup-close :text="tt('Close')" />
                </f7-nav-right>
            </f7-navbar>

            <template v-for="section in sections" :key="section.title">
                <f7-block-title>{{ section.title }}</f7-block-title>
                <f7-block strong inset>
                    <template v-for="(block, i) in section.blocks" :key="i">
                        <p v-if="block.type === 'p'" class="budget-help-p">{{ block.text }}</p>
                        <p v-else-if="block.type === 'note'" class="budget-help-note">{{ block.text }}</p>
                        <ul v-else-if="block.type === 'list'" class="budget-help-list">
                            <li v-for="(item, j) in block.items" :key="j">{{ item }}</li>
                        </ul>
                        <details v-else-if="block.type === 'example'" class="budget-help-example">
                            <summary>{{ block.title }}</summary>
                            <pre>{{ block.lines.join('\n') }}</pre>
                        </details>
                        <div v-else-if="block.type === 'table'" class="budget-help-table-wrap">
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
                </f7-block>
            </template>
        </f7-page>
    </f7-popup>
</template>

<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import { BUDGET_HELP_SECTIONS } from '@/views/base/BudgetHelpContent.ts';

defineProps<{ opened: boolean }>();
const emit = defineEmits<{ 'update:opened': [value: boolean] }>();

const { tt } = useI18n();
const sections = BUDGET_HELP_SECTIONS;
</script>

<style scoped>
.budget-help-p {
    font-size: 0.875rem;
    line-height: 1.5;
    margin: 0 0 10px;
}

.budget-help-p:last-child {
    margin-bottom: 0;
}

.budget-help-note {
    font-size: 0.8125rem;
    line-height: 1.5;
    margin: 0 0 10px;
    padding: 8px 10px;
    border-inline-start: 3px solid var(--f7-theme-color);
    background: var(--f7-list-border-color);
    border-radius: 4px;
}

.budget-help-list {
    font-size: 0.875rem;
    line-height: 1.5;
    margin: 0 0 10px;
    padding-inline-start: 20px;
}

.budget-help-example {
    margin: 0 0 10px;
    border: 1px solid var(--f7-list-border-color);
    border-radius: 6px;
    overflow: hidden;
}

.budget-help-example > summary {
    cursor: pointer;
    padding: 8px 10px;
    font-size: 0.8125rem;
    font-weight: 500;
    opacity: 0.75;
    user-select: none;
}

.budget-help-example > pre {
    margin: 0;
    padding: 10px;
    border-top: 1px solid var(--f7-list-border-color);
    font-size: 0.72rem;
    line-height: 1.55;
    overflow-x: auto;
    white-space: pre;
    font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.budget-help-table-wrap {
    overflow-x: auto;
    margin: 0 0 10px;
}

.budget-help-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
}

.budget-help-table th,
.budget-help-table td {
    text-align: start;
    padding: 6px 8px;
    border-bottom: 1px solid var(--f7-list-border-color);
    vertical-align: top;
}

.budget-help-table th {
    font-weight: 600;
    opacity: 0.6;
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
}
</style>
