<template>
    <f7-page ptr @ptr:refresh="reload" @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-title :title="tt('global.app.title')"></f7-nav-title>
        </f7-navbar>

        <overview-dashboard :layout="layout" :loading="loading" @navigate="onNavigate" />

        <budget-overview-card :loading="loadingBudget" :budget-summary="budgetSummary" :unbudgeted="unbudgeted" :cycle-note="budgetCycleNote" />

        <f7-toolbar tabbar icons bottom class="main-tabbar">
            <f7-link class="link" href="/transaction/list" style="position: relative;" :aria-label="tt('Details')">
                <f7-icon f7="square_list" aria-hidden="true"></f7-icon>
                <span class="tabbar-label">{{ tt('Details') }}</span>
                <f7-badge color="orange" v-if="transactionDraftsStore.draftCount > 0"
                          style="position: absolute; top: 2px; right: 18%; min-width: 14px; height: 14px; font-size: 9px; line-height: 14px;">
                    {{ formatNumberToLocalizedNumerals(transactionDraftsStore.draftCount) }}
                </f7-badge>
            </f7-link>
            <f7-link class="link" href="/account/list" :aria-label="tt('Accounts')">
                <f7-icon f7="creditcard" aria-hidden="true"></f7-icon>
                <span class="tabbar-label">{{ tt('Accounts') }}</span>
            </f7-link>
            <!-- "homepage-add-button" must have the "dragenabled" class, otherwise the popover disappears immediately after the second long press -->
            <f7-link id="homepage-add-button" class="link dragenabled"
                     href="/transaction/add"
                     :aria-label="tt('Add Transaction')"
                     @taphold="openTransactionTemplatePopover">
                <f7-icon f7="plus_square" class="ebk-tarbar-big-icon" aria-hidden="true"></f7-icon>
            </f7-link>
            <f7-link class="link" href="/statistic/transaction" :aria-label="tt('Statistics')">
                <f7-icon f7="chart_pie" aria-hidden="true"></f7-icon>
                <span class="tabbar-label">{{ tt('Statistics') }}</span>
            </f7-link>
            <f7-link id="homepage-financial-control-button" class="link"
                     href="#" :aria-label="tt('Control')" @click="showFinancialControlPopover = true">
                <f7-icon f7="square_grid_2x2" aria-hidden="true"></f7-icon>
                <span class="tabbar-label">{{ tt('Control') }}</span>
            </f7-link>
            <f7-link class="link" href="/settings" :aria-label="tt('Settings')">
                <f7-icon f7="gear_alt" aria-hidden="true"></f7-icon>
                <span class="tabbar-label">{{ tt('Settings') }}</span>
            </f7-link>
        </f7-toolbar>

        <f7-popover class="template-popover-menu" target-el="#homepage-add-button"
                    v-model:opened="showTransactionTemplatePopover">
            <f7-list dividers v-if="hasTransactionAddMenuItems">
                <f7-list-item key="AIClipboardTextRecognition" link="#" no-chevron popover-close
                              :title="tt('AI Clipboard Text Recognition')"
                              @click="addByRecognizingClipboardText"
                              v-if="isTransactionFromAITextRecognitionEnabled()">
                    <template #media>
                        <f7-icon f7="wand_stars"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item key="AIImageRecognition" link="#" no-chevron popover-close
                              :title="tt('AI Image Recognition')"
                              @click="showAIReceiptImageRecognitionSheet = true"
                              v-if="isTransactionFromAIImageRecognitionEnabled()">
                    <template #media>
                        <f7-icon f7="wand_stars"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item popover-close :key="template.id" :title="template.name"
                              :link="'/transaction/add?templateId=' + template.id"
                              v-for="template in allTransactionTemplates">
                    <template #media>
                        <f7-icon f7="doc_plaintext"></f7-icon>
                    </template>
                </f7-list-item>
            </f7-list>
        </f7-popover>

        <!-- Financial control entries; add future pages (e.g. reports) as further list items here -->
        <f7-popover class="financial-control-popover-menu" target-el="#homepage-financial-control-button"
                    v-model:opened="showFinancialControlPopover">
            <f7-list dividers>
                <f7-list-item popover-close link="/budget" :title="tt('Budget')">
                    <template #media>
                        <f7-icon f7="creditcard_fill"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item popover-close link="/goals" :title="tt('Goals')">
                    <template #media>
                        <f7-icon f7="flag_2"></f7-icon>
                    </template>
                </f7-list-item>
                <f7-list-item popover-close link="/subscriptions" :title="tt('Subscriptions')">
                    <template #media>
                        <f7-icon f7="arrow_2_squarepath"></f7-icon>
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
import AIImageRecognitionSheet, { type AIImageRecognitionResult } from '@/components/mobile/AIImageRecognitionSheet.vue';
import BudgetOverviewCard, { type BudgetSummaryItem, type UnbudgetedItem } from '@/views/mobile/budget/BudgetOverviewCard.vue';
import OverviewDashboard from './overview/OverviewDashboard.vue';

import { ref, computed, useTemplateRef } from 'vue';
import type { Router } from 'framework7/types';
import axios from 'axios';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents, isiOS } from '@/lib/ui/mobile.ts';

import { useSettingsStore } from '@/stores/setting.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useOverviewStore } from '@/stores/overview.ts';
import { useUserStore } from '@/stores/user.ts';
import { useTransactionDraftsStore } from '@/stores/transactionDraft.ts';

import { CategoryType } from '@/core/category.ts';
import {
    type MobileOverviewLayout,
    OverviewWidgetDataRequirement,
    MobileOverviewWidgetNavigationType
} from '@/core/overview_layout.ts';
import { TemplateType } from '@/core/template.ts';
import { MOBILE_OVERVIEW_WIDGET_DEFINITIONS, DEFAULT_MOBILE_OVERVIEW_LAYOUT } from '@/consts/overview_layout.ts';

import { TransactionTemplate } from '@/models/transaction_template.ts';
import type { ApiResponse } from '@/core/api.ts';

import { isFunction } from '@/lib/common.ts';
import {
    getOverviewDataRequirements,
    getOverviewTransactionOverviewMonths,
    getOverviewRecentTransactionsQueries,
    getOverviewAssetTrendMonths,
    getOverviewCalendarHeatmapMonths,
    getOverviewTransactionCategoryStatisticDateTypes,
    parseMobileOverviewLayout
} from '@/lib/overview_layout.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';
import { getShareCacheImageBlob } from '@/lib/cache.ts';
import { getThisMonthFirstUnixTime, getThisMonthLastUnixTime } from '@/lib/datetime.ts';
import { addMonths } from '@/views/base/BudgetPageBase.ts';
import {
    isTransactionFromAITextRecognitionEnabled,
    isTransactionFromAIImageRecognitionEnabled
} from '@/lib/server_settings.ts';
import logger from '@/lib/logger.ts';

type AIImageRecognitionSheetType = InstanceType<typeof AIImageRecognitionSheet>;

const props = defineProps<{
    f7router: Router.Router;
}>();

const { tt, formatNumberToLocalizedNumerals } = useI18n();
const { showToast } = useI18nUIComponents();

const settingsStore = useSettingsStore();
const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionTemplatesStore = useTransactionTemplatesStore();
const overviewStore = useOverviewStore();
const userStore = useUserStore();
const transactionDraftsStore = useTransactionDraftsStore();

const aiImageRecognitionSheet = useTemplateRef<AIImageRecognitionSheetType>('aiImageRecognitionSheet');

const loading = ref<boolean>(true);
const loadingBudget = ref<boolean>(true);
const showTransactionTemplatePopover = ref<boolean>(false);
const showFinancialControlPopover = ref<boolean>(false);
const showAIReceiptImageRecognitionSheet = ref<boolean>(false);

const budgetSummary = ref<BudgetSummaryItem[]>([]);
const unbudgeted = ref<UnbudgetedItem[]>([]);
const cycleYear = ref<number>(new Date().getFullYear());
const cycleMonth = ref<number>(new Date().getMonth() + 1);

const budgetCycleNote = computed<string>(() => {
    const endDay = userStore.currentUserBudgetEndDay;
    if (!endDay) return '';
    const { month: prevMonth } = addMonths(cycleYear.value, cycleMonth.value, -1);
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    return `Budget cycle: ${endDay + 1} ${monthNames[prevMonth - 1]} – ${endDay} ${monthNames[cycleMonth.value - 1]}`;
});

interface BudgetTargetRawItem {
    id: string;
    categoryId: string;
    year: number;
    month: number;
    amount: string;
}

interface BudgetActualRawItem {
    categoryId: string;
    section: string;
    amount: number;
}

async function loadBudgetOverview(): Promise<void> {
    const now = new Date();
    const calYear = now.getFullYear();
    const calMonth = now.getMonth() + 1;
    const endDay = userStore.currentUserBudgetEndDay;
    let year: number;
    let month: number;
    let startTime: number;
    let endTime: number;

    if (!endDay) {
        year = calYear;
        month = calMonth;
        startTime = getThisMonthFirstUnixTime();
        endTime = getThisMonthLastUnixTime();
    } else {
        // If today has passed endDay, we're already in the next cycle
        const resolved = now.getDate() > endDay ? addMonths(calYear, calMonth, 1) : { year: calYear, month: calMonth };
        year = resolved.year;
        month = resolved.month;
        const { year: prevYear, month: prevMonth } = addMonths(year, month, -1);
        startTime = Math.floor(new Date(prevYear, prevMonth - 1, endDay + 1, 0, 0, 0, 0).getTime() / 1000);
        endTime = Math.floor(new Date(year, month - 1, endDay + 1, 0, 0, 0, 0).getTime() / 1000) - 1;
    }

    cycleYear.value = year;
    cycleMonth.value = month;

    const [budgetResp, actualsResp] = await Promise.all([
        axios.get<ApiResponse<BudgetTargetRawItem[]>>(`v1/budget/targets.json?year=${year}&month=${month}`),
        axios.get<ApiResponse<{ items: BudgetActualRawItem[] }>>('v1/budget/actuals.json', { params: { startTime, endTime } })
    ]);

    const targets = budgetResp.data?.result ?? [];
    const actualItems = actualsResp.data?.result?.items ?? [];

    // Per-category amounts split by section (the backend already excludes budget-excluded transactions)
    const sectionByCat: Record<string, Partial<Record<string, number>>> = {};
    for (const item of actualItems) {
        const entry = sectionByCat[item.categoryId] ?? (sectionByCat[item.categoryId] = {});
        entry[item.section] = (entry[item.section] ?? 0) + item.amount;
    }

    const spentBySubcategoryId: Record<string, number> = {};
    const savingsNetBySubId: Record<string, number> = {};
    for (const [catId, sections] of Object.entries(sectionByCat)) {
        if (sections['expense']) {
            spentBySubcategoryId[catId] = sections['expense'];
        }
        // Gross set-aside: savings contributions plus card/debt paydowns. Withdrawals are not
        // netted off — they are income, and the overview card tracks progress toward targets.
        const setAside = (sections['savings'] ?? 0) + (sections['debt'] ?? 0);
        if (setAside !== 0) {
            savingsNetBySubId[catId] = setAside;
        }
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

const layout = computed<MobileOverviewLayout>(() => {
    try {
        return parseMobileOverviewLayout(settingsStore.appSettings.mobileOverviewPageLayout);
    } catch (error) {
        logger.warn('failed to parse mobile overview page layout, fallback to default layout', error);
        return DEFAULT_MOBILE_OVERVIEW_LAYOUT;
    }
});

const allTransactionTemplates = computed<TransactionTemplate[]>(() => {
    const allTemplates = transactionTemplatesStore.allVisibleTemplates;
    return allTemplates[TemplateType.Normal.type] || [];
});

const hasTransactionAddMenuItems = computed<boolean>(() => isTransactionFromAITextRecognitionEnabled() || isTransactionFromAIImageRecognitionEnabled() || (allTransactionTemplates.value && allTransactionTemplates.value.length > 0));

function openTransactionTemplatePopover(): void {
    if (hasTransactionAddMenuItems.value) {
        showTransactionTemplatePopover.value = true;
    }
}

function init(): void {
    if (isUserLogined() && isUserUnlocked()) {
        loading.value = true;

        const promises: Promise<unknown>[] = [
            getShareCacheImageBlob(),
            accountsStore.loadAllAccounts({ force: false }),
            transactionCategoriesStore.loadAllCategories({ force: false }).then(() => loadBudgetOverview().finally(() => { loadingBudget.value = false; })),
            transactionTemplatesStore.loadAllTemplates({ templateType: TemplateType.Normal.type,  force: false }),
            transactionDraftsStore.refreshDraftCount(),
            ...reloadOverviewData(false)
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
    const promises: Promise<unknown>[] = reloadOverviewData(force);

    transactionDraftsStore.refreshDraftCount();

    if (promises.length < 1) {
        done?.();
        return;
    }

    Promise.all(promises).then(() => {
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

function reloadOverviewData(force: boolean): Promise<unknown>[] {
    const requirements: OverviewWidgetDataRequirement[] = getOverviewDataRequirements(layout.value, MOBILE_OVERVIEW_WIDGET_DEFINITIONS);
    const promises: Promise<unknown>[] = [];

    if (requirements.includes(OverviewWidgetDataRequirement.TransactionOverview)) {
        promises.push(overviewStore.loadTransactionOverview({
            force: force,
            months: getOverviewTransactionOverviewMonths(layout.value)
        }));
    }

    if (requirements.includes(OverviewWidgetDataRequirement.TransactionCategoryStatistics)) {
        for (const dateType of getOverviewTransactionCategoryStatisticDateTypes(layout.value)) {
            promises.push(overviewStore.loadTransactionCategoryStatistics({
                force: force,
                dateType: dateType
            }));
        }
    }

    if (requirements.includes(OverviewWidgetDataRequirement.AssetTrends)) {
        promises.push(overviewStore.loadTransactionAssetTrends({
            force: force,
            months: getOverviewAssetTrendMonths(layout.value)
        }));
    }

    if (requirements.includes(OverviewWidgetDataRequirement.RecentTransactions)) {
        promises.push(overviewStore.loadRecentTransactions({
            force: force,
            queries: getOverviewRecentTransactionsQueries(layout.value)
        }));
    }

    if (requirements.includes(OverviewWidgetDataRequirement.CurrentMonthTransactions)) {
        promises.push(overviewStore.loadCurrentMonthTransactions({
            force: force
        }));
    }

    if (requirements.includes(OverviewWidgetDataRequirement.DailyTransactionAmounts)) {
        promises.push(overviewStore.loadTransactionDailyAmounts({
            force: force,
            months: getOverviewCalendarHeatmapMonths(layout.value)
        }));
    }

    return promises;
}

function addByRecognizingClipboardText(): void {
    if (navigator.clipboard && isFunction(navigator.clipboard.readText) && !isiOS()) {
        navigator.clipboard.readText().then(text => {
            const clipboardText = text && text.trim() ? text.trim() : '';
            props.f7router.navigate('/transaction/add', {
                props: {
                    autoRecognizeClipboardText: clipboardText,
                }
            });
        }).catch(error => {
            logger.error('failed to read clipboard', error);
            props.f7router.navigate('/transaction/add', {
                props: {
                    autoRecognizeClipboardText: '',
                }
            });
        });
    } else {
        props.f7router.navigate('/transaction/add', {
            props: {
                autoRecognizeClipboardText: '',
            }
        });
    }
}

function onReceiptRecognitionChanged(result: AIImageRecognitionResult): void {
    const recognizedResponse = result.response;
    const autoUploadRecognizedImage = settingsStore.appSettings.autoUploadTransactionPictureForAIRecognition;
    const params: string[] = [];

    if (recognizedResponse.type) {
        params.push(`type=${recognizedResponse.type}`);
    }

    if (recognizedResponse.time) {
        params.push(`time=${recognizedResponse.time}`);
    }

    if (recognizedResponse.categoryId) {
        params.push(`categoryId=${recognizedResponse.categoryId}`);
    }

    if (recognizedResponse.sourceAccountId) {
        params.push(`accountId=${recognizedResponse.sourceAccountId}`);
    }

    if (recognizedResponse.destinationAccountId) {
        params.push(`destinationAccountId=${recognizedResponse.destinationAccountId}`);
    }

    if (recognizedResponse.sourceAmount) {
        params.push(`amount=${recognizedResponse.sourceAmount}`);
    }

    if (recognizedResponse.destinationAmount) {
        params.push(`destinationAmount=${recognizedResponse.destinationAmount}`);
    }

    if (recognizedResponse.tagIds) {
        params.push(`tagIds=${recognizedResponse.tagIds.join(',')}`);
    }

    if (recognizedResponse.comment) {
        params.push(`comment=${encodeURIComponent(recognizedResponse.comment)}`);
    }

    params.push(`noTransactionDraft=true`);

    props.f7router.navigate(`/transaction/add?${params.join('&')}`, {
        props: {
            autoUploadPicture: autoUploadRecognizedImage ? result.imageFile : undefined,
        }
    });
}

function onNavigate(type: MobileOverviewWidgetNavigationType, path?: string): void {
    if (type === MobileOverviewWidgetNavigationType.Url && path) {
        props.f7router.navigate(path);
    } else if (type === MobileOverviewWidgetNavigationType.AIClipboardTextRecognition) {
        addByRecognizingClipboardText();
    } else if (type === MobileOverviewWidgetNavigationType.AIImageRecognition) {
        showAIReceiptImageRecognitionSheet.value = true;
    }
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
