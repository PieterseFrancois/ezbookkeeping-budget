<template>
    <v-card :class="{ 'disabled': loading }">
        <template #title>
            <span>{{ tt('Budget Overview') }}</span>
        </template>

        <v-card-text v-if="loading && !hasAnyData">
            <div v-for="i in 5" :key="i" class="mb-2">
                <v-skeleton-loader width="100%" type="text" :loading="true"/>
            </div>
        </v-card-text>

        <v-card-text v-else-if="!loading && !hasAnyData">
            <div class="d-flex align-center justify-center py-4">
                <span class="text-body-1 text-medium-emphasis">{{ tt('No budget targets for this month') }}</span>
            </div>
        </v-card-text>

        <v-card-text v-else>
            <div class="d-flex align-center pb-2 text-caption text-medium-emphasis">
                <span class="budget-name-col"></span>
                <span class="budget-amount-col text-end">{{ tt('Budgeted') }}</span>
                <span class="budget-amount-col text-end">{{ tt('Spent') }}</span>
                <span class="budget-amount-col text-end">{{ tt('Remaining') }}</span>
            </div>

            <div v-for="item in budgetSummary" :key="item.categoryName"
                 class="d-flex align-center py-1">
                <span class="budget-name-col text-truncate">{{ item.categoryName }}</span>
                <span class="budget-amount-col text-end text-body-2">{{ displayAmount(item.budgeted) }}</span>
                <span class="budget-amount-col text-end text-body-2">{{ displayAmount(item.spent) }}</span>
                <span class="budget-amount-col text-end text-body-2"
                      :class="{ 'text-error': item.remaining < 0 }">{{ displayAmount(item.remaining) }}</span>
            </div>

            <template v-if="unbudgeted.length > 0">
                <v-divider class="my-3"/>
                <div class="d-flex align-center justify-space-between py-1 cursor-pointer"
                     @click="showUnbudgeted = !showUnbudgeted">
                    <span class="text-body-2 text-medium-emphasis">{{ tt('Unbudgeted Categories') }}</span>
                    <v-icon :icon="showUnbudgeted ? mdiChevronUp : mdiChevronDown" size="18"/>
                </div>
                <template v-if="showUnbudgeted">
                    <div v-for="item in unbudgeted" :key="item.categoryName"
                         class="d-flex align-center justify-space-between py-1 ps-2">
                        <span class="text-body-2 text-medium-emphasis text-truncate me-4">{{ item.categoryName }}</span>
                        <span class="text-body-2 text-medium-emphasis">{{ displayAmount(item.spent) }}</span>
                    </div>
                </template>
            </template>
        </v-card-text>
    </v-card>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { mdiChevronUp, mdiChevronDown } from '@mdi/js';

import { useI18n } from '@/locales/helpers.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { useUserStore } from '@/stores/user.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';

export interface BudgetSummaryItem {
    categoryName: string;
    budgeted: number;
    spent: number;
    remaining: number;
}

export interface UnbudgetedItem {
    categoryName: string;
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
    return formatAmountToLocalizedNumeralsWithCurrency(amount, defaultCurrency.value);
}
</script>

<style>
.budget-name-col {
    flex: 1;
    min-width: 0;
    margin-inline-end: 8px;
}

.budget-amount-col {
    width: 90px;
    flex-shrink: 0;
}
</style>
