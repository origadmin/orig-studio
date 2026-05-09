import React from 'react';
import {Link} from '@tanstack/react-router';
import {colorFromName} from '@/lib/utils/tag-color';

const HASHTAG_REGEX = /(#[\p{L}\p{N}_][\p{L}\p{N}_\-]*)/gu;

interface HashtagTextProps {
    text: string;
    className?: string;
}

export function HashtagText({text, className}: HashtagTextProps) {
    if (!text) return null;

    const parts = text.split(HASHTAG_REGEX);

    return (
        <span className={className}>
            {parts.map((part, i) => {
                if (part.startsWith('#') && part.length > 1) {
                    const tagName = part.slice(1);
                    const color = colorFromName(tagName);
                    return (
                        <Link
                            key={i}
                            to="/search"
                            search={{tag: tagName}}
                            className="hover:opacity-80 cursor-pointer transition-opacity"
                            style={{color}}
                        >{part}</Link>
                    );
                }
                return <span key={i}>{part}</span>;
            })}
        </span>
    );
}
