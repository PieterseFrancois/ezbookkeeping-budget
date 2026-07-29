// Subscription billing frequency, mirrors pkg/models/subscription.go's SubscriptionFrequencyType
export const SubscriptionFrequency = {
    Daily: 1,
    Weekly: 2,
    Monthly: 3,
    Quarterly: 4,
    Annual: 5,
} as const;

export type SubscriptionFrequencyValue = typeof SubscriptionFrequency[keyof typeof SubscriptionFrequency];

const SECONDS_PER_DAY = 86400;

function toLocalMidnightUnixTime(date: Date): number {
    return Math.floor(new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime() / 1000);
}

// Adds the given number of months to a date, clamping the day-of-month to the
// last valid day of the target month (e.g. Jan 31 + 1 month -> Feb 28/29,
// not an overflow into March). This is the billing-cycle convention used by
// most real subscription providers.
function addMonthsClamped(date: Date, months: number): Date {
    const day = date.getDate();
    const firstOfTargetMonth = new Date(date.getFullYear(), date.getMonth() + months, 1);
    const lastDayOfTargetMonth = new Date(firstOfTargetMonth.getFullYear(), firstOfTargetMonth.getMonth() + 1, 0).getDate();
    firstOfTargetMonth.setDate(Math.min(day, lastDayOfTargetMonth));
    return firstOfTargetMonth;
}

// Returns the next expected billing date (as a unix timestamp, local midnight)
// for a subscription, given its start date and frequency. If "now" is before
// the start date, the start date itself is returned. Otherwise the date rolls
// forward by whole cycles until it lands on or after "now".
export function getNextExpectedDate(startDate: number, frequency: SubscriptionFrequencyValue, now: Date = new Date()): number {
    const startOfToday = toLocalMidnightUnixTime(now);
    const startDateObj = new Date(startDate * 1000);
    const normalizedStartDate = toLocalMidnightUnixTime(startDateObj);

    if (startOfToday <= normalizedStartDate) {
        return normalizedStartDate;
    }

    if (frequency === SubscriptionFrequency.Daily || frequency === SubscriptionFrequency.Weekly) {
        const cycleSeconds = frequency === SubscriptionFrequency.Daily ? SECONDS_PER_DAY : SECONDS_PER_DAY * 7;
        const elapsedCycles = Math.floor((startOfToday - normalizedStartDate) / cycleSeconds);
        let next = normalizedStartDate + elapsedCycles * cycleSeconds;

        if (next < startOfToday) {
            next += cycleSeconds;
        }

        return next;
    }

    const monthsStep = frequency === SubscriptionFrequency.Monthly ? 1 : frequency === SubscriptionFrequency.Quarterly ? 3 : 12;
    const totalMonthsDiff = (now.getFullYear() - startDateObj.getFullYear()) * 12 + (now.getMonth() - startDateObj.getMonth());
    let elapsedCycles = Math.max(0, Math.floor(totalMonthsDiff / monthsStep));
    let candidate = toLocalMidnightUnixTime(addMonthsClamped(startDateObj, elapsedCycles * monthsStep));

    while (candidate < startOfToday) {
        elapsedCycles += 1;
        candidate = toLocalMidnightUnixTime(addMonthsClamped(startDateObj, elapsedCycles * monthsStep));
    }

    return candidate;
}
