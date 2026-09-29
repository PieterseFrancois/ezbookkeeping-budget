<template>
    <f7-page ptr
             infinite
             :infinite-preloader="loadingMore"
             :infinite-distance="600"
             @ptr:refresh="reload"
             @page:afterin="onPageAfterIn"
             @infinite="loadMoreDrafts">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>{{ tt('Transaction Drafts') }}</f7-nav-title>
            <f7-nav-right>
                <f7-link icon-f7="question_circle" @click="showDraftHelp"></f7-link>
                <f7-link icon-f7="arrow_clockwise" @click="() => reload()"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-top-half skeleton-text" v-if="loading && !drafts.length">
            <f7-list-item title="Draft" footer="Loading draft" :key="i" v-for="i in 3"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-top-half" v-else-if="!loading && !drafts.length">
            <f7-list-item :title="tt('No transaction drafts')"></f7-list-item>
        </f7-list>

        <f7-list strong inset dividers media-list class="transaction-info-list" v-else>
            <f7-list-item
                swipeout
                no-chevron
                class="transaction-info drafts-m-item"
                :class="{ 'drafts-m-item-incomplete': !draft.complete }"
                :key="draft.source"
                v-for="draft in drafts"
                @click="openTransactionEditPage(draft)"
            >
                <template #media>
                    <div class="transaction-icon display-flex align-items-center" style="position: relative;">
                        <ItemIcon icon-type="category" :icon-id="getCategory(draft)?.icon" :color="getCategory(draft)?.color" v-if="getCategory(draft)"></ItemIcon>
                        <f7-icon v-else f7="tray"></f7-icon>
                        <f7-badge v-if="draft.complete" color="green"
                                  style="position: absolute; bottom: -4px; right: -4px; min-width: 14px; height: 14px; font-size: 9px; line-height: 14px;">
                            <f7-icon f7="checkmark_alt" style="font-size: 9px;"></f7-icon>
                        </f7-badge>
                    </div>
                </template>
                <template #title>
                    <div class="item-title-row">
                        <div class="item-title">
                            <div class="transaction-category-name no-padding">
                                <span>{{ getCategoryName(draft) || tt('No category set') }}</span>
                            </div>
                        </div>
                        <div class="item-after">
                            <div class="transaction-amount"
                                 :class="{ 'text-expense': draft.type === TransactionType.Expense, 'text-income': draft.type === TransactionType.Income }">
                                <span>{{ getDisplayAmount(draft) }}</span>
                            </div>
                        </div>
                    </div>
                    <div class="item-text" v-if="draft.comment">
                        <div class="transaction-description">
                            <span>{{ draft.comment }}</span>
                        </div>
                    </div>
                    <div class="item-footer">
                        <div class="transaction-footer drafts-m-footer">
                            <span>{{ getDisplayTime(draft) }}</span>
                            <span class="drafts-m-footer-account">
                                <span>{{ getAccountName(draft) || tt('No account set') }}</span>
                                <f7-icon class="transaction-account-arrow icon-with-direction" f7="arrow_right"
                                         v-if="draft.type === TransactionType.Transfer && getDestinationAccount(draft) && draft.accountId !== draft.destinationAccountId"></f7-icon>
                                <span v-if="draft.type === TransactionType.Transfer && getDestinationAccount(draft) && draft.accountId !== draft.destinationAccountId">{{ getDestinationAccount(draft)?.name }}</span>
                            </span>
                        </div>
                    </div>
                </template>
                <f7-swipeout-actions right>
                    <f7-swipeout-button color="green" close :class="{ 'disabled': !draft.complete }" :text="tt('Confirm')" @click="quickConfirm(draft)"></f7-swipeout-button>
                    <f7-swipeout-button color="red" close :text="tt('Discard')" @click="confirmDiscard(draft)"></f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>
        </f7-list>

        <f7-actions v-model:opened="showDiscardActions">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to discard this transaction draft?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="executeDiscard">{{ tt('Discard') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button @click="showDiscardActions = false">{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionDraftsStore } from '@/stores/transactionDraft.ts';

import { TransactionType } from '@/core/transaction.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionDraftInfoResponse, TransactionDraftConfirmRequest } from '@/models/transaction_draft.ts';

import services from '@/lib/services.ts';
import { parseDateTimeFromUnixTimeWithTimezoneOffset } from '@/lib/datetime.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';

const props = defineProps<{
    f7router: Router.Router;
}>();

const {
    tt,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatDateTimeToLongDateTime
} = useI18n();
const { showAlert, showToast } = useI18nUIComponents();

const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionDraftsStore = useTransactionDraftsStore();

const loading = ref<boolean>(true);
const loadingMore = ref<boolean>(false);
const drafts = ref<TransactionDraftInfoResponse[]>([]);
const nextPage = ref<number>(0);
const pageSize = 20;
const hasMoreDrafts = ref<boolean>(false);

const showDiscardActions = ref<boolean>(false);
const discardingDraft = ref<TransactionDraftInfoResponse | null>(null);

const defaultCurrency = computed<string>(() => accountsStore.allPlainAccounts[0]?.currency ?? 'USD');
const allAccountsMap = computed<Record<string, Account>>(() => accountsStore.allAccountsMap);
const allCategoriesMap = computed<Record<string, TransactionCategory>>(() => transactionCategoriesStore.allTransactionCategoriesMap);

function getDisplayTime(draft: TransactionDraftInfoResponse): string {
    return formatDateTimeToLongDateTime(parseDateTimeFromUnixTimeWithTimezoneOffset(draft.time, draft.utcOffset));
}

function getDisplayAmount(draft: TransactionDraftInfoResponse): string {
    if (draft.hideAmount) {
        return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, defaultCurrency.value);
    }

    const account = draft.accountId ? allAccountsMap.value[draft.accountId] : undefined;
    return formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(draft.amount), account ? account.currency : defaultCurrency.value);
}

function getCategory(draft: TransactionDraftInfoResponse): TransactionCategory | undefined {
    if (!draft.categoryId) {
        return undefined;
    }

    return allCategoriesMap.value[draft.categoryId];
}

function getCategoryName(draft: TransactionDraftInfoResponse): string {
    return getCategory(draft)?.name ?? '';
}

function getAccount(draft: TransactionDraftInfoResponse): Account | undefined {
    if (!draft.accountId) {
        return undefined;
    }

    return allAccountsMap.value[draft.accountId];
}

function getAccountName(draft: TransactionDraftInfoResponse): string {
    return getAccount(draft)?.name ?? '';
}

function getDestinationAccount(draft: TransactionDraftInfoResponse): Account | undefined {
    if (!draft.destinationAccountId) {
        return undefined;
    }

    return allAccountsMap.value[draft.destinationAccountId];
}

function reload(done?: () => void): void {
    if (!done) {
        loading.value = true;
    }

    services.listTransactionDrafts({ page: 0, count: pageSize }).then(response => {
        loading.value = false;
        done?.();

        if (!response.data || !response.data.success || !response.data.result) {
            showToast('Unable to retrieve transaction drafts');
            return;
        }

        drafts.value = response.data.result;
        nextPage.value = 1;
        hasMoreDrafts.value = response.data.result.length >= pageSize;
    }).catch(error => {
        loading.value = false;
        done?.();

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function loadMoreDrafts(): void {
    if (!hasMoreDrafts.value || loadingMore.value || loading.value) {
        return;
    }

    loadingMore.value = true;

    services.listTransactionDrafts({ page: nextPage.value, count: pageSize }).then(response => {
        loadingMore.value = false;

        if (!response.data || !response.data.success || !response.data.result) {
            showToast('Unable to retrieve transaction drafts');
            return;
        }

        drafts.value = drafts.value.concat(response.data.result);
        nextPage.value = nextPage.value + 1;
        hasMoreDrafts.value = response.data.result.length >= pageSize;
    }).catch(error => {
        loadingMore.value = false;

        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function showDraftHelp(): void {
    showAlert('A draft can be saved without a category or account, and reviewed and confirmed later.');
}

function removeDraftFromList(source: string): void {
    drafts.value = drafts.value.filter(draft => draft.source !== source);
}

function openTransactionEditPage(draft: TransactionDraftInfoResponse): void {
    transactionDraftsStore.setPendingEditDraft(draft);
    props.f7router.navigate(`/transaction/add?draftSource=${encodeURIComponent(draft.source)}&noTransactionDraft=true`);
}

function quickConfirm(draft: TransactionDraftInfoResponse): void {
    if (!draft.complete) {
        showToast('Please complete the missing fields before confirming');
        return;
    }

    const req: TransactionDraftConfirmRequest = {
        source: draft.source,
        categoryId: draft.categoryId || undefined,
        accountId: draft.accountId || undefined,
        destinationAccountId: draft.type === TransactionType.Transfer ? (draft.destinationAccountId || undefined) : undefined,
        excludeFromBudget: draft.excludeFromBudget
    };

    services.confirmTransactionDraft(req).then(response => {
        if (!response.data || !response.data.success || !response.data.result) {
            showToast('Unable to confirm transaction draft');
            return;
        }

        showToast('You have confirmed this transaction draft');
        removeDraftFromList(draft.source);
        transactionDraftsStore.refreshDraftCount();
    }).catch(error => {
        if (!error.processed) {
            showToast(error.message || error);
        }
    });
}

function confirmDiscard(draft: TransactionDraftInfoResponse): void {
    discardingDraft.value = draft;
    showDiscardActions.value = true;
}

function executeDiscard(): void {
    if (!discardingDraft.value) {
        return;
    }

    const draft = discardingDraft.value;
    showDiscardActions.value = false;

    services.deleteTransactionDraftBySource({ source: draft.source }).then(response => {
        if (!response.data || !response.data.success) {
            showToast('Unable to discard transaction draft');
            return;
        }

        showToast('You have discarded this transaction draft');
        removeDraftFromList(draft.source);
        transactionDraftsStore.refreshDraftCount();
    }).catch(error => {
        if (!error.processed) {
            showToast(error.message || error);
        }
    }).finally(() => {
        discardingDraft.value = null;
    });
}

let initialized = false;

async function init(): Promise<void> {
    try {
        await accountsStore.loadAllAccounts({ force: false });
    } catch {
        // accounts may already be loaded
    }
    try {
        await transactionCategoriesStore.loadAllCategories({ force: false });
    } catch {
        // categories may already be loaded
    }
    reload();
    initialized = true;
}

function onPageAfterIn(): void {
    if (!isUserLogined() || !isUserUnlocked()) {
        return;
    }

    if (!initialized) {
        init();
    } else {
        // returning from the transaction edit page (e.g. after editing/confirming a draft) - refresh
        reload();
        transactionDraftsStore.refreshDraftCount();
    }
}

</script>

<style>

.drafts-m-item-incomplete .item-inner {
    box-shadow: inset 3px 0 0 0 var(--f7-color-red);
}

.drafts-m-item .transaction-icon {
    margin-inline-end: 12px;
}

.drafts-m-item .item-title-row {
    display: flex !important;
    align-items: baseline !important;
    justify-content: space-between !important;
    gap: 8px;
    width: 100% !important;
}

.drafts-m-item .item-title-row .item-title {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.drafts-m-item .item-title-row .item-after {
    flex: 0 0 auto;
}

.drafts-m-item .drafts-m-footer {
    display: flex !important;
    align-items: baseline !important;
    justify-content: space-between !important;
    gap: 8px;
    width: 100% !important;
}

.drafts-m-item .drafts-m-footer > span:first-child {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.drafts-m-item .drafts-m-footer-account {
    flex: 0 0 auto;
    display: flex;
    align-items: baseline;
}
</style>
