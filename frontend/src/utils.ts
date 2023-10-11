export function secondsToHHMMSS(seconds: number): string {
    const format: any = (value: number) => (value < 10 ? `0${value}` : value);
    var hours: number = Math.floor(seconds / 3600);
    var minutes: number = Math.floor((seconds - (hours * 3600)) / 60);
    seconds %= 60;
    return `${format(hours)}:${format(minutes)}:${format(seconds)}`;
}