<template>
    <v-row class="match-height">
        <!-- Header -->
        <v-col cols="12" class="d-flex align-center pb-2">
            <h5 class="text-h5">{{ tt('Subscriptions') }}</h5>
            <v-spacer />
            <v-btn :prepend-icon="mdiPlus" variant="tonal" @click="openAddDialog">
                {{ tt('Add Subscription') }}
            </v-btn>
        </v-col>

        <v-col cols="12">
            <v-card>
                <v-data-table
                    :loading="loading"
                    :headers="headers"
                    :items="tableItems"
                    :items-per-page="-1"
                    :sort-by="[{ key: 'nextExpectedDate', order: 'asc' }]"
                    hide-default-footer
                    :no-data-text="tt('No subscriptions yet. Add your first subscription to get started.')"
                    item-value="id"
                    :row-props="({ item }) => ({ class: item.isActive ? '' : 'subscription-row-inactive' })"
                >
                    <template #item.categoryName="{ item }">
                        <div v-if="item.primaryCategoryName">{{ item.primaryCategoryName }}</div>
                        <div class="text-medium-emphasis text-caption" v-if="item.subCategoryName">{{ item.subCategoryName }}</div>
                    </template>
                    <template #item.amountDisplay="{ item }">
                        <div>{{ item.amountDisplay }}</div>
                        <div class="text-medium-emphasis text-caption" v-if="item.convertedAmountDisplay">{{ item.convertedAmountDisplay }}</div>
                    </template>
                    <template #item.nextExpectedDate="{ item }">
                        {{ item.nextExpectedDateDisplay }}
                    </template>
                    <template #item.isActive="{ item }">
                        <div class="d-flex justify-center">
                            <v-switch
                                class="subscriptions-active-switch"
                                density="compact"
                                hide-details
                                :model-value="item.isActive"
                                @update:model-value="(value: boolean | null) => toggleActive(item, !!value)"
                            />
                        </div>
                    </template>
                    <template #item.actions="{ item }">
                        <div class="d-flex gap-1 justify-end">
                            <v-btn
                                v-if="item.templateId && item.templateId !== '0'"
                                density="compact"
                                color="primary"
                                variant="text"
                                :icon="true"
                                @click="addTransactionFromSubscription(item.templateId)"
                            >
                                <v-icon :icon="mdiPlusCircle" />
                                <v-tooltip activator="parent">{{ tt('Add Transaction') }}</v-tooltip>
                            </v-btn>
                            <v-btn density="compact" color="default" variant="text" :icon="true" @click="openEditDialog(item.raw)">
                                <v-icon :icon="mdiPencilOutline" />
                                <v-tooltip activator="parent">{{ tt('Edit Subscription') }}</v-tooltip>
                            </v-btn>
                            <v-btn density="compact" color="error" variant="text" :icon="true" @click="confirmDelete(item.raw)">
                                <v-icon :icon="mdiTrashCanOutline" />
                                <v-tooltip activator="parent">{{ tt('Delete Subscription') }}</v-tooltip>
                            </v-btn>
                        </div>
                    </template>
                </v-data-table>
            </v-card>
        </v-col>
    </v-row>

    <edit-dialog ref="editDialog" :type="TransactionEditPageType.Transaction" />

    <!-- Add / Edit Dialog -->
    <v-dialog v-model="showDialog" max-width="520" :persistent="saving">
        <v-card :title="editingSubscription ? tt('Edit Subscription') : tt('Add Subscription')">
            <v-card-text class="d-flex flex-column gap-4 pt-2">
                <v-text-field
                    v-model="form.name"
                    :label="tt('Subscription Name')"
                    density="compact"
                    :error-messages="formErrors.name ? [formErrors.name] : []"
                    autofocus
                />
                <currency-select
                    :label="tt('Currency')"
                    :placeholder="tt('Currency')"
                    v-model="form.currency"
                />
                <div>
                    <amount-input
                        v-model="form.amount"
                        :label="tt('Amount')"
                        :currency="form.currency"
                        :show-currency="true"
                        :persistent-placeholder="true"
                        density="compact"
                        hide-details
                    />
                    <div class="text-error text-caption mt-1" v-if="formErrors.amount">{{ formErrors.amount }}</div>
                </div>
                <div>
                    <date-select
                        :label="tt('Start Date')"
                        v-model="form.startDate"
                    />
                    <div class="text-error text-caption mt-1" v-if="formErrors.startDate">{{ formErrors.startDate }}</div>
                </div>
                <v-select
                    v-model="form.frequency"
                    :items="frequencyOptions"
                    item-title="label"
                    item-value="value"
                    :label="tt('Frequency')"
                    density="compact"
                    hide-details
                />
                <v-select
                    v-model="form.primaryCategoryId"
                    :items="primaryCategoryOptions"
                    item-title="name"
                    item-value="id"
                    :label="tt('Category')"
                    density="compact"
                    hide-details
                    @update:model-value="onPrimaryCategoryChange"
                />
                <div v-if="form.primaryCategoryId !== '0'">
                    <v-select
                        v-model="form.categoryId"
                        :items="secondaryCategoryOptions"
                        item-title="name"
                        item-value="id"
                        :label="tt('Subcategory')"
                        density="compact"
                        :error-messages="formErrors.category ? [formErrors.category] : []"
                    />
                </div>
                <v-select
                    v-model="form.templateId"
                    :items="templateOptions"
                    item-title="name"
                    item-value="id"
                    :label="tt('Transaction Template')"
                    density="compact"
                    hide-details
                />
                <v-switch
                    v-model="form.isActive"
                    :label="tt('Active')"
                    density="compact"
                    hide-details
                />
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn :disabled="saving" @click="showDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" :loading="saving" @click="saveSubscription">{{ tt('Save') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="showDeleteDialog" max-width="420" :persistent="deleting">
        <v-card :title="tt('Delete Subscription')">
            <v-card-text>{{ tt('Are you sure you want to delete this subscription?') }}</v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn :disabled="deleting" @click="showDeleteDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="error" :loading="deleting" @click="deleteSubscription">{{ tt('Confirm') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';
import AmountInput from '@/components/desktop/AmountInput.vue';
import CurrencySelect from '@/components/desktop/CurrencySelect.vue';
import DateSelect from '@/components/desktop/DateSelect.vue';
import EditDialog from '@/views/desktop/transactions/list/dialogs/EditDialog.vue';
import { ref, computed, useTemplateRef } from 'vue';
import axios from 'axios';

import { useI18n } from '@/locales/helpers.ts';
import { useUserStore } from '@/stores/user.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';
import {
    parseDateTimeFromUnixTime,
    getLocalDateFromYearDashMonthDashDay,
    getUnixTimeFromLocalDatetime
} from '@/lib/datetime.ts';
import { getNextExpectedDate, SubscriptionFrequency, type SubscriptionFrequencyValue } from '@/lib/subscription.ts';
import { CategoryType } from '@/core/category.ts';
import { TemplateType } from '@/core/template.ts';
import { TransactionEditPageType } from '@/views/base/transactions/TransactionEditPageBase.ts';
import type { TextualYearMonthDay } from '@/core/datetime.ts';
import type { ApiResponse } from '@/core/api.ts';

import {
    mdiPlus,
    mdiPlusCircle,
    mdiPencilOutline,
    mdiTrashCanOutline,
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;
const snackbar = useTemplateRef<SnackBarType>('snackbar');

type EditDialogType = InstanceType<typeof EditDialog>;
const editDialog = useTemplateRef<EditDialogType>('editDialog');

const {
    tt,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatGregorianTextualYearMonthDayToLongDate,
} = useI18n();

const userStore = useUserStore();
const categoriesStore = useTransactionCategoriesStore();
const templatesStore = useTransactionTemplatesStore();
const exchangeRatesStore = useExchangeRatesStore();

const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

// ── Types ──────────────────────────────────────────────────

interface RawSubscription {
    id: string;
    name: string;
    amount: string;
    currency: string;
    categoryId: string;
    templateId: string;
    startDate: string;
    frequency: SubscriptionFrequencyValue;
    isActive: boolean;
    createdAt: string;
}

interface Subscription {
    id: string;
    name: string;
    amount: number;
    currency: string;
    categoryId: string;
    templateId: string;
    startDate: number;
    frequency: SubscriptionFrequencyValue;
    isActive: boolean;
    createdAt: number;
}

// ── State ──────────────────────────────────────────────────

const loading = ref<boolean>(true);
const saving = ref<boolean>(false);
const deleting = ref<boolean>(false);
const subscriptions = ref<Subscription[]>([]);

const showDialog = ref<boolean>(false);
const showDeleteDialog = ref<boolean>(false);
const editingSubscription = ref<Subscription | null>(null);
const deletingSubscription = ref<Subscription | null>(null);

interface SubscriptionFormErrors {
    name?: string;
    amount?: string;
    startDate?: string;
    category?: string;
}

const formErrors = ref<SubscriptionFormErrors>({});

interface SubscriptionForm {
    name: string;
    amount: number;
    currency: string;
    primaryCategoryId: string;
    categoryId: string;
    templateId: string;
    startDate: TextualYearMonthDay | undefined;
    frequency: SubscriptionFrequencyValue;
    isActive: boolean;
}

function defaultForm(): SubscriptionForm {
    return {
        name: '',
        amount: 0,
        currency: defaultCurrency.value,
        primaryCategoryId: '0',
        categoryId: '0',
        templateId: '0',
        startDate: parseDateTimeFromUnixTime(getUnixTimeFromLocalDatetime(new Date())).getGregorianCalendarYearDashMonthDashDay(),
        frequency: SubscriptionFrequency.Monthly,
        isActive: true,
    };
}

const form = ref<SubscriptionForm>(defaultForm());

// ── Options ────────────────────────────────────────────────

const frequencyOptions = computed(() => [
    { label: tt('Daily'), value: SubscriptionFrequency.Daily },
    { label: tt('Weekly'), value: SubscriptionFrequency.Weekly },
    { label: tt('Monthly'), value: SubscriptionFrequency.Monthly },
    { label: tt('Quarterly'), value: SubscriptionFrequency.Quarterly },
    { label: tt('Annual'), value: SubscriptionFrequency.Annual },
]);

function frequencyLabel(frequency: SubscriptionFrequencyValue): string {
    return frequencyOptions.value.find(option => option.value === frequency)?.label ?? '';
}

const expenseCategories = computed(() => categoriesStore.allTransactionCategories[CategoryType.Expense] || []);

const primaryCategoryOptions = computed(() => {
    const options: { id: string, name: string }[] = [{ id: '0', name: tt('None') }];

    for (const category of expenseCategories.value) {
        options.push({ id: category.id, name: category.name });
    }

    return options;
});

const secondaryCategoryOptions = computed(() => {
    const primaryCategory = expenseCategories.value.find(category => category.id === form.value.primaryCategoryId);
    return (primaryCategory?.subCategories || []).map(subCategory => ({ id: subCategory.id, name: subCategory.name }));
});

function onPrimaryCategoryChange(): void {
    form.value.categoryId = secondaryCategoryOptions.value[0]?.id ?? '0';
}

function categoryNameParts(categoryId: string): { primaryName: string, subName: string } {
    if (!categoryId || categoryId === '0') {
        return { primaryName: '', subName: '' };
    }

    const category = categoriesStore.allTransactionCategoriesMap[categoryId];

    if (!category) {
        return { primaryName: '', subName: '' };
    }

    const parentCategory = category.parentId && category.parentId !== '0' ? categoriesStore.allTransactionCategoriesMap[category.parentId] : undefined;
    return parentCategory ? { primaryName: parentCategory.name, subName: category.name } : { primaryName: category.name, subName: '' };
}

const templateOptions = computed(() => {
    const options: { id: string, name: string }[] = [{ id: '0', name: tt('None') }];
    const normalTemplates = templatesStore.allTransactionTemplates[TemplateType.Normal.type] || [];

    for (const template of normalTemplates) {
        options.push({ id: template.id, name: template.name });
    }

    return options;
});

// ── Table items ────────────────────────────────────────────

const headers = computed(() => [
    { title: tt('Subscription Name'), key: 'name' },
    { title: tt('Category'), key: 'categoryName' },
    { title: tt('Amount'), key: 'amountDisplay' },
    { title: tt('Next Expected Date'), key: 'nextExpectedDate', width: '240px' },
    { title: tt('Frequency'), key: 'frequencyLabel', align: 'center' as const },
    { title: tt('Active'), key: 'isActive', align: 'center' as const, width: '96px' },
    { title: '', key: 'actions', sortable: false, align: 'end' as const },
]);

const tableItems = computed(() => subscriptions.value.map(subscription => {
    const convertedAmount = subscription.currency !== defaultCurrency.value
        ? exchangeRatesStore.getExchangedAmount(subscription.amount, subscription.currency, defaultCurrency.value)
        : null;
    const category = categoryNameParts(subscription.categoryId);

    return {
        raw: subscription,
        id: subscription.id,
        name: subscription.name,
        amountDisplay: formatAmountToLocalizedNumeralsWithCurrency(subscription.amount, subscription.currency),
        convertedAmountDisplay: subscription.currency !== defaultCurrency.value
            ? (convertedAmount !== null ? formatAmountToLocalizedNumeralsWithCurrency(Math.round(convertedAmount), defaultCurrency.value) : '-')
            : '',
        categoryName: [category.primaryName, category.subName].filter(Boolean).join(' '),
        primaryCategoryName: category.primaryName,
        subCategoryName: category.subName,
        templateId: subscription.templateId,
        frequencyLabel: frequencyLabel(subscription.frequency),
        nextExpectedDate: getNextExpectedDate(subscription.startDate, subscription.frequency),
        nextExpectedDateDisplay: formatGregorianTextualYearMonthDayToLongDate(parseDateTimeFromUnixTime(getNextExpectedDate(subscription.startDate, subscription.frequency)).getGregorianCalendarYearDashMonthDashDay()),
        isActive: subscription.isActive,
    };
}));

// ── Dialog helpers ─────────────────────────────────────────

function openAddDialog(): void {
    editingSubscription.value = null;
    form.value = defaultForm();
    formErrors.value = {};
    showDialog.value = true;
}

function openEditDialog(subscription: Subscription): void {
    editingSubscription.value = subscription;
    const category = subscription.categoryId !== '0' ? categoriesStore.allTransactionCategoriesMap[subscription.categoryId] : undefined;
    form.value = {
        name: subscription.name,
        amount: subscription.amount,
        currency: subscription.currency,
        primaryCategoryId: category?.parentId && category.parentId !== '0' ? category.parentId : '0',
        categoryId: subscription.categoryId,
        templateId: subscription.templateId,
        startDate: parseDateTimeFromUnixTime(subscription.startDate).getGregorianCalendarYearDashMonthDashDay(),
        frequency: subscription.frequency,
        isActive: subscription.isActive,
    };
    formErrors.value = {};
    showDialog.value = true;
}

function validateForm(): boolean {
    const errors: SubscriptionFormErrors = {};

    if (!form.value.name.trim()) {
        errors.name = tt('Subscription name is required');
    }

    if (!form.value.amount || form.value.amount <= 0) {
        errors.amount = tt('Amount must be greater than zero');
    }

    if (!form.value.startDate) {
        errors.startDate = tt('Start date is required');
    }

    if (form.value.primaryCategoryId !== '0' && form.value.categoryId === '0') {
        errors.category = tt('Please select a subcategory');
    }

    formErrors.value = errors;
    return Object.keys(errors).length === 0;
}

function confirmDelete(subscription: Subscription): void {
    deletingSubscription.value = subscription;
    showDeleteDialog.value = true;
}

function addTransactionFromSubscription(templateId: string): void {
    const template = templatesStore.allTransactionTemplatesMap[TemplateType.Normal.type]?.[templateId];

    if (!template) {
        snackbar.value?.showError(tt('Unable to find transaction template'));
        return;
    }

    editDialog.value?.open({ template }).then(result => {
        if (result && result.message) {
            snackbar.value?.showMessage(result.message);
        }
    }).catch(error => {
        if (error) {
            snackbar.value?.showError(error);
        }
    });
}

// ── API calls ──────────────────────────────────────────────

async function loadSubscriptions(): Promise<void> {
    loading.value = true;
    try {
        const resp = await axios.get<ApiResponse<RawSubscription[]>>('v1/subscriptions/list.json');
        subscriptions.value = (resp.data?.result ?? []).map(s => ({
            id: s.id,
            name: s.name,
            amount: Number(s.amount),
            currency: s.currency,
            categoryId: s.categoryId,
            templateId: s.templateId,
            startDate: Number(s.startDate),
            frequency: s.frequency,
            isActive: s.isActive,
            createdAt: Number(s.createdAt),
        }));
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        loading.value = false;
    }
}

async function saveSubscription(): Promise<void> {
    if (saving.value) return;

    if (!validateForm()) {
        return;
    }

    saving.value = true;
    try {
        const startDateObj = form.value.startDate ? getLocalDateFromYearDashMonthDashDay(form.value.startDate) : null;
        const startUnix = startDateObj ? getUnixTimeFromLocalDatetime(startDateObj) : getUnixTimeFromLocalDatetime(new Date());

        const payload = {
            name: form.value.name,
            amount: String(form.value.amount),
            currency: form.value.currency,
            categoryId: form.value.categoryId,
            templateId: form.value.templateId,
            startDate: String(startUnix),
            frequency: form.value.frequency,
            isActive: form.value.isActive,
        };

        if (editingSubscription.value) {
            await axios.post('v1/subscriptions/modify.json', { id: String(editingSubscription.value.id), ...payload });
        } else {
            await axios.post('v1/subscriptions/add.json', payload);
        }
        showDialog.value = false;
        await loadSubscriptions();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        saving.value = false;
    }
}

async function toggleActive(item: { raw: Subscription }, isActive: boolean): Promise<void> {
    try {
        await axios.post('v1/subscriptions/toggle.json', { id: String(item.raw.id), isActive });
        item.raw.isActive = isActive;
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    }
}

async function deleteSubscription(): Promise<void> {
    if (!deletingSubscription.value || deleting.value) return;
    deleting.value = true;
    try {
        await axios.post('v1/subscriptions/delete.json', { id: String(deletingSubscription.value.id) });
        showDeleteDialog.value = false;
        await loadSubscriptions();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        deleting.value = false;
    }
}

// ── Boot ───────────────────────────────────────────────────

async function init(): Promise<void> {
    try {
        await categoriesStore.loadAllCategories({ force: false });
    } catch {
        // categories may already be loaded
    }
    try {
        await templatesStore.loadAllTemplates({ templateType: TemplateType.Normal.type, force: false });
    } catch {
        // templates may already be loaded
    }
    try {
        await exchangeRatesStore.getLatestExchangeRates({ silent: true, force: false });
    } catch {
        // exchange rates may be unavailable, converted amounts will be hidden
    }
    await loadSubscriptions();
}

if (isUserLogined() && isUserUnlocked()) {
    init();
}
</script>

<style scoped>
:deep(tr.subscription-row-inactive) {
    opacity: 0.5;
}

:deep(.subscriptions-active-switch) {
    transform: scale(0.8);
    flex: none;
}
</style>
