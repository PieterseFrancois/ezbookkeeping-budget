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

        <!-- Budget table -->
        <v-col cols="12" class="pt-3">
            <div class="budget-table-wrap">
                <div class="budget-table">

                    <!-- ── Row 1: month title headers ── -->
                    <div class="budget-name-col budget-th budget-th-corner"></div>
                    <div
                        v-for="(col, colIdx) in threeMonthColumns"
                        :key="`th1-${col.year}-${col.month}`"
                        class="budget-th budget-month-header"
                        :class="{ 'budget-month-header--current': colIdx === 1 }"
                        style="grid-column: span 3"
                    >{{ formatColumnTitle(col) }}</div>

                    <!-- ── Row 2: B / A / R sub-headers ── -->
                    <div class="budget-name-col budget-th budget-th-sub-corner"></div>
                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`th2-${col.year}-${col.month}`">
                        <div class="budget-th budget-th-sub budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ tt('Budgeted') }}</div>
                        <div class="budget-th budget-th-sub" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ tt('Actual') }}</div>
                        <div class="budget-th budget-th-sub" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ tt('Remaining') }}</div>
                    </template>

                    <!-- ── Summary section top border ── -->
                    <div class="budget-full-span budget-summary-border-top"></div>

                    <!-- Expenses summary row -->
                    <div
                        class="budget-name-col budget-summary-name cursor-pointer"
                        @click="showExpenseSection = !showExpenseSection"
                    >
                        <v-icon class="flex-shrink-0" :icon="showExpenseSection ? mdiChevronDown : mdiChevronRight" size="16" />
                        <span class="budget-row-name text-body-2 font-weight-medium ms-1">{{ tt('Expenses') }}</span>
                    </div>
                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`sexp-${col.year}-${col.month}`">
                        <div class="budget-summary-cell budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(colExpenseBudgeted(col)) }}</div>
                        <div class="budget-summary-cell" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(colExpenseActual(col)) }}</div>
                        <div class="budget-summary-cell" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(colExpenseDiff(col))]">{{ fmt(colExpenseDiff(col)) }}</div>
                    </template>

                    <!-- Income summary row -->
                    <div
                        class="budget-name-col budget-summary-name cursor-pointer"
                        @click="showIncomeSection = !showIncomeSection"
                    >
                        <v-icon class="flex-shrink-0" :icon="showIncomeSection ? mdiChevronDown : mdiChevronRight" size="16" />
                        <span class="budget-row-name text-body-2 font-weight-medium ms-1">{{ tt('Income') }}</span>
                    </div>
                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`sinc-${col.year}-${col.month}`">
                        <div class="budget-summary-cell budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(colIncomeBudgeted(col)) }}</div>
                        <div class="budget-summary-cell" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(colIncomeActual(col)) }}</div>
                        <div class="budget-summary-cell" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(colIncomeDiff(col))]">{{ fmt(colIncomeDiff(col)) }}</div>
                    </template>

                    <!-- Savings summary row -->
                    <div
                        class="budget-name-col budget-summary-name cursor-pointer"
                        @click="showSavingsSection = !showSavingsSection"
                    >
                        <v-icon class="flex-shrink-0" :icon="showSavingsSection ? mdiChevronDown : mdiChevronRight" size="16" />
                        <span class="budget-row-name text-body-2 font-weight-medium ms-1">{{ tt('Savings') }}</span>
                    </div>
                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`ssav-${col.year}-${col.month}`">
                        <div class="budget-summary-cell budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(colSavingsBudgeted(col)) }}</div>
                        <div class="budget-summary-cell" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(colSavingsActual(col)) }}</div>
                        <div class="budget-summary-cell" :class="[{ 'budget-col-current-tint': colIdx === 1 }, savingsDiffClass(colSavingsDiff(col))]">{{ fmt(colSavingsDiff(col)) }}</div>
                    </template>

                    <!-- Net row -->
                    <div class="budget-name-col budget-summary-name">
                        <span class="budget-row-name text-body-2 font-weight-medium ps-5">{{ tt('Net') }}</span>
                    </div>
                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`snet-${col.year}-${col.month}`">
                        <div class="budget-summary-cell budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(colNetBudgeted(col)) }}</div>
                        <div class="budget-summary-cell" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(colNetActual(col))]">{{ fmt(colNetActual(col)) }}</div>
                        <div class="budget-summary-cell" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(colNetDiff(col))]">{{ fmt(colNetDiff(col)) }}</div>
                    </template>

                    <!-- ── Summary section bottom border ── -->
                    <div class="budget-full-span budget-summary-border-bottom"></div>

                    <!-- ── Loading skeleton ── -->
                    <template v-if="loading && !hasAnyData">
                        <div class="budget-full-span px-2 py-3">
                            <div v-for="i in 4" :key="i" class="mb-2">
                                <v-skeleton-loader width="100%" type="text" :loading="true" />
                            </div>
                        </div>
                    </template>

                    <template v-else>
                        <!-- ── EXPENSES section ── -->
                        <template v-if="showExpenseSection">
                            <div class="budget-full-span budget-section-label">{{ tt('Expenses') }}</div>

                            <template v-for="parent in allExpenseParents" :key="parent.id">
                                <template v-if="isParentVisibleInAnyCol(parent)">
                                    <!-- Expense parent row -->
                                    <div class="budget-name-col budget-parent-row">
                                        <div class="budget-expand-cell d-flex align-center justify-center">
                                            <v-btn
                                                density="compact"
                                                variant="text"
                                                size="x-small"
                                                :icon="expandedParents.has(parent.id) ? mdiChevronDown : mdiChevronRight"
                                                @click="toggleExpanded(parent.id)"
                                            />
                                        </div>
                                        <span class="budget-row-name font-weight-bold text-body-2 text-truncate">{{ parent.name }}</span>
                                        <div class="budget-row-eye d-flex align-center justify-center">
                                            <v-btn
                                                v-if="parentCanHide(parent)"
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
                                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`ep-${parent.id}-${col.year}-${col.month}`">
                                        <div class="budget-data-cell budget-parent-body budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(parentBudgeted(parent, col)) }}</div>
                                        <div class="budget-data-cell budget-parent-body" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(parentActual(parent, col)) }}</div>
                                        <div class="budget-data-cell budget-parent-body" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(parentRemaining(parent, col))]">{{ fmt(parentRemaining(parent, col)) }}</div>
                                    </template>

                                    <!-- Expense sub rows -->
                                    <template v-if="expandedParents.has(parent.id)">
                                        <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                            <template v-if="isSubVisibleInAnyCol(sub.id)">
                                                <div class="budget-name-col budget-sub-row">
                                                    <span class="budget-row-name text-body-2 text-medium-emphasis text-truncate">{{ sub.name }}</span>
                                                    <div class="budget-row-eye d-flex align-center justify-center">
                                                        <v-btn
                                                            v-if="subCanHide(sub.id)"
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
                                                <template v-for="(col, colIdx) in threeMonthColumns" :key="`es-${sub.id}-${col.year}-${col.month}`">
                                                    <div class="budget-data-cell budget-sub-body budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">
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
                                                    <div class="budget-data-cell budget-sub-body" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(subActual(sub.id, col)) }}</div>
                                                    <div class="budget-data-cell budget-sub-body" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(subRemaining(sub.id, col))]">{{ fmt(subRemaining(sub.id, col)) }}</div>
                                                </template>
                                            </template>
                                        </template>
                                    </template>
                                </template>
                            </template>
                        </template>

                        <!-- Divider between Expenses and Income -->
                        <div v-if="showExpenseSection && showIncomeSection" class="budget-full-span budget-section-divider"></div>

                        <!-- ── INCOME section ── -->
                        <template v-if="showIncomeSection">
                            <div class="budget-full-span budget-section-label">{{ tt('Income') }}</div>

                            <template v-for="parent in allIncomeParents" :key="parent.id">
                                <template v-if="isParentVisibleInAnyCol(parent)">
                                    <!-- Income parent row -->
                                    <div class="budget-name-col budget-parent-row">
                                        <div class="budget-expand-cell d-flex align-center justify-center">
                                            <v-btn
                                                density="compact"
                                                variant="text"
                                                size="x-small"
                                                :icon="expandedParents.has(parent.id) ? mdiChevronDown : mdiChevronRight"
                                                @click="toggleExpanded(parent.id)"
                                            />
                                        </div>
                                        <span class="budget-row-name font-weight-bold text-body-2 text-truncate">{{ parent.name }}</span>
                                        <div class="budget-row-eye d-flex align-center justify-center">
                                            <v-btn
                                                v-if="parentCanHide(parent)"
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
                                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`ip-${parent.id}-${col.year}-${col.month}`">
                                        <div class="budget-data-cell budget-parent-body budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(parentBudgeted(parent, col)) }}</div>
                                        <div class="budget-data-cell budget-parent-body" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(parentActual(parent, col)) }}</div>
                                        <div class="budget-data-cell budget-parent-body" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(-parentRemaining(parent, col))]">{{ fmt(-parentRemaining(parent, col)) }}</div>
                                    </template>

                                    <!-- Income sub rows -->
                                    <template v-if="expandedParents.has(parent.id)">
                                        <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                            <template v-if="isSubVisibleInAnyCol(sub.id)">
                                                <div class="budget-name-col budget-sub-row">
                                                    <span class="budget-row-name text-body-2 text-medium-emphasis text-truncate">{{ sub.name }}</span>
                                                    <div class="budget-row-eye d-flex align-center justify-center">
                                                        <v-btn
                                                            v-if="subCanHide(sub.id)"
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
                                                <template v-for="(col, colIdx) in threeMonthColumns" :key="`is-${sub.id}-${col.year}-${col.month}`">
                                                    <div class="budget-data-cell budget-sub-body budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">
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
                                                    <div class="budget-data-cell budget-sub-body" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(subActual(sub.id, col)) }}</div>
                                                    <div class="budget-data-cell budget-sub-body" :class="[{ 'budget-col-current-tint': colIdx === 1 }, diffClass(-subRemaining(sub.id, col))]">{{ fmt(-subRemaining(sub.id, col)) }}</div>
                                                </template>
                                            </template>
                                        </template>
                                    </template>
                                </template>
                            </template>
                        </template>

                        <!-- Divider between Income/Expenses and Savings -->
                        <div v-if="(showExpenseSection || showIncomeSection) && showSavingsSection" class="budget-full-span budget-section-divider"></div>

                        <!-- ── SAVINGS section ── -->
                        <template v-if="showSavingsSection">
                            <div class="budget-full-span budget-section-label">{{ tt('Savings') }}</div>

                            <template v-for="parent in allTransferParents" :key="parent.id">
                                <template v-if="isSavingsParentVisibleInAnyCol(parent)">
                                    <!-- Savings parent row -->
                                    <div class="budget-name-col budget-parent-row">
                                        <div class="budget-expand-cell d-flex align-center justify-center">
                                            <v-btn
                                                density="compact"
                                                variant="text"
                                                size="x-small"
                                                :icon="expandedParents.has(parent.id) ? mdiChevronDown : mdiChevronRight"
                                                @click="toggleExpanded(parent.id)"
                                            />
                                        </div>
                                        <span class="budget-row-name font-weight-bold text-body-2 text-truncate">{{ parent.name }}</span>
                                        <div class="budget-row-eye d-flex align-center justify-center">
                                            <v-btn
                                                v-if="savingsParentCanHide(parent)"
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
                                    <template v-for="(col, colIdx) in threeMonthColumns" :key="`sp-${parent.id}-${col.year}-${col.month}`">
                                        <div class="budget-data-cell budget-parent-body budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(parentSavingsBudgeted(parent, col)) }}</div>
                                        <div class="budget-data-cell budget-parent-body" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(parentSavingsActual(parent, col)) }}</div>
                                        <div class="budget-data-cell budget-parent-body" :class="[{ 'budget-col-current-tint': colIdx === 1 }, savingsDiffClass(parentSavingsRemaining(parent, col))]">{{ fmt(parentSavingsRemaining(parent, col)) }}</div>
                                    </template>

                                    <!-- Savings sub rows -->
                                    <template v-if="expandedParents.has(parent.id)">
                                        <template v-for="sub in (parent.subCategories ?? [])" :key="sub.id">
                                            <template v-if="isSavingsSubVisibleInAnyCol(sub.id)">
                                                <div class="budget-name-col budget-sub-row">
                                                    <span class="budget-row-name text-body-2 text-medium-emphasis text-truncate">{{ sub.name }}</span>
                                                    <div class="budget-row-eye d-flex align-center justify-center">
                                                        <v-btn
                                                            v-if="savingsSubCanHide(sub.id)"
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
                                                <template v-for="(col, colIdx) in threeMonthColumns" :key="`ss-${sub.id}-${col.year}-${col.month}`">
                                                    <div class="budget-data-cell budget-sub-body budget-month-first" :class="{ 'budget-col-current-tint': colIdx === 1 }">
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
                                                            :class="{ 'text-medium-emphasis': subSavingsBudgeted(sub.id, col) === 0 }"
                                                            @click="startEdit(sub.id, col)"
                                                        >{{ fmt(subSavingsBudgeted(sub.id, col)) }}</span>
                                                    </div>
                                                    <div class="budget-data-cell budget-sub-body" :class="{ 'budget-col-current-tint': colIdx === 1 }">{{ fmt(subSavingsActual(sub.id, col)) }}</div>
                                                    <div class="budget-data-cell budget-sub-body" :class="[{ 'budget-col-current-tint': colIdx === 1 }, savingsDiffClass(subSavingsRemaining(sub.id, col))]">{{ fmt(subSavingsRemaining(sub.id, col)) }}</div>
                                                </template>
                                            </template>
                                        </template>
                                    </template>
                                </template>
                            </template>
                        </template>

                        <!-- Empty state -->
                        <div v-if="!hasAnyData" class="budget-full-span d-flex align-center justify-center py-4">
                            <span class="text-body-2 text-medium-emphasis">{{ tt('No categories available to add') }}</span>
                        </div>
                    </template>

                    <!-- Add Category button -->
                    <div v-if="hasHiddenItems" class="budget-full-span pa-3">
                        <v-btn
                            variant="tonal"
                            size="small"
                            :prepend-icon="mdiPlus"
                            @click="showAddCategoryDialog = true"
                        >
                            {{ tt('Add Category') }}
                        </v-btn>
                    </div>

                </div>
            </div>
        </v-col>
    </v-row>

    <!-- Copy Budget Dialog -->
    <v-dialog v-model="showCopyDialog" max-width="520" :persistent="copyLoading">
        <v-card :title="tt('Copy Budget')">
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
                    <div class="text-subtitle-2 font-weight-bold mb-2">
                        {{ item.parentCategoryName }} › {{ item.subcategoryName }}
                    </div>

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
                <v-btn :disabled="copyLoading" @click="showCopyDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn v-if="copyStep === 1" color="primary" :loading="copyLoading" @click="advanceCopyStep">{{ tt('Next') }}</v-btn>
                <v-btn v-else color="primary" :loading="copyLoading" @click="executeCopy">{{ tt('Confirm') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <!-- Add Category Dialog -->
    <v-dialog v-model="showAddCategoryDialog" max-width="420">
        <v-card :title="tt('Add Category')">
            <v-card-text>
                <div v-if="!hasHiddenItems" class="text-body-2 text-medium-emphasis">
                    {{ tt('No categories available to add') }}
                </div>
                <v-list v-else density="compact">
                    <template v-for="parent in hiddenIncomeParents" :key="'ip-' + parent.id">
                        <v-list-item :title="parent.name" :subtitle="tt('Income')" class="cursor-pointer" @click="onAddParent(parent)">
                            <template #prepend><v-icon :icon="mdiPlus" size="18" class="me-1" /></template>
                        </v-list-item>
                    </template>
                    <template v-for="item in hiddenIncomeSubsUnderVisibleParent" :key="'is-' + item.sub.id">
                        <v-list-item :title="item.sub.name" :subtitle="item.parent.name + ' (' + tt('Income') + ')'" class="cursor-pointer" @click="onAddSub(item.sub.id)">
                            <template #prepend><v-icon :icon="mdiPlus" size="18" class="me-1" /></template>
                        </v-list-item>
                    </template>
                    <template v-for="parent in hiddenExpenseParents" :key="'ep-' + parent.id">
                        <v-list-item :title="parent.name" :subtitle="tt('Expenses')" class="cursor-pointer" @click="onAddParent(parent)">
                            <template #prepend><v-icon :icon="mdiPlus" size="18" class="me-1" /></template>
                        </v-list-item>
                    </template>
                    <template v-for="item in hiddenExpenseSubsUnderVisibleParent" :key="'es-' + item.sub.id">
                        <v-list-item :title="item.sub.name" :subtitle="item.parent.name + ' (' + tt('Expenses') + ')'" class="cursor-pointer" @click="onAddSub(item.sub.id)">
                            <template #prepend><v-icon :icon="mdiPlus" size="18" class="me-1" /></template>
                        </v-list-item>
                    </template>
                    <template v-for="parent in hiddenSavingsParents" :key="'sp-' + parent.id">
                        <v-list-item :title="parent.name" :subtitle="tt('Savings')" class="cursor-pointer" @click="onAddParent(parent)">
                            <template #prepend><v-icon :icon="mdiPlus" size="18" class="me-1" /></template>
                        </v-list-item>
                    </template>
                    <template v-for="item in hiddenSavingsSubsUnderVisibleParent" :key="'ss-' + item.sub.id">
                        <v-list-item :title="item.sub.name" :subtitle="item.parent.name + ' (' + tt('Savings') + ')'" class="cursor-pointer" @click="onAddSub(item.sub.id)">
                            <template #prepend><v-icon :icon="mdiPlus" size="18" class="me-1" /></template>
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
    savingsActuals,
    threeMonthColumns,
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

// ---------- UI state ----------

const loading = ref<boolean>(true);
const spentByMonth = ref<Record<string, Record<string, number>>>({});
const expandedParents = ref<Set<string>>(new Set());
const saving = ref<boolean>(false);
const showIncomeSection = ref<boolean>(true);
const showExpenseSection = ref<boolean>(true);
const showSavingsSection = ref<boolean>(true);

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

function savingsDiffClass(diff: number): string {
    if (diff > 0) return 'text-error';
    if (diff < 0) return 'text-success';
    return '';
}

// ---------- Per-column visibility ----------

function isSubVisibleInCol(subId: string, col: { year: number; month: number }): boolean {
    return !hiddenCategoryIds.value.has(subId) || subBudgeted(subId, col) > 0 || subActual(subId, col) > 0;
}

function isParentVisibleInCol(parent: TransactionCategory, col: { year: number; month: number }): boolean {
    if (!hiddenCategoryIds.value.has(parent.id)) return true;
    return (parent.subCategories ?? []).some(sub => subBudgeted(sub.id, col) > 0 || subActual(sub.id, col) > 0);
}

function isSavingsSubVisibleInCol(subId: string, col: { year: number; month: number }): boolean {
    return !hiddenCategoryIds.value.has(subId) || subSavingsBudgeted(subId, col) > 0 || subSavingsActual(subId, col) !== 0;
}

function isSavingsParentVisibleInCol(parent: TransactionCategory, col: { year: number; month: number }): boolean {
    if (!hiddenCategoryIds.value.has(parent.id)) return true;
    return (parent.subCategories ?? []).some(sub => subSavingsBudgeted(sub.id, col) > 0 || subSavingsActual(sub.id, col) !== 0);
}

// ---------- Cross-column visibility (shared row layout) ----------

function isParentVisibleInAnyCol(parent: TransactionCategory): boolean {
    return threeMonthColumns.value.some(col => isParentVisibleInCol(parent, col));
}

function isSubVisibleInAnyCol(subId: string): boolean {
    return threeMonthColumns.value.some(col => isSubVisibleInCol(subId, col));
}

function isSavingsParentVisibleInAnyCol(parent: TransactionCategory): boolean {
    return threeMonthColumns.value.some(col => isSavingsParentVisibleInCol(parent, col));
}

function isSavingsSubVisibleInAnyCol(subId: string): boolean {
    return threeMonthColumns.value.some(col => isSavingsSubVisibleInCol(subId, col));
}

// ---------- Hide button eligibility (zero budget across all three months) ----------

function parentCanHide(parent: TransactionCategory): boolean {
    return !hiddenCategoryIds.value.has(parent.id)
        && threeMonthColumns.value.every(col => parentBudgeted(parent, col) === 0);
}

function subCanHide(subId: string): boolean {
    return !hiddenCategoryIds.value.has(subId)
        && threeMonthColumns.value.every(col => subBudgeted(subId, col) === 0);
}

function savingsParentCanHide(parent: TransactionCategory): boolean {
    return !hiddenCategoryIds.value.has(parent.id)
        && threeMonthColumns.value.every(col => parentSavingsBudgeted(parent, col) === 0);
}

function savingsSubCanHide(subId: string): boolean {
    return !hiddenCategoryIds.value.has(subId)
        && threeMonthColumns.value.every(col => subSavingsBudgeted(subId, col) === 0);
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

// ---------- Savings amount helpers ----------

function subSavingsBudgeted(subcatId: string, col: { year: number; month: number }): number {
    return budgetTargets.value[`${col.year}-${col.month}`]?.[subcatId]?.amount ?? 0;
}

function subSavingsActual(subcatId: string, col: { year: number; month: number }): number {
    return getSavingsNet(subcatId, col.year, col.month);
}

function subSavingsRemaining(subcatId: string, col: { year: number; month: number }): number {
    return subSavingsBudgeted(subcatId, col) - subSavingsActual(subcatId, col);
}

function parentSavingsBudgeted(parent: TransactionCategory, col: { year: number; month: number }): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subSavingsBudgeted(sub.id, col), 0);
}

function parentSavingsActual(parent: TransactionCategory, col: { year: number; month: number }): number {
    return (parent.subCategories ?? []).reduce((sum, sub) => sum + subSavingsActual(sub.id, col), 0);
}

function parentSavingsRemaining(parent: TransactionCategory, col: { year: number; month: number }): number {
    return parentSavingsBudgeted(parent, col) - parentSavingsActual(parent, col);
}

// ---------- Column summary totals ----------

function colIncomeBudgeted(col: { year: number; month: number }): number {
    return allIncomeParents.value.flatMap(p => p.subCategories ?? []).reduce((sum, sub) => sum + subBudgeted(sub.id, col), 0);
}

function colIncomeActual(col: { year: number; month: number }): number {
    return allIncomeParents.value.flatMap(p => p.subCategories ?? []).reduce((sum, sub) => sum + subActual(sub.id, col), 0);
}

function colIncomeDiff(col: { year: number; month: number }): number {
    return colIncomeActual(col) - colIncomeBudgeted(col);
}

function colExpenseBudgeted(col: { year: number; month: number }): number {
    return allExpenseParents.value.flatMap(p => p.subCategories ?? []).reduce((sum, sub) => sum + subBudgeted(sub.id, col), 0);
}

function colExpenseActual(col: { year: number; month: number }): number {
    return allExpenseParents.value.flatMap(p => p.subCategories ?? []).reduce((sum, sub) => sum + subActual(sub.id, col), 0);
}

function colExpenseDiff(col: { year: number; month: number }): number {
    return colExpenseBudgeted(col) - colExpenseActual(col);
}

function colSavingsBudgeted(col: { year: number; month: number }): number {
    return allTransferParents.value.flatMap(p => p.subCategories ?? []).reduce((sum, sub) => sum + subSavingsBudgeted(sub.id, col), 0);
}

function colSavingsActual(col: { year: number; month: number }): number {
    return allTransferParents.value.flatMap(p => p.subCategories ?? []).reduce((sum, sub) => sum + subSavingsActual(sub.id, col), 0);
}

function colSavingsDiff(col: { year: number; month: number }): number {
    return colSavingsBudgeted(col) - colSavingsActual(col);
}

function colNetBudgeted(col: { year: number; month: number }): number {
    return colIncomeBudgeted(col) - colExpenseBudgeted(col) - colSavingsBudgeted(col);
}

function colNetActual(col: { year: number; month: number }): number {
    return colIncomeActual(col) - colExpenseActual(col) - colSavingsActual(col);
}

function colNetDiff(col: { year: number; month: number }): number {
    return colNetActual(col) - colNetBudgeted(col);
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

// ---------- Hide/show ----------

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
    if (!savingsActuals.value[key]) tasks.push(loadSavingsActuals(year, month));
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
            ...allTransferParents.value.map(p => p.id),
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
/* ── Timeline ── */
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

/* ── Outer scroll wrapper ── */
.budget-table-wrap {
    overflow-x: auto;
    overflow-y: auto;
    height: calc(100vh - 200px);
    min-height: 320px;
    width: 100%;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 8px;
}

/* ── CSS grid: 1 name col + 9 data cols (3 months × 3 sub-cols) ── */
.budget-table {
    display: grid;
    grid-template-columns: minmax(180px, auto) repeat(9, minmax(90px, 1fr));
}

/* ── Full-width spanning utility ── */
.budget-full-span {
    grid-column: span 10;
}

/* ── Sticky name column (shared base for all name cells) ── */
.budget-name-col {
    position: sticky;
    left: 0;
    z-index: 2;
    background: rgb(var(--v-theme-surface));
    display: flex;
    align-items: center;
    overflow: hidden;
}

/* ── Header row 1: month titles ── */
.budget-th {
    position: sticky;
    top: 0;
    z-index: 3;
    background: rgb(var(--v-theme-surface));
    font-size: 0.75rem;
    color: rgba(var(--v-theme-on-surface), 0.6);
    padding: 8px 8px 6px;
    text-align: right;
    display: flex;
    align-items: center;
    justify-content: flex-end;
}

.budget-th-corner {
    z-index: 4;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.12);
}

.budget-month-header {
    justify-content: center;
    font-size: 0.875rem;
    font-weight: 600;
    color: rgba(var(--v-theme-on-surface), 0.87);
    padding: 12px 8px 10px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-left: 2px solid rgba(var(--v-theme-on-surface), 0.2);
}

/* Left border on every Budgeted (first) cell of each month group */
.budget-month-first {
    border-left: 2px solid rgba(var(--v-theme-on-surface), 0.2);
}

.budget-month-header--current {
    background-color: color-mix(in srgb, rgb(var(--v-theme-surface)) 94%, rgb(var(--v-theme-primary)) 6%);
    border-bottom: 2px solid rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-primary));
}

/* ── Header row 2: B / A / R labels ── */
.budget-th-sub-corner {
    top: 40px;
    z-index: 4;
    border-bottom: 2px solid rgba(var(--v-theme-on-surface), 0.12);
}

.budget-th-sub {
    position: sticky;
    top: 40px;
    z-index: 3;
    font-size: 0.72rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.5);
    padding: 4px 8px 6px;
    border-bottom: 2px solid rgba(var(--v-theme-on-surface), 0.12);
    display: flex;
    align-items: center;
    justify-content: flex-end;
    background: rgb(var(--v-theme-surface));
}

/* ── Current month column tint ── */
.budget-col-current-tint {
    background-color: rgba(var(--v-theme-primary), 0.04);
}

/* Sticky header cells must have an opaque background */
.budget-th-sub.budget-col-current-tint {
    background-color: color-mix(in srgb, rgb(var(--v-theme-surface)) 96%, rgb(var(--v-theme-primary)) 4%);
}

/* ── Summary section borders ── */
.budget-summary-border-top {
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    height: 0;
}

.budget-summary-border-bottom {
    border-bottom: 2px solid rgba(var(--v-theme-on-surface), 0.15);
    height: 0;
    margin-bottom: 4px;
}

/* ── Summary row cells ── */
.budget-summary-name {
    padding: 6px 8px 6px 6px;
    background: rgba(var(--v-theme-on-surface), 0.03);
    /* position sticky already from .budget-name-col */
}

/* Override sticky bg for summary name so tint is visible through solid bg */
.budget-name-col.budget-summary-name {
    background: color-mix(in srgb, rgb(var(--v-theme-surface)) 97%, currentColor 3%);
}

.budget-summary-cell {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 6px 8px;
    font-size: 0.8125rem;
    background-color: rgba(var(--v-theme-on-surface), 0.03);
}

.budget-summary-cell.budget-col-current-tint {
    background-color: rgba(var(--v-theme-primary), 0.06);
}

/* ── Section labels ── */
.budget-section-label {
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: rgba(var(--v-theme-on-surface), 0.45);
    padding: 10px 8px 4px;
}

/* ── Section divider (between Expenses/Income/Savings) ── */
.budget-section-divider {
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.1);
    margin: 8px 0;
    height: 0;
}

/* ── Parent category rows ── */
.budget-parent-row {
    padding: 2px 6px 2px 0;
    min-height: 36px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.07);
}

.budget-parent-body {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 4px 8px;
    font-size: 0.8125rem;
    min-height: 36px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.07);
}

/* ── Sub-category rows ── */
.budget-sub-row {
    padding: 2px 6px 2px 36px;
    min-height: 32px;
}

.budget-sub-body {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    padding: 3px 8px;
    font-size: 0.8125rem;
    min-height: 32px;
}

/* ── Name cell internals ── */
.budget-row-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
}

.budget-row-eye {
    width: 28px;
    min-width: 28px;
    flex-shrink: 0;
}

.budget-expand-cell {
    width: 32px;
    min-width: 32px;
    flex-shrink: 0;
}

/* ── Eye button: fade in on hover of its own name cell ── */
.budget-parent-row .budget-eye-btn,
.budget-sub-row .budget-eye-btn {
    opacity: 0;
    transition: opacity 0.15s;
}

.budget-parent-row:hover .budget-eye-btn,
.budget-sub-row:hover .budget-eye-btn {
    opacity: 1;
}

/* ── Inline edit ── */
.budget-edit-field {
    width: 82px;
}

.budget-budgeted-span {
    display: block;
    text-align: right;
    width: 100%;
}
</style>
