<template>
    <f7-sheet swipe-to-close swipe-handler=".swipe-handler" class="time-selection-sheet" style="height:auto"
              :opened="show" @sheet:open="onSheetOpen" @sheet:closed="onSheetClosed">
        <f7-toolbar class="toolbar-with-swipe-handler">
            <div class="swipe-handler"></div>
            <div class="left"></div>
            <div class="right">
                <f7-button round fill icon-f7="checkmark_alt" :aria-label="tt('Apply')" @click="confirm"></f7-button>
            </div>
        </f7-toolbar>
        <f7-page-content class="margin-top">
            <div class="block no-margin no-padding padding-vertical-half">
                <div class="time-picker-container" ref="timePickerContainer">
                    <div class="picker picker-inline picker-3d">
                        <div class="picker-columns">
                            <div class="picker-column" v-if="!is24Hour && isMeridiemIndicatorFirst">
                                <div class="picker-items picker-items-meridiem-indicator-first"
                                     @scroll="onPickerColumnScroll('picker-items-meridiem-indicator-first', 'picker-meridiem-indicator', false)">
                                    <div :class="{ 'picker-item': true, 'picker-meridiem-indicator': true, 'picker-item-selected': currentMeridiemIndicator === item.value }"
                                         :key="item.value" :data-value="item.value"
                                         @click="currentMeridiemIndicator = item.value; scrollToSelectedItem('picker-items-meridiem-indicator-first', 'picker-meridiem-indicator', item.value)"
                                         v-for="item in meridiemItems">
                                        <span>{{ item.name }}</span>
                                    </div>
                                </div>
                            </div>
                            <div class="picker-column">
                                <div class="picker-items picker-items-hour"
                                     @scroll="onPickerColumnScroll('picker-items-hour', 'picker-hour', false)"
                                     @scrollend="onPickerColumnScroll('picker-items-hour', 'picker-hour', true)">
                                    <div :class="{ 'picker-item': true, 'picker-hour': true, 'picker-item-selected': currentHour === item.value }"
                                         :key="`${item.itemsIndex}_${item.value}`" :data-items-index="item.itemsIndex" :data-value="item.value"
                                         @click="currentHour = item.value; scrollToSelectedItem('picker-items-hour', 'picker-hour', item.value)"
                                         v-for="item in hourItems">
                                        <span :style="getTimerPickerItemStyle(item.value, currentHour, item.itemsIndex, hourItems)">{{ item.value }}</span>
                                    </div>
                                </div>
                            </div>
                            <div class="picker-column picker-column-divider">:</div>
                            <div class="picker-column">
                                <div class="picker-items picker-items-minute"
                                     @scroll="onPickerColumnScroll('picker-items-minute', 'picker-minute', false)"
                                     @scrollend="onPickerColumnScroll('picker-items-minute', 'picker-minute', true)">
                                    <div :class="{ 'picker-item': true, 'picker-minute': true, 'picker-item-selected': currentMinute === item.value }"
                                         :key="`${item.itemsIndex}_${item.value}`" :data-items-index="item.itemsIndex" :data-value="item.value"
                                         @click="currentMinute = item.value; scrollToSelectedItem('picker-items-minute', 'picker-minute', item.value)"
                                         v-for="item in minuteItems">
                                        <span :style="getTimerPickerItemStyle(item.value, currentMinute, item.itemsIndex, minuteItems)">{{ item.value }}</span>
                                    </div>
                                </div>
                            </div>
                            <div class="picker-column" v-if="!is24Hour && !isMeridiemIndicatorFirst">
                                <div class="picker-items picker-items-meridiem-indicator-last"
                                     @scroll="onPickerColumnScroll('picker-items-meridiem-indicator-last', 'picker-meridiem-indicator', false)">
                                    <div :class="{ 'picker-item': true, 'picker-meridiem-indicator': true, 'picker-item-selected': currentMeridiemIndicator === item.value }"
                                         :key="item.value" :data-value="item.value"
                                         @click="currentMeridiemIndicator = item.value; scrollToSelectedItem('picker-items-meridiem-indicator-last', 'picker-meridiem-indicator', item.value)"
                                         v-for="item in meridiemItems">
                                        <span>{{ item.name }}</span>
                                    </div>
                                </div>
                            </div>
                            <div class="picker-center-highlight"></div>
                        </div>
                    </div>
                </div>
            </div>
        </f7-page-content>
    </f7-sheet>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { type TimePickerValue, useDateTimeSelectionBase } from '@/components/base/DateTimeSelectionBase.ts';

import { NumeralSystem } from '@/core/numeral.ts';

import { isDefined } from '@/lib/common.ts';
import { getHourIn12HourFormat, getAMOrPM, getCombinedDateAndTimeValues } from '@/lib/datetime.ts';

const props = defineProps<{
    modelValue?: string; // "HH:mm" in 24-hour format
    show: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
    (e: 'update:show', value: boolean): void;
}>();

const { tt, getCurrentNumeralSystemType } = useI18n();

const {
    is24Hour,
    isHourTwoDigits,
    isMinuteTwoDigits,
    isMeridiemIndicatorFirst,
    meridiemItems,
    getDisplayTimeValue,
    generateAllHours,
    generateAllMinutesOrSeconds
} = useDateTimeSelectionBase();

const timePickerContainer = useTemplateRef<HTMLDivElement>('timePickerContainer');

let resetTimePickerItemPositionItemsClass: string | undefined = undefined;
let resetTimePickerItemPositionItemClass: string | undefined = undefined;
let resetTimePickerItemPositionItemsLastOffsetTop: number | undefined = undefined;
let resetTimePickerItemPositionCheckedFrames: number | undefined = undefined;

const draftTimeOfDay = ref<string>('00:00');
const timePickerContainerHeight = ref<number | undefined>(undefined);
const timePickerItemHeight = ref<number | undefined>(undefined);

const numeralSystem = computed<NumeralSystem>(() => getCurrentNumeralSystemType());

const hourItems = computed<TimePickerValue[]>(() => generateAllHours(3, isHourTwoDigits.value));
const minuteItems = computed<TimePickerValue[]>(() => generateAllMinutesOrSeconds(3, isMinuteTwoDigits.value));

function parseHour24(): number {
    const parts = draftTimeOfDay.value.split(':');
    const hour = parseInt(parts[0] ?? '0', 10);
    return isNaN(hour) ? 0 : Math.min(23, Math.max(0, hour));
}

function parseMinute(): number {
    const parts = draftTimeOfDay.value.split(':');
    const minute = parseInt(parts[1] ?? '0', 10);
    return isNaN(minute) ? 0 : Math.min(59, Math.max(0, minute));
}

function updateTime(hour: string, minute: string, meridiemIndicator: string): void {
    const dummyDate = new Date(2020, 0, 1, parseHour24(), parseMinute());
    const combined = getCombinedDateAndTimeValues(dummyDate, numeralSystem.value, hour, minute, '0', meridiemIndicator, is24Hour.value);
    draftTimeOfDay.value = `${combined.getHours().toString().padStart(2, '0')}:${combined.getMinutes().toString().padStart(2, '0')}`;
}

const currentMeridiemIndicator = computed<string>({
    get: () => {
        return getAMOrPM(parseHour24());
    },
    set: (value: string) => {
        updateTime(currentHour.value, currentMinute.value, value);
    }
});
const currentHour = computed<string>({
    get: () => {
        return getDisplayTimeValue(is24Hour.value ? parseHour24() : getHourIn12HourFormat(parseHour24()), isHourTwoDigits.value);
    },
    set: (value: string) => {
        updateTime(value, currentMinute.value, currentMeridiemIndicator.value);
    }
});
const currentMinute = computed<string>({
    get: () => {
        return getDisplayTimeValue(parseMinute(), isMinuteTwoDigits.value);
    },
    set: (value: string) => {
        updateTime(currentHour.value, value, currentMeridiemIndicator.value);
    }
});

function confirm(): void {
    emit('update:modelValue', draftTimeOfDay.value);
    emit('update:show', false);
}

function getTimerPickerItemStyle(textualValue: string, textualCurrentValue: string, itemsIndex: number, values: TimePickerValue[]): string {
    if (!timePickerContainerHeight.value || !timePickerItemHeight.value) {
        return '';
    }

    const minValue = parseInt(values[0]!.value);
    const maxValue = parseInt(values[values.length - 1]!.value);
    const value = parseInt(textualValue, 10);
    const currentValue = parseInt(textualCurrentValue, 10);
    let valueDiff = value - currentValue;

    if (Math.abs(valueDiff) >= 5) {
        if (itemsIndex === 0 && maxValue - 5 < value && currentValue < minValue + 5) {
            valueDiff = value - (maxValue + currentValue + 1);
        } else if (itemsIndex === 2 && maxValue - 5 < currentValue && value < minValue + 5) {
            valueDiff = (maxValue + value + 1) - currentValue;
        }
    }

    const angle = -24 * valueDiff;

    if (angle > 180) {
        return '';
    }
    if (angle < -180) {
        return '';
    }

    return `transform: translate3d(0, ${-valueDiff * timePickerItemHeight.value}px, -100px) rotateX(${angle}deg)`;
}

function initTimePickerStyle(): void {
    const pickerItems = timePickerContainer.value?.querySelectorAll('.picker-item');
    const firstPickerItem = pickerItems ? pickerItems[0] : null;

    if (timePickerContainer.value) {
        timePickerContainerHeight.value = timePickerContainer.value.offsetHeight as number;
    }

    if (firstPickerItem && 'offsetHeight' in firstPickerItem) {
        timePickerItemHeight.value = firstPickerItem.offsetHeight as number;
    }

    if (timePickerContainer.value && firstPickerItem && 'offsetHeight' in firstPickerItem) {
        timePickerContainer.value.style.setProperty('--f7-picker-scroll-padding', `${(timePickerContainer.value.offsetHeight - (firstPickerItem.offsetHeight as number)) / 2}px`);
    }
}

function scrollAllTimeSelectedItems(): void {
    scrollToSelectedItem('picker-items-hour', 'picker-hour', currentHour.value);
    scrollToSelectedItem('picker-items-minute', 'picker-minute', currentMinute.value);
    scrollToSelectedItem('picker-items-meridiem-indicator-first', 'picker-meridiem-indicator', currentMeridiemIndicator.value);
    scrollToSelectedItem('picker-items-meridiem-indicator-last', 'picker-meridiem-indicator', currentMeridiemIndicator.value);
}

function scrollTimeSelectedItems(itemsClass: string, itemClass: string): void {
    switch (resetTimePickerItemPositionItemClass) {
        case 'picker-hour':
            scrollToSelectedItem(itemsClass, itemClass, currentHour.value);
            break;
        case 'picker-minute':
            scrollToSelectedItem(itemsClass, itemClass, currentMinute.value);
            break;
    }
}

function scrollToSelectedItem(itemsClass: string, itemClass: string, value: string): void {
    const itemsElement = timePickerContainer.value?.querySelector(`.${itemsClass}`);
    const itemElements = itemsElement?.querySelectorAll(`.${itemClass}`);

    if (!itemsElement || !itemElements || !itemElements.length) {
        return;
    }

    for (let i = 0; i < itemElements.length; i++) {
        const itemElement = itemElements[i] as HTMLElement;

        if ('offsetHeight' in itemsElement && 'offsetTop' in itemElement && 'offsetHeight' in itemElement
            && (!itemElement.hasAttribute('data-items-index') || itemElement.getAttribute('data-items-index') === '1')
            && itemElement.getAttribute('data-value') === value) {
            itemsElement.scrollTop = (itemElement.offsetTop as number) - ((itemsElement.offsetHeight as number) / 2) + ((itemElement.offsetHeight as number) / 2);
            break;
        }
    }
}

function onPickerColumnScroll(itemsClass: string, itemClass: string, scrollEnd: boolean): void {
    const itemsElement = timePickerContainer.value?.querySelector(`.${itemsClass}`);
    const itemElements = itemsElement?.querySelectorAll(`.${itemClass}`);
    const firstPickerElement = itemElements ? itemElements[0] : null;

    if (!itemsElement || !itemElements || !itemElements.length || !firstPickerElement || !('offsetHeight' in firstPickerElement)) {
        return;
    }

    const itemHeight = firstPickerElement.offsetHeight as number;
    const scrollTop = itemsElement?.scrollTop || 0;
    const index = Math.round(scrollTop / itemHeight);
    const selectedItem = itemElements[index];

    if (selectedItem) {
        const value = selectedItem.getAttribute('data-value');
        const itemsIndex = selectedItem.getAttribute('data-items-index');

        if (value) {
            switch (itemClass) {
                case 'picker-hour':
                    currentHour.value = value;
                    break;
                case 'picker-minute':
                    currentMinute.value = value;
                    break;
                case 'picker-meridiem-indicator':
                    currentMeridiemIndicator.value = value;
                    break;
            }

            if (itemsIndex === '0' || itemsIndex === '2') {
                if (scrollEnd) {
                    scrollToSelectedItem(itemsClass, itemClass, value);
                } else {
                    if (resetTimePickerItemPositionItemsClass && resetTimePickerItemPositionItemClass
                        && resetTimePickerItemPositionItemsClass !== itemsClass && resetTimePickerItemPositionItemClass !== itemClass) {
                        scrollTimeSelectedItems(resetTimePickerItemPositionItemsClass, resetTimePickerItemPositionItemClass);
                        resetTimePickerItemPositionItemsClass = undefined;
                        resetTimePickerItemPositionItemClass = undefined;
                        resetTimePickerItemPositionItemsLastOffsetTop = undefined;
                        resetTimePickerItemPositionCheckedFrames = undefined;
                    }

                    if (!resetTimePickerItemPositionCheckedFrames && window.requestAnimationFrame) {
                        resetTimePickerItemPositionItemsClass = itemsClass;
                        resetTimePickerItemPositionItemClass = itemClass;
                        resetTimePickerItemPositionItemsLastOffsetTop = itemsElement.scrollTop;
                        resetTimePickerItemPositionCheckedFrames = 1;
                        window.requestAnimationFrame(delayCheckAndResetTimePickerItemPosition);
                    }
                }
            }
        }
    }
}

function delayCheckAndResetTimePickerItemPosition(): void {
    if (!resetTimePickerItemPositionItemsClass || !resetTimePickerItemPositionItemClass || !isDefined(resetTimePickerItemPositionItemsLastOffsetTop) || !isDefined(resetTimePickerItemPositionCheckedFrames)) {
        return;
    }

    const itemsElement = timePickerContainer.value?.querySelector(`.${resetTimePickerItemPositionItemsClass}`);

    if (!itemsElement) {
        return;
    }

    if (itemsElement.scrollTop === resetTimePickerItemPositionItemsLastOffsetTop) {
        resetTimePickerItemPositionCheckedFrames++;
    } else {
        resetTimePickerItemPositionItemsLastOffsetTop = itemsElement.scrollTop;
        resetTimePickerItemPositionCheckedFrames = 0;
    }

    if (resetTimePickerItemPositionCheckedFrames > 3) {
        scrollTimeSelectedItems(resetTimePickerItemPositionItemsClass, resetTimePickerItemPositionItemClass);
        resetTimePickerItemPositionItemsClass = undefined;
        resetTimePickerItemPositionItemClass = undefined;
        resetTimePickerItemPositionItemsLastOffsetTop = undefined;
        resetTimePickerItemPositionCheckedFrames = undefined;
        return;
    }

    window.requestAnimationFrame(delayCheckAndResetTimePickerItemPosition);
}

function onSheetOpen(): void {
    draftTimeOfDay.value = props.modelValue || '00:00';

    nextTick(() => {
        initTimePickerStyle();
        scrollAllTimeSelectedItems();
    });
}

function onSheetClosed(): void {
    emit('update:show', false);
}
</script>

<style>
.time-selection-sheet .time-picker-container .picker-columns {
    justify-content: space-evenly;
}

.time-selection-sheet .picker-hour,
.time-selection-sheet .picker-minute {
    font-variant-numeric: tabular-nums;
}
</style>
