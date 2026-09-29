<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>{{ tt('Subscriptions') }}</f7-nav-title>
            <f7-nav-right>
                <f7-link @click="showAmountInSubscriptionsPage = !showAmountInSubscriptionsPage">
                    <f7-icon class="ebk-hide-icon" :f7="showAmountInSubscriptionsPage ? 'eye_slash_fill' : 'eye_fill'"></f7-icon>
                </f7-link>
                <f7-link icon-f7="plus" @click="openAddPopup" />
            </f7-nav-right>
        </f7-navbar>

        <!-- Skeleton while loading -->
        <f7-list strong inset dividers class="margin-top-half skeleton-text" v-if="loading">
            <f7-list-item title="Subscription Name" footer="Loading subscription" :key="i" v-for="i in 3">
                <template #media>
                    <f7-icon f7="creditcard"></f7-icon>
                </template>
            </f7-list-item>
        </f7-list>

        <!-- Empty state -->
        <f7-list strong inset dividers class="margin-top-half" v-else-if="subscriptions.length === 0">
            <f7-list-item :title="tt('No subscriptions yet. Add your first subscription to get started.')"></f7-list-item>
        </f7-list>

        <!-- Subscription list -->
        <f7-list strong inset dividers class="margin-top-half" v-else>
            <f7-list-item
                swipeout
                no-chevron
                class="subscriptions-m-list-item"
                :class="{ 'subscriptions-m-inactive': !subscription.isActive }"
                :key="subscription.id"
                v-for="subscription in tableItems"
                @click="openEditPopup(subscription.raw)"
            >
                <template #title>
                    <div class="subscriptions-m-row">
                        <span class="subscriptions-m-name">{{ subscription.name }}</span>
                        <div class="subscriptions-m-amount-col">
                            <span class="subscriptions-m-converted subscriptions-m-nowrap" v-if="subscription.convertedAmountDisplay">({{ subscription.convertedAmountDisplay }})</span>
                            <span class="subscriptions-m-nowrap subscriptions-m-primary-amount">{{ subscription.amountDisplay }}</span>
                        </div>
                    </div>
                    <div class="subscriptions-m-row subscriptions-m-secondary">
                        <span class="subscriptions-m-next-date">{{ subscription.nextExpectedDateDisplay }}</span>
                        <span>{{ subscription.frequencyLabel }}</span>
                    </div>
                    <div class="subscriptions-m-row subscriptions-m-category" v-if="subscription.categoryName || subscription.accountCoverage">
                        <span>{{ subscription.categoryName }}</span>
                        <span class="subscriptions-m-account-coverage" v-if="subscription.accountCoverage">
                            <f7-icon
                                :f7="subscription.accountCoverage.covered ? 'checkmark_circle_fill' : 'xmark_circle_fill'"
                                :class="subscription.accountCoverage.covered ? 'text-color-green' : 'text-color-red'"
                                size="14"
                            ></f7-icon>
                            {{ subscription.accountCoverage.accountName }}
                        </span>
                    </div>
                </template>
                <f7-swipeout-actions right>
                    <f7-swipeout-button color="orange" close :text="tt('Edit')" @click="openEditPopup(subscription.raw)"></f7-swipeout-button>
                    <f7-swipeout-button color="red" class="padding-horizontal" @click="confirmDelete(subscription.raw)">
                        <f7-icon f7="trash"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
                <f7-swipeout-actions left v-if="subscription.templateId && subscription.templateId !== '0'">
                    <f7-swipeout-button color="blue" close :href="'/transaction/add?templateId=' + subscription.templateId">
                        <f7-icon f7="plus_circle_fill"></f7-icon>
                    </f7-swipeout-button>
                </f7-swipeout-actions>
            </f7-list-item>
        </f7-list>

        <!-- Add / Edit Popup -->
        <f7-popup v-model:opened="showFormPopup" tablet-fullscreen>
            <f7-page v-if="showFormPopup" :key="editingSubscription ? editingSubscription.id : 'new'">
                <f7-navbar :title="editingSubscription ? tt('Edit Subscription') : tt('Add Subscription')">
                    <f7-nav-right>
                        <f7-link popup-close :text="tt('Cancel')" />
                    </f7-nav-right>
                </f7-navbar>

                <f7-list strong inset dividers>
                    <f7-list-input
                        type="text"
                        :label="tt('Subscription Name')"
                        :value="form.name"
                        @input="form.name = ($event.target as HTMLInputElement).value"
                        clear-button
                    />
                    <f7-list-item
                        link="#"
                        :header="tt('Currency')"
                        :title="form.currency"
                        @click="showCurrencyPopup = true"
                    >
                        <list-item-selection-popup value-type="item"
                                                   key-field="currencyCode" value-field="currencyCode"
                                                   title-field="displayName" after-field="currencyCode"
                                                   :title="tt('Currency Name')"
                                                   :enable-filter="true"
                                                   :filter-placeholder="tt('Currency')"
                                                   :filter-no-items-text="tt('No results')"
                                                   :items="allCurrencies"
                                                   v-model:show="showCurrencyPopup"
                                                   v-model="form.currency">
                        </list-item-selection-popup>
                    </f7-list-item>
                    <f7-list-input
                        type="number"
                        :label="tt('Amount')"
                        :value="form.amountRaw"
                        min="0"
                        @input="form.amountRaw = ($event.target as HTMLInputElement).value"
                    />
                    <f7-list-item
                        link="#"
                        :header="tt('Start Date')"
                        :title="form.startDate ? formatGregorianTextualYearMonthDayToLongDate(form.startDate) : ''"
                        @click="showStartDateSheet = true"
                    >
                        <date-selection-sheet v-model:show="showStartDateSheet"
                                              v-model="form.startDate">
                        </date-selection-sheet>
                    </f7-list-item>
                    <f7-list-item
                        link="#"
                        :header="tt('Frequency')"
                        :title="frequencyLabel(form.frequency)"
                        @click="showFrequencyPopup = true"
                    >
                        <list-item-selection-popup value-type="item"
                                                   key-field="value" value-field="value"
                                                   title-field="label"
                                                   :title="tt('Frequency')"
                                                   :items="frequencyOptions"
                                                   v-model:show="showFrequencyPopup"
                                                   :model-value="form.frequency"
                                                   @update:model-value="(value) => form.frequency = value as SubscriptionFrequencyValue">
                        </list-item-selection-popup>
                    </f7-list-item>
                    <f7-list-item
                        link="#"
                        :header="tt('Transaction Template')"
                        :title="templateName"
                        :footer="tt('Category and account are taken from the linked template, if any')"
                        @click="showTemplatePopup = true"
                    >
                        <list-item-selection-popup value-type="item"
                                                   key-field="id" value-field="id"
                                                   title-field="name"
                                                   :title="tt('Transaction Template')"
                                                   :items="templateOptions"
                                                   v-model:show="showTemplatePopup"
                                                   v-model="form.templateId">
                        </list-item-selection-popup>
                    </f7-list-item>
                    <f7-list-item>
                        <template #title>{{ tt('Active') }}</template>
                        <template #after>
                            <f7-toggle :checked="form.isActive" @toggle:change="(value: boolean) => form.isActive = value"></f7-toggle>
                        </template>
                    </f7-list-item>
                </f7-list>

                <div class="margin">
                    <f7-button fill :loading="saving" @click="saveSubscription">{{ tt('Save') }}</f7-button>
                </div>
            </f7-page>
        </f7-popup>

        <!-- Delete confirmation actions -->
        <f7-actions v-model:opened="showDeleteActions">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this subscription?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="executeDelete">{{ tt('Delete Subscription') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button @click="showDeleteActions = false">{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import ListItemSelectionPopup from '@/components/mobile/ListItemSelectionPopup.vue';
import DateSelectionSheet from '@/components/mobile/DateSelectionSheet.vue';
import { ref, computed } from 'vue';
import axios from 'axios';

import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useUserStore } from '@/stores/user.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useTransactionTemplatesStore } from '@/stores/transactionTemplate.ts';
import { useExchangeRatesStore } from '@/stores/exchangeRates.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';
import {
    parseDateTimeFromUnixTime,
    getLocalDateFromYearDashMonthDashDay,
    getUnixTimeFromLocalDatetime
} from '@/lib/datetime.ts';
import { getNextExpectedDate, SubscriptionFrequency, type SubscriptionFrequencyValue } from '@/lib/subscription.ts';
import { TemplateType } from '@/core/template.ts';
import type { TransactionTemplate } from '@/models/transaction_template.ts';
import type { LocalizedCurrencyInfo } from '@/core/currency.ts';
import type { TextualYearMonthDay } from '@/core/datetime.ts';
import type { ApiResponse } from '@/core/api.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';

const {
    tt,
    parseAmountFromLocalizedNumerals,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatAmountToLocalizedNumeralsWithoutDigitGrouping,
    formatGregorianTextualYearMonthDayToLongDate,
    getAllCurrencies,
} = useI18n();
const { showToast } = useI18nUIComponents();

const userStore = useUserStore();
const accountsStore = useAccountsStore();
const categoriesStore = useTransactionCategoriesStore();
const templatesStore = useTransactionTemplatesStore();
const exchangeRatesStore = useExchangeRatesStore();
const settingsStore = useSettingsStore();

const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

const showAmountInSubscriptionsPage = computed<boolean>({
    get: () => settingsStore.appSettings.showAmountInSubscriptionsPage,
    set: (value) => settingsStore.setShowAmountInSubscriptionsPage(value)
});
const allCurrencies = computed<LocalizedCurrencyInfo[]>(() => getAllCurrencies());

// ── Types ──────────────────────────────────────────────────

interface RawSubscription {
    id: string;
    name: string;
    amount: string;
    currency: string;
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
    templateId: string;
    startDate: number;
    frequency: SubscriptionFrequencyValue;
    isActive: boolean;
    createdAt: number;
}

// ── State ──────────────────────────────────────────────────

const loading = ref<boolean>(true);
const saving = ref<boolean>(false);
const subscriptions = ref<Subscription[]>([]);

const showFormPopup = ref<boolean>(false);
const showDeleteActions = ref<boolean>(false);
const showCurrencyPopup = ref<boolean>(false);
const showStartDateSheet = ref<boolean>(false);
const showFrequencyPopup = ref<boolean>(false);
const showTemplatePopup = ref<boolean>(false);
const editingSubscription = ref<Subscription | null>(null);
const deletingSubscription = ref<Subscription | null>(null);

interface SubscriptionForm {
    name: string;
    amountRaw: string;
    currency: string;
    templateId: string;
    startDate: TextualYearMonthDay | undefined;
    frequency: SubscriptionFrequencyValue;
    isActive: boolean;
}

function defaultForm(): SubscriptionForm {
    return {
        name: '',
        amountRaw: '',
        currency: defaultCurrency.value,
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

const templateOptions = computed(() => {
    const options: { id: string, name: string }[] = [{ id: '0', name: tt('None') }];
    const normalTemplates = templatesStore.allTransactionTemplates[TemplateType.Normal.type] || [];

    for (const template of normalTemplates) {
        options.push({ id: template.id, name: template.name });
    }

    return options;
});

const templateName = computed(() => templateOptions.value.find(option => option.id === form.value.templateId)?.name ?? '');

function getTemplate(templateId: string): TransactionTemplate | undefined {
    if (!templateId || templateId === '0') {
        return undefined;
    }

    return templatesStore.allTransactionTemplatesMap[TemplateType.Normal.type]?.[templateId];
}

function categoryNameFromTemplate(template: TransactionTemplate | undefined): string {
    const categoryId = template?.getCategoryId();

    if (!categoryId || categoryId === '0') {
        return '';
    }

    const category = categoriesStore.allTransactionCategoriesMap[categoryId];

    if (!category) {
        return '';
    }

    const parentCategory = category.parentId && category.parentId !== '0' ? categoriesStore.allTransactionCategoriesMap[category.parentId] : undefined;
    return parentCategory ? `${parentCategory.name} > ${category.name}` : category.name;
}

interface AccountCoverage {
    covered: boolean;
    accountName: string;
}

function accountCoverage(subscription: Subscription, template: TransactionTemplate | undefined): AccountCoverage | undefined {
    if (!template || !template.sourceAccountId || template.sourceAccountId === '0') {
        return undefined;
    }

    const account = accountsStore.allPlainAccounts.find(a => a.id === template.sourceAccountId);

    if (!account) {
        return undefined;
    }

    const neededInAccountCurrency = account.currency === subscription.currency
        ? parseBigDecimal(subscription.amount)
        : exchangeRatesStore.getExchangedAmount(parseBigDecimal(subscription.amount), subscription.currency, account.currency);

    if (neededInAccountCurrency === null) {
        return undefined;
    }

    return { covered: parseBigDecimal(account.balance).greaterThanOrEqual(neededInAccountCurrency), accountName: account.name };
}

// ── List items ─────────────────────────────────────────────

const tableItems = computed(() => subscriptions.value.map(subscription => {
    const convertedAmount = subscription.currency !== defaultCurrency.value
        ? exchangeRatesStore.getExchangedAmount(parseBigDecimal(subscription.amount), subscription.currency, defaultCurrency.value)
        : null;
    const template = getTemplate(subscription.templateId);

    return {
        raw: subscription,
        id: subscription.id,
        name: subscription.name,
        amountDisplay: showAmountInSubscriptionsPage.value
            ? formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(subscription.amount), subscription.currency)
            : formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, subscription.currency),
        convertedAmountDisplay: subscription.currency !== defaultCurrency.value && convertedAmount !== null
            ? (showAmountInSubscriptionsPage.value
                ? formatAmountToLocalizedNumeralsWithCurrency(convertedAmount, defaultCurrency.value)
                : formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, defaultCurrency.value))
            : '',
        categoryName: categoryNameFromTemplate(template),
        accountCoverage: accountCoverage(subscription, template),
        templateId: subscription.templateId,
        frequencyLabel: frequencyLabel(subscription.frequency),
        nextExpectedDateDisplay: formatGregorianTextualYearMonthDayToLongDate(parseDateTimeFromUnixTime(getNextExpectedDate(subscription.startDate, subscription.frequency)).getGregorianCalendarYearDashMonthDashDay()),
        isActive: subscription.isActive,
    };
}));

// ── Popup helpers ──────────────────────────────────────────

function openAddPopup(): void {
    editingSubscription.value = null;
    form.value = defaultForm();
    showFormPopup.value = true;
}

function openEditPopup(subscription: Subscription): void {
    editingSubscription.value = subscription;
    form.value = {
        name: subscription.name,
        amountRaw: formatAmountToLocalizedNumeralsWithoutDigitGrouping(parseBigDecimal(subscription.amount), subscription.currency),
        currency: subscription.currency,
        templateId: subscription.templateId,
        startDate: parseDateTimeFromUnixTime(subscription.startDate).getGregorianCalendarYearDashMonthDashDay(),
        frequency: subscription.frequency,
        isActive: subscription.isActive,
    };
    showFormPopup.value = true;
}

function confirmDelete(subscription: Subscription): void {
    deletingSubscription.value = subscription;
    showDeleteActions.value = true;
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
            templateId: s.templateId,
            startDate: Number(s.startDate),
            frequency: s.frequency,
            isActive: s.isActive,
            createdAt: Number(s.createdAt),
        }));
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        loading.value = false;
    }
}

async function saveSubscription(): Promise<void> {
    if (saving.value) return;
    saving.value = true;
    try {
        const startDateObj = form.value.startDate ? getLocalDateFromYearDashMonthDashDay(form.value.startDate) : null;
        const startUnix = startDateObj ? getUnixTimeFromLocalDatetime(startDateObj) : getUnixTimeFromLocalDatetime(new Date());

        const payload = {
            name: form.value.name,
            amount: String(parseAmountFromLocalizedNumerals(form.value.amountRaw)),
            currency: form.value.currency,
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
        showFormPopup.value = false;
        await loadSubscriptions();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        saving.value = false;
    }
}

async function deleteSubscription(subscription: Subscription): Promise<void> {
    try {
        await axios.post('v1/subscriptions/delete.json', { id: String(subscription.id) });
        await loadSubscriptions();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    }
}

async function executeDelete(): Promise<void> {
    if (!deletingSubscription.value) return;
    showDeleteActions.value = false;
    await deleteSubscription(deletingSubscription.value);
    deletingSubscription.value = null;
}

// ── Boot ───────────────────────────────────────────────────

let initialized = false;

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
        await accountsStore.loadAllAccounts({ force: false });
    } catch {
        // accounts may already be loaded
    }
    try {
        await exchangeRatesStore.getLatestExchangeRates({ silent: true, force: false });
    } catch {
        // exchange rates may be unavailable, converted amounts will be hidden
    }
    await loadSubscriptions();
    initialized = true;
}

function onPageAfterIn(): void {
    if (!initialized) {
        if (isUserLogined() && isUserUnlocked()) {
            init();
        }
    }
}
</script>

<style>
.subscriptions-m-list-item .item-inner > .item-title {
    width: 100%;
}

.subscriptions-m-inactive {
    opacity: 0.5;
}

.subscriptions-m-nowrap {
    white-space: nowrap;
}

.subscriptions-m-row {
    display: flex;
    width: 100%;
    justify-content: space-between;
    align-items: baseline;
    gap: 8px;
}

.subscriptions-m-name {
    font-weight: 600;
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
}

.subscriptions-m-amount-col {
    display: flex;
    align-items: baseline;
    gap: 4px;
    flex-shrink: 0;
}

.subscriptions-m-primary-amount {
    font-size: 1.0625rem;
    font-weight: 600;
}

.subscriptions-m-converted {
    font-size: 0.8125rem;
    opacity: 0.6;
}

.subscriptions-m-secondary {
    font-size: 0.9375rem;
    opacity: 0.75;
    margin-top: 2px;
}

.subscriptions-m-next-date {
    color: var(--f7-theme-color);
    font-weight: 500;
}

.subscriptions-m-category {
    font-size: 0.8125rem;
    opacity: 0.7;
    margin-top: 2px;
}

.subscriptions-m-account-coverage {
    display: inline-flex;
    align-items: center;
    gap: 3px;
}
</style>
