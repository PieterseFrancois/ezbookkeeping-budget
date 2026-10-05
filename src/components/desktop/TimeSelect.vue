<template>
    <v-select
        persistent-placeholder
        :readonly="readonly"
        :disabled="disabled"
        :label="label"
        :menu-props="{ contentClass: 'time-select-menu' }"
        v-model="timeOfDay"
    >
        <template #selection>
            <span class="text-truncate cursor-pointer">{{ displayTime }}</span>
        </template>

        <template #no-data>
            <div class="time-select-time-picker-container">
                <v-btn class="px-3" color="primary" variant="flat"
                       v-if="!is24Hour && isMeridiemIndicatorFirst"
                       @click="toggleMeridiemIndicator">
                    {{ tt(`datetime.${currentMeridiemIndicator}.content`) }}
                </v-btn>
                <v-autocomplete eager
                                density="compact"
                                max-width="70px"
                                item-title="value"
                                item-value="value"
                                auto-select-first="exact"
                                :items="hourItems"
                                :hide-no-data="true"
                                v-model="currentHour"
                />
                <span>:</span>
                <v-autocomplete eager
                                density="compact"
                                max-width="70px"
                                item-title="value"
                                item-value="value"
                                auto-select-first="exact"
                                :items="minuteItems"
                                :hide-no-data="true"
                                v-model="currentMinute"
                />
                <v-btn class="px-3" color="primary" variant="flat"
                       v-if="!is24Hour && !isMeridiemIndicatorFirst"
                       @click="toggleMeridiemIndicator">
                    {{ tt(`datetime.${currentMeridiemIndicator}.content`) }}
                </v-btn>
            </div>
        </template>
    </v-select>
</template>

<script setup lang="ts">
import { computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { type TimePickerValue, useDateTimeSelectionBase } from '@/components/base/DateTimeSelectionBase.ts';

import { MeridiemIndicator } from '@/core/datetime.ts';
import { type NumeralSystem } from '@/core/numeral.ts';
import { getHourIn12HourFormat, getAMOrPM, getCombinedDateAndTimeValues } from '@/lib/datetime.ts';

const props = defineProps<{
    modelValue?: string; // "HH:mm" in 24-hour format
    disabled?: boolean;
    readonly?: boolean;
    label?: string;
}>();

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
}>();

const { tt, getCurrentNumeralSystemType } = useI18n();

const {
    is24Hour,
    isHourTwoDigits,
    isMinuteTwoDigits,
    isMeridiemIndicatorFirst,
    getDisplayTimeValue,
    generateAllHours,
    generateAllMinutesOrSeconds
} = useDateTimeSelectionBase();

const numeralSystem = computed<NumeralSystem>(() => getCurrentNumeralSystemType());

const timeOfDay = computed<string>({
    get: () => props.modelValue || '00:00',
    set: (value: string) => {
        emit('update:modelValue', value);
    }
});

const hourItems = computed<TimePickerValue[]>(() => generateAllHours(1, isHourTwoDigits.value));
const minuteItems = computed<TimePickerValue[]>(() => generateAllMinutesOrSeconds(1, isMinuteTwoDigits.value));

function parseHour24(): number {
    const parts = timeOfDay.value.split(':');
    const hour = parseInt(parts[0] ?? '0', 10);
    return isNaN(hour) ? 0 : Math.min(23, Math.max(0, hour));
}

function parseMinute(): number {
    const parts = timeOfDay.value.split(':');
    const minute = parseInt(parts[1] ?? '0', 10);
    return isNaN(minute) ? 0 : Math.min(59, Math.max(0, minute));
}

function updateTime(hour: string, minute: string, meridiemIndicator: string): void {
    const dummyDate = new Date(2020, 0, 1, parseHour24(), parseMinute());
    const combined = getCombinedDateAndTimeValues(dummyDate, numeralSystem.value, hour, minute, '0', meridiemIndicator, is24Hour.value);
    timeOfDay.value = `${combined.getHours().toString().padStart(2, '0')}:${combined.getMinutes().toString().padStart(2, '0')}`;
}

const currentMeridiemIndicator = computed<string>({
    get: () => {
        return getAMOrPM(parseHour24());
    },
    set: (value: string) => {
        if (value !== MeridiemIndicator.AM.name && value !== MeridiemIndicator.PM.name) {
            return;
        }

        updateTime(currentHour.value, currentMinute.value, value);
    }
});
const currentHour = computed<string>({
    get: () => {
        return getDisplayTimeValue(is24Hour.value ? parseHour24() : getHourIn12HourFormat(parseHour24()), isHourTwoDigits.value);
    },
    set: (value: string) => {
        const hour = numeralSystem.value.parseInt(value);

        if (isNaN(hour) || hour < 0 || (is24Hour.value ? hour > 23 : hour > 12)) {
            return;
        }

        updateTime(value, currentMinute.value, currentMeridiemIndicator.value);
    }
});
const currentMinute = computed<string>({
    get: () => {
        return getDisplayTimeValue(parseMinute(), isMinuteTwoDigits.value);
    },
    set: (value: string) => {
        const minute = numeralSystem.value.parseInt(value);

        if (isNaN(minute) || minute < 0 || minute > 59) {
            return;
        }

        updateTime(currentHour.value, value, currentMeridiemIndicator.value);
    }
});

const displayTime = computed<string>(() => {
    if (is24Hour.value) {
        return `${currentHour.value}:${currentMinute.value}`;
    }

    const meridiemText = tt(`datetime.${currentMeridiemIndicator.value}.content`);

    return isMeridiemIndicatorFirst.value
        ? `${meridiemText} ${currentHour.value}:${currentMinute.value}`
        : `${currentHour.value}:${currentMinute.value} ${meridiemText}`;
});

function toggleMeridiemIndicator(): void {
    if (currentMeridiemIndicator.value === MeridiemIndicator.AM.name) {
        currentMeridiemIndicator.value = MeridiemIndicator.PM.name;
    } else {
        currentMeridiemIndicator.value = MeridiemIndicator.AM.name;
    }
}
</script>

<style>
.time-select-menu {
    max-height: inherit !important;
}

.time-select-time-picker-container {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: var(--dp-menu-padding);
    column-gap: 8px;
}

.time-select-time-picker-container .v-autocomplete.v-input--density-compact {
    --v-input-control-height: 38px;
    --v-field-input-padding-top: 4px;
    --v-field-input-padding-bottom: 4px;
}

.time-select-time-picker-container .v-autocomplete.v-input--density-compact .v-field {
    --v-field-padding-start: 12px;
    --v-field-padding-end: 0;
}

.time-select-time-picker-container .v-autocomplete.v-input--density-compact .v-field__input {
    min-height: 38px !important;
}

.time-select-time-picker-container .v-autocomplete.v-input--density-compact .v-field__append-inner .v-autocomplete__menu-icon {
    margin-inline-start: 0;
}

.time-select-time-picker-container .v-autocomplete .v-field--appended {
    padding-inline-end: 8px;
}
</style>
