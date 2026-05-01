import {parseWebVTT, findCueAtTime, parseVTTTime} from './parseWebVTT';

describe('parseVTTTime', () => {
    it('parses HH:MM:SS.mmm format', () => {
        expect(parseVTTTime('00:01:23.456')).toBe(83.456);
    });

    it('parses MM:SS.mmm format', () => {
        expect(parseVTTTime('01:23.456')).toBe(83.456);
    });

    it('parses zero time', () => {
        expect(parseVTTTime('00:00:00.000')).toBe(0);
    });

    it('parses one hour', () => {
        expect(parseVTTTime('01:00:00.000')).toBe(3600);
    });
});

describe('parseWebVTT', () => {
    const standardVTT = `WEBVTT

00:00:00.000 --> 00:00:10.000
sprites/abc123.jpg#xywh=0,0,160,90

00:00:10.000 --> 00:00:20.000
sprites/abc123.jpg#xywh=160,0,160,90

00:00:20.000 --> 00:00:30.000
sprites/abc123.jpg#xywh=320,0,160,90`;

    it('parses standard VTT with multiple cues', () => {
        const result = parseWebVTT(standardVTT, 'http://example.com/medias/abc/sprite.vtt');
        expect(result).not.toBeNull();
        expect(result!.cues).toHaveLength(3);
        expect(result!.cues[0]).toEqual({
            startTime: 0,
            endTime: 10,
            x: 0,
            y: 0,
            w: 160,
            h: 90,
        });
        expect(result!.cues[1]).toEqual({
            startTime: 10,
            endTime: 20,
            x: 160,
            y: 0,
            w: 160,
            h: 90,
        });
        expect(result!.cues[2]).toEqual({
            startTime: 20,
            endTime: 30,
            x: 320,
            y: 0,
            w: 160,
            h: 90,
        });
    });

    it('computes totalWidth and totalHeight from cues', () => {
        const result = parseWebVTT(standardVTT, 'http://example.com/medias/abc/sprite.vtt');
        expect(result!.totalWidth).toBe(480); // 320 + 160
        expect(result!.totalHeight).toBe(90); // 0 + 90
    });

    it('resolves imageUrl from baseUrl and relative path', () => {
        const result = parseWebVTT(standardVTT, 'http://example.com/medias/abc/sprite.vtt');
        expect(result!.imageUrl).toBe('http://example.com/medias/abc/sprites/abc123.jpg');
    });

    it('returns null for empty string', () => {
        expect(parseWebVTT('', 'http://example.com/')).toBeNull();
    });

    it('returns null for text without WEBVTT header', () => {
        expect(parseWebVTT('random text', 'http://example.com/')).toBeNull();
    });

    it('skips NOTE blocks', () => {
        const vttWithNote = `WEBVTT

NOTE This is a comment

00:00:00.000 --> 00:00:10.000
sprites/abc.jpg#xywh=0,0,160,90`;
        const result = parseWebVTT(vttWithNote, 'http://example.com/medias/abc/sprite.vtt');
        expect(result).not.toBeNull();
        expect(result!.cues).toHaveLength(1);
    });

    it('skips malformed cues without breaking parsing', () => {
        const vttWithBadCue = `WEBVTT

invalid cue line

00:00:00.000 --> 00:00:10.000
sprites/abc.jpg#xywh=0,0,160,90`;
        const result = parseWebVTT(vttWithBadCue, 'http://example.com/medias/abc/sprite.vtt');
        expect(result).not.toBeNull();
        expect(result!.cues).toHaveLength(1);
    });

    it('handles multi-row sprite sheets', () => {
        const multiRowVTT = `WEBVTT

00:00:00.000 --> 00:00:10.000
sprites/abc.jpg#xywh=0,0,160,90

00:00:10.000 --> 00:00:20.000
sprites/abc.jpg#xywh=160,0,160,90

00:00:20.000 --> 00:00:30.000
sprites/abc.jpg#xywh=0,90,160,90

00:00:30.000 --> 00:00:40.000
sprites/abc.jpg#xywh=160,90,160,90`;
        const result = parseWebVTT(multiRowVTT, 'http://example.com/medias/abc/sprite.vtt');
        expect(result!.totalWidth).toBe(320); // 160 + 160
        expect(result!.totalHeight).toBe(180); // 90 + 90
    });

    it('handles BOM in VTT content', () => {
        const vttWithBom = '\uFEFFWEBVTT\n\n00:00:00.000 --> 00:00:10.000\nsprites/abc.jpg#xywh=0,0,160,90';
        const result = parseWebVTT(vttWithBom, 'http://example.com/medias/abc/sprite.vtt');
        expect(result).not.toBeNull();
        expect(result!.cues).toHaveLength(1);
    });

    it('handles MM:SS.mmm timestamp format', () => {
        const shortFormatVTT = `WEBVTT

00:00.000 --> 00:10.000
sprites/abc.jpg#xywh=0,0,160,90`;
        const result = parseWebVTT(shortFormatVTT, 'http://example.com/medias/abc/sprite.vtt');
        expect(result).not.toBeNull();
        expect(result!.cues[0].startTime).toBe(0);
        expect(result!.cues[0].endTime).toBe(10);
    });
});

describe('findCueAtTime', () => {
    const cues = [
        {startTime: 0, endTime: 10, x: 0, y: 0, w: 160, h: 90},
        {startTime: 10, endTime: 20, x: 160, y: 0, w: 160, h: 90},
        {startTime: 20, endTime: 30, x: 320, y: 0, w: 160, h: 90},
    ];

    it('finds cue in the middle', () => {
        expect(findCueAtTime(cues, 15)).toEqual(cues[1]);
    });

    it('finds cue at boundary (startTime inclusive)', () => {
        expect(findCueAtTime(cues, 10)).toEqual(cues[1]);
    });

    it('finds first cue at time 0', () => {
        expect(findCueAtTime(cues, 0)).toEqual(cues[0]);
    });

    it('finds last cue', () => {
        expect(findCueAtTime(cues, 25)).toEqual(cues[2]);
    });

    it('returns null for time beyond range', () => {
        expect(findCueAtTime(cues, 30)).toBeNull();
    });

    it('returns null for negative time', () => {
        expect(findCueAtTime(cues, -1)).toBeNull();
    });

    it('returns null for empty cues array', () => {
        expect(findCueAtTime([], 5)).toBeNull();
    });

    it('handles single cue', () => {
        const singleCue = [{startTime: 0, endTime: 10, x: 0, y: 0, w: 160, h: 90}];
        expect(findCueAtTime(singleCue, 5)).toEqual(singleCue[0]);
        expect(findCueAtTime(singleCue, 10)).toBeNull();
    });
});
