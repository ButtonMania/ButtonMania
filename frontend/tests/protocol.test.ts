import { ButtonPhase, ButtonType, GameState, UserLocale } from '../src/protocol/enums';
import { GameplayMessageRequest } from '../src/protocol/requests';
import { RoomStats, UpdateResponse, RecordResponse, ErrorResponse, SessionContext, Record, Error as GameError } from '../src/protocol/respons';

describe('Protocol Enums', () => {
    it('should have correct GameState values', () => {
        expect(GameState.Update).toBe(0);
        expect(GameState.Record).toBe(1);
        expect(GameState.Error).toBe(3);
    });

    it('should have correct ButtonType values', () => {
        expect(ButtonType.NewYear).toBe('newyear');
        expect(ButtonType.Peace).toBe('peace');
        expect(ButtonType.Love).toBe('love');
        expect(ButtonType.Fortune).toBe('fortune');
        expect(ButtonType.Prestige).toBe('prestige');
    });

    it('should have correct ButtonPhase values', () => {
        expect(ButtonPhase.Idle).toBe('idle');
        expect(ButtonPhase.Push).toBe('push');
        expect(ButtonPhase.Hold).toBe('hold');
        expect(ButtonPhase.Release).toBe('release');
    });

    it('should have correct UserLocale values', () => {
        expect(UserLocale.EN).toBe('en');
        expect(UserLocale.RU).toBe('ru');
    });
});

describe('Protocol Messages', () => {
    it('should create GameplayMessageRequest', () => {
        const req = new GameplayMessageRequest(ButtonPhase.Hold);
        expect(req.buttonPhase).toBe(ButtonPhase.Hold);
    });

    it('should create RoomStats', () => {
        const stats = new RoomStats(5, 100, 3600, 1800, ['p1', 'p2']);
        expect(stats.countActive).toBe(5);
        expect(stats.countLeaderboard).toBe(100);
        expect(stats.bestOverallDuration).toBe(3600);
        expect(stats.bestTodaysDuration).toBe(1800);
        expect(stats.bestUsersPayloads).toEqual(['p1', 'p2']);
    });

    it('should create UpdateResponse', () => {
        const ctx = new SessionContext(ButtonPhase.Hold, 1000, 60);
        const update = new UpdateResponse(ctx, 1, 10, 'Keep going!');
        expect(update.context.buttonPhase).toBe(ButtonPhase.Hold);
        expect(update.placeActive).toBe(1);
        expect(update.countActive).toBe(10);
        expect(update.message).toBe('Keep going!');
    });

    it('should create RecordResponse', () => {
        const record = new Record(1000, 60);
        const resp = new RecordResponse(record, 5, 100, true);
        expect(resp.record.duration).toBe(60);
        expect(resp.placeLeaderboard).toBe(5);
        expect(resp.worldRecord).toBe(true);
    });

    it('should create ErrorResponse', () => {
        const err = new GameError('test error');
        const resp = new ErrorResponse(err);
        expect(resp.error.message).toBe('test error');
    });
});
