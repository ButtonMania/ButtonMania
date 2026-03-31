import { secondsToHHMMSS } from '../src/utils';

describe('secondsToHHMMSS', () => {
    it('should format 0 seconds', () => {
        expect(secondsToHHMMSS(0)).toBe('00:00:00');
    });

    it('should format seconds only', () => {
        expect(secondsToHHMMSS(45)).toBe('00:00:45');
    });

    it('should format minutes and seconds', () => {
        expect(secondsToHHMMSS(125)).toBe('00:02:05');
    });

    it('should format hours, minutes, and seconds', () => {
        expect(secondsToHHMMSS(3661)).toBe('01:01:01');
    });

    it('should handle large values', () => {
        expect(secondsToHHMMSS(86400)).toBe('24:00:00');
    });

    it('should handle negative values gracefully', () => {
        const result = secondsToHHMMSS(-1);
        expect(typeof result).toBe('string');
    });
});
