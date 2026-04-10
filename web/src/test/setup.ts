import '@testing-library/jest-dom';
import {TextEncoder, TextDecoder} from 'util';

// Polyfill for jsdom
global.TextEncoder = TextEncoder;
global.TextDecoder = TextDecoder as any;

// Mock localStorage
global.localStorage = {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
} as any;

// Mock document
global.document = {
    body: {},
    createElement: jest.fn(),
    querySelector: jest.fn(),
    querySelectorAll: jest.fn(),
} as any;

// 清理 localStorage 每个测试后
afterEach(() => {
    localStorage.clear();
    // 清理所有定时器
    jest.clearAllTimers();
    // 清理所有 mock
    jest.clearAllMocks();
});

// Mock window.matchMedia
global.matchMedia = global.matchMedia || function () {
    return {
        matches: false,
        addListener: jest.fn(),
        removeListener: jest.fn(),
    };
};

// 防止测试卡住
global.setTimeout = jest.useRealTimers ? setTimeout : jest.fn().mockReturnValue(123) as any;
global.clearTimeout = jest.useRealTimers ? clearTimeout : jest.fn() as any;

// Mock window.scrollTo
global.scrollTo = jest.fn();
(global as any).window = {
    ...(global as any).window,
    scrollTo: jest.fn(),
    localStorage: global.localStorage,
    document: global.document,
};
