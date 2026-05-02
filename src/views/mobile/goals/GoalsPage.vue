<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>{{ tt('Goals') }}</f7-nav-title>
            <f7-nav-right>
                <f7-link icon-f7="plus" @click="openAddPopup" />
            </f7-nav-right>
        </f7-navbar>

        <!-- Skeleton while loading -->
        <template v-if="loading">
            <f7-list strong inset dividers class="skeleton-text margin-top">
                <f7-list-item v-for="i in 3" :key="i">
                    <template #title>Loading goal name</template>
                    <template #footer>Loading account balance</template>
                </f7-list-item>
            </f7-list>
        </template>

        <!-- Empty state -->
        <div v-else-if="goals.length === 0" class="goals-m-empty">
            <f7-icon f7="flag_2" size="40" class="goals-m-empty-icon" />
            <p>{{ tt('No goals yet. Add your first goal to get started.') }}</p>
        </div>

        <!-- Goal cards -->
        <div v-else class="goals-m-cards">
            <f7-card v-for="goal in goals" :key="goal.id">
                <f7-card-content :padding="false">
                    <div class="goals-m-card">
                        <!-- Name + action buttons -->
                        <div class="goals-m-header">
                            <span class="goals-m-name">{{ goal.name }}</span>
                            <div class="goals-m-btn-row">
                                <f7-link icon-f7="pencil" @click="openEditPopup(goal)" />
                                <f7-link icon-f7="trash" color="red" @click="confirmDelete(goal)" />
                            </div>
                        </div>

                        <!-- Account + balance -->
                        <div class="goals-m-account">
                            <span>{{ accountName(goal.accountId) }}</span>
                            <span class="goals-m-balance" :class="balanceColorClass(goal)">
                                {{ fmtBalance(goal) }}
                            </span>
                        </div>

                        <!-- Progress bar -->
                        <div class="goals-m-progress-wrap">
                            <div class="goals-m-progress-track">
                                <div
                                    class="goals-m-progress-fill"
                                    :class="progressColorClass(goal)"
                                    :style="{ width: progressPct(goal) + '%' }"
                                />
                            </div>
                            <span class="goals-m-progress-pct">{{ progressPct(goal) }}%</span>
                        </div>

                        <!-- Goal Reached badge -->
                        <div v-if="isGoalReached(goal)" class="goals-m-reached-badge">
                            <f7-icon f7="checkmark_seal_fill" size="14" />
                            {{ tt('Goal Reached') }}
                        </div>

                        <!-- Target & suggestion -->
                        <div class="goals-m-meta">
                            <div class="goals-m-meta-row">
                                <span class="goals-m-label">{{ tt('Target Amount') }}</span>
                                <span>{{ fmtAmount(goal.targetAmount) }}</span>
                            </div>
                            <div class="goals-m-meta-row">
                                <span class="goals-m-label">{{ tt('Target Date') }}</span>
                                <span>{{ fmtTargetDate(goal.targetDate) }}</span>
                            </div>
                            <div class="goals-m-meta-row">
                                <span class="goals-m-label">{{ tt('Suggested Monthly Contribution') }}</span>
                                <span :class="isOverdue(goal) ? 'text-color-red' : ''">
                                    {{ suggestionLabel(goal) }}
                                </span>
                            </div>
                        </div>

                        <!-- Muted future note -->
                        <div class="goals-m-future-note">{{ tt('Future: auto-link contribution to budget') }}</div>
                    </div>
                </f7-card-content>
            </f7-card>
        </div>

        <!-- Add / Edit Popup -->
        <f7-popup v-model:opened="showFormPopup" tablet-fullscreen>
            <f7-page>
                <f7-navbar :title="editingGoal ? tt('Edit Goal') : tt('Add Goal')">
                    <f7-nav-right>
                        <f7-link popup-close :text="tt('Cancel')" />
                    </f7-nav-right>
                </f7-navbar>

                <f7-list strong inset dividers>
                    <f7-list-input
                        type="text"
                        :label="tt('Goal Name')"
                        :value="form.name"
                        @input="form.name = ($event.target as HTMLInputElement).value"
                        clear-button
                    />
                    <f7-list-input
                        type="select"
                        :label="tt('Accounts')"
                        @change="form.accountId = ($event.target as HTMLSelectElement).value"
                    >
                        <option
                            v-for="acc in accountOptions"
                            :key="acc.id"
                            :value="acc.id"
                            :selected="acc.id === form.accountId"
                        >{{ acc.name }}</option>
                    </f7-list-input>
                    <f7-list-input
                        type="number"
                        :label="tt('Target Amount')"
                        :value="form.targetAmountRaw"
                        min="1"
                        @input="form.targetAmountRaw = ($event.target as HTMLInputElement).value"
                    />
                    <f7-list-input
                        type="select"
                        :label="tt('Year')"
                        @change="form.targetYear = Number(($event.target as HTMLSelectElement).value)"
                    >
                        <option
                            v-for="y in yearOptions"
                            :key="y"
                            :value="y"
                            :selected="y === form.targetYear"
                        >{{ y }}</option>
                    </f7-list-input>
                    <f7-list-input
                        type="select"
                        :label="tt('Month')"
                        @change="form.targetMonth = Number(($event.target as HTMLSelectElement).value)"
                    >
                        <option
                            v-for="m in monthOptions"
                            :key="m.value"
                            :value="m.value"
                            :selected="m.value === form.targetMonth"
                        >{{ m.label }}</option>
                    </f7-list-input>
                </f7-list>

                <div class="margin">
                    <f7-button fill :loading="saving" @click="saveGoal">{{ tt('Save') }}</f7-button>
                </div>
            </f7-page>
        </f7-popup>

        <!-- Delete confirmation actions -->
        <f7-actions v-model:opened="showDeleteActions">
            <f7-actions-group>
                <f7-actions-label>{{ tt('Are you sure you want to delete this goal?') }}</f7-actions-label>
                <f7-actions-button color="red" @click="executeDelete">{{ tt('Delete Goal') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button @click="showDeleteActions = false">{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import axios from 'axios';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useAccountsStore } from '@/stores/account.ts';
import { useUserStore } from '@/stores/user.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';
import type { ApiResponse } from '@/core/api.ts';

const {
    tt,
    parseAmountFromLocalizedNumerals,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatAmountToLocalizedNumeralsWithoutDigitGrouping,
    formatDateTimeToGregorianLikeLongYearMonth,
} = useI18n();
const { showToast } = useI18nUIComponents();

const accountsStore = useAccountsStore();
const userStore = useUserStore();
const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

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
const goals = ref<Goal[]>([]);

const showFormPopup = ref<boolean>(false);
const showDeleteActions = ref<boolean>(false);
const editingGoal = ref<Goal | null>(null);
const deletingGoal = ref<Goal | null>(null);

interface GoalForm {
    name: string;
    accountId: string;
    targetAmountRaw: string;
    targetYear: number;
    targetMonth: number;
}

const now = new Date();

const form = ref<GoalForm>({
    name: '',
    accountId: '',
    targetAmountRaw: '',
    targetYear: now.getFullYear() + 1,
    targetMonth: now.getMonth() + 1,
});

// ── Account helpers ────────────────────────────────────────

const accountOptions = computed(() =>
    accountsStore.allPlainAccounts.map(a => ({ id: a.id, name: a.name }))
);

function accountName(accountId: string): string {
    return accountsStore.allPlainAccounts.find(a => a.id === accountId)?.name ?? accountId;
}

function accountBalance(accountId: string): number {
    return accountsStore.allPlainAccounts.find(a => a.id === accountId)?.balance ?? 0;
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
    return formatAmountToLocalizedNumeralsWithCurrency(cents, defaultCurrency.value)
        .replace(/[,.]00$/, '');
}

function fmtBalance(goal: Goal): string {
    const bal = accountBalance(goal.accountId);
    const cur = accountCurrency(goal.accountId);
    return formatAmountToLocalizedNumeralsWithCurrency(bal, cur).replace(/[,.]00$/, '');
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

function progressColorClass(goal: Goal): string {
    const pct = progressPct(goal);
    if (pct >= 100) return 'goals-m-progress-green';
    if (pct >= 80) return 'goals-m-progress-blue';
    if (pct >= 50) return 'goals-m-progress-amber';
    return 'goals-m-progress-red';
}

function balanceColorClass(goal: Goal): string {
    const pct = progressPct(goal);
    if (pct >= 100) return 'text-color-green';
    if (pct >= 50) return 'text-color-orange';
    return 'text-color-red';
}

function isGoalReached(goal: Goal): boolean {
    return accountBalance(goal.accountId) >= goal.targetAmount;
}

// ── Months remaining & suggestion ─────────────────────────

function monthsRemaining(goal: Goal): number {
    const target = new Date(goal.targetDate * 1000);
    const diff = (target.getFullYear() - now.getFullYear()) * 12 + (target.getMonth() - now.getMonth());
    return diff;
}

function isOverdue(goal: Goal): boolean {
    return !isGoalReached(goal) && monthsRemaining(goal) <= 0;
}

function suggestionLabel(goal: Goal): string {
    if (isGoalReached(goal)) return fmtAmount(0);
    const remaining = monthsRemaining(goal);
    if (remaining <= 0) return tt('Overdue');
    const needed = goal.targetAmount - accountBalance(goal.accountId);
    const monthly = needed > 0 ? Math.round(needed / remaining) : 0;
    return fmtAmount(monthly);
}

// ── Popup helpers ──────────────────────────────────────────

function openAddPopup(): void {
    editingGoal.value = null;
    form.value = {
        name: '',
        accountId: accountOptions.value[0]?.id ?? '',
        targetAmountRaw: '',
        targetYear: now.getFullYear() + 1,
        targetMonth: now.getMonth() + 1,
    };
    showFormPopup.value = true;
}

function openEditPopup(goal: Goal): void {
    editingGoal.value = goal;
    const d = new Date(goal.targetDate * 1000);
    form.value = {
        name: goal.name,
        accountId: goal.accountId,
        targetAmountRaw: formatAmountToLocalizedNumeralsWithoutDigitGrouping(goal.targetAmount, defaultCurrency.value),
        targetYear: d.getFullYear(),
        targetMonth: d.getMonth() + 1,
    };
    showFormPopup.value = true;
}

function confirmDelete(goal: Goal): void {
    deletingGoal.value = goal;
    showDeleteActions.value = true;
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
            showToast((error as Error).message || String(error));
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
                targetAmount: String(parseAmountFromLocalizedNumerals(form.value.targetAmountRaw)),
                targetDate: String(targetUnix),
            });
        } else {
            await axios.post('v1/goals/add.json', {
                name: form.value.name,
                accountId: String(form.value.accountId),
                targetAmount: String(parseAmountFromLocalizedNumerals(form.value.targetAmountRaw)),
                targetDate: String(targetUnix),
            });
        }
        showFormPopup.value = false;
        await loadGoals();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        saving.value = false;
    }
}

async function deleteGoal(goal: Goal): Promise<void> {
    try {
        await axios.post('v1/goals/delete.json', { id: String(goal.id) });
        await loadGoals();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    }
}

async function executeDelete(): Promise<void> {
    if (!deletingGoal.value) return;
    showDeleteActions.value = false;
    await deleteGoal(deletingGoal.value);
    deletingGoal.value = null;
}

// ── Boot ───────────────────────────────────────────────────

let initialized = false;

async function init(): Promise<void> {
    try {
        await accountsStore.loadAllAccounts({ force: false });
    } catch {
        // accounts may already be loaded
    }
    await loadGoals();
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
.goals-m-empty {
    padding: 48px 24px;
    text-align: center;
    opacity: 0.5;
    font-size: 0.9rem;
}

.goals-m-empty-icon {
    display: block;
    margin: 0 auto 12px;
}

.goals-m-cards {
    padding-top: 8px;
}

.goals-m-card {
    padding: 12px 16px;
}

.goals-m-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 4px;
}

.goals-m-name {
    font-size: 1rem;
    font-weight: 600;
    flex: 1;
    min-width: 0;
}

.goals-m-btn-row {
    display: flex;
    gap: 12px;
    flex-shrink: 0;
    align-items: center;
}

.goals-m-account {
    font-size: 0.8125rem;
    opacity: 0.7;
    display: flex;
    justify-content: space-between;
    gap: 8px;
    margin-bottom: 10px;
}

.goals-m-balance {
    font-weight: 500;
    opacity: 1;
    flex-shrink: 0;
}

.goals-m-progress-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
}

.goals-m-progress-track {
    flex: 1;
    height: 6px;
    border-radius: 3px;
    background-color: rgba(0, 0, 0, 0.12);
    overflow: hidden;
}

.goals-m-progress-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s ease;
}

.goals-m-progress-red    { background-color: var(--f7-color-red); }
.goals-m-progress-amber  { background-color: var(--f7-color-orange); }
.goals-m-progress-blue   { background-color: var(--f7-theme-color); }
.goals-m-progress-green  { background-color: var(--f7-color-green); }

.goals-m-progress-pct {
    font-size: 0.75rem;
    width: 36px;
    text-align: right;
    flex-shrink: 0;
    opacity: 0.6;
}

.goals-m-reached-badge {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--f7-color-green);
    margin-bottom: 8px;
}

.goals-m-meta {
    margin-top: 4px;
}

.goals-m-meta-row {
    display: flex;
    justify-content: space-between;
    gap: 8px;
    font-size: 0.8125rem;
    padding: 2px 0;
    flex-wrap: wrap;
}

.goals-m-label {
    opacity: 0.55;
    flex-shrink: 1;
    min-width: 0;
}

.goals-m-future-note {
    margin-top: 10px;
    font-size: 0.7rem;
    opacity: 0.4;
    font-style: italic;
}
</style>
