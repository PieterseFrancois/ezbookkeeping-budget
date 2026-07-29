import { ref, computed } from 'vue';
import axios from 'axios';
import type { ApiResponse } from '@/core/api.ts';
import { useUserStore } from '@/stores/user.ts';
import { useSettingsStore } from '@/stores/setting.ts';

const HIDDEN_CATEGORIES_KEY = 'budget_hidden_categories';
// Withdrawal rows are opt-in (the inverse of normal categories, which are opt-out), so they
// need their own set rather than reusing the hidden-category list.
const WITHDRAWAL_CATEGORIES_KEY = 'budget_withdrawal_categories';

export interface BudgetTargetEntry {
    id: string;
    amount: number;
}

interface RawBudgetTarget {
    id: string;
    categoryId: string;
    section: BudgetSection;
    year: number;
    month: number;
    amount: string;
}

export interface CopyDecision {
    subcategoryId: string;
    section: BudgetSection;
    parentCategoryId: string;
    amount: number;
    action: 'copy' | 'copy_unhide' | 'overwrite' | 'overwrite_unhide' | 'skip';
}

// Section names returned by the unified budget actuals endpoint (mirror pkg/models BUDGET_SECTION_*)
export type BudgetSection = 'income' | 'expense' | 'savings' | 'debt';

// Targets are keyed per (category, section) so one category can hold independent targets —
// e.g. a savings category budgeting both a contribution and an expected withdrawal.
export function targetKey(categoryId: string, section: BudgetSection): string {
    return `${categoryId}|${section}`;
}

// Per-category actual amounts, split by section. A transfer category can appear in more than one section
// (e.g. contributions land in "savings" while withdrawals land in "income").
export type CategoryActuals = Partial<Record<BudgetSection, number>>;

interface RawBudgetActualItem {
    categoryId: string;
    section: BudgetSection;
    amount: number;
}

function loadHiddenIds(): string[] {
    try {
        const raw = localStorage.getItem(HIDDEN_CATEGORIES_KEY);
        return raw ? (JSON.parse(raw) as string[]) : [];
    } catch {
        return [];
    }
}

function persistHiddenIds(ids: Set<string>): void {
    localStorage.setItem(HIDDEN_CATEGORIES_KEY, JSON.stringify([...ids]));
}

function loadWithdrawalIds(): string[] {
    try {
        const raw = localStorage.getItem(WITHDRAWAL_CATEGORIES_KEY);
        return raw ? (JSON.parse(raw) as string[]) : [];
    } catch {
        return [];
    }
}

function persistWithdrawalIds(ids: Set<string>): void {
    localStorage.setItem(WITHDRAWAL_CATEGORIES_KEY, JSON.stringify([...ids]));
}

export function addMonths(year: number, month: number, delta: number): { year: number; month: number } {
    const d = new Date(year, month - 1 + delta, 1);
    return { year: d.getFullYear(), month: d.getMonth() + 1 };
}

export function useBudgetPageBase() {
    const now = new Date();
    const userStore = useUserStore();
    const settingsStore = useSettingsStore();
    const showAmountInBudgetPage = computed<boolean>({
        get: () => settingsStore.appSettings.showAmountInBudgetPage,
        set: (value) => settingsStore.setShowAmountInBudgetPage(value)
    });
    const endDay = userStore.currentUserBudgetEndDay;
    const activeMonth = (endDay > 0 && now.getDate() > endDay)
        ? addMonths(now.getFullYear(), now.getMonth() + 1, 1)
        : { year: now.getFullYear(), month: now.getMonth() + 1 };
    const selectedYear = ref<number>(activeMonth.year);
    const selectedMonth = ref<number>(activeMonth.month);
    const hiddenCategoryIds = ref<Set<string>>(new Set(loadHiddenIds()));
    // Savings categories the user has explicitly surfaced as budgetable withdrawal (income) rows
    const withdrawalCategoryIds = ref<Set<string>>(new Set(loadWithdrawalIds()));
    // budgetTargets: outer key = `${year}-${month}`, inner key = targetKey(categoryId, section)
    const budgetTargets = ref<Record<string, Record<string, BudgetTargetEntry>>>({});
    // budgetActuals: outer key = `${year}-${month}`, inner key = categoryId, value = per-section amounts
    const budgetActuals = ref<Record<string, Record<string, CategoryActuals>>>({});

    const threeMonthColumns = computed<{ year: number; month: number }[]>(() => [
        addMonths(selectedYear.value, selectedMonth.value, -1),
        { year: selectedYear.value, month: selectedMonth.value },
        addMonths(selectedYear.value, selectedMonth.value, 1),
    ]);

    function selectMonth(year: number, month: number): void {
        selectedYear.value = year;
        selectedMonth.value = month;
    }

    async function loadBudgetTargets(year: number, month: number): Promise<void> {
        const resp = await axios.get<ApiResponse<RawBudgetTarget[]>>(
            `v1/budget/targets.json?year=${year}&month=${month}`
        );
        const targets = resp.data?.result ?? [];
        const monthMap: Record<string, BudgetTargetEntry> = {};
        for (const t of targets) {
            monthMap[targetKey(t.categoryId, t.section)] = { id: t.id, amount: Number(t.amount) };
        }
        budgetTargets.value[`${year}-${month}`] = monthMap;
    }

    // Budgeted amount for a category within a specific section (0 when no target is set).
    function getTargetAmount(categoryId: string, section: BudgetSection, year: number, month: number): number {
        return budgetTargets.value[`${year}-${month}`]?.[targetKey(categoryId, section)]?.amount ?? 0;
    }

    function monthFirstUnixTime(year: number, month: number): number {
        return Math.floor(new Date(year, month - 1, 1, 0, 0, 0, 0).getTime() / 1000);
    }

    function monthLastUnixTime(year: number, month: number): number {
        return Math.floor(new Date(year, month, 1, 0, 0, 0, 0).getTime() / 1000) - 1;
    }

    // The budget cycle runs from (endDay+1) of the previous month to endDay of the given month (0 = calendar month).
    function cycleFirstUnixTime(year: number, month: number): number {
        const ed = userStore.currentUserBudgetEndDay;
        if (!ed) return monthFirstUnixTime(year, month);
        const { year: prevYear, month: prevMonth } = addMonths(year, month, -1);
        return Math.floor(new Date(prevYear, prevMonth - 1, ed + 1, 0, 0, 0, 0).getTime() / 1000);
    }

    function cycleLastUnixTime(year: number, month: number): number {
        const ed = userStore.currentUserBudgetEndDay;
        if (!ed) return monthLastUnixTime(year, month);
        return Math.floor(new Date(year, month - 1, ed + 1, 0, 0, 0, 0).getTime() / 1000) - 1;
    }

    async function loadBudgetActuals(year: number, month: number): Promise<void> {
        const resp = await axios.get<ApiResponse<{ items: RawBudgetActualItem[] }>>(
            'v1/budget/actuals.json',
            { params: { startTime: cycleFirstUnixTime(year, month), endTime: cycleLastUnixTime(year, month) } }
        );
        const items = resp.data?.result?.items ?? [];
        const monthMap: Record<string, CategoryActuals> = {};
        for (const item of items) {
            const entry = monthMap[item.categoryId] ?? (monthMap[item.categoryId] = {});
            entry[item.section] = (entry[item.section] ?? 0) + item.amount;
        }
        budgetActuals.value[`${year}-${month}`] = monthMap;
    }

    // Actual spent/received for an expense or income category (both stored positive).
    function getExpenseIncomeActual(categoryId: string, year: number, month: number): number {
        const entry = budgetActuals.value[`${year}-${month}`]?.[categoryId];
        if (!entry) return 0;
        return (entry.expense ?? 0) + (entry.income ?? 0);
    }

    // Actual amount for a category within a specific section (income/expense/savings/debt).
    function getSectionActual(categoryId: string, section: BudgetSection, year: number, month: number): number {
        return budgetActuals.value[`${year}-${month}`]?.[categoryId]?.[section] ?? 0;
    }

    async function saveBudgetTarget(
        categoryId: string,
        section: BudgetSection,
        year: number,
        month: number,
        amount: number
    ): Promise<void> {
        const key = `${year}-${month}`;
        const entryKey = targetKey(categoryId, section);
        const existing = budgetTargets.value[key]?.[entryKey];

        if (existing) {
            await axios.post<ApiResponse<RawBudgetTarget>>(
                'v1/budget/targets/modify.json',
                { id: existing.id, amount: String(amount) }
            );
            if (!budgetTargets.value[key]) budgetTargets.value[key] = {};
            budgetTargets.value[key]![entryKey] = { id: existing.id, amount };
        } else {
            const resp = await axios.post<ApiResponse<RawBudgetTarget>>(
                'v1/budget/targets/add.json',
                { categoryId, section, year, month, amount: String(amount) }
            );
            const created = resp.data?.result;
            if (created) {
                if (!budgetTargets.value[key]) budgetTargets.value[key] = {};
                budgetTargets.value[key]![entryKey] = { id: created.id, amount };
            }
        }
    }

    async function deleteBudgetTarget(id: string): Promise<void> {
        await axios.post('v1/budget/targets/delete.json', { id });
        for (const key of Object.keys(budgetTargets.value)) {
            const monthMap = budgetTargets.value[key];
            if (!monthMap) continue;
            for (const catId of Object.keys(monthMap)) {
                if (monthMap[catId]?.id === id) {
                    delete monthMap[catId];
                    break;
                }
            }
        }
    }

    async function copyBudgetFromMonth(
        _sourceYear: number,
        _sourceMonth: number,
        decisions: CopyDecision[]
    ): Promise<void> {
        for (const item of decisions) {
            if (item.action === 'skip') continue;
            if (item.action === 'copy_unhide' || item.action === 'overwrite_unhide') {
                const next = new Set(hiddenCategoryIds.value);
                next.delete(item.parentCategoryId);
                next.delete(item.subcategoryId);
                hiddenCategoryIds.value = next;
                persistHiddenIds(next);
            }
            await saveBudgetTarget(
                item.subcategoryId,
                item.section,
                selectedYear.value,
                selectedMonth.value,
                item.amount
            );
        }
    }

    function addWithdrawalCategory(categoryId: string): void {
        const next = new Set(withdrawalCategoryIds.value);
        next.add(categoryId);
        withdrawalCategoryIds.value = next;
        persistWithdrawalIds(next);
    }

    function removeWithdrawalCategory(categoryId: string): void {
        const next = new Set(withdrawalCategoryIds.value);
        next.delete(categoryId);
        withdrawalCategoryIds.value = next;
        persistWithdrawalIds(next);
    }

    function toggleCategoryHidden(categoryId: string): void {
        const next = new Set(hiddenCategoryIds.value);
        if (next.has(categoryId)) {
            next.delete(categoryId);
        } else {
            next.add(categoryId);
        }
        hiddenCategoryIds.value = next;
        persistHiddenIds(next);
    }

    function hideCategory(categoryId: string): void {
        const next = new Set(hiddenCategoryIds.value);
        next.add(categoryId);
        hiddenCategoryIds.value = next;
        persistHiddenIds(next);
    }

    function hideCategoryWithChildren(parentId: string, childIds: string[]): void {
        const next = new Set(hiddenCategoryIds.value);
        next.add(parentId);
        for (const id of childIds) next.add(id);
        hiddenCategoryIds.value = next;
        persistHiddenIds(next);
    }

    function unhideCategory(categoryId: string): void {
        const next = new Set(hiddenCategoryIds.value);
        next.delete(categoryId);
        hiddenCategoryIds.value = next;
        persistHiddenIds(next);
    }

    function unhideCategoryWithChildren(parentId: string, childIds: string[]): void {
        const next = new Set(hiddenCategoryIds.value);
        next.delete(parentId);
        for (const id of childIds) next.delete(id);
        hiddenCategoryIds.value = next;
        persistHiddenIds(next);
    }

    return {
        showAmountInBudgetPage,
        selectedYear,
        selectedMonth,
        hiddenCategoryIds,
        withdrawalCategoryIds,
        budgetTargets,
        budgetActuals,
        threeMonthColumns,
        selectMonth,
        loadBudgetTargets,
        loadBudgetActuals,
        getExpenseIncomeActual,
        getSectionActual,
        getTargetAmount,
        cycleFirstUnixTime,
        cycleLastUnixTime,
        saveBudgetTarget,
        deleteBudgetTarget,
        copyBudgetFromMonth,
        addWithdrawalCategory,
        removeWithdrawalCategory,
        toggleCategoryHidden,
        hideCategory,
        hideCategoryWithChildren,
        unhideCategory,
        unhideCategoryWithChildren,
    };
}
