<template>
    <v-row class="match-height">
        <v-col cols="12" class="d-flex align-center pb-2">
            <h5 class="text-h5">{{ tt('Transaction Drafts') }}</h5>
            <v-btn density="compact" color="default" variant="text" size="24"
                   class="ms-2" :icon="true">
                <v-icon :icon="mdiHelpCircleOutline" size="20" />
                <v-tooltip activator="parent">{{ tt('A draft can be saved without a category or account, and reviewed and confirmed later.') }}</v-tooltip>
            </v-btn>
            <v-spacer />
            <v-btn density="compact" color="default" variant="text" size="24"
                   class="me-2" :icon="true" :loading="loading" @click="reload">
                <template #loader>
                    <v-progress-circular indeterminate size="20"/>
                </template>
                <v-icon :icon="mdiRefresh" size="24" />
                <v-tooltip activator="parent">{{ tt('Refresh') }}</v-tooltip>
            </v-btn>
        </v-col>

        <v-col cols="12">
            <v-card>
                <v-table class="draft-table" :hover="!loading">
                    <thead>
                    <tr>
                        <th>{{ tt('Time') }}</th>
                        <th>{{ tt('Type') }}</th>
                        <th>{{ tt('Amount') }}</th>
                        <th>{{ tt('Category') }}</th>
                        <th>{{ tt('Account') }}</th>
                        <th>{{ tt('Description') }}</th>
                        <th></th>
                    </tr>
                    </thead>
                    <tbody v-if="loading && !drafts.length">
                        <tr :key="idx" v-for="idx in 3">
                            <td colspan="7"><v-skeleton-loader type="text" :loading="true"></v-skeleton-loader></td>
                        </tr>
                    </tbody>
                    <tbody v-if="!loading && !drafts.length">
                        <tr>
                            <td colspan="7">{{ tt('No transaction drafts') }}</td>
                        </tr>
                    </tbody>
                    <tbody :key="draft.id" :class="{ 'disabled': loading }" v-for="draft in drafts">
                        <tr class="draft-row" :class="{ 'draft-row-incomplete': !draft.complete }" @click="openTransactionEditDialog(draft)">
                            <td>{{ getDisplayTime(draft) }}</td>
                            <td>{{ getTransactionTypeName(draft.type) }}</td>
                            <td :class="{ 'text-expense': draft.type === TransactionType.Expense, 'text-income': draft.type === TransactionType.Income }">
                                {{ getDisplayAmount(draft) }}
                            </td>
                            <td>
                                <div class="d-flex align-center">
                                    <ItemIcon class="me-2" size="20px" icon-type="category"
                                              :icon-id="getCategory(draft)?.icon" :color="getCategory(draft)?.color"
                                              v-if="getCategory(draft)"></ItemIcon>
                                    <span>{{ getCategoryName(draft) || tt('No category set') }}</span>
                                </div>
                            </td>
                            <td>
                                <div class="d-flex align-center">
                                    <ItemIcon class="me-2" size="20px" icon-type="account"
                                              :icon-id="getAccount(draft)?.icon" :color="getAccount(draft)?.color"
                                              v-if="getAccount(draft)"></ItemIcon>
                                    <span>{{ getAccountName(draft) || tt('No account set') }}</span>
                                    <v-icon class="icon-with-direction mx-1" size="13" :icon="mdiArrowRight"
                                            v-if="draft.type === TransactionType.Transfer && getDestinationAccount(draft) && draft.accountId !== draft.destinationAccountId"></v-icon>
                                    <span v-if="draft.type === TransactionType.Transfer && getDestinationAccount(draft) && draft.accountId !== draft.destinationAccountId">{{ getDestinationAccount(draft)?.name }}</span>
                                </div>
                            </td>
                            <td class="text-truncate">{{ draft.comment }}</td>
                            <td>
                                <div class="d-flex gap-1 justify-end" @click.stop>
                                    <v-btn density="compact" color="primary" variant="text" :icon="true"
                                           :disabled="loading || processingSource === draft.source || !draft.complete" @click="quickConfirm(draft)">
                                        <v-icon :icon="mdiCheck" />
                                        <v-tooltip activator="parent">{{ draft.complete ? tt('Confirm') : tt('Complete the missing fields before confirming') }}</v-tooltip>
                                    </v-btn>
                                    <v-btn density="compact" color="error" variant="text" :icon="true"
                                           :disabled="loading || processingSource === draft.source" @click="discardDraft(draft)">
                                        <v-icon :icon="mdiTrashCanOutline" />
                                        <v-tooltip activator="parent">{{ tt('Discard') }}</v-tooltip>
                                    </v-btn>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </v-table>

                <div class="d-flex justify-center align-center mt-2 mb-4 gap-4">
                    <v-btn density="compact" variant="tonal" :disabled="loading || currentPage <= 1" @click="changePage(currentPage - 1)">{{ tt('Previous') }}</v-btn>
                    <span class="text-body-2">{{ formatNumberToLocalizedNumerals(currentPage) }}</span>
                    <v-btn density="compact" variant="tonal" :disabled="loading || !hasNextPage" @click="changePage(currentPage + 1)">{{ tt('Next') }}</v-btn>
                </div>
            </v-card>
        </v-col>
    </v-row>

    <edit-dialog ref="transactionEditDialog" :type="TransactionEditPageType.Transaction" />
    <confirm-dialog ref="confirmDialog"/>
    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import { ref, computed, useTemplateRef, onMounted } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';

import ConfirmDialog from '@/components/desktop/ConfirmDialog.vue';
import SnackBar from '@/components/desktop/SnackBar.vue';
import EditDialog from './list/dialogs/EditDialog.vue';

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionDraftsStore } from '@/stores/transactionDraft.ts';

import { TransactionType } from '@/core/transaction.ts';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { Account } from '@/models/account.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import type { TransactionDraftInfoResponse, TransactionDraftConfirmRequest } from '@/models/transaction_draft.ts';

import services from '@/lib/services.ts';
import { parseDateTimeFromUnixTimeWithTimezoneOffset } from '@/lib/datetime.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';

import { mdiRefresh, mdiCheck, mdiTrashCanOutline, mdiHelpCircleOutline, mdiArrowRight } from '@mdi/js';

type EditDialogType = InstanceType<typeof EditDialog>;
type ConfirmDialogType = InstanceType<typeof ConfirmDialog>;
type SnackBarType = InstanceType<typeof SnackBar>;

const {
    tt,
    formatNumberToLocalizedNumerals,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatDateTimeToLongDateTime
} = useI18n();

const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();
const transactionDraftsStore = useTransactionDraftsStore();

const transactionEditDialog = useTemplateRef<EditDialogType>('transactionEditDialog');
const confirmDialog = useTemplateRef<ConfirmDialogType>('confirmDialog');
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const loading = ref<boolean>(true);
const drafts = ref<TransactionDraftInfoResponse[]>([]);
const currentPage = ref<number>(1);
const pageSize = 20;
const hasNextPage = ref<boolean>(false);
const processingSource = ref<string>('');

const defaultCurrency = computed<string>(() => accountsStore.allPlainAccounts[0]?.currency ?? 'USD');
const allAccountsMap = computed<Record<string, Account>>(() => accountsStore.allAccountsMap);
const allCategoriesMap = computed<Record<string, TransactionCategory>>(() => transactionCategoriesStore.allTransactionCategoriesMap);

function getTransactionTypeName(type: number): string {
    if (type === TransactionType.Income) {
        return tt('Income');
    } else if (type === TransactionType.Expense) {
        return tt('Expense');
    } else if (type === TransactionType.Transfer) {
        return tt('Transfer');
    }

    return tt('Transaction');
}

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

function reload(): void {
    loadDrafts(currentPage.value);
}

function changePage(page: number): void {
    if (page < 1 || loading.value) {
        return;
    }

    loadDrafts(page);
}

function loadDrafts(page: number): void {
    loading.value = true;

    services.listTransactionDrafts({ page: page - 1, count: pageSize }).then(response => {
        loading.value = false;

        if (!response.data || !response.data.success || !response.data.result) {
            snackbar.value?.showMessage('Unable to retrieve transaction drafts');
            return;
        }

        drafts.value = response.data.result;
        currentPage.value = page;
        hasNextPage.value = response.data.result.length >= pageSize;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function openTransactionEditDialog(draft: TransactionDraftInfoResponse): void {
    transactionEditDialog.value?.open({
        type: draft.type,
        categoryId: draft.categoryId || '',
        accountId: draft.accountId || '',
        destinationAccountId: draft.destinationAccountId || '',
        amount: draft.amount,
        destinationAmount: draft.destinationAmount,
        time: draft.time,
        tagIds: draft.tagIds ? draft.tagIds.join(',') : '',
        comment: draft.comment,
        hideAmount: draft.hideAmount,
        excludeFromBudget: draft.excludeFromBudget,
        pictures: draft.pictures,
        draftSource: draft.source,
        noTransactionDraft: true
    }).then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }

        loadDrafts(currentPage.value);
        transactionDraftsStore.refreshDraftCount();
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

function quickConfirm(draft: TransactionDraftInfoResponse): void {
    if (!draft.complete || processingSource.value) {
        return;
    }

    processingSource.value = draft.source;

    const req: TransactionDraftConfirmRequest = {
        source: draft.source,
        categoryId: draft.categoryId || undefined,
        accountId: draft.accountId || undefined,
        destinationAccountId: draft.type === TransactionType.Transfer ? (draft.destinationAccountId || undefined) : undefined,
        excludeFromBudget: draft.excludeFromBudget
    };

    services.confirmTransactionDraft(req).then(response => {
        processingSource.value = '';

        if (!response.data || !response.data.success || !response.data.result) {
            snackbar.value?.showMessage('Unable to confirm transaction draft');
            return;
        }

        snackbar.value?.showMessage('You have confirmed this transaction draft');
        loadDrafts(currentPage.value);
        transactionDraftsStore.refreshDraftCount();
    }).catch(error => {
        processingSource.value = '';

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function discardDraft(draft: TransactionDraftInfoResponse): void {
    confirmDialog.value?.open('Are you sure you want to discard this transaction draft?').then(() => {
        processingSource.value = draft.source;

        services.deleteTransactionDraftBySource({ source: draft.source }).then(response => {
            processingSource.value = '';

            if (!response.data || !response.data.success) {
                snackbar.value?.showMessage('Unable to discard transaction draft');
                return;
            }

            snackbar.value?.showMessage('You have discarded this transaction draft');
            loadDrafts(currentPage.value);
            transactionDraftsStore.refreshDraftCount();
        }).catch(error => {
            processingSource.value = '';

            if (!error.processed) {
                snackbar.value?.showError(error);
            }
        });
    });
}

onMounted(() => {
    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false })
    ]).finally(() => {
        loadDrafts(1);
    });
});
</script>

<style>
.draft-table td {
    max-width: 220px;
}

.draft-table .draft-row {
    cursor: pointer;
}

.draft-table .draft-row-incomplete {
    box-shadow: inset 3px 0 0 0 rgb(var(--v-theme-error));
}
</style>
