<template>
    <v-row class="match-height">
        <v-col cols="12" lg="4" md="12">
            <v-card :class="{ 'disabled': loadingOverview }">
                <template #title>
                    <div class="d-flex align-center">
                        <div class="d-flex align-baseline">
                            <span class="text-2xl font-weight-bold">{{ displayDateRange?.thisMonth?.displayTime }}</span>
                            <span>·</span>
                            <span style="font-size: 1rem">{{ tt('Expense') }}</span>
                        </div>
                        <v-btn density="compact" color="default" variant="text" size="24"
                               class="ms-2" :icon="true" :loading="loadingOverview" @click="reload(true)">
                            <template #loader>
                                <v-progress-circular indeterminate size="20"/>
                            </template>
                            <v-icon :icon="mdiRefresh" size="24" />
                            <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
                        </v-btn>
                    </div>
                </template>

                <v-card-text>
                    <h4 class="text-2xl font-weight-medium text-primary">
                        <span v-if="!loadingOverview || (transactionOverview && transactionOverview.thisMonth && transactionOverview.thisMonth.valid)">{{ transactionOverview && transactionOverview.thisMonth ? getDisplayExpenseAmount(transactionOverview.thisMonth) : '-' }}</span>
                        <v-skeleton-loader class="d-inline-block skeleton-no-margin mt-3 pb-1" width="120px" type="text" :loading="true" v-else-if="loadingOverview && (!transactionOverview || !transactionOverview.thisMonth || !transactionOverview.thisMonth.valid)"></v-skeleton-loader>
                        <v-btn class="ms-1" density="compact" color="default" variant="text"
                               :icon="true" @click="showAmountInHomePage = !showAmountInHomePage">
                            <v-icon :icon="showAmountInHomePage ? mdiEyeOffOutline : mdiEyeOutline" size="20" />
                        </v-btn>
                    </h4>
                    <div class="mt-1 mb-3">
                        <span class="me-2">{{ tt('Monthly income') }}</span>
                        <span v-if="!loadingOverview || (transactionOverview && transactionOverview.thisMonth && transactionOverview.thisMonth.valid)">{{ transactionOverview && transactionOverview.thisMonth ? getDisplayIncomeAmount(transactionOverview.thisMonth) : '-' }}</span>
                        <v-skeleton-loader class="d-inline-block skeleton-no-margin mt-2" width="120px" type="text" :loading="true" v-else-if="loadingOverview && (!transactionOverview || !transactionOverview.thisMonth || !transactionOverview.thisMonth.valid)"></v-skeleton-loader>
                    </div>
                    <v-btn size="small" :to="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.ThisMonth.type })}`">{{ tt('View Details') }}</v-btn>
                    <v-img class="overview-card-background img-with-direction"
                           src="img/desktop/card-background.png"/>
                    <v-img class="overview-card-background-image img-with-direction"
                           width="116px" src="img/desktop/document.svg"/>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" lg="8" md="12">
            <v-card :class="{ 'disabled': loadingOverview }">
                <template #title>
                    <span>{{ tt('Asset Summary') }}</span>
                </template>

                <v-card-text>
                    <div class="mb-8">
                        <span class="text-body-1" v-if="!loadingOverview || (allAccounts && allAccounts.length)">{{ tt('format.misc.youHaveAccounts', { count: displayAccountCount }) }}</span>
                        <v-skeleton-loader class="skeleton-no-margin mt-1 mb-2 pb-1" width="200px" type="text" :loading="true" v-else-if="loadingOverview && (!allAccounts || !allAccounts.length)"></v-skeleton-loader>
                    </div>

                    <v-row>
                        <v-col cols="12" md="4">
                            <div class="d-flex align-center">
                                <div class="me-3">
                                    <v-avatar rounded color="grey" size="42" class="elevation-1">
                                        <v-icon size="24" :icon="mdiBankOutline"/>
                                    </v-avatar>
                                </div>

                                <div class="d-flex flex-column">
                                    <span class="text-caption">{{ tt('Total assets') }}</span>
                                    <span class="text-h5" v-if="!loadingOverview || (allAccounts && allAccounts.length)">{{ totalAssets }}</span>
                                    <v-skeleton-loader class="skeleton-no-margin mt-3 mb-2" width="120px" type="text" :loading="true" v-else-if="loadingOverview && (!allAccounts || !allAccounts.length)"></v-skeleton-loader>
                                </div>
                            </div>
                        </v-col>

                        <v-col cols="12" md="4">
                            <div class="d-flex align-center">
                                <div class="me-3">
                                    <v-avatar rounded color="expense" size="42" class="elevation-1">
                                        <v-icon size="24" :icon="mdiCreditCardOutline"/>
                                    </v-avatar>
                                </div>

                                <div class="d-flex flex-column">
                                    <span class="text-caption">{{ tt('Total liabilities') }}</span>
                                    <span class="text-h5" v-if="!loadingOverview || (allAccounts && allAccounts.length)">{{ totalLiabilities }}</span>
                                    <v-skeleton-loader class="skeleton-no-margin mt-3 mb-2" width="120px" type="text" :loading="true" v-else-if="loadingOverview && (!allAccounts || !allAccounts.length)"></v-skeleton-loader>
                                </div>
                            </div>
                        </v-col>

                        <v-col cols="12" md="4">
                            <div class="d-flex align-center">
                                <div class="me-3">
                                    <v-avatar rounded color="primary" size="42" class="elevation-1">
                                        <v-icon size="24" :icon="mdiPiggyBankOutline"/>
                                    </v-avatar>
                                </div>

                                <div class="d-flex flex-column">
                                    <span class="text-caption">{{ tt('Net assets') }}</span>
                                    <span class="text-h5" v-if="!loadingOverview || (allAccounts && allAccounts.length)">{{ netAssets }}</span>
                                    <v-skeleton-loader class="skeleton-no-margin mt-3 mb-2" width="120px" type="text" :loading="true" v-else-if="loadingOverview && (!allAccounts || !allAccounts.length)"></v-skeleton-loader>
                                </div>
                            </div>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" md="6">
            <v-row>
                <v-col cols="12" md="6">
                    <income-expense-overview-card
                        :loading="loadingOverview" :disabled="loadingOverview" :icon="mdiCalendarTodayOutline"
                        :title="tt('Today')"
                        :expense-amount="transactionOverview.today && transactionOverview.today.valid ? getDisplayExpenseAmount(transactionOverview.today) : ''"
                        :income-amount="transactionOverview.today && transactionOverview.today.valid ? getDisplayIncomeAmount(transactionOverview.today) : ''"
                        :datetime="displayDateRange?.today?.displayTime || ''"
                    >
                        <template #menus>
                            <v-list-item :prepend-icon="mdiListBoxOutline" :to="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.Today.type })}`">
                                <v-list-item-title>{{ tt('View Details') }}</v-list-item-title>
                            </v-list-item>
                        </template>
                    </income-expense-overview-card>
                </v-col>

                <v-col cols="12" md="6">
                    <income-expense-overview-card
                        :loading="loadingOverview" :disabled="loadingOverview" :icon="mdiCalendarWeekOutline"
                        :title="tt('This Week')"
                        :expense-amount="transactionOverview.thisWeek && transactionOverview.thisWeek.valid ? getDisplayExpenseAmount(transactionOverview.thisWeek) : ''"
                        :income-amount="transactionOverview.thisWeek && transactionOverview.thisWeek.valid ? getDisplayIncomeAmount(transactionOverview.thisWeek) : ''"
                        :datetime="displayDateRange?.thisWeek?.startTime + '-' + displayDateRange?.thisWeek?.endTime"
                    >
                        <template #menus>
                            <v-list-item :prepend-icon="mdiListBoxOutline" :to="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.ThisWeek.type })}`">
                                <v-list-item-title>{{ tt('View Details') }}</v-list-item-title>
                            </v-list-item>
                        </template>
                    </income-expense-overview-card>
                </v-col>

                <v-col cols="12" md="6">
                    <income-expense-overview-card
                        :loading="loadingOverview" :disabled="loadingOverview" :icon="mdiCalendarMonthOutline"
                        :title="tt('This Month')"
                        :expense-amount="transactionOverview.thisMonth && transactionOverview.thisMonth.valid ? getDisplayExpenseAmount(transactionOverview.thisMonth) : ''"
                        :income-amount="transactionOverview.thisMonth && transactionOverview.thisMonth.valid ? getDisplayIncomeAmount(transactionOverview.thisMonth) : ''"
                        :datetime="displayDateRange?.thisMonth?.startTime + '-' + displayDateRange?.thisMonth?.endTime"
                    >
                        <template #menus>
                            <v-list-item :prepend-icon="mdiListBoxOutline" :to="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.ThisMonth.type })}`">
                                <v-list-item-title>{{ tt('View Details') }}</v-list-item-title>
                            </v-list-item>
                        </template>
                    </income-expense-overview-card>
                </v-col>

                <v-col cols="12" md="6">
                    <income-expense-overview-card
                        :loading="loadingOverview" :disabled="loadingOverview" :icon="mdiLayersTripleOutline"
                        :title="tt('This Year')"
                        :expense-amount="transactionOverview.thisYear && transactionOverview.thisYear.valid ? getDisplayExpenseAmount(transactionOverview.thisYear) : ''"
                        :income-amount="transactionOverview.thisYear && transactionOverview.thisYear.valid ? getDisplayIncomeAmount(transactionOverview.thisYear) : ''"
                        :datetime="displayDateRange?.thisYear?.displayTime || ''"
                    >
                        <template #menus>
                            <v-list-item :prepend-icon="mdiListBoxOutline" :to="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.ThisYear.type })}`">
                                <v-list-item-title>{{ tt('View Details') }}</v-list-item-title>
                            </v-list-item>
                        </template>
                    </income-expense-overview-card>
                </v-col>
            </v-row>
        </v-col>

        <v-col cols="12" md="6">
            <monthly-income-and-expense-card :data="monthlyIncomeAndExpenseData" :is-dark-mode="isDarkMode"
                                             :loading="loadingOverview" :disabled="loadingOverview"
                                             :enable-click-item="true" @click="clickMonthlyIncomeOrExpense" />
        </v-col>

        <v-col cols="12">
            <budget-overview-card :loading="loadingOverview" :budget-summary="budgetSummary" :unbudgeted="unbudgeted"/>
        </v-col>
    </v-row>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';
import IncomeExpenseOverviewCard from './overview/cards/IncomeExpenseOverviewCard.vue';
import MonthlyIncomeAndExpenseCard, { type MonthlyIncomeAndExpenseCardClickEvent } from './overview/cards/MonthlyIncomeAndExpenseCard.vue';
import BudgetOverviewCard, { type BudgetSummaryItem, type UnbudgetedItem } from './overview/cards/BudgetOverviewCard.vue';

import { ref, computed, useTemplateRef } from 'vue';
import { useRouter } from 'vue-router';
import { useTheme } from 'vuetify';

import { useI18n } from '@/locales/helpers.ts';
import { useHomePageBase } from '@/views/base/HomePageBase.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useOverviewStore } from '@/stores/overview.ts';

import { DateRange } from '@/core/datetime.ts';
import { CategoryType } from '@/core/category.ts';
import { ThemeType } from '@/core/theme.ts';
import {
    type TransactionMonthlyIncomeAndExpenseData,
    LATEST_12MONTHS_TRANSACTION_AMOUNTS_REQUEST_TYPES
} from '@/models/transaction.ts';

import { getUnixTimeBeforeUnixTime, getUnixTimeAfterUnixTime, getThisMonthFirstUnixTime, getThisMonthLastUnixTime } from '@/lib/datetime.ts';
import axios from 'axios';
import type { ApiResponse } from '@/core/api.ts';
import services from '@/lib/services.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';

import {
    mdiRefresh,
    mdiEyeOutline,
    mdiEyeOffOutline,
    mdiBankOutline,
    mdiCreditCardOutline,
    mdiPiggyBankOutline,
    mdiCalendarTodayOutline,
    mdiCalendarWeekOutline,
    mdiCalendarMonthOutline,
    mdiLayersTripleOutline,
    mdiListBoxOutline
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;

const router = useRouter();
const theme = useTheme();

const { tt, formatNumberToLocalizedNumerals } = useI18n();
const {
    showAmountInHomePage,
    allAccounts,
    netAssets,
    totalAssets,
    totalLiabilities,
    displayDateRange,
    transactionOverview,
    getDisplayIncomeAmount,
    getDisplayExpenseAmount
} = useHomePageBase();

const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const overviewStore = useOverviewStore();

const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loadingOverview = ref<boolean>(true);
const budgetSummary = ref<BudgetSummaryItem[]>([]);
const unbudgeted = ref<UnbudgetedItem[]>([]);

interface BudgetTargetRawItem {
    id: string;
    categoryId: string;
    year: number;
    month: number;
    amount: string;
}

const isDarkMode = computed<boolean>(() => theme.global.name.value === ThemeType.Dark);

const displayAccountCount = computed<string>(() => formatNumberToLocalizedNumerals(allAccounts.value?.length ?? 0));

function clickMonthlyIncomeOrExpense(e: MonthlyIncomeAndExpenseCardClickEvent): void {
    const minTime = e.monthStartTime;
    const maxTime = getUnixTimeBeforeUnixTime(getUnixTimeAfterUnixTime(minTime, 1, 'months'), 1, 'seconds');
    const type = e.transactionType;

    router.push(`/transaction/list?${overviewStore.getTransactionListPageParams({
        type: type,
        dateType: DateRange.Custom.type,
        minTime: minTime,
        maxTime: maxTime
    })}`);
}

const monthlyIncomeAndExpenseData = computed<TransactionMonthlyIncomeAndExpenseData[]>(() => {
    const data: TransactionMonthlyIncomeAndExpenseData[] = [];

    if (!transactionOverview.value || !transactionOverview.value.thisMonth || !transactionOverview.value.thisMonth.valid) {
        return data;
    }

    for (const amountRequestType of LATEST_12MONTHS_TRANSACTION_AMOUNTS_REQUEST_TYPES) {
        const dateRange = overviewStore.transactionDataRange[amountRequestType];

        if (!dateRange) {
            continue;
        }

        const item = transactionOverview.value[amountRequestType];

        data.push({
            monthStartTime: dateRange.startTime,
            incomeAmount: item?.incomeAmount || 0,
            expenseAmount: item?.expenseAmount || 0,
            incompleteIncomeAmount: item ? item.incompleteIncomeAmount : true,
            incompleteExpenseAmount: item ? item.incompleteExpenseAmount : true
        });
    }

    return data;
});

async function loadBudgetOverview(): Promise<void> {
    const now = new Date();
    const year = now.getFullYear();
    const month = now.getMonth() + 1;
    const startTime = getThisMonthFirstUnixTime();
    const endTime = getThisMonthLastUnixTime();

    const [budgetResp, statsResp] = await Promise.all([
        axios.get<ApiResponse<BudgetTargetRawItem[]>>(`v1/budget/targets.json?year=${year}&month=${month}`),
        services.getTransactionStatistics({ startTime, endTime, tagFilter: '', keyword: '', useTransactionTimezone: false })
    ]);

    const targets = budgetResp.data?.result ?? [];
    const statsItems = statsResp.data?.result?.items ?? [];

    // subcategoryId -> amount spent this month (expense categories only)
    const spentBySubcategoryId: Record<string, number> = {};
    for (const item of statsItems) {
        const cat = transactionCategoriesStore.allTransactionCategoriesMap[item.categoryId];
        if (cat && cat.type === CategoryType.Expense) {
            spentBySubcategoryId[item.categoryId] = (spentBySubcategoryId[item.categoryId] ?? 0) + item.amount;
        }
    }

    // subcategoryId -> budgeted amount
    const budgetedSubcategoryIds = new Set<string>();
    for (const target of targets) {
        budgetedSubcategoryIds.add(target.categoryId);
    }

    // Group budget targets by parent category; only parents with ≥1 budgeted subcategory are included
    const parentGroups: Record<string, { name: string; icon: string; color: string; totalBudgeted: number; totalSpent: number }> = {};

    for (const target of targets) {
        const subCat = transactionCategoriesStore.allTransactionCategoriesMap[target.categoryId];
        if (!subCat || !subCat.parentId || subCat.parentId === '0') continue;

        const parentId = subCat.parentId;
        const parentCat = transactionCategoriesStore.allTransactionCategoriesMap[parentId];
        if (!parentCat) continue;

        const group = parentGroups[parentId] ?? (parentGroups[parentId] = { name: parentCat.name, icon: parentCat.icon, color: parentCat.color, totalBudgeted: 0, totalSpent: 0 });
        group.totalBudgeted += Number(target.amount);
    }

    // Sum spending across ALL subcategories of each budgeted parent (not just the budgeted subs)
    for (const [parentId, group] of Object.entries(parentGroups)) {
        const parentCat = transactionCategoriesStore.allTransactionCategoriesMap[parentId];
        for (const subCat of parentCat?.subCategories ?? []) {
            group.totalSpent += spentBySubcategoryId[subCat.id] ?? 0;
        }
    }

    const summary: BudgetSummaryItem[] = Object.values(parentGroups).map(g => ({
        categoryName: g.name,
        iconId: g.icon,
        color: g.color,
        budgeted: g.totalBudgeted,
        spent: g.totalSpent,
        remaining: g.totalBudgeted - g.totalSpent
    }));

    // Subcategories with spending but no budget target → unbudgeted list
    const unbudgetedList: UnbudgetedItem[] = [];
    for (const [subcatId, spent] of Object.entries(spentBySubcategoryId)) {
        if (spent <= 0 || budgetedSubcategoryIds.has(subcatId)) continue;

        const subCat = transactionCategoriesStore.allTransactionCategoriesMap[subcatId];
        if (!subCat) continue;

        const parentCat = subCat.parentId && subCat.parentId !== '0'
            ? transactionCategoriesStore.allTransactionCategoriesMap[subCat.parentId]
            : undefined;

        const categoryName = parentCat ? `${parentCat.name} > ${subCat.name}` : subCat.name;
        const iconSource = parentCat ?? subCat;
        unbudgetedList.push({ categoryName, iconId: iconSource.icon, color: iconSource.color, spent });
    }

    budgetSummary.value = summary;
    unbudgeted.value = unbudgetedList;
}

function reload(force: boolean): void {
    loadingOverview.value = true;

    const promises = [
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }).then(() => loadBudgetOverview()),
        overviewStore.loadTransactionOverview({ force: force, loadLast11Months: true })
    ];

    Promise.all(promises).then(() => {
        loadingOverview.value = false;

        if (force) {
            snackbar.value?.showMessage('Data has been updated');
        }
    }).catch(error => {
        loadingOverview.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

if (isUserLogined() && isUserUnlocked()) {
    reload(false);
}
</script>

<style>
.overview-card-background {
    position: absolute;
    inline-size: 9rem;
    inset-block-end: 0;
    inset-inline-end: 0;
}

.overview-card-background-image {
    position: absolute;
    inline-size: 5rem;
    inset-block-end: 0.5rem;
    inset-inline-end: 1rem;
}
</style>
