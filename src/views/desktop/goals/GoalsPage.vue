<template>
    <main-page-layout>
        <template #nav-items>
            <div class="my-2">
                <span class="mx-3 text-body-medium">{{ tt('Status') }}</span>
                <v-select
                    item-title="displayName"
                    item-value="value"
                    class="mt-1"
                    density="compact"
                    :disabled="loading"
                    :items="[
                        { displayName: tt('All'), value: 0 },
                        { displayName: tt('In Progress'), value: 1 },
                        { displayName: tt('Goal Reached'), value: 2 },
                        { displayName: tt('Overdue'), value: 3 }
                    ]"
                    v-model="queryStatus"
                />
            </div>
            <div class="my-2">
                <span class="mx-3 text-body-medium">{{ tt('Account') }}</span>
                <v-select
                    item-title="name"
                    item-value="id"
                    class="mt-1"
                    density="compact"
                    :disabled="loading"
                    :items="[{ id: '', name: tt('All Accounts') }, ...availableAccounts]"
                    v-model="queryAccountId"
                />
            </div>
        </template>

        <template #content>
            <v-card min-height="920">
                <template #title>
                    <div class="title-and-toolbar d-flex align-center text-no-wrap">
                        <span>{{ tt('Goals') }}</span>
                        <v-spacer />
                        <v-btn
                            class="me-2 px-2"
                            min-width="0"
                            variant="tonal"
                            @click="showAmountInGoalsPage = !showAmountInGoalsPage"
                        >
                            <v-icon :icon="showAmountInGoalsPage ? mdiEyeOffOutline : mdiEyeOutline" size="20" />
                            <v-tooltip activator="parent">{{ showAmountInGoalsPage ? tt('Hide Amount') : tt('Show Amount') }}</v-tooltip>
                        </v-btn>
                        <v-btn :prepend-icon="mdiPlus" variant="tonal" @click="openAddDialog">
                            {{ tt('Add Goal') }}
                        </v-btn>
                    </div>
                </template>

                <v-card-text>
    <v-row class="match-height">
        <!-- Loading skeletons -->
        <template v-if="loading">
            <v-col v-for="i in 3" :key="i" cols="12" sm="6" lg="4">
                <v-skeleton-loader type="card" :loading="true" />
            </v-col>
        </template>

        <!-- Empty state -->
        <v-col v-else-if="filteredGoals.length === 0" cols="12" class="d-flex justify-center py-10">
            <span class="text-body-2 text-medium-emphasis">{{ tt('No goals yet. Add your first goal to get started.') }}</span>
        </v-col>

        <!-- Goal cards -->
        <v-col
            v-for="goal in filteredGoals"
            v-else
            :key="goal.id"
            cols="12"
            sm="6"
            lg="4"
        >
            <v-card class="goal-card h-100" elevation="1">
                <v-card-text class="d-flex flex-column gap-2 pa-4">
                    <!-- Name -->
                    <div class="d-flex align-center justify-space-between">
                        <span class="text-subtitle-1 font-weight-semibold text-truncate">{{ goal.name }}</span>
                        <div class="d-flex gap-1 flex-shrink-0 ms-2">
                            <v-btn density="compact" color="default" variant="text" :icon="true" @click="openEditDialog(goal)">
                                <v-icon :icon="mdiPencilOutline" />
                                <v-tooltip activator="parent">{{ tt('Edit Goal') }}</v-tooltip>
                            </v-btn>
                            <v-btn density="compact" color="error" variant="text" :icon="true" @click="confirmDelete(goal)">
                                <v-icon :icon="mdiTrashCanOutline" />
                                <v-tooltip activator="parent">{{ tt('Delete Goal') }}</v-tooltip>
                            </v-btn>
                        </div>
                    </div>

                    <!-- Account name & balance -->
                    <div class="text-body-2 text-medium-emphasis">
                        <span>{{ accountName(goal.accountId) }}</span>
                        <span class="ms-2 font-weight-medium" :class="currentBalanceColorClass(goal)">
                            {{ fmtBalance(goal) }}
                        </span>
                    </div>

                    <!-- Progress bar -->
                    <div class="mt-1">
                        <div class="d-flex justify-space-between text-caption mb-1">
                            <span>{{ fmtAmount(goal.targetAmount) }}</span>
                            <span>{{ progressPct(goal) }}%</span>
                        </div>
                        <v-progress-linear
                            :model-value="progressPct(goal)"
                            :color="progressColor(goal)"
                            rounded
                            height="8"
                        />
                    </div>

                    <!-- Goal Reached badge -->
                    <v-chip
                        v-if="isGoalReached(goal)"
                        color="success"
                        size="small"
                        variant="tonal"
                        prepend-icon="mdi-trophy-outline"
                        class="align-self-start mt-1"
                    >{{ tt('Goal Reached') }}</v-chip>

                    <!-- Target date & suggested contribution -->
                    <div class="text-body-2 mt-1">
                        <div class="d-flex justify-space-between">
                            <span class="text-medium-emphasis">{{ tt('Target Date') }}</span>
                            <span>{{ fmtTargetDate(goal.targetDate) }}</span>
                        </div>
                        <div class="d-flex justify-space-between mt-1">
                            <span class="text-medium-emphasis">{{ tt('Suggested Monthly Contribution') }}</span>
                            <span :class="isOverdue(goal) ? 'text-error' : ''">
                                {{ suggestionLabel(goal) }}
                            </span>
                        </div>
                    </div>

                    <!-- Muted future note -->
                    <div class="mt-2">
                        <v-chip size="x-small" variant="tonal" color="secondary" class="text-caption opacity-60">
                            {{ tt('Future: auto-link contribution to budget') }}
                        </v-chip>
                    </div>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>
                </v-card-text>
            </v-card>
        </template>
    </main-page-layout>

    <!-- Add / Edit Dialog -->
    <v-dialog v-model="showDialog" max-width="480" :persistent="saving">
        <v-card :title="editingGoal ? tt('Edit Goal') : tt('Add Goal')">
            <v-card-text class="d-flex flex-column gap-4 pt-2">
                <v-text-field
                    v-model="form.name"
                    :label="tt('Goal Name')"
                    density="compact"
                    hide-details
                    autofocus
                />
                <v-select
                    v-model="form.accountId"
                    :items="accountOptions"
                    item-title="name"
                    item-value="id"
                    :label="tt('Accounts')"
                    density="compact"
                    hide-details
                />
                <amount-input
                    v-model="form.targetAmount"
                    :label="tt('Target Amount')"
                    :currency="defaultCurrency"
                    :show-currency="true"
                    density="compact"
                />
                <div class="d-flex gap-3">
                    <v-select
                        v-model="form.targetYear"
                        :items="yearOptions"
                        :label="tt('Year')"
                        density="compact"
                        hide-details
                        class="flex-1"
                    />
                    <v-select
                        v-model="form.targetMonth"
                        :items="monthOptions"
                        item-title="label"
                        item-value="value"
                        :label="tt('Month')"
                        density="compact"
                        hide-details
                        class="flex-1"
                    />
                </div>
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn :disabled="saving" @click="showDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" :loading="saving" @click="saveGoal">{{ tt('Save') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="showDeleteDialog" max-width="420" :persistent="deleting">
        <v-card :title="tt('Delete Goal')">
            <v-card-text>{{ tt('Are you sure you want to delete this goal?') }}</v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn :disabled="deleting" @click="showDeleteDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="error" :loading="deleting" @click="deleteGoal">{{ tt('Confirm') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';
import AmountInput from '@/components/desktop/AmountInput.vue';
import { ref, computed, useTemplateRef } from 'vue';
import axios from 'axios';

import { useI18n } from '@/locales/helpers.ts';
import { parseBigDecimal } from '@/lib/numeral.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useUserStore } from '@/stores/user.ts';
import { useSettingsStore } from '@/stores/setting.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';
import type { ApiResponse } from '@/core/api.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';

import {
    mdiPlus,
    mdiPencilOutline,
    mdiTrashCanOutline,
    mdiEyeOutline,
    mdiEyeOffOutline,
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;
const snackbar = useTemplateRef<SnackBarType>('snackbar');

const {
    tt,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatDateTimeToGregorianLikeLongYearMonth,
} = useI18n();

const accountsStore = useAccountsStore();
const userStore = useUserStore();
const settingsStore = useSettingsStore();
const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

const showAmountInGoalsPage = computed<boolean>({
    get: () => settingsStore.appSettings.showAmountInGoalsPage,
    set: (value) => settingsStore.setShowAmountInGoalsPage(value)
});

// ── Types ──────────────────────────────────────────────────

interface RawGoal {
    id: string;
    name: string;
    accountId: string;
    targetAmount: string;
    targetDate: string;
    createdAt: string;
}

interface Goal {
    id: string;
    name: string;
    accountId: string;
    targetAmount: number;
    targetDate: number;
    createdAt: number;
}

// ── State ──────────────────────────────────────────────────

const loading = ref<boolean>(true);
const saving = ref<boolean>(false);
const deleting = ref<boolean>(false);
const goals = ref<Goal[]>([]);

const queryStatus = ref<number>(0);
const queryAccountId = ref<string>('');

const showDialog = ref<boolean>(false);
const showDeleteDialog = ref<boolean>(false);
const editingGoal = ref<Goal | null>(null);
const deletingGoal = ref<Goal | null>(null);

interface GoalForm {
    name: string;
    accountId: string;
    targetAmount: number;
    targetYear: number;
    targetMonth: number;
}

const now = new Date();

const form = ref<GoalForm>({
    name: '',
    accountId: '',
    targetAmount: 0,
    targetYear: now.getFullYear() + 1,
    targetMonth: now.getMonth() + 1,
});

// ── Account helpers ────────────────────────────────────────

const accountOptions = computed(() =>
    accountsStore.allPlainAccounts.map(a => ({ id: a.id, name: a.name }))
);

const availableAccounts = computed(() => {
    const seen = new Set<string>();
    const result: { id: string, name: string }[] = [];

    for (const goal of goals.value) {
        if (!seen.has(goal.accountId)) {
            seen.add(goal.accountId);
            result.push({ id: goal.accountId, name: accountName(goal.accountId) });
        }
    }

    return result;
});

function accountName(accountId: string): string {
    return accountsStore.allPlainAccounts.find(a => a.id === accountId)?.name ?? accountId;
}

function accountBalance(accountId: string): number {
    return Number(accountsStore.allPlainAccounts.find(a => a.id === accountId)?.balance ?? 0);
}

function accountCurrency(accountId: string): string {
    return accountsStore.allPlainAccounts.find(a => a.id === accountId)?.currency ?? defaultCurrency.value;
}

// ── Date / amount helpers ──────────────────────────────────

const MONTH_LABELS = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December',
];

const yearOptions = computed<number[]>(() => {
    const years: number[] = [];
    for (let y = now.getFullYear(); y <= now.getFullYear() + 15; y++) years.push(y);
    return years;
});

const monthOptions = computed(() =>
    MONTH_LABELS.map((label, i) => ({ label, value: i + 1 }))
);

function fmtAmount(cents: number): string {
    if (!showAmountInGoalsPage.value) {
        return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, defaultCurrency.value);
    }
    return formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(cents), defaultCurrency.value)
        .replace(/[,.]00$/, '');
}

function fmtBalance(goal: Goal): string {
    const bal = accountBalance(goal.accountId);
    const cur = accountCurrency(goal.accountId);
    if (!showAmountInGoalsPage.value) {
        return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, cur);
    }
    return formatAmountToLocalizedNumeralsWithCurrency(parseBigDecimal(bal), cur).replace(/[,.]00$/, '');
}

function fmtTargetDate(unixTs: number): string {
    return formatDateTimeToGregorianLikeLongYearMonth(parseDateTimeFromUnixTime(unixTs));
}

// ── Progress ───────────────────────────────────────────────

function progressPct(goal: Goal): number {
    if (goal.targetAmount <= 0) return 100;
    const bal = accountBalance(goal.accountId);
    return Math.min(100, Math.round((bal / goal.targetAmount) * 100));
}

function progressColor(goal: Goal): string {
    const pct = progressPct(goal);
    if (pct >= 100) return 'success';
    if (pct >= 80) return 'primary';
    if (pct >= 50) return 'warning';
    return 'error';
}

function currentBalanceColorClass(goal: Goal): string {
    const pct = progressPct(goal);
    if (pct >= 100) return 'text-success';
    if (pct >= 50) return 'text-warning';
    return 'text-error';
}

function isGoalReached(goal: Goal): boolean {
    return accountBalance(goal.accountId) >= goal.targetAmount;
}

// ── Months remaining & suggestion ─────────────────────────

function monthsRemaining(goal: Goal): number {
    const target = new Date(goal.targetDate * 1000);
    const targetYear = target.getFullYear();
    const targetMonth = target.getMonth();
    const diff = (targetYear - now.getFullYear()) * 12 + (targetMonth - now.getMonth());
    return diff;
}

function isOverdue(goal: Goal): boolean {
    return !isGoalReached(goal) && monthsRemaining(goal) <= 0;
}

const filteredGoals = computed(() => goals.value.filter(goal => {
    if (queryStatus.value === 1 && (isGoalReached(goal) || isOverdue(goal))) {
        return false;
    }

    if (queryStatus.value === 2 && !isGoalReached(goal)) {
        return false;
    }

    if (queryStatus.value === 3 && !isOverdue(goal)) {
        return false;
    }

    if (queryAccountId.value && goal.accountId !== queryAccountId.value) {
        return false;
    }

    return true;
}));

function suggestionLabel(goal: Goal): string {
    if (isGoalReached(goal)) return fmtAmount(0);
    const remaining = monthsRemaining(goal);
    if (remaining <= 0) return tt('Overdue');
    const bal = accountBalance(goal.accountId);
    const needed = goal.targetAmount - bal;
    const monthly = needed > 0 ? Math.round(needed / remaining) : 0;
    return fmtAmount(monthly);
}

// ── Dialog helpers ─────────────────────────────────────────

function openAddDialog(): void {
    editingGoal.value = null;
    form.value = {
        name: '',
        accountId: accountOptions.value[0]?.id ?? '',
        targetAmount: 0,
        targetYear: now.getFullYear() + 1,
        targetMonth: now.getMonth() + 1,
    };
    showDialog.value = true;
}

function openEditDialog(goal: Goal): void {
    editingGoal.value = goal;
    const d = new Date(goal.targetDate * 1000);
    form.value = {
        name: goal.name,
        accountId: goal.accountId,
        targetAmount: goal.targetAmount,
        targetYear: d.getFullYear(),
        targetMonth: d.getMonth() + 1,
    };
    showDialog.value = true;
}

function confirmDelete(goal: Goal): void {
    deletingGoal.value = goal;
    showDeleteDialog.value = true;
}

// ── API calls ──────────────────────────────────────────────

async function loadGoals(): Promise<void> {
    loading.value = true;
    try {
        const resp = await axios.get<ApiResponse<RawGoal[]>>('v1/goals/list.json');
        goals.value = (resp.data?.result ?? []).map(g => ({
            id: g.id,
            name: g.name,
            accountId: g.accountId,
            targetAmount: Number(g.targetAmount),
            targetDate: Number(g.targetDate),
            createdAt: Number(g.createdAt),
        }));
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        loading.value = false;
    }
}

async function saveGoal(): Promise<void> {
    if (saving.value) return;
    saving.value = true;
    try {
        const targetUnix = Math.floor(new Date(form.value.targetYear, form.value.targetMonth - 1, 1).getTime() / 1000);
        if (editingGoal.value) {
            await axios.post('v1/goals/modify.json', {
                id: String(editingGoal.value.id),
                name: form.value.name,
                accountId: String(form.value.accountId),
                targetAmount: String(form.value.targetAmount),
                targetDate: String(targetUnix),
            });
        } else {
            await axios.post('v1/goals/add.json', {
                name: form.value.name,
                accountId: String(form.value.accountId),
                targetAmount: String(form.value.targetAmount),
                targetDate: String(targetUnix),
            });
        }
        showDialog.value = false;
        await loadGoals();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        saving.value = false;
    }
}

async function deleteGoal(): Promise<void> {
    if (!deletingGoal.value || deleting.value) return;
    deleting.value = true;
    try {
        await axios.post('v1/goals/delete.json', { id: String(deletingGoal.value.id) });
        showDeleteDialog.value = false;
        await loadGoals();
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
        await accountsStore.loadAllAccounts({ force: false });
    } catch {
        // accounts may already be loaded
    }
    await loadGoals();
}

if (isUserLogined() && isUserUnlocked()) {
    init();
}
</script>

<style scoped>
.goal-card {
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    background:
        linear-gradient(180deg, rgba(var(--v-theme-primary), 0.045), transparent 13rem),
        rgba(var(--ebk-surface), 0.97);
}
</style>
