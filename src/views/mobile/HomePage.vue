<template>
    <f7-page ptr @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-title :title="tt('global.app.title')"></f7-nav-title>
        </f7-navbar>

        <f7-card class="home-summary-card no-margin-top" :class="{ 'skeleton-text': loading }">
            <f7-card-header class="display-block" style="padding-top: 120px;">
                <p class="no-margin">
                    <span class="card-header-content" v-if="loading">
                        <span class="home-summary-month">Month</span>
                        <span>·</span>
                        <small>Expense</small>
                    </span>
                    <span class="card-header-content" v-else-if="!loading">
                        <span class="home-summary-month">{{ displayDateRange?.thisMonth?.displayTime }}</span>
                        <span>·</span>
                        <small>{{ tt('Expense') }}</small>
                    </span>
                </p>
                <p class="no-margin">
                    <span class="month-expense" v-if="loading">0.00 USD</span>
                    <span class="month-expense" v-else-if="!loading">{{ transactionOverview && transactionOverview.thisMonth ? getDisplayExpenseAmount(transactionOverview.thisMonth) : '-' }}</span>
                    <f7-link class="display-inline-flex margin-inline-start-half" @click="showAmountInHomePage = !showAmountInHomePage">
                        <f7-icon class="ebk-hide-icon" :f7="showAmountInHomePage ? 'eye_slash_fill' : 'eye_fill'"></f7-icon>
                    </f7-link>
                </p>
                <p class="no-margin">
                    <small class="home-summary-misc" v-if="loading">Monthly income 0.00 USD</small>
                    <small class="home-summary-misc" v-else-if="!loading">
                        <span>{{ tt('Monthly income') }}</span>
                        <span>{{ transactionOverview && transactionOverview.thisMonth ? getDisplayIncomeAmount(transactionOverview.thisMonth) : '-' }}</span>
                    </small>
                </p>
            </f7-card-header>
        </f7-card>

        <f7-list strong inset dividers class="margin-top overview-transaction-list" :class="{ 'skeleton-text': loading }">
            <f7-list-item :link="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.Today.type })}`" chevron-center>
                <template #media>
                    <f7-icon f7="calendar_today"></f7-icon>
                </template>
                <template #title>
                    <div class="padding-top-half">
                        <span v-if="loading">Today</span>
                        <span v-else-if="!loading">{{ tt('Today') }}</span>
                    </div>
                </template>
                <template #footer>
                    <div class="overview-transaction-footer padding-bottom-half">
                        <span v-if="loading">MM/DD/YYYY</span>
                        <span v-else-if="!loading">{{ displayDateRange?.today?.displayTime }}</span>
                    </div>
                </template>
                <template #after>
                    <div class="overview-transaction-amount">
                        <div class="text-income text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.today && transactionOverview.today.valid">{{ getDisplayIncomeAmount(transactionOverview.today) }}</small>
                        </div>
                        <div class="text-expense text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.today && transactionOverview.today.valid">{{ getDisplayExpenseAmount(transactionOverview.today) }}</small>
                        </div>
                    </div>
                </template>
            </f7-list-item>

            <f7-list-item :link="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.ThisWeek.type })}`" chevron-center>
                <template #media>
                    <f7-icon f7="calendar"></f7-icon>
                </template>
                <template #title>
                    <div class="padding-top-half">
                        <span v-if="loading">This Week</span>
                        <span v-else-if="!loading">{{ tt('This Week') }}</span>
                    </div>
                </template>
                <template #footer>
                    <div class="overview-transaction-footer padding-bottom-half">
                        <span v-if="loading">MM/DD</span>
                        <span v-else-if="!loading">{{ displayDateRange?.thisWeek?.startTime }}</span>
                        <span>-</span>
                        <span v-if="loading">MM/DD</span>
                        <span v-else-if="!loading">{{ displayDateRange?.thisWeek?.endTime }}</span>
                    </div>
                </template>
                <template #after>
                    <div class="overview-transaction-amount">
                        <div class="text-income text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.thisWeek && transactionOverview.thisWeek.valid">{{ getDisplayIncomeAmount(transactionOverview.thisWeek) }}</small>
                        </div>
                        <div class="text-expense text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.thisWeek && transactionOverview.thisWeek.valid">{{ getDisplayExpenseAmount(transactionOverview.thisWeek) }}</small>
                        </div>
                    </div>
                </template>
            </f7-list-item>

            <f7-list-item :link="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.ThisMonth.type })}`" chevron-center>
                <template #media>
                    <f7-icon f7="calendar"></f7-icon>
                </template>
                <template #title>
                    <div class="padding-top-half">
                        <span v-if="loading">This Month</span>
                        <span v-else-if="!loading">{{ tt('This Month') }}</span>
                    </div>
                </template>
                <template #footer>
                    <div class="overview-transaction-footer padding-bottom-half">
                        <span v-if="loading">MM/DD</span>
                        <span v-else-if="!loading">{{ displayDateRange?.thisMonth?.startTime }}</span>
                        <span>-</span>
                        <span v-if="loading">MM/DD</span>
                        <span v-else-if="!loading">{{ displayDateRange?.thisMonth?.endTime }}</span>
                    </div>
                </template>
                <template #after>
                    <div class="overview-transaction-amount">
                        <div class="text-income text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.thisMonth && transactionOverview.thisMonth.valid">{{ getDisplayIncomeAmount(transactionOverview.thisMonth) }}</small>
                        </div>
                        <div class="text-expense text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.thisMonth && transactionOverview.thisMonth.valid">{{ getDisplayExpenseAmount(transactionOverview.thisMonth) }}</small>
                        </div>
                    </div>
                </template>
            </f7-list-item>

            <f7-list-item :link="`/transaction/list?${overviewStore.getTransactionListPageParams({ dateType: DateRange.ThisYear.type })}`" chevron-center>
                <template #media>
                    <f7-icon f7="square_stack_3d_up"></f7-icon>
                </template>
                <template #title>
                    <div class="padding-top-half">
                        <span v-if="loading">This Year</span>
                        <span v-else-if="!loading">{{ tt('This Year') }}</span>
                    </div>
                </template>
                <template #footer>
                    <div class="overview-transaction-footer padding-bottom-half">
                        <span v-if="loading">YYYY</span>
                        <span v-else-if="!loading">{{ displayDateRange?.thisYear?.displayTime }}</span>
                    </div>
                </template>
                <template #after>
                    <div class="overview-transaction-amount">
                        <div class="text-income text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.thisYear && transactionOverview.thisYear.valid">{{ getDisplayIncomeAmount(transactionOverview.thisYear) }}</small>
                        </div>
                        <div class="text-expense text-align-right">
                            <small v-if="loading">0.00 USD</small>
                            <small v-else-if="!loading && transactionOverview.thisYear && transactionOverview.thisYear.valid">{{ getDisplayExpenseAmount(transactionOverview.thisYear) }}</small>
                        </div>
                    </div>
                </template>
            </f7-list-item>
        </f7-list>

        <budget-overview-card :loading="loadingBudget" :budget-summary="budgetSummary" :unbudgeted="unbudgeted" />

        <f7-toolbar tabbar icons bottom class="main-tabbar">
            <f7-link class="link" href="/transaction/list">
                <f7-icon f7="square_list"></f7-icon>
                <span class="tabbar-label">{{ tt('Details') }}</span>
            </f7-link>
            <f7-link class="link" href="/account/list">
                <f7-icon f7="creditcard"></f7-icon>
                <span class="tabbar-label">{{ tt('Accounts') }}</span>
            </f7-link>
            <f7-link id="homepage-add-button" class="link dragenabled"
                     href="/transaction/add" @taphold="openTransactionTemplatePopover">
                <f7-icon f7="plus_square" class="ebk-tarbar-big-icon"></f7-icon>
            </f7-link>
            <f7-link class="link" href="/statistic/transaction">
                <f7-icon f7="chart_pie"></f7-icon>
                <span class="tabbar-label">{{ tt('Statistics') }}</span>
            </f7-link>
            <f7-link class="link" href="/settings">
                <f7-icon f7="gear_alt"></f7-icon>
                <span class="tabbar-label">{{ tt('Settings') }}</span>
            </f7-link>
        </f7-toolbar>

        <f7-popover class="template-popover-menu" target-el="#homepage-add-button"
                    v-model:opened="showTransactionTemplatePopover">
            <f7-list dividers v-if="allTransactionTemplates">
                <f7-list-item key="AIImageRecognition" :title="tt('AI Image Recognition')"
                              @click="showAIReceiptImageRecognitionSheet = true; showTransactionTemplatePopover = false"
                              v-if="isTransactionFromAIImageRecognitionEnabled()">
                    <template #media>
                        <f7-icon f7="wand_stars"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item :key="template.id" :title="template.name"
                              :link="'/transaction/add?templateId=' + template.id"
                              v-for="template in allTransactionTemplates">
                    <template #media>
                        <f7-icon f7="doc_plaintext"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <a-i-image-recognition-sheet ref="aiImageRecognitionSheet"
                                     v-model:show="showAIReceiptImageRecognitionSheet"
                                     @recognition:change="onReceiptRecognitionChanged"/>
    </f7-page>
</template>

<script setup lang="ts">
import AIImageRecognitionSheet from '@/components/mobile/AIImageRecognitionSheet.vue';
import BudgetOverviewCard, { type BudgetSummaryItem, type UnbudgetedItem } from '@/views/mobile/budget/BudgetOverviewCard.vue';

import { ref, computed, useTemplateRef } from 'vue';
import type { Router } from 'framework7/types';
import axios from 'axios';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useHomePageBase } from '@/views/base/HomePageBase.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useOverviewStore } from '@/stores/overview.ts';

import { DateRange } from '@/core/datetime.ts';
import { CategoryType } from '@/core/category.ts';
import { TemplateType } from '@/core/template.ts';
import { TransactionTemplate } from '@/models/transaction_template.ts';
import type { RecognizedReceiptImageResponse } from '@/models/large_language_model.ts';
import type { ApiResponse } from '@/core/api.ts';
import services from '@/lib/services.ts';

import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';
import { getShareCacheImageBlob } from '@/lib/cache.ts';
import { getThisMonthFirstUnixTime, getThisMonthLastUnixTime } from '@/lib/datetime.ts';
import { isTransactionFromAIImageRecognitionEnabled } from '@/lib/server_settings.ts';

type AIImageRecognitionSheetType = InstanceType<typeof AIImageRecognitionSheet>;

const props = defineProps<{
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const { showToast } = useI18nUIComponents();

const {
    showAmountInHomePage,
    displayDateRange,
    transactionOverview,
    getDisplayIncomeAmount,
    getDisplayExpenseAmount
} = useHomePageBase();

const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionTemplatesStore = useTransactionTemplatesStore();
const overviewStore = useOverviewStore();

const aiImageRecognitionSheet = useTemplateRef<AIImageRecognitionSheetType>('aiImageRecognitionSheet');

const loading = ref<boolean>(true);
const loadingBudget = ref<boolean>(true);
const showTransactionTemplatePopover = ref<boolean>(false);
const showAIReceiptImageRecognitionSheet = ref<boolean>(false);

const budgetSummary = ref<BudgetSummaryItem[]>([]);
const unbudgeted = ref<UnbudgetedItem[]>([]);

interface BudgetTargetRawItem {
    id: string;
    categoryId: string;
    year: number;
    month: number;
    amount: string;
}

interface SavingsActualRawItem {
    categoryId: string;
    transferOut: string;
    transferIn: string;
    net: string;
}

async function loadBudgetOverview(): Promise<void> {
    const now = new Date();
    const year = now.getFullYear();
    const month = now.getMonth() + 1;
    const startTime = getThisMonthFirstUnixTime();
    const endTime = getThisMonthLastUnixTime();

    const [budgetResp, statsResp, savingsResp] = await Promise.all([
        axios.get<ApiResponse<BudgetTargetRawItem[]>>(`v1/budget/targets.json?year=${year}&month=${month}`),
        services.getTransactionStatistics({ startTime, endTime, tagFilter: '', keyword: '', useTransactionTimezone: false }),
        axios.get<ApiResponse<{ items: SavingsActualRawItem[] }>>(`v1/budget/savings-actuals.json?year=${year}&month=${month}`)
    ]);

    const targets = budgetResp.data?.result ?? [];
    const statsItems = statsResp.data?.result?.items ?? [];
    const savingsItems = savingsResp.data?.result?.items ?? [];

    const spentBySubcategoryId: Record<string, number> = {};
    for (const item of statsItems) {
        const cat = transactionCategoriesStore.allTransactionCategoriesMap[item.categoryId];
        if (cat && cat.type === CategoryType.Expense) {
            spentBySubcategoryId[item.categoryId] = (spentBySubcategoryId[item.categoryId] ?? 0) + item.amount;
        }
    }

    const savingsNetBySubId: Record<string, number> = {};
    for (const item of savingsItems) {
        savingsNetBySubId[item.categoryId] = Number(item.net);
    }

    const budgetedSubcategoryIds = new Set<string>();
    for (const target of targets) {
        budgetedSubcategoryIds.add(target.categoryId);
    }

    const parentGroups: Record<string, { name: string; icon: string; color: string; isSavings: boolean; totalBudgeted: number; totalSpent: number }> = {};
    for (const target of targets) {
        const subCat = transactionCategoriesStore.allTransactionCategoriesMap[target.categoryId];
        if (!subCat || !subCat.parentId || subCat.parentId === '0') continue;

        const parentId = subCat.parentId;
        const parentCat = transactionCategoriesStore.allTransactionCategoriesMap[parentId];
        if (!parentCat) continue;

        const group = parentGroups[parentId] ?? (parentGroups[parentId] = { name: parentCat.name, icon: parentCat.icon, color: parentCat.color, isSavings: parentCat.type === CategoryType.Transfer, totalBudgeted: 0, totalSpent: 0 });
        group.totalBudgeted += Number(target.amount);
    }

    for (const [parentId, group] of Object.entries(parentGroups)) {
        const parentCat = transactionCategoriesStore.allTransactionCategoriesMap[parentId];
        const isTransfer = parentCat?.type === CategoryType.Transfer;
        for (const subCat of parentCat?.subCategories ?? []) {
            group.totalSpent += isTransfer
                ? (savingsNetBySubId[subCat.id] ?? 0)
                : (spentBySubcategoryId[subCat.id] ?? 0);
        }
    }

    budgetSummary.value = Object.values(parentGroups).map(g => ({
        categoryName: g.name,
        icon: g.icon,
        color: g.color,
        budgeted: g.totalBudgeted,
        spent: g.totalSpent,
        remaining: g.totalBudgeted - g.totalSpent,
        isSavings: g.isSavings,
    }));

    const unbudgetedList: UnbudgetedItem[] = [];
    for (const [subcatId, spent] of Object.entries(spentBySubcategoryId)) {
        if (spent <= 0 || budgetedSubcategoryIds.has(subcatId)) continue;
        const subCat = transactionCategoriesStore.allTransactionCategoriesMap[subcatId];
        if (!subCat) continue;
        const parentCat = subCat.parentId && subCat.parentId !== '0'
            ? transactionCategoriesStore.allTransactionCategoriesMap[subCat.parentId]
            : undefined;
        const iconSource = parentCat ?? subCat;
        unbudgetedList.push({ categoryName: parentCat ? `${parentCat.name} > ${subCat.name}` : subCat.name, icon: iconSource.icon, color: iconSource.color, spent });
    }
    unbudgeted.value = unbudgetedList;
}

const allTransactionTemplates = computed<TransactionTemplate[]>(() => {
    const allTemplates = transactionTemplatesStore.allVisibleTemplates;
    return allTemplates[TemplateType.Normal.type] || [];
});

function openTransactionTemplatePopover(): void {
    if (isTransactionFromAIImageRecognitionEnabled() || (allTransactionTemplates.value && allTransactionTemplates.value.length)) {
        showTransactionTemplatePopover.value = true;
    }
}

function init(): void {
    if (isUserLogined() && isUserUnlocked()) {
        loading.value = true;

        const promises = [
            getShareCacheImageBlob(),
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }).then(() => loadBudgetOverview().finally(() => { loadingBudget.value = false; })),
            transactionTemplatesStore.loadAllTemplates({ templateType: TemplateType.Normal.type,  force: false }),
            overviewStore.loadTransactionOverview({ force: false })
        ];

        Promise.all(promises).then(responses => {
            if (responses[0] && responses[0] instanceof Blob) {
                aiImageRecognitionSheet.value?.loadImage(responses[0]);
                showAIReceiptImageRecognitionSheet.value = true;
            }

            loading.value = false;
        }).catch(error => {
            loading.value = false;

            if (!error.processed) {
                showToast(error.message || error);
            }
        });
    }
}

function reload(done?: () => void): void {
    const force = !!done;

    overviewStore.loadTransactionOverview({
        force: force
    }).then(() => {
        done?.();

        if (force) {
            showToast('Data has been updated');
        }
    }).catch(error => {
        done?.();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function onReceiptRecognitionChanged(result: RecognizedReceiptImageResponse): void {
    const params: string[] = [];

    if (result.type) {
        params.push(`type=${result.type}`);
    }

    if (result.time) {
        params.push(`time=${result.time}`);
    }

    if (result.categoryId) {
        params.push(`categoryId=${result.categoryId}`);
    }

    if (result.sourceAccountId) {
        params.push(`accountId=${result.sourceAccountId}`);
    }

    if (result.destinationAccountId) {
        params.push(`destinationAccountId=${result.destinationAccountId}`);
    }

    if (result.sourceAmount) {
        params.push(`amount=${result.sourceAmount}`);
    }

    if (result.destinationAmount) {
        params.push(`destinationAmount=${result.destinationAmount}`);
    }

    if (result.tagIds) {
        params.push(`tagIds=${result.tagIds.join(',')}`);
    }

    if (result.comment) {
        params.push(`comment=${encodeURIComponent(result.comment)}`);
    }

    params.push(`noTransactionDraft=true`);

    props.f7router.navigate(`/transaction/add?${params.join('&')}`);
}

function onPageAfterIn(): void {
    if (!loading.value) {
        reload();
    }
}

init();
</script>

<style>
.home-summary-card {
    background-color: var(--f7-color-yellow);
}

.home-summary-card .home-summary-month {
    font-size: 1.3em;
}

.home-summary-card .month-expense {
    font-size: 1.5em;
}

.home-summary-card .home-summary-misc {
    opacity: 0.6;
}

.home-summary-misc > span {
    margin-inline-end: 4px;
}

.home-summary-misc > span:last-child {
    margin-inline-end: 0;
}

.dark .home-summary-card {
    background-color: var(--f7-theme-color);
}

.dark .home-summary-card a {
    color: var(--f7-text-color);
    opacity: 0.6;
}

.overview-transaction-list .item-title > div {
    overflow: hidden;
    text-overflow: ellipsis;
}

.overview-transaction-list .item-after {
    max-width: 100%;
}

.overview-transaction-list .overview-transaction-footer {
    padding-top: 6px;
    font-size: var(--ebk-large-footer-font-size);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.overview-transaction-list .overview-transaction-footer > span {
    margin-inline-end: 4px;
}

.overview-transaction-list .overview-transaction-amount {
    max-width: 100%;
}

.overview-transaction-list .overview-transaction-amount > div {
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
}

.tabbar.main-tabbar .link i + span.tabbar-label {
    margin-top: var(--ebk-icon-text-margin);
}

.tabbar.main-tabbar .link i.ebk-tarbar-big-icon {
    font-size: var(--ebk-big-icon-button-size);
    width: var(--ebk-big-icon-button-size);
    height: var(--ebk-big-icon-button-size);
    line-height: var(--ebk-big-icon-button-size);
}

.template-popover-menu .popover-inner {
    max-height: 400px;
    overflow-y: auto;
}
</style>
