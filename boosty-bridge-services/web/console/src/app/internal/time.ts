import appConfig from '@/app/configs/appConfig.json'

enum WeekDays {
    'Sun',
    'Mon',
    'Tue',
    'Wed',
    'Thu',
    'Fri',
    'Sat',
};

/** Parces date from ISO date string and returns day and month. */
export const getDayAndMonth = (ISODateString: string) => {
    const date = new Date(ISODateString);
    const day = date.getUTCDate();
    const weekDayNumber = date.getDay();
    const month = date.toLocaleString('en-US', {
        month: 'short',
    });

    return `${WeekDays[weekDayNumber]} ${day} ${month}`
};

/** Parces date from ISO date string and returns hours and minutes. */
export const getHoursAndMinutes: (ISODateString: string) => string = (ISODateString) => {
    const date = new Date(ISODateString);
    const hours = date.getUTCHours();
    let minutes: number | string = date.getUTCMinutes();

    if (minutes < appConfig.numbers.TEN_NUMBER) {
        minutes = `0${minutes}`;
    }

    return `${hours} : ${minutes}`;
};
