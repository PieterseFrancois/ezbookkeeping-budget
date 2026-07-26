import { ref, computed } from 'vue';
import axios from 'axios';
import type { ApiResponse } from '@/core/api.ts';
import { useUserStore } from '@/stores/user.ts';

const HIDDEN_CATEGORIES_KEY = 'budget_hidden_categories';

export interface BudgetTargetEntry {
    id: string;
    amount: number;
}

interface RawBudgetTarget {
    id: string;
    categoryId: string;
    year: number;
    month: number;
    amount: string;
}

export interface CopyDecision {
    subcategoryId: string;
    parentCategoryId: string;
    amount: number;
    action: 'copy' | 'copy_unhide' | 'overwrite' | 'overwrite_unhide' | 'skip';
}

// Section names returned by the unified budget actuals endpoint (mirror pkg/models BUDGET_SECTION_*)
export type BudgetSection = 'income' | 'expense' | 'savings' | 'debt';

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

export function addMonths(year: number, month: number, delta: number): { year: number; month: number } {
    const d = new Date(year, month - 1 + delta, 1);
    return { year: d.getFullYear(), month: d.getMonth() + 1 };
}

export function useBudgetPageBase() {
    const now = new Date();
    const userStore = useUserStore();
    const endDay = userStore.currentUserBudgetEndDay;
    const activeMonth = (endDay > 0 && now.getDate() > endDay)
        ? addMonths(now.getFullYear(), now.getMonth() + 1, 1)
        : { year: now.getFullYear(), month: now.getMonth() + 1 };
    const selectedYear = ref<number>(activeMonth.year);
    const selectedMonth = ref<number>(activeMonth.month);
    const hiddenCategoryIds = ref<Set<string>>(new Set(loadHiddenIds()));
    // budgetTargets: outer key = `${year}-${month}`, inner key = subcategory id
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
            monthMap[t.categoryId] = { id: t.id, amount: Number(t.amount) };
        }
        budgetTargets.value[`${year}-${month}`] = monthMap;
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

    // Net savings for a transfer category = contributions (into savings) minus withdrawals (out, classified as income).
    function getSavingsNet(categoryId: string, year: number, month: number): number {
        const entry = budgetActuals.value[`${year}-${month}`]?.[categoryId];
        if (!entry) return 0;
        return (entry.savings ?? 0) - (entry.income ?? 0);
    }

    async function saveBudgetTarget(
        categoryId: string,
        year: number,
        month: number,
        amount: number
    ): Promise<void> {
        const key = `${year}-${month}`;
        const existing = budgetTargets.value[key]?.[categoryId];

        if (existing) {
            await axios.post<ApiResponse<RawBudgetTarget>>(
                'v1/budget/targets/modify.json',
                { id: existing.id, amount: String(amount) }
            );
            if (!budgetTargets.value[key]) budgetTargets.value[key] = {};
            budgetTargets.value[key]![categoryId] = { id: existing.id, amount };
        } else {
            const resp = await axios.post<ApiResponse<RawBudgetTarget>>(
                'v1/budget/targets/add.json',
                { categoryId, year, month, amount: String(amount) }
            );
            const created = resp.data?.result;
            if (created) {
                if (!budgetTargets.value[key]) budgetTargets.value[key] = {};
                budgetTargets.value[key]![categoryId] = { id: created.id, amount };
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
                selectedYear.value,
                selectedMonth.value,
                item.amount
            );
        }
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
        selectedYear,
        selectedMonth,
        hiddenCategoryIds,
        budgetTargets,
        budgetActuals,
        threeMonthColumns,
        selectMonth,
        loadBudgetTargets,
        loadBudgetActuals,
        getExpenseIncomeActual,
        getSavingsNet,
        cycleFirstUnixTime,
        cycleLastUnixTime,
        saveBudgetTarget,
        deleteBudgetTarget,
        copyBudgetFromMonth,
        toggleCategoryHidden,
        hideCategory,
        hideCategoryWithChildren,
        unhideCategory,
        unhideCategoryWithChildren,
    };
}
