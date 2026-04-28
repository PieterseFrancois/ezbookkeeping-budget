<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>{{ tt('Budget') }}</f7-nav-title>
            <f7-nav-right>
                <f7-link icon-f7="doc_on_doc" :class="{ disabled: loading }" @click="openCopyPopup" />
            </f7-nav-right>
        </f7-navbar>

        <!-- Month timeline strip -->
        <div class="budget-m-timeline-wrap">
            <div class="budget-m-timeline">
                <span
                    v-for="m in timelineMonths"
                    :key="`${m.year}-${m.month}`"
                    :ref="(el) => setChipRef(m, el)"
                    class="budget-m-chip"
                    :class="{ 'budget-m-chip--active': isSelectedMonth(m) }"
                    @click="onTimelineClick(m)"
                >{{ formatTimelineLabel(m) }}</span>
            </div>
        </div>

        <!-- Summary card -->
        <f7-card class="budget-m-summary-card">
            <f7-card-content :padding="false">
                <!-- Column labels -->
                <div class="budget-m-row budget-m-header-row">
                    <span class="budget-m-name-cell"></span>
                    <span class="budget-m-amt-cell budget-m-col-label">{{ tt('Budgeted') }}</span>
                    <span class="budget-m-amt-cell budget-m-col-label">{{ tt('Actual') }}</span>
                    <span class="budget-m-amt-cell budget-m-col-label">{{ tt('Difference') }}</span>
                </div>
                <!-- Expenses row -->
                <div class="budget-m-row budget-m-summary-row" @click="showExpenseSection = !showExpenseSection">
                    <div class="budget-m-name-cell budget-m-summary-name">
                        <f7-icon :f7="showExpenseSection ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                        <span>{{ tt('Expenses') }}</span>
                    </div>
                    <span class="budget-m-amt-cell">{{ fmt(colExpenseBudgeted) }}</span>
                    <span class="budget-m-amt-cell">{{ fmt(colExpenseActual) }}</span>
                    <span class="budget-m-amt-cell" :class="diffClass(colExpenseDiff)">{{ fmt(colExpenseDiff) }}</span>
                </div>
                <!-- Income row -->
                <div class="budget-m-row budget-m-summary-row" @click="showIncomeSection = !showIncomeSection">
                    <div class="budget-m-name-cell budget-m-summary-name">
                        <f7-icon :f7="showIncomeSection ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                        <span>{{ tt('Income') }}</span>
                    </div>
                    <span class="budget-m-amt-cell">{{ fmt(colIncomeBudgeted) }}</span>
                    <span class="budget-m-amt-cell">{{ fmt(colIncomeActual) }}</span>
                    <span class="budget-m-amt-cell" :class="diffClass(colIncomeDiff)">{{ fmt(colIncomeDiff) }}</span>
                </div>
                <!-- Savings row -->
                <div class="budget-m-row budget-m-summary-row" @click="showSavingsSection = !showSavingsSection">
                    <div class="budget-m-name-cell budget-m-summary-name">
                        <f7-icon :f7="showSavingsSection ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                        <span>{{ tt('Savings') }}</span>
                    </div>
                    <span class="budget-m-amt-cell">{{ fmt(colSavingsBudgeted) }}</span>
                    <span class="budget-m-amt-cell">{{ fmt(colSavingsActual) }}</span>
                    <span class="budget-m-amt-cell" :class="savingsDiffClass(colSavingsDiff)">{{ fmt(colSavingsDiff) }}</span>
                </div>
                <!-- Net row -->
                <div class="budget-m-row">
                    <div class="budget-m-name-cell budget-m-net-name">{{ tt('Net') }}</div>
                    <span class="budget-m-amt-cell">{{ fmt(colNetBudgeted) }}</span>
                    <span class="budget-m-amt-cell">{{ fmt(colNetActual) }}</span>
                    <span class="budget-m-amt-cell" :class="diffClass(colNetDiff)">{{ fmt(colNetDiff) }}</span>
                </div>
            </f7-card-content>
        </f7-card>

        <!-- Category column header labels -->
        <div class="budget-m-row budget-m-cat-header">
            <span class="budget-m-name-cell"></span>
            <span class="budget-m-amt-cell budget-m-col-label">{{ tt('Budgeted') }}</span>
            <span class="budget-m-amt-cell budget-m-col-label">{{ tt('Actual') }}</span>
            <span class="budget-m-amt-cell budget-m-col-label">{{ tt('Remaining') }}</span>
            <span class="budget-m-eye-cell"></span>
        </div>

        <!-- Skeleton while loading -->
        <template v-if="loading && !hasAnyData">
            <f7-list strong inset dividers class="skeleton-text margin-top">
                <f7-list-item v-for="i in 5" :key="i">
                    <template #title>Loading budget category</template>
                    <template #after>0.00</template>
                </f7-list-item>
            </f7-list>
        </template>

        <template v-else>
            <!-- Expense section -->
            <template v-if="showExpenseSection">
                <div class="budget-m-section-label">{{ tt('Expenses') }}</div>
                <template v-for="parent in allExpenseParents" :key="parent.id">
                    <template v-if="isParentVisible(parent)">
                        <div class="budget-m-row budget-m-parent-row" @click="toggleExpanded(parent.id)">
                            <div class="budget-m-name-cell budget-m-parent-name">
                                <f7-icon :f7="expandedParents.has(parent.id) ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                                <span class="budget-m-parent-label">{{ parent.name }}</span>
                            </div>
                            <span class="budget-m-amt-cell">{{ fmt(parentBudgeted(parent)) }}</span>
                            <span class="budget-m-amt-cell">{{ fmt(parentActual(parent)) }}</span>
                            <span class="budget-m-amt-cell" :class="{ 'budget-m-negative': parentRemaining(parent) < 0 }">{{ fmt(parentRemaining(parent)) }}</span>
                            <div class="budget-m-eye-cell">
                                <f7-link
                                    v-if="!hiddenCategoryIds.has(parent.id) && parentBudgeted(parent) === 0"
                                    class="budget-m-eye-btn"
                                    @click.stop="onHideParent(parent)"
                                >
                                    <f7-icon f7="eye_slash" size="18" />
                                </f7-link>
                            </div>
                        </div>
                        <template v-if="expandedParents.has(parent.id)">
                            <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                <div
                                    v-if="isSubVisible(sub.id)"
                                    class="budget-m-row budget-m-sub-row"
                                >
                                    <span class="budget-m-name-cell budget-m-sub-name">{{ sub.name }}</span>
                                    <div class="budget-m-amt-cell budget-m-budgeted-cell" @click="startEdit(sub.id, sub.name)">
                                        <span :class="{ 'budget-m-zero': subBudgeted(sub.id) === 0 }">{{ fmt(subBudgeted(sub.id)) }}</span>
                                    </div>
                                    <span class="budget-m-amt-cell">{{ fmt(subActual(sub.id)) }}</span>
                                    <span class="budget-m-amt-cell" :class="{ 'budget-m-negative': subRemaining(sub.id) < 0 }">{{ fmt(subRemaining(sub.id)) }}</span>
                                    <div class="budget-m-eye-cell">
                                        <f7-link
                                            v-if="!hiddenCategoryIds.has(sub.id) && subBudgeted(sub.id) === 0"
                                            class="budget-m-eye-btn"
                                            @click.stop="onHideSub(sub.id)"
                                        >
                                            <f7-icon f7="eye_slash" size="18" />
                                        </f7-link>
                                    </div>
                                </div>
                            </template>
                        </template>
                    </template>
                </template>
            </template>

            <div v-if="showExpenseSection && showIncomeSection" class="budget-m-section-divider" />

            <!-- Income section -->
            <template v-if="showIncomeSection">
                <div class="budget-m-section-label">{{ tt('Income') }}</div>
                <template v-for="parent in allIncomeParents" :key="parent.id">
                    <template v-if="isParentVisible(parent)">
                        <div class="budget-m-row budget-m-parent-row" @click="toggleExpanded(parent.id)">
                            <div class="budget-m-name-cell budget-m-parent-name">
                                <f7-icon :f7="expandedParents.has(parent.id) ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                                <span class="budget-m-parent-label">{{ parent.name }}</span>
                            </div>
                            <span class="budget-m-amt-cell">{{ fmt(parentBudgeted(parent)) }}</span>
                            <span class="budget-m-amt-cell">{{ fmt(parentActual(parent)) }}</span>
                            <span class="budget-m-amt-cell" :class="{ 'budget-m-negative': parentRemaining(parent) < 0 }">{{ fmt(parentRemaining(parent)) }}</span>
                            <div class="budget-m-eye-cell">
                                <f7-link
                                    v-if="!hiddenCategoryIds.has(parent.id) && parentBudgeted(parent) === 0"
                                    class="budget-m-eye-btn"
                                    @click.stop="onHideParent(parent)"
                                >
                                    <f7-icon f7="eye_slash" size="18" />
                                </f7-link>
                            </div>
                        </div>
                        <template v-if="expandedParents.has(parent.id)">
                            <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                <div
                                    v-if="isSubVisible(sub.id)"
                                    class="budget-m-row budget-m-sub-row"
                                >
                                    <span class="budget-m-name-cell budget-m-sub-name">{{ sub.name }}</span>
                                    <div class="budget-m-amt-cell budget-m-budgeted-cell" @click="startEdit(sub.id, sub.name)">
                                        <span :class="{ 'budget-m-zero': subBudgeted(sub.id) === 0 }">{{ fmt(subBudgeted(sub.id)) }}</span>
                                    </div>
                                    <span class="budget-m-amt-cell">{{ fmt(subActual(sub.id)) }}</span>
                                    <span class="budget-m-amt-cell" :class="{ 'budget-m-negative': subRemaining(sub.id) < 0 }">{{ fmt(subRemaining(sub.id)) }}</span>
                                    <div class="budget-m-eye-cell">
                                        <f7-link
                                            v-if="!hiddenCategoryIds.has(sub.id) && subBudgeted(sub.id) === 0"
                                            class="budget-m-eye-btn"
                                            @click.stop="onHideSub(sub.id)"
                                        >
                                            <f7-icon f7="eye_slash" size="18" />
                                        </f7-link>
                                    </div>
                                </div>
                            </template>
                        </template>
                    </template>
                </template>
            </template>

            <div v-if="(showExpenseSection || showIncomeSection) && showSavingsSection" class="budget-m-section-divider" />

            <!-- Savings section -->
            <template v-if="showSavingsSection">
                <div class="budget-m-section-label">{{ tt('Savings') }}</div>
                <template v-for="parent in allTransferParents" :key="parent.id">
                    <template v-if="isSavingsParentVisible(parent)">
                        <div class="budget-m-row budget-m-parent-row" @click="toggleExpanded(parent.id)">
                            <div class="budget-m-name-cell budget-m-parent-name">
                                <f7-icon :f7="expandedParents.has(parent.id) ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                                <span class="budget-m-parent-label">{{ parent.name }}</span>
                            </div>
                            <span class="budget-m-amt-cell">{{ fmt(parentSavingsBudgeted(parent)) }}</span>
                            <span class="budget-m-amt-cell">{{ fmt(parentSavingsActual(parent)) }}</span>
                            <span class="budget-m-amt-cell" :class="savingsDiffClass(parentSavingsRemaining(parent))">{{ fmt(parentSavingsRemaining(parent)) }}</span>
                            <div class="budget-m-eye-cell">
                                <f7-link
                                    v-if="!hiddenCategoryIds.has(parent.id) && parentSavingsBudgeted(parent) === 0"
                                    class="budget-m-eye-btn"
                                    @click.stop="onHideParent(parent)"
                                >
                                    <f7-icon f7="eye_slash" size="18" />
                                </f7-link>
                            </div>
                        </div>
                        <template v-if="expandedParents.has(parent.id)">
                            <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                <div
                                    v-if="isSavingsSubVisible(sub.id)"
                                    class="budget-m-row budget-m-sub-row"
                                >
                                    <span class="budget-m-name-cell budget-m-sub-name">{{ sub.name }}</span>
                                    <div class="budget-m-amt-cell budget-m-budgeted-cell" @click="startEdit(sub.id, sub.name)">
                                        <span :class="{ 'budget-m-zero': subSavingsBudgeted(sub.id) === 0 }">{{ fmt(subSavingsBudgeted(sub.id)) }}</span>
                                    </div>
                                    <span class="budget-m-amt-cell">{{ fmt(subSavingsActual(sub.id)) }}</span>
                                    <span class="budget-m-amt-cell" :class="savingsDiffClass(subSavingsRemaining(sub.id))">{{ fmt(subSavingsRemaining(sub.id)) }}</span>
                                    <div class="budget-m-eye-cell">
                                        <f7-link
                                            v-if="!hiddenCategoryIds.has(sub.id) && subSavingsBudgeted(sub.id) === 0"
                                            class="budget-m-eye-btn"
                                            @click.stop="onHideSub(sub.id)"
                                        >
                                            <f7-icon f7="eye_slash" size="18" />
                                        </f7-link>
                                    </div>
                                </div>
                            </template>
                        </template>
                    </template>
                </template>
            </template>

            <!-- Empty state -->
            <div v-if="!hasAnyData" class="budget-m-empty">
                <span>{{ tt('No categories available to add') }}</span>
            </div>
        </template>

        <!-- Add category button -->
        <div v-if="hasHiddenItems" class="margin">
            <f7-button fill @click="showAddSheet = true">{{ tt('Add Category') }}</f7-button>
        </div>

        <!-- Number pad sheet for editing budgeted amount -->
        <number-pad-sheet
            :hint="editingSubName"
            :currency="defaultCurrency"
            v-model:show="showNumPad"
            v-model="editingAmount"
        />

        <!-- Copy Budget Popup -->
        <f7-popup v-model:opened="showCopyPopup" tablet-fullscreen>
            <f7-page>
                <f7-navbar :title="tt('Copy Budget')">
                    <f7-nav-right>
                        <f7-link popup-close :text="tt('Cancel')" />
                    </f7-nav-right>
                </f7-navbar>

                <!-- Step 1: source month picker -->
                <template v-if="copyStep === 1">
                    <f7-block-title>{{ tt('Select Source Month') }}</f7-block-title>
                    <f7-list strong inset dividers>
                        <f7-list-input
                            type="select"
                            :label="tt('Year')"
                            @change="copySourceYear = Number(($event.target as HTMLSelectElement).value)"
                        >
                            <option
                                v-for="y in copyYearOptions"
                                :key="y"
                                :value="y"
                                :selected="y === copySourceYear"
                            >{{ y }}</option>
                        </f7-list-input>
                        <f7-list-input
                            type="select"
                            :label="tt('Month')"
                            @change="copySourceMonth = Number(($event.target as HTMLSelectElement).value)"
                        >
                            <option
                                v-for="m in copyMonthOptions"
                                :key="m.value"
                                :value="m.value"
                                :selected="m.value === copySourceMonth"
                            >{{ m.label }}</option>
                        </f7-list-input>
                    </f7-list>
                    <div class="margin">
                        <f7-button fill :loading="copyLoading" @click="advanceCopyStep">{{ tt('Next') }}</f7-button>
                    </div>
                </template>

                <!-- Step 2: conflict resolution -->
                <template v-else>
                    <f7-block-title>
                        {{ formatMonthTitle(copySourceYear, copySourceMonth) }} → {{ formatMonthTitle(selectedYear, selectedMonth) }}
                    </f7-block-title>

                    <template v-for="item in copyConflictItems" :key="item.subcategoryId">
                        <f7-block-title medium>{{ item.parentCategoryName }} › {{ item.subcategoryName }}</f7-block-title>

                        <!-- Hidden, no existing target -->
                        <f7-list v-if="item.isHidden && !item.hasExistingTarget" strong inset dividers>
                            <f7-list-item link="#" no-chevron :title="tt('Skip')" @click="item.action = 'skip'">
                                <template #after>
                                    <f7-icon f7="checkmark_alt" v-if="item.action === 'skip'" />
                                </template>
                            </f7-list-item>
                            <f7-list-item
                                link="#" no-chevron
                                :title="tt('Copy and Unhide')"
                                :footer="`${tt('Source Amount')}: ${fmt(item.amount)}`"
                                @click="item.action = 'copy_unhide'"
                            >
                                <template #after>
                                    <f7-icon f7="checkmark_alt" v-if="item.action === 'copy_unhide'" />
                                </template>
                            </f7-list-item>
                        </f7-list>

                        <!-- Visible, has existing target -->
                        <f7-list v-else-if="!item.isHidden && item.hasExistingTarget" strong inset dividers>
                            <f7-list-item
                                link="#" no-chevron
                                :title="tt('Keep Existing')"
                                :footer="`${tt('Destination Amount')}: ${fmt(item.existingAmount)}`"
                                @click="item.action = 'skip'"
                            >
                                <template #after>
                                    <f7-icon f7="checkmark_alt" v-if="item.action === 'skip'" />
                                </template>
                            </f7-list-item>
                            <f7-list-item
                                link="#" no-chevron
                                :title="tt('Overwrite')"
                                :footer="`${tt('Source Amount')}: ${fmt(item.amount)}`"
                                @click="item.action = 'overwrite'"
                            >
                                <template #after>
                                    <f7-icon f7="checkmark_alt" v-if="item.action === 'overwrite'" />
                                </template>
                            </f7-list-item>
                        </f7-list>

                        <!-- Hidden AND has existing target -->
                        <f7-list v-else strong inset dividers>
                            <f7-list-item link="#" no-chevron :title="tt('Skip')" @click="item.action = 'skip'">
                                <template #after>
                                    <f7-icon f7="checkmark_alt" v-if="item.action === 'skip'" />
                                </template>
                            </f7-list-item>
                            <f7-list-item
                                link="#" no-chevron
                                :title="tt('Overwrite and Unhide')"
                                :footer="`${tt('Source Amount')}: ${fmt(item.amount)} · ${tt('Destination Amount')}: ${fmt(item.existingAmount)}`"
                                @click="item.action = 'overwrite_unhide'"
                            >
                                <template #after>
                                    <f7-icon f7="checkmark_alt" v-if="item.action === 'overwrite_unhide'" />
                                </template>
                            </f7-list-item>
                        </f7-list>
                    </template>

                    <f7-block v-if="copyAutoItems.length > 0">
                        <p class="text-color-gray">{{ copyAutoItems.length }} {{ tt('categories will be copied automatically') }}</p>
                    </f7-block>

                    <div class="margin">
                        <f7-button fill :loading="copyLoading" @click="executeCopy">{{ tt('Confirm') }}</f7-button>
                    </div>
                </template>
            </f7-page>
        </f7-popup>

        <!-- Add Category Sheet -->
        <f7-sheet v-model:opened="showAddSheet" swipe-to-close backdrop style="height: auto; max-height: 70vh">
            <f7-page-content>
                <f7-block-title>{{ tt('Add Category') }}</f7-block-title>
                <f7-list strong inset dividers v-if="hasHiddenItems">
                    <f7-list-item
                        v-for="parent in hiddenExpenseParents"
                        :key="'ep-' + parent.id"
                        link="#" no-chevron
                        :title="parent.name"
                        :footer="tt('Expenses')"
                        @click="onAddParent(parent)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                    <f7-list-item
                        v-for="item in hiddenExpenseSubsUnderVisibleParent"
                        :key="'es-' + item.sub.id"
                        link="#" no-chevron
                        :title="item.sub.name"
                        :footer="item.parent.name + ' · ' + tt('Expenses')"
                        @click="onAddSub(item.sub.id)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                    <f7-list-item
                        v-for="parent in hiddenIncomeParents"
                        :key="'ip-' + parent.id"
                        link="#" no-chevron
                        :title="parent.name"
                        :footer="tt('Income')"
                        @click="onAddParent(parent)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                    <f7-list-item
                        v-for="item in hiddenIncomeSubsUnderVisibleParent"
                        :key="'is-' + item.sub.id"
                        link="#" no-chevron
                        :title="item.sub.name"
                        :footer="item.parent.name + ' · ' + tt('Income')"
                        @click="onAddSub(item.sub.id)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                    <f7-list-item
                        v-for="parent in hiddenSavingsParents"
                        :key="'sp-' + parent.id"
                        link="#" no-chevron
                        :title="parent.name"
                        :footer="tt('Savings')"
                        @click="onAddParent(parent)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                    <f7-list-item
                        v-for="item in hiddenSavingsSubsUnderVisibleParent"
                        :key="'ss-' + item.sub.id"
                        link="#" no-chevron
                        :title="item.sub.name"
                        :footer="item.parent.name + ' · ' + tt('Savings')"
                        @click="onAddSub(item.sub.id)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                </f7-list>
                <f7-block v-else>
                    <p>{{ tt('No categories available to add') }}</p>
                </f7-block>
            </f7-page-content>
        </f7-sheet>
    </f7-page>
</template>

<script setup lang="ts">
import NumberPadSheet from '@/components/mobile/NumberPadSheet.vue';

import { ref, computed, watch } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useBudgetPageBase, type CopyDecision, addMonths } from '@/views/base/BudgetPageBase.ts';

import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useUserStore } from '@/stores/user.ts';

import { CategoryType } from '@/core/category.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import services from '@/lib/services.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';

const {
    tt,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatDateTimeToGregorianLikeLongYearMonth,
    formatDateTimeToGregorianLikeShortYearMonth,
} = useI18n();
const { showToast } = useI18nUIComponents();

const {
    selectedYear,
    selectedMonth,
    hiddenCategoryIds,
    budgetTargets,
    savingsActuals,
    selectMonth,
    loadBudgetTargets,
    loadSavingsActuals,
    getSavingsNet,
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

const col = computed(() => ({ year: selectedYear.value, month: selectedMonth.value }));

// ---------- UI state ----------

const loading = ref<boolean>(true);
const spentByMonth = ref<Record<string, Record<string, number>>>({});
const expandedParents = ref<Set<string>>(new Set());
const saving = ref<boolean>(false);
const showIncomeSection = ref<boolean>(true);
const showExpenseSection = ref<boolean>(true);
const showSavingsSection = ref<boolean>(true);

const showNumPad = ref<boolean>(false);
const editingSubId = ref<string>('');
const editingSubName = ref<string>('');
const editingAmount = ref<number>(0);

const showCopyPopup = ref<boolean>(false);
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

const showAddSheet = ref<boolean>(false);

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

const allTransferParents = computed<TransactionCategory[]>(() =>
    ((categoriesStore.allTransactionCategories[CategoryType.Transfer] ?? []) as TransactionCategory[])
        .filter(p => p.name === 'Savings & Investments')
);

const hiddenExpenseParents = computed<TransactionCategory[]>(() =>
    allExpenseParents.value.filter(p => hiddenCategoryIds.value.has(p.id))
);

const hiddenIncomeParents = computed<TransactionCategory[]>(() =>
    allIncomeParents.value.filter(p => hiddenCategoryIds.value.has(p.id))
);

const hiddenSavingsParents = computed<TransactionCategory[]>(() =>
    allTransferParents.value.filter(p => hiddenCategoryIds.value.has(p.id))
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

const hiddenSavingsSubsUnderVisibleParent = computed<HiddenSubItem[]>(() => {
    const result: HiddenSubItem[] = [];
    for (const parent of allTransferParents.value) {
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
    hiddenIncomeSubsUnderVisibleParent.value.length > 0 ||
    hiddenSavingsParents.value.length > 0 ||
    hiddenSavingsSubsUnderVisibleParent.value.length > 0
);

const hasAnyData = computed<boolean>(() =>
    allExpenseParents.value.length > 0 || allIncomeParents.value.length > 0 || allTransferParents.value.length > 0
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

function formatMonthTitle(year: number, month: number): string {
    return formatDateTimeToGregorianLikeLongYearMonth(
        parseDateTimeFromUnixTime(monthFirstUnixTime(year, month))
    );
}

function formatTimelineLabel(m: { year: number; month: number }): string {
    return formatDateTimeToGregorianLikeShortYearMonth(
        parseDateTimeFromUnixTime(monthFirstUnixTime(m.year, m.month))
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
    if (diff > 0) return 'text-color-green';
    if (diff < 0) return 'text-color-red';
    return '';
}

// ---------- Amount calculations (single-column helpers using col) ----------

function subBudgeted(subcatId: string): number {
    const c = col.value;
    return budgetTargets.value[`${c.year}-${c.month}`]?.[subcatId]?.amount ?? 0;
}

function subActual(subcatId: string): number {
    const c = col.value;
    return spentByMonth.value[`${c.year}-${c.month}`]?.[subcatId] ?? 0;
}

function subRemaining(subcatId: string): number {
    return subBudgeted(subcatId) - subActual(subcatId);
}

function parentBudgeted(parent: TransactionCategory): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subBudgeted(sub.id), 0);
}

function parentActual(parent: TransactionCategory): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subActual(sub.id), 0);
}

function parentRemaining(parent: TransactionCategory): number {
    return parentBudgeted(parent) - parentActual(parent);
}

// ---------- Visibility (same logic as desktop: hidden overridden by non-zero budget OR actual) ----------

function isSubVisible(subId: string): boolean {
    return !hiddenCategoryIds.value.has(subId) || subBudgeted(subId) > 0 || subActual(subId) > 0;
}

function isParentVisible(parent: TransactionCategory): boolean {
    if (!hiddenCategoryIds.value.has(parent.id)) return true;
    return (parent.subCategories ?? []).some(sub => subBudgeted(sub.id) > 0 || subActual(sub.id) > 0);
}

// ---------- Savings single-column helpers ----------

function subSavingsBudgeted(subcatId: string): number {
    const c = col.value;
    return budgetTargets.value[`${c.year}-${c.month}`]?.[subcatId]?.amount ?? 0;
}

function subSavingsActual(subcatId: string): number {
    const c = col.value;
    return getSavingsNet(subcatId, c.year, c.month);
}

function subSavingsRemaining(subcatId: string): number {
    return subSavingsBudgeted(subcatId) - subSavingsActual(subcatId);
}

function parentSavingsBudgeted(parent: TransactionCategory): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subSavingsBudgeted(sub.id), 0);
}

function parentSavingsActual(parent: TransactionCategory): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subSavingsActual(sub.id), 0);
}

function parentSavingsRemaining(parent: TransactionCategory): number {
    return parentSavingsBudgeted(parent) - parentSavingsActual(parent);
}

function isSavingsSubVisible(subId: string): boolean {
    return !hiddenCategoryIds.value.has(subId) || subSavingsBudgeted(subId) > 0 || subSavingsActual(subId) !== 0;
}

function isSavingsParentVisible(parent: TransactionCategory): boolean {
    if (!hiddenCategoryIds.value.has(parent.id)) return true;
    return (parent.subCategories ?? []).some(sub => subSavingsBudgeted(sub.id) > 0 || subSavingsActual(sub.id) !== 0);
}

function savingsDiffClass(amount: number): string {
    if (amount > 0) return 'text-color-red';
    if (amount < 0) return 'text-color-green';
    return '';
}

// ---------- Summary totals ----------

const colExpenseBudgeted = computed<number>(() =>
    allExpenseParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subBudgeted(sub.id), 0)
);
const colExpenseActual = computed<number>(() =>
    allExpenseParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subActual(sub.id), 0)
);
const colExpenseDiff = computed<number>(() => colExpenseBudgeted.value - colExpenseActual.value);

const colIncomeBudgeted = computed<number>(() =>
    allIncomeParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subBudgeted(sub.id), 0)
);
const colIncomeActual = computed<number>(() =>
    allIncomeParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subActual(sub.id), 0)
);
const colIncomeDiff = computed<number>(() => colIncomeBudgeted.value - colIncomeActual.value);

const colSavingsBudgeted = computed<number>(() =>
    allTransferParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subSavingsBudgeted(sub.id), 0)
);
const colSavingsActual = computed<number>(() =>
    allTransferParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subSavingsActual(sub.id), 0)
);
const colSavingsDiff = computed<number>(() => colSavingsBudgeted.value - colSavingsActual.value);

const colNetBudgeted = computed<number>(() => colIncomeBudgeted.value - colExpenseBudgeted.value - colSavingsBudgeted.value);
const colNetActual = computed<number>(() => colIncomeActual.value - colExpenseActual.value - colSavingsActual.value);
const colNetDiff = computed<number>(() => colNetActual.value - colNetBudgeted.value);

// ---------- Expand/collapse ----------

function toggleExpanded(parentId: string): void {
    const next = new Set(expandedParents.value);
    if (next.has(parentId)) next.delete(parentId);
    else next.add(parentId);
    expandedParents.value = next;
}

// ---------- Hide/show ----------

function onHideParent(parent: TransactionCategory): void {
    hideCategoryWithChildren(parent.id, (parent.subCategories ?? []).map(s => s.id));
}

function onHideSub(subId: string): void {
    hideCategory(subId);
}

function onAddParent(parent: TransactionCategory): void {
    unhideCategoryWithChildren(parent.id, (parent.subCategories ?? []).map(s => s.id));
    showAddSheet.value = false;
}

function onAddSub(subId: string): void {
    unhideCategory(subId);
    showAddSheet.value = false;
}

// ---------- Inline edit via number pad ----------

function startEdit(subcatId: string, subName: string): void {
    editingSubId.value = subcatId;
    editingSubName.value = subName;
    editingAmount.value = subBudgeted(subcatId);
    showNumPad.value = true;
}

watch(showNumPad, async (isOpen) => {
    if (isOpen || !editingSubId.value) return;
    const subId = editingSubId.value;
    const amount = editingAmount.value;
    editingSubId.value = '';
    if (saving.value) return;
    saving.value = true;
    try {
        await saveBudgetTarget(subId, col.value.year, col.value.month, amount);
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        saving.value = false;
    }
});

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

async function loadCurrentMonthData(): Promise<void> {
    const { year, month } = col.value;
    const key = `${year}-${month}`;
    const tasks: Promise<void>[] = [];
    if (!budgetTargets.value[key]) tasks.push(loadBudgetTargets(year, month));
    if (!spentByMonth.value[key]) tasks.push(loadStatsForMonth(year, month));
    if (!savingsActuals.value[key]) tasks.push(loadSavingsActuals(year, month));
    await Promise.all(tasks);
}

function scrollToSelected(): void {
    const el = chipRefs.get(`${selectedYear.value}-${selectedMonth.value}`);
    if (el) el.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'nearest' });
}

let initialized = false;

async function init(): Promise<void> {
    loading.value = true;
    try {
        await categoriesStore.loadAllCategories({ force: false });
        expandedParents.value = new Set([
            ...allExpenseParents.value.map(p => p.id),
            ...allIncomeParents.value.map(p => p.id),
            ...allTransferParents.value.map(p => p.id),
        ]);
        await loadCurrentMonthData();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        loading.value = false;
        initialized = true;
    }
    scrollToSelected();
}

watch(col, async (newCol, oldCol) => {
    if (!initialized) return;
    if (newCol.year === oldCol.year && newCol.month === oldCol.month) return;
    loading.value = true;
    try {
        await loadCurrentMonthData();
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        loading.value = false;
    }
    scrollToSelected();
});

// ---------- Timeline ----------

function onTimelineClick(m: { year: number; month: number }): void {
    selectMonth(m.year, m.month);
}

// ---------- Copy popup ----------

function openCopyPopup(): void {
    const prev = addMonths(selectedYear.value, selectedMonth.value, -1);
    copySourceYear.value = prev.year;
    copySourceMonth.value = prev.month;
    copyStep.value = 1;
    copyItems.value = [];
    showCopyPopup.value = true;
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
            showCopyPopup.value = false;
            showToast(tt('Budget copied successfully'));
        } else {
            copyStep.value = 2;
        }
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
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
        showCopyPopup.value = false;
        showToast(tt('Budget copied successfully'));
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        copyLoading.value = false;
    }
}

// ---------- Boot ----------

function onPageAfterIn(): void {
    if (!initialized) {
        if (isUserLogined() && isUserUnlocked()) {
            init();
        }
    } else {
        scrollToSelected();
    }
}
</script>

<style>
.budget-m-timeline-wrap {
    overflow-x: auto;
    padding: 8px 16px;
    scrollbar-width: none;
}

.budget-m-timeline-wrap::-webkit-scrollbar {
    display: none;
}

.budget-m-timeline {
    display: flex;
    gap: 8px;
    white-space: nowrap;
}

.budget-m-chip {
    display: inline-flex;
    align-items: center;
    padding: 5px 14px;
    border-radius: 20px;
    font-size: 0.8125rem;
    cursor: pointer;
    background-color: var(--f7-block-strong-bg-color);
    border: 1px solid var(--f7-list-border-color);
    user-select: none;
    flex-shrink: 0;
}

.budget-m-chip--active {
    background-color: var(--f7-theme-color);
    color: #fff;
    border-color: transparent;
}

.budget-m-summary-card {
    margin-top: 8px;
}

.budget-m-row {
    display: flex;
    align-items: center;
    padding: 7px 16px;
    min-height: 40px;
}

.budget-m-name-cell {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.875rem;
}

.budget-m-amt-cell {
    width: 68px;
    flex-shrink: 0;
    font-size: 0.8rem;
    text-align: right;
}

.budget-m-eye-cell {
    width: 30px;
    flex-shrink: 0;
    display: flex;
    justify-content: center;
}

.budget-m-header-row {
    padding-top: 6px;
    padding-bottom: 2px;
}

.budget-m-col-label {
    font-size: 0.68rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    opacity: 0.5;
}

.budget-m-summary-row {
    cursor: pointer;
    border-radius: 6px;
}

.budget-m-summary-name {
    display: flex;
    align-items: center;
    gap: 4px;
    font-weight: 500;
}

.budget-m-net-name {
    padding-left: 22px;
    font-weight: 600;
    font-size: 0.875rem;
}

.budget-m-chevron {
    flex-shrink: 0;
    color: var(--f7-list-chevron-icon-color);
}

.budget-m-cat-header {
    padding-top: 6px;
    padding-bottom: 2px;
    margin-top: 8px;
}

.budget-m-section-label {
    padding: 8px 16px 4px;
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    opacity: 0.45;
}

.budget-m-section-divider {
    height: 1px;
    background-color: var(--f7-list-border-color);
    margin: 10px 0;
}

.budget-m-parent-row {
    cursor: pointer;
    background-color: var(--f7-block-strong-bg-color);
    border-top: 1px solid var(--f7-list-border-color);
}

.budget-m-parent-name {
    display: flex;
    align-items: center;
    gap: 4px;
}

.budget-m-parent-label {
    font-weight: 600;
    font-size: 0.875rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.budget-m-sub-row {
    border-top: 1px solid var(--f7-list-border-color);
    padding-left: 36px;
}

.budget-m-sub-name {
    font-size: 0.875rem;
    color: var(--f7-list-item-subtitle-text-color);
}

.budget-m-budgeted-cell {
    text-align: right;
    cursor: pointer;
}

.budget-m-budgeted-cell span {
    text-decoration: underline;
    text-decoration-style: dotted;
    text-underline-offset: 2px;
}

.budget-m-zero {
    opacity: 0.4;
}

.budget-m-negative {
    color: var(--f7-color-red);
}

.budget-m-eye-btn {
    color: var(--f7-list-item-subtitle-text-color);
    line-height: 1;
}

.budget-m-empty {
    padding: 32px 16px;
    text-align: center;
    opacity: 0.5;
    font-size: 0.875rem;
}
</style>
