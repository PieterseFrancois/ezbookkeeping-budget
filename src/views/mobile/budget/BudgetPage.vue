<template>
    <f7-page @page:afterin="onPageAfterIn">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title>{{ tt('Budget') }}</f7-nav-title>
            <f7-nav-right>
                <f7-link icon-f7="question_circle" @click="showHelpPopup = true" />
                <f7-link @click="showAmountInBudgetPage = !showAmountInBudgetPage">
                    <f7-icon class="ebk-hide-icon" :f7="showAmountInBudgetPage ? 'eye_slash_fill' : 'eye_fill'"></f7-icon>
                </f7-link>
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

        <!-- Budget cycle note -->
        <div v-if="budgetCycleNote" class="budget-m-cycle-note">{{ budgetCycleNote }}</div>

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
                <!-- Cards & Debt row -->
                <div class="budget-m-row budget-m-summary-row" @click="showDebtSection = !showDebtSection">
                    <div class="budget-m-name-cell budget-m-summary-name">
                        <f7-icon :f7="showDebtSection ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                        <span>{{ tt('Cards & Debt') }}</span>
                    </div>
                    <span class="budget-m-amt-cell">{{ fmt(colDebtBudgeted) }}</span>
                    <span class="budget-m-amt-cell">{{ fmt(colDebtActual) }}</span>
                    <span class="budget-m-amt-cell" :class="savingsDiffClass(colDebtDiff)">{{ fmt(colDebtDiff) }}</span>
                </div>
                <!-- Net row -->
                <div class="budget-m-row">
                    <div class="budget-m-name-cell budget-m-net-name">{{ tt('Net') }}</div>
                    <span class="budget-m-amt-cell">{{ fmt(colNetBudgeted) }}</span>
                    <span class="budget-m-amt-cell" :class="diffClass(colNetActual)">{{ fmt(colNetActual) }}</span>
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
            <div class="budget-m-section" v-if="showExpenseSection">
                <div class="budget-m-section-label">
                    {{ tt('Expenses') }}<span class="budget-m-section-hint">{{ tt('money out') }}</span>
                </div>
                <template v-for="parent in allExpenseParents" :key="parent.id">
                    <template v-if="isParentVisible(parent)">
                        <div class="budget-m-row budget-m-parent-row" @click="toggleExpanded(parent.id)">
                            <div class="budget-m-name-cell budget-m-parent-name">
                                <f7-icon :f7="expandedParents.has(parent.id) ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                                <span class="budget-m-parent-label">{{ parent.name }}</span>
                            </div>
                            <span class="budget-m-amt-cell">{{ fmt(parentBudgeted(parent)) }}</span>
                            <span class="budget-m-amt-cell">{{ fmt(parentActual(parent)) }}</span>
                            <span class="budget-m-amt-cell" :class="diffClass(parentRemaining(parent))">{{ fmt(parentRemaining(parent)) }}</span>
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
                                    <span class="budget-m-amt-cell" :class="diffClass(subRemaining(sub.id))">{{ fmt(subRemaining(sub.id)) }}</span>
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
            </div>

            <div v-if="showExpenseSection && showIncomeSection" class="budget-m-section-divider" />

            <!-- Income section -->
            <div class="budget-m-section" v-if="showIncomeSection">
                <div class="budget-m-section-label">
                    {{ tt('Income') }}<span class="budget-m-section-hint">{{ tt('money in') }}</span>
                </div>
                <template v-for="parent in allIncomeParents" :key="parent.id">
                    <template v-if="isParentVisible(parent)">
                        <div class="budget-m-row budget-m-parent-row" @click="toggleExpanded(parent.id)">
                            <div class="budget-m-name-cell budget-m-parent-name">
                                <f7-icon :f7="expandedParents.has(parent.id) ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                                <span class="budget-m-parent-label">{{ parent.name }}</span>
                            </div>
                            <span class="budget-m-amt-cell">{{ fmt(parentBudgeted(parent)) }}</span>
                            <span class="budget-m-amt-cell">{{ fmt(parentActual(parent)) }}</span>
                            <span class="budget-m-amt-cell" :class="diffClass(-parentRemaining(parent))">{{ fmt(-parentRemaining(parent)) }}</span>
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
                                    <span class="budget-m-amt-cell" :class="diffClass(-subRemaining(sub.id))">{{ fmt(-subRemaining(sub.id)) }}</span>
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

                <!-- Savings withdrawals surface here as income, budgetable on the same category -->
                <div v-for="w in withdrawalItems" :key="'wd-' + w.id" class="budget-m-row budget-m-sub-row">
                    <span class="budget-m-name-cell budget-m-sub-name">{{ w.parentName }} › {{ w.name }}</span>
                    <div class="budget-m-amt-cell budget-m-budgeted-cell" @click="startEdit(w.id, w.name, 'income')">
                        <span :class="{ 'budget-m-zero': subWithdrawalBudgeted(w.id) === 0 }">{{ fmt(subWithdrawalBudgeted(w.id)) }}</span>
                    </div>
                    <span class="budget-m-amt-cell">{{ fmt(w.amount) }}</span>
                    <span class="budget-m-amt-cell" :class="diffClass(w.amount - subWithdrawalBudgeted(w.id))">{{ fmt(w.amount - subWithdrawalBudgeted(w.id)) }}</span>
                    <div class="budget-m-eye-cell">
                        <f7-link
                            v-if="w.amount === 0 && subWithdrawalBudgeted(w.id) === 0"
                            class="budget-m-eye-btn"
                            @click.stop="removeWithdrawalCategory(w.id)"
                        >
                            <f7-icon f7="eye_slash" size="18" />
                        </f7-link>
                    </div>
                </div>
            </div>

            <div v-if="(showExpenseSection || showIncomeSection) && showSavingsSection" class="budget-m-section-divider" />

            <!-- Savings section -->
            <div class="budget-m-section" v-if="showSavingsSection">
                <div class="budget-m-section-label">
                    {{ tt('Savings') }}<span class="budget-m-section-hint">{{ tt('money out — set aside') }}</span>
                </div>
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
                                    <div class="budget-m-amt-cell budget-m-budgeted-cell" @click="startEdit(sub.id, sub.name, 'savings')">
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
            </div>

            <div v-if="(showExpenseSection || showIncomeSection || showSavingsSection) && showDebtSection" class="budget-m-section-divider" />

            <!-- Cards & Debt section -->
            <div class="budget-m-section" v-if="showDebtSection">
                <div class="budget-m-section-label">
                    {{ tt('Cards & Debt') }}<span class="budget-m-section-hint">{{ tt('money out — paydown') }}</span>
                </div>
                <template v-for="parent in allDebtParents" :key="parent.id">
                    <template v-if="isDebtParentVisible(parent)">
                        <div class="budget-m-row budget-m-parent-row" @click="toggleExpanded(parent.id)">
                            <div class="budget-m-name-cell budget-m-parent-name">
                                <f7-icon :f7="expandedParents.has(parent.id) ? 'chevron_down' : 'chevron_right'" size="14" class="budget-m-chevron" />
                                <span class="budget-m-parent-label">{{ parent.name }}</span>
                            </div>
                            <span class="budget-m-amt-cell">{{ fmt(parentDebtBudgeted(parent)) }}</span>
                            <span class="budget-m-amt-cell">{{ fmt(parentDebtActual(parent)) }}</span>
                            <span class="budget-m-amt-cell" :class="savingsDiffClass(parentDebtRemaining(parent))">{{ fmt(parentDebtRemaining(parent)) }}</span>
                            <div class="budget-m-eye-cell">
                                <f7-link
                                    v-if="!hiddenCategoryIds.has(parent.id) && parentDebtBudgeted(parent) === 0"
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
                                    v-if="isDebtSubVisible(sub.id)"
                                    class="budget-m-row budget-m-sub-row"
                                >
                                    <span class="budget-m-name-cell budget-m-sub-name">{{ sub.name }}</span>
                                    <div class="budget-m-amt-cell budget-m-budgeted-cell" @click="startEdit(sub.id, sub.name, 'debt')">
                                        <span :class="{ 'budget-m-zero': subDebtBudgeted(sub.id) === 0 }">{{ fmt(subDebtBudgeted(sub.id)) }}</span>
                                    </div>
                                    <span class="budget-m-amt-cell">{{ fmt(subDebtActual(sub.id)) }}</span>
                                    <span class="budget-m-amt-cell" :class="savingsDiffClass(subDebtRemaining(sub.id))">{{ fmt(subDebtRemaining(sub.id)) }}</span>
                                    <div class="budget-m-eye-cell">
                                        <f7-link
                                            v-if="!hiddenCategoryIds.has(sub.id) && subDebtBudgeted(sub.id) === 0"
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
            </div>

            <!-- Empty state -->
            <div v-if="!hasAnyData" class="budget-m-empty">
                <span>{{ tt('No categories available to add') }}</span>
            </div>
        </template>

        <!-- Cards & Debt reserve panel -->
        <f7-card v-if="!loading && liabilityReserves.length > 0" class="budget-m-reserve-card">
            <f7-card-header class="budget-m-reserve-header">{{ tt('Cards & Debt') }}</f7-card-header>
            <f7-card-content :padding="false">
                <div v-for="r in liabilityReserves" :key="r.accountId" class="budget-m-reserve-row">
                    <div class="budget-m-reserve-main">
                        <span class="budget-m-reserve-name">{{ r.name }}</span>
                        <span class="budget-m-reserve-owed" :class="r.owed > 0 ? 'text-color-red' : 'text-color-green'">
                            {{ r.owed > 0 ? tt('Owed') : tt('Settled') }} {{ fmtCurrency(Math.abs(r.owed), r.currency) }}
                        </span>
                    </div>
                    <div class="budget-m-reserve-sub">
                        <span>{{ tt('Charged') }} {{ fmtCurrency(r.cycleSpend, r.currency) }}</span>
                        <span>{{ tt('Paid') }} {{ fmtCurrency(r.cyclePayments, r.currency) }}</span>
                    </div>
                </div>
            </f7-card-content>
        </f7-card>

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

        <!-- Help Popup -->
        <budget-help-popup v-model:opened="showHelpPopup" />

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
                    <f7-list-item
                        v-for="parent in hiddenDebtParents"
                        :key="'dp-' + parent.id"
                        link="#" no-chevron
                        :title="parent.name"
                        :footer="tt('Cards & Debt')"
                        @click="onAddParent(parent)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                    <f7-list-item
                        v-for="item in hiddenDebtSubsUnderVisibleParent"
                        :key="'ds-' + item.sub.id"
                        link="#" no-chevron
                        :title="item.sub.name"
                        :footer="item.parent.name + ' · ' + tt('Cards & Debt')"
                        @click="onAddSub(item.sub.id)"
                    >
                        <template #media><f7-icon f7="plus_circle" /></template>
                    </f7-list-item>
                    <f7-list-item
                        v-for="item in availableWithdrawalSubs"
                        :key="'wa-' + item.sub.id"
                        link="#" no-chevron
                        :title="item.sub.name"
                        :footer="item.parent.name + ' · ' + tt('Withdrawal') + ' (' + tt('Income') + ')'"
                        @click="onAddWithdrawal(item.sub.id)"
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
import BudgetHelpPopup from '@/views/mobile/budget/BudgetHelpPopup.vue';

import { ref, computed, watch } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';
import { useBudgetPageBase, type CopyDecision, type BudgetSection, addMonths } from '@/views/base/BudgetPageBase.ts';

import { useTransactionCategoriesStore } from '@/stores/transactionCategory.ts';
import { useUserStore } from '@/stores/user.ts';

import { CategoryType } from '@/core/category.ts';
import type { TransactionCategory } from '@/models/transaction_category.ts';
import { parseDateTimeFromUnixTime } from '@/lib/datetime.ts';
import { isUserLogined, isUserUnlocked } from '@/lib/userstate.ts';
import axios from 'axios';
import type { ApiResponse } from '@/core/api.ts';
import { DISPLAY_HIDDEN_AMOUNT } from '@/consts/numeral.ts';

// Transfer parent categories that map to the Savings and Cards & Debt sections
const SAVINGS_PARENT_NAME = 'Savings & Investments';
const DEBT_PARENT_NAME = 'Loan & Debt';

interface LiabilityReserve {
    accountId: string;
    name: string;
    icon: string;
    color: string;
    currency: string;
    owed: number;
    cycleSpend: number;
    cyclePayments: number;
}

const {
    tt,
    formatAmountToLocalizedNumeralsWithCurrency,
    formatDateTimeToGregorianLikeLongYearMonth,
    formatDateTimeToGregorianLikeShortYearMonth,
} = useI18n();
const { showToast } = useI18nUIComponents();

const {
    showAmountInBudgetPage,
    selectedYear,
    selectedMonth,
    hiddenCategoryIds,
    withdrawalCategoryIds,
    budgetTargets,
    budgetActuals,
    selectMonth,
    loadBudgetTargets,
    loadBudgetActuals,
    getExpenseIncomeActual,
    getSectionActual,
    getTargetAmount,
    cycleFirstUnixTime,
    cycleLastUnixTime,
    saveBudgetTarget,
    copyBudgetFromMonth,
    addWithdrawalCategory,
    removeWithdrawalCategory,
    hideCategoryWithChildren,
    hideCategory,
    unhideCategory,
    unhideCategoryWithChildren,
} = useBudgetPageBase();

const categoriesStore = useTransactionCategoriesStore();
const userStore = useUserStore();

const defaultCurrency = computed<string>(() => userStore.currentUserDefaultCurrency);

const col = computed(() => ({ year: selectedYear.value, month: selectedMonth.value }));

const budgetCycleNote = computed<string>(() => {
    const endDay = userStore.currentUserBudgetEndDay;
    if (!endDay) return '';
    const { month: prevMonth } = addMonths(selectedYear.value, selectedMonth.value, -1);
    const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    const startLabel = `${endDay + 1} ${monthNames[prevMonth - 1]}`;
    const endLabel = `${endDay} ${monthNames[selectedMonth.value - 1]}`;
    return `Budget cycle: ${startLabel} – ${endLabel}`;
});

// ---------- UI state ----------

const loading = ref<boolean>(true);
const expandedParents = ref<Set<string>>(new Set());
const saving = ref<boolean>(false);
const liabilityReserves = ref<LiabilityReserve[]>([]);
const showIncomeSection = ref<boolean>(true);
const showExpenseSection = ref<boolean>(true);
const showSavingsSection = ref<boolean>(true);
const showDebtSection = ref<boolean>(true);

const showNumPad = ref<boolean>(false);
const editingSubId = ref<string>('');
const editingSubSection = ref<BudgetSection>('expense');
const editingSubName = ref<string>('');
const editingAmount = ref<number>(0);

const showCopyPopup = ref<boolean>(false);
const copyStep = ref<1 | 2>(1);
const copySourceYear = ref<number>(new Date().getFullYear());
const copySourceMonth = ref<number>(new Date().getMonth() === 0 ? 12 : new Date().getMonth());
const copyLoading = ref<boolean>(false);

interface CopyItem {
    subcategoryId: string;
    section: BudgetSection;
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
const showHelpPopup = ref<boolean>(false);

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
        .filter(p => p.name === SAVINGS_PARENT_NAME)
);

const allDebtParents = computed<TransactionCategory[]>(() =>
    ((categoriesStore.allTransactionCategories[CategoryType.Transfer] ?? []) as TransactionCategory[])
        .filter(p => p.name === DEBT_PARENT_NAME)
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

const hiddenDebtParents = computed<TransactionCategory[]>(() =>
    allDebtParents.value.filter(p => hiddenCategoryIds.value.has(p.id))
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

const hiddenDebtSubsUnderVisibleParent = computed<HiddenSubItem[]>(() => {
    const result: HiddenSubItem[] = [];
    for (const parent of allDebtParents.value) {
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
    hiddenSavingsSubsUnderVisibleParent.value.length > 0 ||
    hiddenDebtParents.value.length > 0 ||
    hiddenDebtSubsUnderVisibleParent.value.length > 0 ||
    availableWithdrawalSubs.value.length > 0
);

const hasAnyData = computed<boolean>(() =>
    allExpenseParents.value.length > 0 || allIncomeParents.value.length > 0 ||
    allTransferParents.value.length > 0 || allDebtParents.value.length > 0
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
    if (!showAmountInBudgetPage.value) {
        return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, defaultCurrency.value);
    }
    const rands = Math.round(amount / 100) * 100;
    return formatAmountToLocalizedNumeralsWithCurrency(rands, defaultCurrency.value)
        .replace(/[,.]00$/, '');
}

function fmtCurrency(amount: number, currency: string): string {
    if (!showAmountInBudgetPage.value) {
        return formatAmountToLocalizedNumeralsWithCurrency(DISPLAY_HIDDEN_AMOUNT, currency);
    }
    return formatAmountToLocalizedNumeralsWithCurrency(amount, currency);
}

function diffClass(diff: number): string {
    if (diff > 0) return 'text-color-green';
    if (diff < 0) return 'text-color-red';
    return '';
}

// ---------- Amount calculations (single-column helpers using col) ----------

// Expense and income categories map to their section unambiguously by type, so the
// expense/income helpers can resolve it rather than threading it through every call site.
function sectionForCategory(subcatId: string): BudgetSection {
    const cat = categoriesStore.allTransactionCategoriesMap[subcatId];
    return cat?.type === CategoryType.Income ? 'income' : 'expense';
}

function subBudgeted(subcatId: string): number {
    const c = col.value;
    return getTargetAmount(subcatId, sectionForCategory(subcatId), c.year, c.month);
}

function subActual(subcatId: string): number {
    const c = col.value;
    return getExpenseIncomeActual(subcatId, c.year, c.month);
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
    return getTargetAmount(subcatId, 'savings', c.year, c.month);
}

// Gross contribution into savings. Withdrawals are NOT netted off here — they surface
// separately as income lines (see subWithdrawalActual), so a R2000 deposit still reads
// as R2000 toward the goal even if R500 was later withdrawn.
function subSavingsActual(subcatId: string): number {
    const c = col.value;
    return getSectionActual(subcatId, 'savings', c.year, c.month);
}

// Money taken back out of a savings/investment account, classified as income.
function subWithdrawalActual(subcatId: string): number {
    const c = col.value;
    return getSectionActual(subcatId, 'income', c.year, c.month);
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

// Budgeted withdrawal for a savings category — a target on the same category but in the income
// section, independent of its contribution target.
function subWithdrawalBudgeted(subcatId: string): number {
    const c = col.value;
    return getTargetAmount(subcatId, 'income', c.year, c.month);
}

// Savings subcategories shown in the Income section: those withdrawn from this cycle, any with a
// planned withdrawal target, plus any explicitly added via the Add Category sheet (which is how
// you budget a withdrawal before one has happened).
const withdrawalItems = computed<{ id: string; name: string; parentName: string; amount: number }[]>(() => {
    const items: { id: string; name: string; parentName: string; amount: number }[] = [];
    for (const parent of allTransferParents.value) {
        for (const sub of (parent.subCategories ?? [])) {
            const amount = subWithdrawalActual(sub.id);
            if (amount !== 0 || subWithdrawalBudgeted(sub.id) > 0 || withdrawalCategoryIds.value.has(sub.id)) {
                items.push({ id: sub.id, name: sub.name, parentName: parent.name, amount });
            }
        }
    }
    return items;
});

// Savings subcategories not yet surfaced as withdrawal rows — offered in the Add Category sheet
const availableWithdrawalSubs = computed<HiddenSubItem[]>(() => {
    const result: HiddenSubItem[] = [];
    for (const parent of allTransferParents.value) {
        for (const sub of (parent.subCategories ?? [])) {
            if (withdrawalCategoryIds.value.has(sub.id)) continue;
            if (subWithdrawalActual(sub.id) !== 0 || subWithdrawalBudgeted(sub.id) > 0) continue;
            result.push({ sub, parent });
        }
    }
    return result;
});

const withdrawalTotal = computed<number>(() =>
    withdrawalItems.value.reduce((sum, i) => sum + i.amount, 0)
);

function savingsDiffClass(amount: number): string {
    if (amount > 0) return 'text-color-red';
    if (amount < 0) return 'text-color-green';
    return '';
}

// ---------- Cards & Debt single-column helpers (paydown, goal-style like savings) ----------

function subDebtBudgeted(subcatId: string): number {
    const c = col.value;
    return getTargetAmount(subcatId, 'debt', c.year, c.month);
}

function subDebtActual(subcatId: string): number {
    const c = col.value;
    return getSectionActual(subcatId, 'debt', c.year, c.month);
}

function subDebtRemaining(subcatId: string): number {
    return subDebtBudgeted(subcatId) - subDebtActual(subcatId);
}

function parentDebtBudgeted(parent: TransactionCategory): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subDebtBudgeted(sub.id), 0);
}

function parentDebtActual(parent: TransactionCategory): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subDebtActual(sub.id), 0);
}

function parentDebtRemaining(parent: TransactionCategory): number {
    return parentDebtBudgeted(parent) - parentDebtActual(parent);
}

function isDebtSubVisible(subId: string): boolean {
    return !hiddenCategoryIds.value.has(subId) || subDebtBudgeted(subId) > 0 || subDebtActual(subId) !== 0;
}

function isDebtParentVisible(parent: TransactionCategory): boolean {
    if (!hiddenCategoryIds.value.has(parent.id)) return true;
    return (parent.subCategories ?? []).some(sub => subDebtBudgeted(sub.id) > 0 || subDebtActual(sub.id) !== 0);
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
    + withdrawalItems.value.reduce((s, w) => s + subWithdrawalBudgeted(w.id), 0)
);
const colIncomeActual = computed<number>(() =>
    allIncomeParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subActual(sub.id), 0)
    + withdrawalTotal.value
);
const colIncomeDiff = computed<number>(() => colIncomeActual.value - colIncomeBudgeted.value);

const colSavingsBudgeted = computed<number>(() =>
    allTransferParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subSavingsBudgeted(sub.id), 0)
);
const colSavingsActual = computed<number>(() =>
    allTransferParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subSavingsActual(sub.id), 0)
);
const colSavingsDiff = computed<number>(() => colSavingsBudgeted.value - colSavingsActual.value);

const colDebtBudgeted = computed<number>(() =>
    allDebtParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subDebtBudgeted(sub.id), 0)
);
const colDebtActual = computed<number>(() =>
    allDebtParents.value.flatMap(p => p.subCategories ?? []).reduce((s, sub) => s + subDebtActual(sub.id), 0)
);
const colDebtDiff = computed<number>(() => colDebtBudgeted.value - colDebtActual.value);

const colNetBudgeted = computed<number>(() => colIncomeBudgeted.value - colExpenseBudgeted.value - colSavingsBudgeted.value - colDebtBudgeted.value);
const colNetActual = computed<number>(() => colIncomeActual.value - colExpenseActual.value - colSavingsActual.value - colDebtActual.value);
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

function onAddWithdrawal(subId: string): void {
    addWithdrawalCategory(subId);
    showAddSheet.value = false;
}

// ---------- Inline edit via number pad ----------

function startEdit(subcatId: string, subName: string, section?: BudgetSection): void {
    const resolved = section ?? sectionForCategory(subcatId);
    editingSubId.value = subcatId;
    editingSubSection.value = resolved;
    editingSubName.value = subName;
    editingAmount.value = getTargetAmount(subcatId, resolved, col.value.year, col.value.month);
    showNumPad.value = true;
}

watch(showNumPad, async (isOpen) => {
    if (isOpen || !editingSubId.value) return;
    const subId = editingSubId.value;
    const section = editingSubSection.value;
    const amount = editingAmount.value;
    editingSubId.value = '';
    if (saving.value) return;
    saving.value = true;
    try {
        await saveBudgetTarget(subId, section, col.value.year, col.value.month, amount);
    } catch (error: unknown) {
        if (!((error as { processed?: boolean }).processed)) {
            showToast((error as Error).message || String(error));
        }
    } finally {
        saving.value = false;
    }
});

// ---------- Data loading ----------

async function loadLiabilityReserves(year: number, month: number): Promise<void> {
    const resp = await axios.get<ApiResponse<{ items: LiabilityReserve[] }>>(
        'v1/budget/liabilities.json',
        { params: { startTime: cycleFirstUnixTime(year, month), endTime: cycleLastUnixTime(year, month) } }
    );
    liabilityReserves.value = resp.data?.result?.items ?? [];
}

async function loadCurrentMonthData(): Promise<void> {
    const { year, month } = col.value;
    const key = `${year}-${month}`;
    const tasks: Promise<void>[] = [];
    if (!budgetTargets.value[key]) tasks.push(loadBudgetTargets(year, month));
    if (!budgetActuals.value[key]) tasks.push(loadBudgetActuals(year, month));
    tasks.push(loadLiabilityReserves(year, month));
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
            ...allDebtParents.value.map(p => p.id),
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
        // Target keys are `${categoryId}|${section}` so one category can appear once per section
        for (const [entryKey, entry] of Object.entries(sourceTargets)) {
            const [subcatId, section] = entryKey.split('|') as [string, BudgetSection];
            const subCat = categoriesStore.allTransactionCategoriesMap[subcatId];
            if (!subCat || !subCat.parentId || subCat.parentId === '0') continue;
            const parentCat = categoriesStore.allTransactionCategoriesMap[subCat.parentId];
            if (!parentCat) continue;

            const isHidden = hiddenCategoryIds.value.has(subcatId) || hiddenCategoryIds.value.has(parentCat.id);
            const hasExistingTarget = !!destTargets[entryKey];
            const existingAmount = destTargets[entryKey]?.amount ?? 0;

            items.push({
                subcategoryId: subcatId,
                section,
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
                section: i.section,
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
            section: i.section,
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
.budget-m-cycle-note {
    font-size: 12px;
    color: var(--f7-list-item-subtitle-text-color, #888);
    padding: 2px 16px 6px;
}

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

/* Shares the section label's tint so the two form one continuous sticky bar */
.budget-m-cat-header {
    position: sticky;
    top: 0;
    z-index: 10;
    background-color: color-mix(in srgb, var(--f7-page-bg-color) 94%, var(--f7-theme-color) 6%);
    padding-top: 14px;
    padding-bottom: 2px;
}

/* The .budget-m-section wrapper in the template is deliberately unstyled: it exists only to bound
   the sticky label so the next section pushes it off, and must stay unpositioned so it does not
   create a stacking context around the label. */
.budget-m-section-label {
    /* Sits just below the sticky column header (40px tall) */
    position: sticky;
    top: 40px;
    z-index: 9;
    /* Must be a full opaque band, otherwise rows scrolling underneath show through
       above and below the short label text */
    /* Slightly raised off the page background so the sticky band reads as a bar rather than
       a plain slab, with a soft shadow onto the rows scrolling beneath it */
    background-color: color-mix(in srgb, var(--f7-page-bg-color) 94%, var(--f7-theme-color) 6%);
    border-top: 1px solid var(--f7-list-border-color);
    border-bottom: 1px solid var(--f7-list-border-color);
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.14);
    padding: 12px 16px 10px;
    min-height: 36px;
    display: flex;
    align-items: center;
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    /* Dim the text only — an `opacity` here would make the background translucent
       and let rows scrolling underneath show through */
    color: color-mix(in srgb, currentColor 60%, transparent);
}

/* Small accent tick so each section reads as its own heading */
.budget-m-section-label::before {
    content: '';
    flex-shrink: 0;
    width: 3px;
    height: 13px;
    border-radius: 2px;
    margin-inline-end: 8px;
    background-color: var(--f7-theme-color);
}

.budget-m-section-hint {
    font-weight: 400;
    letter-spacing: 0;
    text-transform: none;
    margin-inline-start: 6px;
    font-style: italic;
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

.budget-m-reserve-header {
    font-weight: 600;
    font-size: 0.9rem;
}

.budget-m-reserve-row {
    padding: 8px 16px;
    border-top: 1px solid var(--f7-list-border-color);
}

.budget-m-reserve-main {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
}

.budget-m-reserve-name {
    font-size: 0.875rem;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.budget-m-reserve-owed {
    font-size: 0.8rem;
    font-variant-numeric: tabular-nums;
    flex-shrink: 0;
    padding-inline-start: 8px;
}

.budget-m-reserve-sub {
    display: flex;
    gap: 16px;
    margin-top: 2px;
    font-size: 0.72rem;
    opacity: 0.6;
    font-variant-numeric: tabular-nums;
}
</style>
