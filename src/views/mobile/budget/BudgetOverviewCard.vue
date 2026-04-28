<template>
    <f7-card class="budget-ov-card" :class="{ 'skeleton-text': loading && !hasAnyData }">
        <f7-card-header>
            <span>{{ tt('Budget Overview') }}</span>
            <f7-link href="/budget" class="budget-ov-view-all">{{ tt('View All') }}</f7-link>
        </f7-card-header>

        <f7-card-content :padding="false">
            <div v-if="loading && !hasAnyData" class="budget-ov-skeleton">
                <div v-for="i in 3" :key="i" class="budget-ov-skeleton-row skeleton-effect-wave"></div>
            </div>

            <div v-else-if="!loading && !hasAnyData" class="budget-ov-empty">
                <span class="text-color-gray">{{ tt('No budget targets for this month') }}</span>
            </div>

            <template v-else>
                <!-- Column headers -->
                <div class="budget-ov-header-row">
                    <span class="budget-ov-icon-placeholder"></span>
                    <span class="budget-ov-name-cell"></span>
                    <span class="budget-ov-amt-cell">{{ tt('Budgeted') }}</span>
                    <span class="budget-ov-amt-cell">{{ tt('Remaining') }}</span>
                </div>

                <!-- Budget rows -->
                <div v-for="item in budgetSummary" :key="item.categoryName" class="budget-ov-row">
                    <div class="budget-ov-icon-circle" :style="{ backgroundColor: `#${item.color || 'c67e48'}` }">
                        <ItemIcon icon-type="category" :icon-id="item.icon || '1'" default-color="#ffffff" />
                    </div>
                    <span class="budget-ov-name-cell">{{ item.categoryName }}</span>
                    <span class="budget-ov-amt-cell">{{ displayAmount(item.budgeted) }}</span>
                    <span
                        class="budget-ov-amt-cell budget-ov-remaining"
                        :class="item.isSavings ? (item.remaining > 0 ? 'text-color-red' : item.remaining < 0 ? 'text-color-green' : '') : (item.remaining < 0 ? 'text-color-red' : 'text-color-green')"
                    >{{ displayAmount(item.remaining) }}</span>
                </div>

                <!-- Unbudgeted section -->
                <template v-if="unbudgeted.length > 0">
                    <div class="budget-ov-section-divider"></div>
                    <div class="budget-ov-unbudgeted-toggle" @click="showUnbudgeted = !showUnbudgeted">
                        <span>{{ tt('Unbudgeted Categories') }}</span>
                        <f7-icon :f7="showUnbudgeted ? 'chevron_up' : 'chevron_down'" size="13"></f7-icon>
                    </div>
                    <template v-if="showUnbudgeted">
                        <div v-for="item in unbudgeted" :key="item.categoryName" class="budget-ov-row budget-ov-unbudgeted-row">
                            <div class="budget-ov-icon-circle" :style="{ backgroundColor: `#${item.color || 'c67e48'}` }">
                                <ItemIcon icon-type="category" :icon-id="item.icon || '1'" default-color="#ffffff" />
                            </div>
                            <span class="budget-ov-name-cell">{{ item.categoryName }}</span>
                            <span class="budget-ov-amt-cell"></span>
                            <span class="budget-ov-amt-cell">{{ displayAmount(item.spent) }}</span>
                        </div>
                    </template>
                </template>
            </template>
        </f7-card-content>
    </f7-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import ItemIcon from '@/components/mobile/ItemIcon.vue';

import { useI18n } from '@/locales/helpers.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';

export interface BudgetSummaryItem {
    categoryName: string;
    icon: string;
    color: string;
    budgeted: number;
    spent: number;
    remaining: number;
    isSavings?: boolean;
}

export interface UnbudgetedItem {
    categoryName: string;
    icon: string;
    color: string;
    spent: number;
}

const props = defineProps<{
    loading: boolean;
    budgetSummary: BudgetSummaryItem[];
    unbudgeted: UnbudgetedItem[];
}>();

const { tt, formatAmountToLocalizedNumeralsWithCurrency } = useI18n();
const settingsStore = useSettingsStore();
const userStore = useUserStore();

const showAmountInHomePage = computed<boolean>(() => settingsStore.appSettings.showAmountInHomePage);
const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);
const hasAnyData = computed<boolean>(() => props.budgetSummary && props.budgetSummary.length > 0);

const showUnbudgeted = ref<boolean>(false);

function displayAmount(amount: number): string {
    if (!showAmountInHomePage.value) {
        return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, defaultCurrency.value);
    }
    const rands = Math.round(amount / 100) * 100;
    return formatAmountToLocalizedNumeralsWithCurrency(rands, defaultCurrency.value)
        .replace(/[,.]00$/, '');
}
</script>

<style>
.budget-ov-card .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.budget-ov-view-all {
    font-size: 13px;
}

.budget-ov-skeleton {
    padding: 8px 16px 12px;
}

.budget-ov-skeleton-row {
    height: 32px;
    border-radius: 16px;
    background: var(--f7-list-border-color);
    margin-bottom: 10px;
}

.budget-ov-empty {
    padding: 20px 16px;
    text-align: center;
    font-size: 14px;
}

.budget-ov-header-row,
.budget-ov-row {
    display: flex;
    align-items: center;
    padding: 8px 16px;
}

.budget-ov-header-row {
    padding-bottom: 4px;
    font-size: 11px;
    opacity: 0.55;
    border-bottom: 1px solid var(--f7-list-border-color);
    margin-bottom: 2px;
}

.budget-ov-icon-placeholder {
    width: 36px;
    flex-shrink: 0;
}

.budget-ov-icon-circle {
    width: 30px;
    height: 30px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    margin-inline-end: 6px;
}

.budget-ov-icon-circle i.icon {
    color: #fff !important;
    font-size: 15px;
    width: 15px;
    height: 15px;
    line-height: 15px;
}

.budget-ov-name-cell {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    padding-inline-end: 6px;
    font-size: 13px;
}

.budget-ov-amt-cell {
    width: 76px;
    flex-shrink: 0;
    text-align: right;
    font-size: 12px;
    font-variant-numeric: tabular-nums;
}

.budget-ov-remaining {
    font-weight: 500;
}

.budget-ov-section-divider {
    height: 1px;
    background: var(--f7-list-border-color);
    margin: 6px 16px;
}

.budget-ov-unbudgeted-toggle {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 16px;
    font-size: 12px;
    opacity: 0.6;
    cursor: pointer;
}

.budget-ov-unbudgeted-row {
    opacity: 0.7;
}

.budget-ov-card .card-content {
    padding-bottom: 10px;
}
</style>
