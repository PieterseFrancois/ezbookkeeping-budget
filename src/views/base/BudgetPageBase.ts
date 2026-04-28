import { ref, computed } from 'vue';
import axios from 'axios';
import type { ApiResponse } from '@/core/api.ts';

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

export interface SavingsCategoryActual {
    categoryId: string;
    transferOut: string;
    transferIn: string;
    net: string;
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
    const selectedYear = ref<number>(now.getFullYear());
    const selectedMonth = ref<number>(now.getMonth() + 1);
    const hiddenCategoryIds = ref<Set<string>>(new Set(loadHiddenIds()));
    // budgetTargets: outer key = `${year}-${month}`, inner key = subcategory id
    const budgetTargets = ref<Record<string, Record<string, BudgetTargetEntry>>>({});
    // savingsActuals: outer key = `${year}-${month}`, inner key = categoryId string
    const savingsActuals = ref<Record<string, Record<string, SavingsCategoryActual>>>({});

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

    async function loadSavingsActuals(year: number, month: number): Promise<void> {
        const resp = await axios.get<ApiResponse<{ items: SavingsCategoryActual[] }>>(
            `v1/budget/savings-actuals.json?year=${year}&month=${month}`
        );
        const items = resp.data?.result?.items ?? [];
        const monthMap: Record<string, SavingsCategoryActual> = {};
        for (const item of items) {
            monthMap[item.categoryId] = item;
        }
        savingsActuals.value[`${year}-${month}`] = monthMap;
    }

    function getSavingsNet(categoryId: string, year: number, month: number): number {
        const entry = savingsActuals.value[`${year}-${month}`]?.[categoryId];
        return entry ? Number(entry.net) : 0;
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
        savingsActuals,
        threeMonthColumns,
        selectMonth,
        loadBudgetTargets,
        loadSavingsActuals,
        getSavingsNet,
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
