import {type ClassValue, clsx} from "clsx"
import {twMerge} from "tailwind-merge"

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:9090";

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
}

export const getFullUrl = (path?: string): string => {
    if (!path) return '';
    if (path.startsWith('http')) return path;
    const base = API_BASE_URL.replace(/\/$/, '');
    const sep = path.startsWith('/') ? '' : '/';
    return `${base}${sep}${path}`;
};