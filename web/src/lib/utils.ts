import {type ClassValue, clsx} from "clsx"
import {twMerge} from "tailwind-merge"
import config from '@/config'

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
}

export const getFullUrl = (path?: string): string => {
    if (!path) return '';
    if (path.startsWith('http')) return path;
    const base = config.api.baseUrl.replace(/\/$/, '');
    const sep = path.startsWith('/') ? '' : '/';
    return `${base}${sep}${path}`;
};