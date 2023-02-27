import { getDayAndMonth, getHoursAndMinutes } from "@app/internal/time";

describe('Parces ISO date string and returns month, day, hours and minutes', () => {
    const CURRENT_DATE: string = '2022-11-10T13:30:00.000Z';
    const CURRENT_DATE_DAY_AND_MONTH: string = 'Thu 10 Nov';
    const CURRENT_DATE_HOURS_AND_MINUTES: string = '13 : 30';
    const INVALID_DATES: string[] = [
        '2022-11-11T12:48:00.000Z',
        '2022-10-11T19:12:45.000Z',
        '2022-01-01T01:30:10.000Z',
    ];

    it('should return current month and day', () => {
        expect(getDayAndMonth(CURRENT_DATE)).toBe(CURRENT_DATE_DAY_AND_MONTH);
    });

    it ('should be an error when try to get current month and day', () => {
        INVALID_DATES.forEach((date: string) => {
            expect(getDayAndMonth(date)).not.toBe(CURRENT_DATE_DAY_AND_MONTH);
        });
    });

    it('should return current date hours and minutes', () => {
        expect(getHoursAndMinutes(CURRENT_DATE)).toBe(CURRENT_DATE_HOURS_AND_MINUTES);
    });

    it ('should be an error when try to get current date hours and minutes', () => {
        INVALID_DATES.forEach((date: string) => {
            expect(getHoursAndMinutes(date)).not.toBe(CURRENT_DATE_HOURS_AND_MINUTES);
        });
    });
});
