<template>
    <v-row class="match-height">
        <!-- Header -->
        <v-col cols="12" class="d-flex align-center pb-2">
            <h5 class="text-h5">{{ tt('Budget') }}</h5>
            <v-spacer />
            <v-btn :prepend-icon="mdiContentCopy" variant="tonal" @click="openCopyDialog">
                {{ tt('Copy Budget') }}
            </v-btn>
        </v-col>

        <!-- Timeline strip -->
        <v-col cols="12" class="py-0">
            <div class="budget-timeline-wrap">
                <div class="budget-timeline">
                    <span
                        v-for="m in timelineMonths"
                        :key="`${m.year}-${m.month}`"
                        :ref="(el) => setChipRef(m, el)"
                        class="budget-timeline-chip"
                        :class="{ 'budget-timeline-chip--active': isSelectedMonth(m) }"
                        @click="onTimelineClick(m)"
                    >{{ formatTimelineLabel(m) }}</span>
                </div>
            </div>
        </v-col>

        <!-- Three-column grid -->
        <v-col cols="12" class="pt-3">
            <div class="budget-columns-grid">
                <v-card
                    v-for="(col, colIdx) in threeMonthColumns"
                    :key="`${col.year}-${col.month}`"
                    class="budget-month-card"
                    :class="{ 'budget-col-current': colIdx === 1 }"
                >
                    <template #title>
                        <span class="text-subtitle-1">{{ formatColumnTitle(col) }}</span>
                    </template>

                    <v-card-text class="pt-0">
                        <!-- Column summary header -->
                        <div class="budget-summary-section mb-2 pa-2 rounded">
                            <div class="d-flex align-center text-caption text-medium-emphasis mb-1">
                                <span class="budget-summary-name-cell"></span>
                                <span class="budget-amt-cell text-end">{{ tt('Budgeted') }}</span>
                                <span class="budget-amt-cell text-end">{{ tt('Actual') }}</span>
                                <span class="budget-amt-cell text-end">{{ tt('Difference') }}</span>
                            </div>

                            <!-- Expenses summary row (click to collapse/expand expense section) -->
                            <div
                                class="d-flex align-center budget-summary-row cursor-pointer py-1 rounded"
                                @click="showExpenseSection = !showExpenseSection"
                            >
                                <div class="budget-summary-name-cell d-flex align-center">
                                    <v-icon
                                        class="flex-shrink-0"
                                        :icon="showExpenseSection ? mdiChevronDown : mdiChevronRight"
                                        size="16"
                                    />
                                    <span class="text-body-2 font-weight-medium ms-1">{{ tt('Expenses') }}</span>
                                </div>
                                <span class="budget-amt-cell text-end text-body-2">{{ fmt(colExpenseBudgeted(col)) }}</span>
                                <span class="budget-amt-cell text-end text-body-2">{{ fmt(colExpenseActual(col)) }}</span>
                                <span
                                    class="budget-amt-cell text-end text-body-2"
                                    :class="diffClass(colExpenseDiff(col))"
                                >{{ fmt(colExpenseDiff(col)) }}</span>
                            </div>

                            <!-- Income summary row (click to collapse/expand income section) -->
                            <div
                                class="d-flex align-center budget-summary-row cursor-pointer py-1 rounded"
                                @click="showIncomeSection = !showIncomeSection"
                            >
                                <div class="budget-summary-name-cell d-flex align-center">
                                    <v-icon
                                        class="flex-shrink-0"
                                        :icon="showIncomeSection ? mdiChevronDown : mdiChevronRight"
                                        size="16"
                                    />
                                    <span class="text-body-2 font-weight-medium ms-1">{{ tt('Income') }}</span>
                                </div>
                                <span class="budget-amt-cell text-end text-body-2">{{ fmt(colIncomeBudgeted(col)) }}</span>
                                <span class="budget-amt-cell text-end text-body-2">{{ fmt(colIncomeActual(col)) }}</span>
                                <span
                                    class="budget-amt-cell text-end text-body-2"
                                    :class="diffClass(colIncomeDiff(col))"
                                >{{ fmt(colIncomeDiff(col)) }}</span>
                            </div>

                            <!-- Net row (always visible, not collapsible) -->
                            <div class="d-flex align-center py-1">
                                <div class="budget-summary-name-cell">
                                    <span class="text-body-2 font-weight-medium ps-5">{{ tt('Net') }}</span>
                                </div>
                                <span class="budget-amt-cell text-end text-body-2">{{ fmt(colNetBudgeted(col)) }}</span>
                                <span class="budget-amt-cell text-end text-body-2">{{ fmt(colNetActual(col)) }}</span>
                                <span
                                    class="budget-amt-cell text-end text-body-2"
                                    :class="diffClass(colNetDiff(col))"
                                >{{ fmt(colNetDiff(col)) }}</span>
                            </div>
                        </div>

                        <v-divider class="mb-2" />

                        <!-- Column sub-header labels -->
                        <div class="d-flex align-center mb-1 text-caption text-medium-emphasis">
                            <span class="budget-name-cell pe-1"></span>
                            <span class="budget-amt-cell text-end">{{ tt('Budgeted') }}</span>
                            <span class="budget-amt-cell text-end">{{ tt('Actual') }}</span>
                            <span class="budget-amt-cell text-end">{{ tt('Remaining') }}</span>
                            <span class="budget-eye-cell"></span>
                        </div>

                        <!-- Skeleton while loading -->
                        <template v-if="loading && !hasAnyData">
                            <div v-for="i in 4" :key="i" class="mb-2">
                                <v-skeleton-loader width="100%" type="text" :loading="true" />
                            </div>
                        </template>

                        <template v-else>
                            <!-- Expense section -->
                            <template v-if="showExpenseSection">
                                <div class="text-caption text-medium-emphasis mt-1 mb-1 ps-1 budget-section-label">
                                    {{ tt('Expenses') }}
                                </div>
                                <template v-for="parent in allExpenseParents" :key="parent.id">
                                    <template v-if="isParentVisibleInCol(parent, col)">
                                        <v-divider class="mb-1" />
                                        <div class="d-flex align-center budget-parent-row py-1">
                                            <v-btn
                                                density="compact"
                                                variant="text"
                                                size="x-small"
                                                class="flex-shrink-0 me-1"
                                                :icon="expandedParents.has(parent.id) ? mdiChevronDown : mdiChevronRight"
                                                @click="toggleExpanded(parent.id)"
                                            />
                                            <span class="budget-name-cell font-weight-bold text-body-2 text-truncate pe-1">
                                                {{ parent.name }}
                                            </span>
                                            <span class="budget-amt-cell text-end text-body-2">
                                                {{ fmt(parentBudgeted(parent, col)) }}
                                            </span>
                                            <span class="budget-amt-cell text-end text-body-2">
                                                {{ fmt(parentActual(parent, col)) }}
                                            </span>
                                            <span
                                                class="budget-amt-cell text-end text-body-2"
                                                :class="{ 'text-error': parentRemaining(parent, col) < 0 }"
                                            >{{ fmt(parentRemaining(parent, col)) }}</span>
                                            <div class="budget-eye-cell d-flex justify-center">
                                                <v-btn
                                                    v-if="!hiddenCategoryIds.has(parent.id) && parentBudgeted(parent, col) === 0"
                                                    density="compact"
                                                    variant="text"
                                                    size="x-small"
                                                    class="budget-eye-btn"
                                                    @click="onHideParent(parent)"
                                                >
                                                    <v-icon :icon="mdiEyeOff" size="16" />
                                                    <v-tooltip activator="parent">{{ tt('Hide') }}</v-tooltip>
                                                </v-btn>
                                            </div>
                                        </div>

                                        <template v-if="expandedParents.has(parent.id)">
                                            <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                                <div
                                                    v-if="isSubVisibleInCol(sub.id, col)"
                                                    class="d-flex align-center budget-sub-row py-1 ps-6"
                                                >
                                                    <span class="budget-name-cell text-body-2 text-medium-emphasis text-truncate pe-1">
                                                        {{ sub.name }}
                                                    </span>
                                                    <div class="budget-amt-cell d-flex justify-end align-center">
                                                        <v-text-field
                                                            v-if="isEditing(sub.id, col)"
                                                            v-model="editingText"
                                                            density="compact"
                                                            variant="plain"
                                                            hide-details
                                                            class="budget-edit-field"
                                                            autofocus
                                                            @focus="($event.target as HTMLInputElement).select()"
                                                            @keydown.enter="commitEdit"
                                                            @keydown.escape="cancelEdit"
                                                            @blur="commitEdit"
                                                        />
                                                        <span
                                                            v-else
                                                            class="cursor-pointer text-body-2 budget-budgeted-span"
                                                            :class="{ 'text-medium-emphasis': subBudgeted(sub.id, col) === 0 }"
                                                            @click="startEdit(sub.id, col)"
                                                        >{{ fmt(subBudgeted(sub.id, col)) }}</span>
                                                    </div>
                                                    <span class="budget-amt-cell text-end text-body-2">
                                                        {{ fmt(subActual(sub.id, col)) }}
                                                    </span>
                                                    <span
                                                        class="budget-amt-cell text-end text-body-2"
                                                        :class="{ 'text-error': subRemaining(sub.id, col) < 0 }"
                                                    >{{ fmt(subRemaining(sub.id, col)) }}</span>
                                                    <div class="budget-eye-cell d-flex justify-center">
                                                        <v-btn
                                                            v-if="!hiddenCategoryIds.has(sub.id) && subBudgeted(sub.id, col) === 0"
                                                            density="compact"
                                                            variant="text"
                                                            size="x-small"
                                                            class="budget-eye-btn"
                                                            @click="onHideSub(sub.id)"
                                                        >
                                                            <v-icon :icon="mdiEyeOff" size="16" />
                                                            <v-tooltip activator="parent">{{ tt('Hide') }}</v-tooltip>
                                                        </v-btn>
                                                    </div>
                                                </div>
                                            </template>
                                        </template>
                                    </template>
                                </template>
                            </template>

                            <v-divider v-if="showExpenseSection && showIncomeSection" class="my-3" />

                            <!-- Income section -->
                            <template v-if="showIncomeSection">
                                <div class="text-caption text-medium-emphasis mt-1 mb-1 ps-1 budget-section-label">
                                    {{ tt('Income') }}
                                </div>
                                <template v-for="parent in allIncomeParents" :key="parent.id">
                                    <template v-if="isParentVisibleInCol(parent, col)">
                                        <v-divider class="mb-1" />
                                        <div class="d-flex align-center budget-parent-row py-1">
                                            <v-btn
                                                density="compact"
                                                variant="text"
                                                size="x-small"
                                                class="flex-shrink-0 me-1"
                                                :icon="expandedParents.has(parent.id) ? mdiChevronDown : mdiChevronRight"
                                                @click="toggleExpanded(parent.id)"
                                            />
                                            <span class="budget-name-cell font-weight-bold text-body-2 text-truncate pe-1">
                                                {{ parent.name }}
                                            </span>
                                            <span class="budget-amt-cell text-end text-body-2">
                                                {{ fmt(parentBudgeted(parent, col)) }}
                                            </span>
                                            <span class="budget-amt-cell text-end text-body-2">
                                                {{ fmt(parentActual(parent, col)) }}
                                            </span>
                                            <span
                                                class="budget-amt-cell text-end text-body-2"
                                                :class="{ 'text-error': parentRemaining(parent, col) < 0 }"
                                            >{{ fmt(parentRemaining(parent, col)) }}</span>
                                            <div class="budget-eye-cell d-flex justify-center">
                                                <v-btn
                                                    v-if="!hiddenCategoryIds.has(parent.id) && parentBudgeted(parent, col) === 0"
                                                    density="compact"
                                                    variant="text"
                                                    size="x-small"
                                                    class="budget-eye-btn"
                                                    @click="onHideParent(parent)"
                                                >
                                                    <v-icon :icon="mdiEyeOff" size="16" />
                                                    <v-tooltip activator="parent">{{ tt('Hide') }}</v-tooltip>
                                                </v-btn>
                                            </div>
                                        </div>

                                        <template v-if="expandedParents.has(parent.id)">
                                            <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                                <div
                                                    v-if="isSubVisibleInCol(sub.id, col)"
                                                    class="d-flex align-center budget-sub-row py-1 ps-6"
                                                >
                                                    <span class="budget-name-cell text-body-2 text-medium-emphasis text-truncate pe-1">
                                                        {{ sub.name }}
                                                    </span>
                                                    <div class="budget-amt-cell d-flex justify-end align-center">
                                                        <v-text-field
                                                            v-if="isEditing(sub.id, col)"
                                                            v-model="editingText"
                                                            density="compact"
                                                            variant="plain"
                                                            hide-details
                                                            class="budget-edit-field"
                                                            autofocus
                                                            @focus="($event.target as HTMLInputElement).select()"
                                                            @keydown.enter="commitEdit"
                                                            @keydown.escape="cancelEdit"
                                                            @blur="commitEdit"
                                                        />
                                                        <span
                                                            v-else
                                                            class="cursor-pointer text-body-2 budget-budgeted-span"
                                                            :class="{ 'text-medium-emphasis': subBudgeted(sub.id, col) === 0 }"
                                                            @click="startEdit(sub.id, col)"
                                                        >{{ fmt(subBudgeted(sub.id, col)) }}</span>
                                                    </div>
                                                    <span class="budget-amt-cell text-end text-body-2">
                                                        {{ fmt(subActual(sub.id, col)) }}
                                                    </span>
                                                    <span
                                                        class="budget-amt-cell text-end text-body-2"
                                                        :class="{ 'text-error': subRemaining(sub.id, col) < 0 }"
                                                    >{{ fmt(subRemaining(sub.id, col)) }}</span>
                                                    <div class="budget-eye-cell d-flex justify-center">
                                                        <v-btn
                                                            v-if="!hiddenCategoryIds.has(sub.id) && subBudgeted(sub.id, col) === 0"
                                                            density="compact"
                                                            variant="text"
                                                            size="x-small"
                                                            class="budget-eye-btn"
                                                            @click="onHideSub(sub.id)"
                                                        >
                                                            <v-icon :icon="mdiEyeOff" size="16" />
                                                            <v-tooltip activator="parent">{{ tt('Hide') }}</v-tooltip>
                                                        </v-btn>
                                                    </div>
                                                </div>
                                            </template>
                                        </template>
                                    </template>
                                </template>
                            </template>

                            <!-- Empty state -->
                            <div
                                v-if="!hasAnyData"
                                class="d-flex align-center justify-center py-4"
                            >
                                <span class="text-body-2 text-medium-emphasis">
                                    {{ tt('No categories available to add') }}
                                </span>
                            </div>
                        </template>

                        <!-- Add Category button (center column only) -->
                        <div v-if="colIdx === 1 && hasHiddenItems" class="mt-3">
                            <v-btn
                                variant="tonal"
                                size="small"
                                :prepend-icon="mdiPlus"
                                @click="showAddCategoryDialog = true"
                            >
                                {{ tt('Add Category') }}
                            </v-btn>
                        </div>
                    </v-card-text>
                </v-card>
            </div>
        </v-col>
    </v-row>

    <!-- Copy Budget Dialog -->
    <v-dialog v-model="showCopyDialog" max-width="520" :persistent="copyLoading">
        <v-card :title="tt('Copy Budget')">
            <!-- Step 1: source month picker -->
            <v-card-text v-if="copyStep === 1">
                <div class="text-body-2 text-medium-emphasis mb-4">{{ tt('Select Source Month') }}</div>
                <div class="d-flex gap-4">
                    <v-select
                        v-model="copySourceYear"
                        :items="copyYearOptions"
                        :label="tt('Year')"
                        density="compact"
                        hide-details
                        class="flex-1"
                    />
                    <v-select
                        v-model="copySourceMonth"
                        :items="copyMonthOptions"
                        :label="tt('Month')"
                        density="compact"
                        hide-details
                        class="flex-1"
                        item-title="label"
                        item-value="value"
                    />
                </div>
            </v-card-text>

            <!-- Step 2: conflict resolution -->
            <v-card-text v-else>
                <div class="text-body-2 text-medium-emphasis mb-4">
                    {{ formatColumnTitle({ year: copySourceYear, month: copySourceMonth }) }}
                    &rarr;
                    {{ formatColumnTitle({ year: selectedYear, month: selectedMonth }) }}
                </div>

                <div
                    v-for="item in copyConflictItems"
                    :key="item.subcategoryId"
                    class="mb-4 pb-3 border-b"
                >
                    <!-- Category name prominent -->
                    <div class="text-subtitle-2 font-weight-bold mb-2">
                        {{ item.parentCategoryName }} › {{ item.subcategoryName }}
                    </div>

                    <!-- Hidden, no existing target -->
                    <template v-if="item.isHidden && !item.hasExistingTarget">
                        <div class="text-body-2 text-medium-emphasis mb-2">
                            {{ tt('This category is hidden in the destination month.') }}
                            {{ tt('Source Amount') }}: <strong>{{ fmt(item.amount) }}</strong>
                        </div>
                        <v-radio-group v-model="item.action" inline hide-details density="compact">
                            <v-radio :label="tt('Skip')" value="skip" />
                            <v-radio :label="tt('Copy and Unhide')" value="copy_unhide" />
                        </v-radio-group>
                    </template>

                    <!-- Visible, has existing target -->
                    <template v-else-if="!item.isHidden && item.hasExistingTarget">
                        <div class="text-body-2 text-medium-emphasis mb-2">
                            {{ tt('This category already has a budget in the destination month.') }}
                            {{ tt('Destination Amount') }}: <strong>{{ fmt(item.existingAmount) }}</strong>
                            &nbsp;{{ tt('Source Amount') }}: <strong>{{ fmt(item.amount) }}</strong>
                        </div>
                        <v-radio-group v-model="item.action" inline hide-details density="compact">
                            <v-radio :label="tt('Keep Existing')" value="skip" />
                            <v-radio :label="tt('Overwrite')" value="overwrite" />
                        </v-radio-group>
                    </template>

                    <!-- Hidden AND has existing target -->
                    <template v-else>
                        <div class="text-body-2 text-medium-emphasis mb-2">
                            {{ tt('This category is hidden in the destination month.') }}
                            {{ tt('Destination Amount') }}: <strong>{{ fmt(item.existingAmount) }}</strong>
                            &nbsp;{{ tt('Source Amount') }}: <strong>{{ fmt(item.amount) }}</strong>
                        </div>
                        <v-radio-group v-model="item.action" inline hide-details density="compact">
                            <v-radio :label="tt('Skip')" value="skip" />
                            <v-radio :label="tt('Overwrite and Unhide')" value="overwrite_unhide" />
                        </v-radio-group>
                    </template>
                </div>

                <div v-if="copyAutoItems.length > 0" class="text-caption text-medium-emphasis mt-2">
                    {{ copyAutoItems.length }} {{ tt('categories will be copied automatically') }}
                </div>
            </v-card-text>

            <v-card-actions>
                <v-spacer />
                <v-btn :disabled="copyLoading" @click="showCopyDialog = false">
                    {{ tt('Cancel') }}
                </v-btn>
                <v-btn
                    v-if="copyStep === 1"
                    color="primary"
                    :loading="copyLoading"
                    @click="advanceCopyStep"
                >{{ tt('Next') }}</v-btn>
                <v-btn
                    v-else
                    color="primary"
                    :loading="copyLoading"
                    @click="executeCopy"
                >{{ tt('Confirm') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Add Category Dialog -->
    <v-dialog v-model="showAddCategoryDialog" max-width="420">
        <v-card :title="tt('Add Category')">
            <v-card-text>
                <div
                    v-if="!hasHiddenItems"
                    class="text-body-2 text-medium-emphasis"
                >
                    {{ tt('No categories available to add') }}
                </div>
                <v-list v-else density="compact">
                    <!-- Hidden income parent categories -->
                    <template v-for="parent in hiddenIncomeParents" :key="'ip-' + parent.id">
                        <v-list-item
                            :title="parent.name"
                            :subtitle="tt('Income')"
                            class="cursor-pointer"
                            @click="onAddParent(parent)"
                        >
                            <template #prepend>
                                <v-icon :icon="mdiPlus" size="18" class="me-1" />
                            </template>
                        </v-list-item>
                    </template>

                    <!-- Hidden income sub-categories (parent is visible) -->
                    <template v-for="item in hiddenIncomeSubsUnderVisibleParent" :key="'is-' + item.sub.id">
                        <v-list-item
                            :title="item.sub.name"
                            :subtitle="item.parent.name + ' (' + tt('Income') + ')'"
                            class="cursor-pointer"
                            @click="onAddSub(item.sub.id)"
                        >
                            <template #prepend>
                                <v-icon :icon="mdiPlus" size="18" class="me-1" />
                            </template>
                        </v-list-item>
                    </template>

                    <!-- Hidden expense parent categories -->
                    <template v-for="parent in hiddenExpenseParents" :key="'ep-' + parent.id">
                        <v-list-item
                            :title="parent.name"
                            :subtitle="tt('Expenses')"
                            class="cursor-pointer"
                            @click="onAddParent(parent)"
                        >
                            <template #prepend>
                                <v-icon :icon="mdiPlus" size="18" class="me-1" />
                            </template>
                        </v-list-item>
                    </template>

                    <!-- Hidden expense sub-categories (parent is visible) -->
                    <template v-for="item in hiddenExpenseSubsUnderVisibleParent" :key="'es-' + item.sub.id">
                        <v-list-item
                            :title="item.sub.name"
                            :subtitle="item.parent.name + ' (' + tt('Expenses') + ')'"
                            class="cursor-pointer"
                            @click="onAddSub(item.sub.id)"
                        >
                            <template #prepend>
                                <v-icon :icon="mdiPlus" size="18" class="me-1" />
                            </template>
                        </v-list-item>
                    </template>
                </v-list>
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn @click="showAddCategoryDialog = false">{{ tt('Close') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';
import { ref, computed, watch, nextTick, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useBudgetPageBase, type CopyDecision, addMonths } from '@/views/base/BudgetPageBase.ts';

import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useUserStore } from '@/stores/user.ts';

import { CategoryType } from '@/core/category.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import services from '@/lib/services.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';

import {
    mdiChevronDown,
    mdiChevronRight,
    mdiEyeOff,
    mdiPlus,
    mdiContentCopy,
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;

const snackbar = useTemplateRef<SnackBarType>('snackbar');

const {
    tt,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatAmountToLocalizedNumerals,
    parseAmountFromLocalizedNumerals,
    formatDateTimeToGregorianLikeLongYearMonth,
    formatDateTimeToGregorianLikeShortYearMonth,
} = useI18n();

const {
    selectedYear,
    selectedMonth,
    hiddenCategoryIds,
    budgetTargets,
    threeMonthColumns,
    selectMonth,
    loadBudgetTargets,
    saveBudgetTarget,
    copyBudgetFromMonth,
    hideCategoryWithChildren,
    hideCategory,
    unhideCategory,
    unhideCategoryWithChildren,
} = useBudgetPageBase();

const categoriesStore = useTransactionCategoriesStore();
const userStore = useUserStore();

const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

// ---------- UI state ----------

const loading = ref<boolean>(true);
const spentByMonth = ref<Record<string, Record<string, number>>>({});
const expandedParents = ref<Set<string>>(new Set());
const saving = ref<boolean>(false);
const showIncomeSection = ref<boolean>(true);
const showExpenseSection = ref<boolean>(true);

interface EditingCell { categoryId: string; year: number; month: number; }
const editingCell = ref<EditingCell | null>(null);
const editingText = ref<string>('');

const showCopyDialog = ref<boolean>(false);
const copyStep = ref<1 | 2>(1);
const copySourceYear = ref<number>(new Date().getFullYear());
const copySourceMonth = ref<number>(new Date().getMonth() === 0 ? 12 : new Date().getMonth());
const copyLoading = ref<boolean>(false);

interface CopyItem {
    subcategoryId: string;
    subcategoryName: string;
    parentCategoryId: string;
    parentCategoryName: string;
    amount: number;
    existingAmount: number;
    isHidden: boolean;
    hasExistingTarget: boolean;
    action: CopyDecision['action'];
}
const copyItems = ref<CopyItem[]>([]);

const showAddCategoryDialog = ref<boolean>(false);

// ---------- Timeline chip refs ----------

const chipRefs = new Map<string, Element>();
function setChipRef(m: { year: number; month: number }, el: unknown): void {
    if (el) chipRefs.set(`${m.year}-${m.month}`, el as Element);
}

// ---------- Computed: categories ----------

const allExpenseParents = computed<TransactionCategory[]>(() =>
    (categoriesStore.allTransactionCategories[CategoryType.Expense] ?? []) as TransactionCategory[]
);

const allIncomeParents = computed<TransactionCategory[]>(() =>
    (categoriesStore.allTransactionCategories[CategoryType.Income] ?? []) as TransactionCategory[]
);

const hiddenExpenseParents = computed<TransactionCategory[]>(() =>
    allExpenseParents.value.filter(p => hiddenCategoryIds.value.has(p.id))
);

const hiddenIncomeParents = computed<TransactionCategory[]>(() =>
    allIncomeParents.value.filter(p => hiddenCategoryIds.value.has(p.id))
);

interface HiddenSubItem { sub: TransactionCategory; parent: TransactionCategory; }

const hiddenExpenseSubsUnderVisibleParent = computed<HiddenSubItem[]>(() => {
    const result: HiddenSubItem[] = [];
    for (const parent of allExpenseParents.value) {
        if (hiddenCategoryIds.value.has(parent.id)) continue;
        for (const sub of (parent.subCategories ?? [])) {
            if (hiddenCategoryIds.value.has(sub.id)) result.push({ sub, parent });
        }
    }
    return result;
});

const hiddenIncomeSubsUnderVisibleParent = computed<HiddenSubItem[]>(() => {
    const result: HiddenSubItem[] = [];
    for (const parent of allIncomeParents.value) {
        if (hiddenCategoryIds.value.has(parent.id)) continue;
        for (const sub of (parent.subCategories ?? [])) {
            if (hiddenCategoryIds.value.has(sub.id)) result.push({ sub, parent });
        }
    }
    return result;
});

const hasHiddenItems = computed<boolean>(() =>
    hiddenExpenseParents.value.length > 0 ||
    hiddenIncomeParents.value.length > 0 ||
    hiddenExpenseSubsUnderVisibleParent.value.length > 0 ||
    hiddenIncomeSubsUnderVisibleParent.value.length > 0
);

const hasAnyData = computed<boolean>(() =>
    allExpenseParents.value.length > 0 || allIncomeParents.value.length > 0
);

const nowDate = new Date();
const nowYear = nowDate.getFullYear();
const nowMonth = nowDate.getMonth() + 1;

const timelineMonths = computed<{ year: number; month: number }[]>(() => {
    const months: { year: number; month: number }[] = [];
    for (let delta = -24; delta <= 12; delta++) {
        months.push(addMonths(nowYear, nowMonth, delta));
    }
    return months;
});

const copyYearOptions = computed<number[]>(() => {
    const years: number[] = [];
    for (let y = nowYear - 3; y <= nowYear + 1; y++) years.push(y);
    return years;
});

const MONTH_LABELS = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December',
];
const copyMonthOptions = computed(() =>
    MONTH_LABELS.map((label, i) => ({ label, value: i + 1 }))
);

const copyConflictItems = computed<CopyItem[]>(() =>
    copyItems.value.filter(i => i.isHidden || i.hasExistingTarget)
);
const copyAutoItems = computed<CopyItem[]>(() =>
    copyItems.value.filter(i => !i.isHidden && !i.hasExistingTarget)
);

// ---------- Helpers ----------

function monthFirstUnixTime(year: number, month: number): number {
    return Math.floor(new Date(year, month - 1, 1, 0, 0, 0, 0).getTime() / 1000);
}

function monthLastUnixTime(year: number, month: number): number {
    return Math.floor(new Date(year, month, 1, 0, 0, 0, 0).getTime() / 1000) - 1;
}

function formatColumnTitle(col: { year: number; month: number }): string {
    return formatDateTimeToGregorianLikeLongYearMonth(
        parseDateTimeFromUnixTime(monthFirstUnixTime(col.year, col.month))
    );
}

function formatTimelineLabel(col: { year: number; month: number }): string {
    return formatDateTimeToGregorianLikeShortYearMonth(
        parseDateTimeFromUnixTime(monthFirstUnixTime(col.year, col.month))
    );
}

function isSelectedMonth(m: { year: number; month: number }): boolean {
    return m.year === selectedYear.value && m.month === selectedMonth.value;
}

function fmt(amount: number): string {
    const rands = Math.round(amount / 100) * 100;
    return formatAmountToLocalizedNumeralsWithCurrency(rands, defaultCurrency.value)
        .replace(/[,.]00$/, '');
}

function diffClass(diff: number): string {
    if (diff > 0) return 'text-success';
    if (diff < 0) return 'text-error';
    return '';
}

// ---------- Per-column visibility (change #4) ----------

function isSubVisibleInCol(subId: string, col: { year: number; month: number }): boolean {
    return !hiddenCategoryIds.value.has(subId) || subBudgeted(subId, col) > 0 || subActual(subId, col) > 0;
}

function isParentVisibleInCol(parent: TransactionCategory, col: { year: number; month: number }): boolean {
    if (!hiddenCategoryIds.value.has(parent.id)) return true;
    return (parent.subCategories ?? []).some(sub => subBudgeted(sub.id, col) > 0 || subActual(sub.id, col) > 0);
}

// ---------- Amount calculations ----------

function subBudgeted(subcatId: string, col: { year: number; month: number }): number {
    return budgetTargets.value[`${col.year}-${col.month}`]?.[subcatId]?.amount ?? 0;
}

function subActual(subcatId: string, col: { year: number; month: number }): number {
    return spentByMonth.value[`${col.year}-${col.month}`]?.[subcatId] ?? 0;
}

function subRemaining(subcatId: string, col: { year: number; month: number }): number {
    return subBudgeted(subcatId, col) - subActual(subcatId, col);
}

function parentBudgeted(parent: TransactionCategory, col: { year: number; month: number }): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subBudgeted(sub.id, col), 0);
}

function parentActual(parent: TransactionCategory, col: { year: number; month: number }): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subActual(sub.id, col), 0);
}

function parentRemaining(parent: TransactionCategory, col: { year: number; month: number }): number {
    return parentBudgeted(parent, col) - parentActual(parent, col);
}

// ---------- Column summary totals (change #6) ----------

function colIncomeBudgeted(col: { year: number; month: number }): number {
    return allIncomeParents.value
        .flatMap(p => p.subCategories ?? [])
        .reduce((sum, sub) => sum + subBudgeted(sub.id, col), 0);
}

function colIncomeActual(col: { year: number; month: number }): number {
    return allIncomeParents.value
        .flatMap(p => p.subCategories ?? [])
        .reduce((sum, sub) => sum + subActual(sub.id, col), 0);
}

function colIncomeDiff(col: { year: number; month: number }): number {
    return colIncomeBudgeted(col) - colIncomeActual(col);
}

function colExpenseBudgeted(col: { year: number; month: number }): number {
    return allExpenseParents.value
        .flatMap(p => p.subCategories ?? [])
        .reduce((sum, sub) => sum + subBudgeted(sub.id, col), 0);
}

function colExpenseActual(col: { year: number; month: number }): number {
    return allExpenseParents.value
        .flatMap(p => p.subCategories ?? [])
        .reduce((sum, sub) => sum + subActual(sub.id, col), 0);
}

function colExpenseDiff(col: { year: number; month: number }): number {
    return colExpenseBudgeted(col) - colExpenseActual(col);
}

function colNetBudgeted(col: { year: number; month: number }): number {
    return colIncomeBudgeted(col) - colExpenseBudgeted(col);
}

function colNetActual(col: { year: number; month: number }): number {
    return colIncomeActual(col) - colExpenseActual(col);
}

function colNetDiff(col: { year: number; month: number }): number {
    return colNetBudgeted(col) - colNetActual(col);
}

// ---------- Expand/collapse ----------

function toggleExpanded(parentId: string): void {
    const next = new Set(expandedParents.value);
    if (next.has(parentId)) {
        next.delete(parentId);
    } else {
        next.add(parentId);
    }
    expandedParents.value = next;
}

// ---------- Hide/show (change #2) ----------

function onHideParent(parent: TransactionCategory): void {
    const childIds = (parent.subCategories ?? []).map(s => s.id);
    hideCategoryWithChildren(parent.id, childIds);
}

function onHideSub(subId: string): void {
    hideCategory(subId);
}

function onAddParent(parent: TransactionCategory): void {
    const childIds = (parent.subCategories ?? []).map(s => s.id);
    unhideCategoryWithChildren(parent.id, childIds);
    showAddCategoryDialog.value = false;
}

function onAddSub(subId: string): void {
    unhideCategory(subId);
    showAddCategoryDialog.value = false;
}

// ---------- Inline edit ----------

function isEditing(subcatId: string, col: { year: number; month: number }): boolean {
    return editingCell.value?.categoryId === subcatId
        && editingCell.value?.year === col.year
        && editingCell.value?.month === col.month;
}

function startEdit(subcatId: string, col: { year: number; month: number }): void {
    editingText.value = formatAmountToLocalizedNumerals(subBudgeted(subcatId, col));
    editingCell.value = { categoryId: subcatId, year: col.year, month: col.month };
}

async function commitEdit(): Promise<void> {
    if (!editingCell.value || saving.value) return;
    const cell = { ...editingCell.value };
    const text = editingText.value;
    editingCell.value = null;
    const amount = parseAmountFromLocalizedNumerals(text);
    saving.value = true;
    try {
        await saveBudgetTarget(cell.categoryId, cell.year, cell.month, amount);
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        saving.value = false;
    }
}

function cancelEdit(): void {
    editingCell.value = null;
}

// ---------- Data loading ----------

async function loadStatsForMonth(year: number, month: number): Promise<void> {
    const resp = await services.getTransactionStatistics({
        startTime: monthFirstUnixTime(year, month),
        endTime: monthLastUnixTime(year, month),
        tagFilter: '',
        keyword: '',
        useTransactionTimezone: false,
    });
    const items = resp.data?.result?.items ?? [];
    const monthActual: Record<string, number> = {};
    for (const item of items) {
        monthActual[item.categoryId] = (monthActual[item.categoryId] ?? 0) + item.amount;
    }
    spentByMonth.value[`${year}-${month}`] = monthActual;
}

async function loadMonthData(year: number, month: number): Promise<void> {
    const key = `${year}-${month}`;
    const tasks: Promise<void>[] = [];
    if (!budgetTargets.value[key]) tasks.push(loadBudgetTargets(year, month));
    if (!spentByMonth.value[key]) tasks.push(loadStatsForMonth(year, month));
    await Promise.all(tasks);
}

function scrollToSelected(): void {
    nextTick(() => {
        const el = chipRefs.get(`${selectedYear.value}-${selectedMonth.value}`);
        if (el) el.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'nearest' });
    });
}

let initialized = false;

async function init(): Promise<void> {
    loading.value = true;
    try {
        await categoriesStore.loadAllCategories({ force: false });
        expandedParents.value = new Set([
            ...allExpenseParents.value.map(p => p.id),
            ...allIncomeParents.value.map(p => p.id),
        ]);
        await Promise.all(threeMonthColumns.value.map(col => loadMonthData(col.year, col.month)));
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        loading.value = false;
        initialized = true;
    }
    scrollToSelected();
}

watch(threeMonthColumns, async (cols) => {
    if (!initialized) return;
    try {
        await Promise.all(cols.map(col => loadMonthData(col.year, col.month)));
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    }
    scrollToSelected();
});

// ---------- Timeline ----------

function onTimelineClick(m: { year: number; month: number }): void {
    selectMonth(m.year, m.month);
}

// ---------- Copy dialog ----------

function openCopyDialog(): void {
    const prev = addMonths(selectedYear.value, selectedMonth.value, -1);
    copySourceYear.value = prev.year;
    copySourceMonth.value = prev.month;
    copyStep.value = 1;
    copyItems.value = [];
    showCopyDialog.value = true;
}

async function advanceCopyStep(): Promise<void> {
    copyLoading.value = true;
    try {
        await loadBudgetTargets(copySourceYear.value, copySourceMonth.value);
        const destKey = `${selectedYear.value}-${selectedMonth.value}`;
        if (!budgetTargets.value[destKey]) {
            await loadBudgetTargets(selectedYear.value, selectedMonth.value);
        }

        const sourceTargets = budgetTargets.value[`${copySourceYear.value}-${copySourceMonth.value}`] ?? {};
        const destTargets = budgetTargets.value[destKey] ?? {};

        const items: CopyItem[] = [];
        for (const [subcatId, entry] of Object.entries(sourceTargets)) {
            const subCat = categoriesStore.allTransactionCategoriesMap[subcatId];
            if (!subCat || !subCat.parentId || subCat.parentId === '0') continue;
            const parentCat = categoriesStore.allTransactionCategoriesMap[subCat.parentId];
            if (!parentCat) continue;

            const isHidden = hiddenCategoryIds.value.has(subcatId) || hiddenCategoryIds.value.has(parentCat.id);
            const hasExistingTarget = !!destTargets[subcatId];
            const existingAmount = destTargets[subcatId]?.amount ?? 0;

            items.push({
                subcategoryId: subcatId,
                subcategoryName: subCat.name,
                parentCategoryId: parentCat.id,
                parentCategoryName: parentCat.name,
                amount: entry.amount,
                existingAmount,
                isHidden,
                hasExistingTarget,
                action: (isHidden || hasExistingTarget) ? 'skip' : 'copy',
            });
        }
        copyItems.value = items;

        if (items.every(i => !i.isHidden && !i.hasExistingTarget)) {
            const decisions: CopyDecision[] = items.map(i => ({
                subcategoryId: i.subcategoryId,
                parentCategoryId: i.parentCategoryId,
                amount: i.amount,
                action: i.action,
            }));
            await copyBudgetFromMonth(copySourceYear.value, copySourceMonth.value, decisions);
            showCopyDialog.value = false;
            snackbar.value?.showMessage(tt('Budget copied successfully'));
        } else {
            copyStep.value = 2;
        }
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        copyLoading.value = false;
    }
}

async function executeCopy(): Promise<void> {
    copyLoading.value = true;
    try {
        const decisions: CopyDecision[] = copyItems.value.map(i => ({
            subcategoryId: i.subcategoryId,
            parentCategoryId: i.parentCategoryId,
            amount: i.amount,
            action: i.action,
        }));
        await copyBudgetFromMonth(copySourceYear.value, copySourceMonth.value, decisions);
        showCopyDialog.value = false;
        snackbar.value?.showMessage(tt('Budget copied successfully'));
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            snackbar.value?.showError(error as string);
        }
    } finally {
        copyLoading.value = false;
    }
}

// ---------- Boot ----------

if (isUserLogined() && isUserUnlocked()) {
    init();
}
</script>

<style>
.budget-timeline-wrap {
    overflow-x: auto;
    width: 100%;
    scrollbar-width: thin;
}

.budget-timeline {
    display: flex;
    gap: 8px;
    padding: 6px 2px;
    white-space: nowrap;
}

.budget-timeline-chip {
    display: inline-flex;
    align-items: center;
    padding: 4px 12px;
    border-radius: 16px;
    font-size: 0.8125rem;
    cursor: pointer;
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    transition: background-color 0.15s;
    user-select: none;
    flex-shrink: 0;
}

.budget-timeline-chip:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.14);
}

.budget-timeline-chip--active {
    background-color: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
}

.budget-timeline-chip--active:hover {
    filter: brightness(0.92);
}

.budget-columns-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 16px;
}

.budget-month-card {
    min-width: 260px;
}

.budget-col-current {
    border: 2px solid rgb(var(--v-theme-primary)) !important;
}

.budget-name-cell {
    flex: 1;
    min-width: 0;
}

.budget-amt-cell {
    width: 88px;
    flex-shrink: 0;
}

.budget-eye-cell {
    width: 32px;
    flex-shrink: 0;
}

.budget-summary-name-cell {
    flex: 1;
    min-width: 0;
}

.budget-summary-section {
    background-color: rgba(var(--v-theme-on-surface), 0.03);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
}

.budget-summary-row:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.04);
}

.budget-section-label {
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: rgba(var(--v-theme-on-surface), 0.45);
}

.budget-parent-row .budget-eye-btn {
    opacity: 0;
    transition: opacity 0.15s;
}

.budget-parent-row:hover .budget-eye-btn {
    opacity: 1;
}

.budget-sub-row .budget-eye-btn {
    opacity: 0;
    transition: opacity 0.15s;
}

.budget-sub-row:hover .budget-eye-btn {
    opacity: 1;
}

.budget-edit-field {
    width: 88px;
}

.budget-budgeted-span {
    display: block;
    text-align: right;
    width: 100%;
}
</style>
